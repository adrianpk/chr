package app

import (
	"maps"
	"path/filepath"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

// Even an absent session needs a captured revision before Start. A delayed TUI
// mutation with no binding must never fall back to observing a newer question.
func TestTUIQuizBinding(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	views, err := loadTUIQuizzes(operations)
	if err != nil || len(views[0].Binding.ExpectedRevision) != 64 {
		t.Fatal("absent quiz omitted its observed revision", views, err)
	}

	if _, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.ModeFast}); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))

	_, err = newTUIQuizActionExecutor(operations).ExecuteQuizAction(tui.QuizActionRequest{Action: tui.QuizActionAnswer, Answer: "Unbound delayed answer"})
	if err == nil || !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("unbound adapter request mutated the current question", err)
	}
}

func observedTUIQuizBinding(t *testing.T, operations service, slug, operationID string) tui.QuizActionBinding {
	t.Helper()

	observed, err := operations.quizRepository().Observe(t.Context(), operations.quizScope(quizcore.Scope{CaseSlug: slug}))
	if err != nil {
		t.Fatal(err)
	}

	binding := quizStateBinding(observed.State, operationID)
	binding.ExpectedRevision = observed.Revision

	return tui.QuizActionBinding{SessionID: binding.SessionID, QuestionID: binding.QuestionID, ExpectedRevision: binding.ExpectedRevision, OperationID: operationID}
}
