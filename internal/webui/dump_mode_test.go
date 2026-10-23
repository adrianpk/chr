package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type dumpModeProvider struct{ guidedCaptureProvider }

func (p dumpModeProvider) ArtifactDetailView(context.Context, ArtifactDetailRequest) (ArtifactDetailView, error) {
	return ArtifactDetailView{Family: "dumps", ID: "manual", Content: "Existing dump"}, nil
}

// TestDumpModeActions keeps the wizard forward-only and the manual editor independent.
func TestDumpModeActions(t *testing.T) {
	provider := dumpModeProvider{guidedCaptureProvider{view: GuidedUseCaseView{RunID: "run", CaseSlug: "onboarding", Stage: "dump"}}}

	handler, err := NewHandler(HandlerOptions{Provider: provider})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct{ name, path, want, absent string }{
		{"wizard", "/dumps?case=onboarding", `name="step" value="finish"`, "Save dump"},
		{"manual", "/dumps/detail?id=manual", `value="dump.save"`, "Continue to quiz"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))

			body := response.Body.String()
			if response.Code != http.StatusOK || !strings.Contains(body, test.want) || strings.Contains(body, test.absent) || strings.Contains(body, `form="workflow-finish"`) {
				t.Fatal("editor mixed manual and wizard actions")
			}
		})
	}
}
