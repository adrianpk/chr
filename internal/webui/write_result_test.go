package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui/htmx"
)

// htmx must swap expected write failures so operators can see the conflict and reload.
func TestHandlerWriteFailure(t *testing.T) {
	for _, tt := range []struct {
		name     string
		path     string
		htmx     bool
		surface  string
		status   int
		modal    bool
		retarget string
	}{
		{name: "page", path: "/actions", status: http.StatusBadRequest},
		{name: "fragment", path: "/actions", htmx: true, status: http.StatusOK},
		{name: "modal", path: "/actions", htmx: true, surface: "modal", status: http.StatusOK, modal: true},
		{name: "comment", path: "/board/ticket/comment", htmx: true, status: http.StatusOK, retarget: "#ticket-comment-result"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewHandler(HandlerOptions{Data: testShellData(), Provider: staleWriteProvider{}})
			if err != nil {
				t.Fatal(err)
			}

			form := url.Values{"action": {"ticket.comment"}, "id": {"CHR-1"}, "confirm": {"true"}, "surface": {tt.surface}}
			request := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set(htmx.HeaderRequest, strconv.FormatBool(tt.htmx))

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.status)
			}

			if !strings.Contains(recorder.Body.String(), "Ticket changed. Reload before saving.") {
				t.Fatal("write failure was not rendered")
			}

			assertWriteFailureSurface(t, recorder, tt.modal, tt.retarget)
		})
	}
}

func assertWriteFailureSurface(t *testing.T, recorder *httptest.ResponseRecorder, wantModal bool, wantRetarget string) {
	t.Helper()

	if got := recorder.Header().Get(htmx.HeaderRetarget); got != wantRetarget {
		t.Fatalf("retarget = %q, want %q", got, wantRetarget)
	}

	if modal := strings.Contains(recorder.Body.String(), `role="dialog"`); modal != wantModal {
		t.Fatalf("modal = %v, want %v", modal, wantModal)
	}
}

type staleWriteProvider struct{ testDataProvider }

func (staleWriteProvider) WriteAction(context.Context, WriteActionRequest) (WriteActionResponse, error) {
	return WriteActionResponse{
		Status: WriteStatusStale, Summary: "Ticket changed. Reload before saving.",
		ReloadURL: "/board", DiscardURL: "/board",
	}, nil
}
