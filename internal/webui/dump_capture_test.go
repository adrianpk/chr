package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

// Capture failures and pending projection acknowledgements render into the form's
// result region; they never replace the form or redirect away from retained input.
func TestDumpCaptureResult(t *testing.T) {
	for _, status := range []string{WriteStatusValidation, WriteStatusOK} {
		t.Run(status, func(t *testing.T) {
			handler, err := NewHandler(HandlerOptions{Data: testShellData(), Provider: dumpCaptureResultProvider{status: status}})
			if err != nil {
				t.Fatal(err)
			}

			form := url.Values{"action": {"dump.create"}, "surface": {"modal"}, "confirm": {"true"}}
			request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set(htmx.HeaderRequest, "true")

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK || recorder.Header().Get(htmx.HeaderRetarget) != "" || recorder.Header().Get(htmx.HeaderRedirect) != "" {
				t.Fatal("capture result discarded retained form", recorder.Code, recorder.Header())
			}

			if !strings.Contains(recorder.Body.String(), `data-write-status="`+status+`"`) || strings.Contains(recorder.Body.String(), `role="dialog"`) {
				t.Fatal("capture result replaced form dialog", recorder.Body.String())
			}
		})
	}
}

type dumpCaptureResultProvider struct {
	testDataProvider
	status string
}

func (p dumpCaptureResultProvider) WriteAction(context.Context, WriteActionRequest) (WriteActionResponse, error) {
	return WriteActionResponse{Status: p.status, Summary: "Retained capture result", RedirectURL: "/dumps", ProjectionStale: p.status == WriteStatusOK}, nil
}
