package webui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type harnessViewProvider struct {
	testDataProvider
	failure bool
}

func (p harnessViewProvider) HarnessSelectionView(context.Context) (HarnessSelectionView, error) {
	if p.failure {
		return HarnessSelectionView{}, errors.New("private /credential/location must not be rendered")
	}

	return HarnessSelectionView{ProjectID: "project-1", Worker: HarnessRoleView{Current: "alpha · Codex · model-a"}, Generation: HarnessRoleView{Current: "Legacy dispatch · varies by operation"}}, nil
}

// Unsupported providers hide the control, and expected setup failures render
// a bounded public diagnostic instead of a private provider error or raw page.
func TestWebHarnessSelectionView(t *testing.T) {
	for _, test := range []struct {
		name     string
		provider DataProvider
		status   int
		text     string
	}{
		{"unsupported", testDataProvider{}, http.StatusNotFound, "404"},
		{"defaults", harnessViewProvider{}, http.StatusOK, "alpha · Codex · model-a"},
		{"unavailable", harnessViewProvider{failure: true}, http.StatusOK, "Harness choices could not be loaded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(HandlerOptions{Provider: test.provider, Data: testShellData()})
			if err != nil {
				t.Fatal(err)
			}

			response := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/project/harnesses", nil)
			request.Header.Set("HX-Request", "true")
			handler.ServeHTTP(response, request)

			if response.Code != test.status || !strings.Contains(response.Body.String(), test.text) || strings.Contains(response.Body.String(), "/credential/location") {
				t.Fatal("incorrect project harness view", response.Code, response.Body.String())
			}
		})
	}
}
