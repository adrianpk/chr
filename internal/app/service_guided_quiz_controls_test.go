package app

import (
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// TestGuidedUseCaseQuizControls exercises native modes, skipping, sufficient
// completion and lost acknowledgements without replacing the committed input.
func TestGuidedUseCaseQuizControls(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "research")
	now := serviceQuizClock()()
	batch := serviceQuizModelResponse("batch", "question-2", now)
	batch.Question.Prompt = "Which boundary must be preserved?"
	model.responses = append(model.responses[:1], batch, model.responses[1])

	current, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	current = switchGuidedQuizModes(t, operations, current)

	// The generated fast batch stays retained and is consumed only in fast mode.
	current, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, "fast-again", "quiz-mode", "fast"))
	if err != nil {
		t.Fatal(err)
	}

	current, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, "skip", "quiz-skip", ""))
	if err != nil || current.Guided.Quiz.QuestionID != "question-2" {
		t.Fatal("skip did not use the native queue", err)
	}

	current, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, "answer", "answer", "Preserve the existing contract."))
	if err != nil || current.Guided.Stage != "quiz" {
		t.Fatal("batch exhaustion bypassed sufficiency", err)
	}

	command := guidedQuizCommand(current, "done", "quiz-done", "")

	completed, err := operations.ConductGuidedQuiz(t.Context(), command)
	if err != nil {
		t.Fatal(err)
	}

	assertGuidedQuizCompletion(t, completed, session.Guided.Checkpoints[0])

	assertGuidedQuizControlReplay(t, operations, command, completed, model)
}

// TestGuidedUseCaseQuizDone refuses user-only completion, retaining a question
// until the model confirms that the saved answers suffice for the specification.
func TestGuidedUseCaseQuizDone(t *testing.T) {
	operations, session, model := guidedQuizFixture(t, "default")
	now := serviceQuizClock()()
	more := serviceQuizModelResponse("more", "question-2", now)
	more.Question.Prompt = "Which existing behavior must remain?"
	model.responses = append(model.responses[:1], more, more, model.responses[1])

	current, err := operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(session, "start", "quiz", ""))
	if err != nil {
		t.Fatal(err)
	}

	for _, command := range []struct{ id, action, content string }{
		{"empty-done", "quiz-done", ""},
		{"answer", "answer", "Keep the current public contract."},
		{"insufficient-done", "quiz-done", ""},
	} {
		current, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, command.id, command.action, command.content))
		if err != nil || current.Guided.Stage != "quiz" || current.Guided.Quiz.QuestionID == "" {
			t.Fatal("unfinished clarification advanced or lost its question", command.id, err)
		}
	}

	current, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, "sufficient-done", "quiz-done", ""))
	if err != nil {
		t.Fatal(err)
	}

	assertGuidedQuizCompletion(t, current, session.Guided.Checkpoints[0])
}

func observeGuidedQuiz(t *testing.T, operations service, session authoringcore.AuthoringSession) quizcore.Observation {
	t.Helper()

	observed, err := operations.quizRepository().Observe(t.Context(), quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug})
	if err != nil {
		t.Fatal(err)
	}

	return observed
}

func switchGuidedQuizModes(t *testing.T, operations service, current authoringcore.AuthoringSession) authoringcore.AuthoringSession {
	t.Helper()

	var err error
	for _, mode := range []string{"fast", "deep"} {
		current, err = operations.ConductGuidedQuiz(t.Context(), guidedQuizCommand(current, mode, "quiz-mode", mode))
		if err != nil {
			t.Fatal(err)
		}

		observed := observeGuidedQuiz(t, operations, current)
		if string(observed.State.Mode) != mode || observed.State.ActiveQuestion.ID != "question-1" {
			t.Fatal("mode switch lost the current question", observed.State)
		}
	}

	return current
}

func assertGuidedQuizControlReplay(t *testing.T, operations service, command authoringcore.GuidedUseCaseCommand, completed authoringcore.AuthoringSession, model *fakeServiceQuizModel) {
	t.Helper()

	replay, err := operations.ConductGuidedQuiz(t.Context(), command)
	if err != nil || replay.Revision != completed.Revision || len(model.requests) != 3 || model.requests[1].Mode != quizcore.ModeFast {
		t.Fatal("controls repeated generation or used the wrong strategy", err)
	}
}
