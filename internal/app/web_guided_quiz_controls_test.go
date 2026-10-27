package app

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// TestGuidedUseCaseSharedQuizWeb checks that original quiz forms reach the
// guided owner, retaining both bindings and rejecting stale or foreign targets.
func TestGuidedUseCaseSharedQuizWeb(t *testing.T) {
	provider, model := guidedWebFixture(t)
	handler := guidedWebHandler(t, provider)
	current := captureGuidedWebInput(t, provider, handler)
	operations := provider.service(t.Context())
	observed := observeGuidedQuiz(t, operations, current)
	form := quizHTTPForm(observed.State, "quiz.mode", "shared-mode")
	form.Set("guided_run", current.ID)
	form.Set("guided_revision", "0")
	form.Set("namespace", current.Namespace)
	form.Set("input_digest", current.Guided.Quiz.InputDigest)
	form.Set("mode", "fast")

	for _, field := range []string{"guided_revision", "revision", "session_id", "question_id", "input_digest", "namespace", "case_slug", "id", "mode"} {
		t.Run(field, func(t *testing.T) {
			request := url.Values{}
			for key, values := range form {
				request[key] = append([]string(nil), values...)
			}

			request.Set("guided_revision", strconv.FormatInt(current.Revision, 10))
			request.Set(field, "foreign")

			response := guidedCaptureHTTP(t, handler, http.MethodPost, "/actions", request)
			if response.Code == http.StatusSeeOther {
				t.Fatal("invalid shared binding accepted", field)
			}

			saved := readWebGuidedSession(t, operations, current.ID)
			if saved.Revision != current.Revision || len(model.requests) != 1 {
				t.Fatal("invalid binding reserved or generated work", field)
			}
		})
	}

	batch := serviceQuizModelResponse("fast-batch", "question-2", serviceQuizClock()())
	model.responses = append(model.responses[:1], batch)

	form.Set("guided_revision", strconv.FormatInt(current.Revision, 10))

	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/actions", form)
	if response.Code != http.StatusSeeOther {
		t.Fatal("shared mode form failed", response.Code, response.Body.String())
	}

	updated := readWebGuidedSession(t, operations, current.ID)

	quiz := observeGuidedQuiz(t, operations, updated)
	if quiz.State.Mode != quizcore.QuizModeFast || quiz.State.ActiveQuestion.ID != observed.State.ActiveQuestion.ID || model.requests[1].Mode != quizcore.QuizModeFast {
		t.Fatal("mode form bypassed the shared native quiz")
	}
}
