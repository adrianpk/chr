package webui

import (
	"context"
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
