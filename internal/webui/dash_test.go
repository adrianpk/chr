package webui

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type dashControlProvider struct {
	testDataProvider
	last *WriteActionRequest
}

func (p dashControlProvider) WriteAction(_ context.Context, request WriteActionRequest) (WriteActionResponse, error) {
	*p.last = request

	return WriteActionResponse{Status: WriteStatusOK, Summary: "Preview retained", RedirectURL: "/dash"}, nil
}

// Retained controls target their existing operations and confirmations while
// rendering only their result region, so monitoring is not replaced or hidden.
func TestDashControls(t *testing.T) {
	for _, action := range []string{"work.preview", "work.daemon.stop"} {
		var captured WriteActionRequest

		handler, err := NewHandler(HandlerOptions{Data: testShellData(), Provider: dashControlProvider{last: &captured}})
		if err != nil {
			t.Fatal(err)
		}

		form := url.Values{"action": {action}, "surface": {"dash"}, "confirm": {"true"}}
		request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("HX-Request", "true")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != 200 || response.Header().Get("HX-Redirect") != "" || captured.Action != action || !captured.Confirm {
			t.Fatal(response.Code, captured, response.Header())
		}

		if !strings.Contains(response.Body.String(), "Preview retained") {
			t.Fatal(response.Body.String())
		}
	}
}

type dashReadProvider struct {
	testDataProvider
	unavailable bool
}

func (p dashReadProvider) DashView(ctx context.Context) (DashView, error) {
	if p.unavailable {
		return DashView{}, fmt.Errorf("temporary source failure")
	}

	view, err := p.testDataProvider.DashView(ctx)
	view.ObservedAt = "2026-10-18T12:00:00Z"

	return view, err
}

// Semantic ETags avoid replacing unchanged observations, while bad pages and
// source outages remain distinguishable from a genuinely empty queue.
func TestDashRefresh(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Data: testShellData(), Provider: dashReadProvider{}})
	if err != nil {
		t.Fatal(err)
	}

	response := dashReadRequest(handler, "/dash/fragment", "")
	if response.Code != 200 || response.Header().Get("ETag") != `"dash-revision"` || response.Header().Get("X-CHR-Observed-At") == "" {
		t.Fatal(response.Code, response.Header())
	}

	unchanged := dashReadRequest(handler, "/dash/fragment", response.Header().Get("ETag"))
	if unchanged.Code != http.StatusNotModified || unchanged.Body.Len() != 0 {
		t.Fatal(unchanged.Code, unchanged.Body.String())
	}

	for _, offset := range []string{"-1", "4097", "invalid"} {
		rejected := dashReadRequest(handler, "/dash/fragment?queue_offset="+offset, "")
		if rejected.Code != http.StatusBadRequest {
			t.Fatal(offset, rejected.Code)
		}
	}
}

// A first-load failure shows unavailable state; polling failure leaves an
// explicit non-success response so the browser can retain its last observation.
func TestDashUnavailable(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Data: testShellData(), Provider: dashReadProvider{unavailable: true}})
	if err != nil {
		t.Fatal(err)
	}

	page := dashReadRequest(handler, "/dash", "")
	if page.Code != 200 || !strings.Contains(page.Body.String(), `data-state="unavailable"`) || strings.Contains(page.Body.String(), "No pending eligible work") {
		t.Fatal(page.Code, page.Body.String())
	}

	fragment := dashReadRequest(handler, "/dash/fragment", "")
	if fragment.Code != http.StatusServiceUnavailable {
		t.Fatal(fragment.Code)
	}
}

func dashReadRequest(handler http.Handler, path, etag string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("HX-Request", "true")
	request.Header.Set("If-None-Match", etag)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}
