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

// Save acknowledgements and expected failures must update only the result
// region, leaving the live editor and its pending body in place.
func TestDumpSaveResult(t *testing.T) {
	for _, status := range []string{WriteStatusValidation, WriteStatusStale, WriteStatusBlocked, WriteStatusOK} {
		t.Run(status, func(t *testing.T) {
			handler, err := NewHandler(HandlerOptions{Data: testShellData(), Provider: dumpCaptureResultProvider{status: status}})
			if err != nil {
				t.Fatal(err)
			}

			form := url.Values{"action": {"dump.save"}, "surface": {"modal"}, "confirm": {"true"}}
			request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set(htmx.HeaderRequest, "true")

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK || recorder.Header().Get(htmx.HeaderRetarget) != "" || recorder.Header().Get(htmx.HeaderRedirect) != "" || strings.Contains(recorder.Body.String(), `role="dialog"`) {
				t.Fatal("Save result replaced the editor", recorder.Code, recorder.Header(), recorder.Body.String())
			}

			if !strings.Contains(recorder.Body.String(), `data-write-status="`+status+`"`) {
				t.Fatal("Save result omitted its acknowledgement or failure", recorder.Body.String())
			}
		})
	}
}

type dumpCaptureResultProvider struct {
	testDataProvider
	status string
}

// Full-page editing must have the same one-body surface and cannot poll over
// pending input; explicit Save or reload supplies the next stored revision.
func TestHandlerDumpPageKeepsEditor(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Data: testShellData(), Provider: testDataProvider{}})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/dumps/detail?id=.chr/default/dumps/project-notes.md", nil))

	body := recorder.Body.String()
	if recorder.Code != http.StatusOK || strings.Count(body, "<textarea") != 1 || strings.Count(body, `value="dump.save"`) != 1 {
		t.Fatal("dump page did not render one editor and Save", recorder.Code, body)
	}

	for _, forbidden := range []string{"dump.append", "artifact-content", "Append note", "hx-trigger="} {
		if strings.Contains(body, forbidden) {
			t.Fatal("full-page editor can duplicate or replace pending input", forbidden)
		}
	}
}

func (p dumpCaptureResultProvider) WriteAction(context.Context, WriteActionRequest) (WriteActionResponse, error) {
	return WriteActionResponse{Status: p.status, Summary: "Retained capture result", RedirectURL: "/dumps", ProjectionStale: p.status == WriteStatusOK}, nil
}
