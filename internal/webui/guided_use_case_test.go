package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type guidedCaptureProvider struct {
	staticDataProvider
	view GuidedUseCaseView
}

func (p guidedCaptureProvider) GuidedUseCaseView(context.Context, string, string) (GuidedUseCaseView, error) {
	p.view.Enabled = true

	return p.view, nil
}

// TestGuidedUseCaseCaptureView renders only current input and escapes captured content.
func TestGuidedUseCaseCaptureView(t *testing.T) {
	for _, test := range []struct {
		name   string
		view   GuidedUseCaseView
		want   string
		absent string
	}{
		{"start", GuidedUseCaseView{Namespace: "default", RequestID: "start"}, "Develop a new use case", "Finish capture"},
		{"name", GuidedUseCaseView{RunID: "run", Stage: "name"}, "Use-case name", "Finish capture"},
		{"dump", GuidedUseCaseView{RunID: "run", Stage: "dump", Capture: "<script>bad()</script>"}, "Finish capture", "<script>bad()"},
		{"sealed", GuidedUseCaseView{RunID: "run", Stage: "quiz", Commit: "commit", CommittedDump: "Original capture"}, "Inspect committed dump", "Finish capture"},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{view: test.view}})
			if err != nil {
				t.Fatal(err)
			}

			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/guided", nil))

			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.want) || strings.Contains(response.Body.String(), test.absent) {
				t.Fatal("incorrect current-stage form", response.Code, response.Body.String())
			}
		})
	}
}

// TestGuidedUseCaseCaptureBounds rejects duplicate parameters and oversized capture requests.
func TestGuidedUseCaseCaptureBounds(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{}})
	if err != nil {
		t.Fatal(err)
	}

	for _, body := range []string{"action=guided.start&action=guided.input", "content=" + strings.Repeat("x", 73<<10)} {
		request := httptest.NewRequest(http.MethodPost, "/guided/actions", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("unbounded or duplicate form accepted: %d", response.Code)
		}
	}
}

// TestGuidedUseCaseCaptureUnavailable keeps static navigation useful without commands.
func TestGuidedUseCaseCaptureUnavailable(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/guided", nil))

	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, "Guided use cases are unavailable") || strings.Contains(body, `action="/guided/actions"`) {
		t.Fatal("unavailable commands exposed", response.Code, body)
	}
}
