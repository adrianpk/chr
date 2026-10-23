package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// TestCaseWorkflowPresentation keeps transport guards hidden while reusing the quiz.
func TestCaseWorkflowPresentation(t *testing.T) {
	view := GuidedUseCaseView{RunID: "opaque-run", Namespace: "private-storage", Stage: "quiz", Name: "Improve checkout", Question: "What must checkout preserve?", QuestionID: "opaque-question", InputDigest: "opaque-input", Commit: "opaque-commit", Policy: []string{"trusted_auto_merge"}, CanRetainIdea: true}

	view.Quiz = &QuizAnswerView{Question: view.Question, QuestionID: view.QuestionID, CanAnswer: true, Open: true, Started: true, Mode: "deep"}

	handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{view: view}})
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/cases/workflow?id=opaque-run", nil))

	body := response.Body.String()

	visible := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(body, "")
	for _, hidden := range []string{"opaque-", "private-storage", "trusted_auto_merge", "Namespace", "Ideas for later", "Project policy"} {
		if strings.Contains(visible, hidden) {
			t.Fatalf("internal detail is visible: %s", hidden)
		}
	}

	for _, required := range []string{`class="quiz-panel"`, `class="quiz-response"`, `name="question_id" value="opaque-question"`, `name="input_digest" value="opaque-input"`, "What must checkout preserve?"} {
		if !strings.Contains(body, required) {
			t.Fatalf("missing shared quiz or original guard: %s", required)
		}
	}
}

// TestCaseWorkflowObservation ignores fresh request tokens and rejects ambiguous reads.
func TestCaseWorkflowObservation(t *testing.T) {
	view := GuidedUseCaseView{RunID: "run", Stage: "quiz", Observation: "2", RequestID: "first"}
	first := guidedStateToken(view)

	view.RequestID = "second"
	if guidedStateToken(view) != first {
		t.Fatal("fresh form token triggered navigation")
	}

	view.Observation = "3"
	if guidedStateToken(view) == first {
		t.Fatal("accepted state change was invisible")
	}

	handler, err := NewHandler(HandlerOptions{Provider: guidedCaptureProvider{view: view}})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/cases/workflow/state?id=run", "/cases/workflow/state?id=run&id=other", "/cases/workflow/state?id=run&case=checkout"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))

		if path != "/cases/workflow/state?id=run" {
			if response.Code != http.StatusBadRequest {
				t.Fatal("ambiguous target accepted", path)
			}

			continue
		}

		var state struct{ State string }
		if err := json.Unmarshal(response.Body.Bytes(), &state); err != nil || state.State != guidedStateToken(view) || response.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("incorrect read-only observation", err)
		}
	}
}
