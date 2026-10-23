package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// TestGuidedUseCaseQuizSufficiency requires an answered, owner-confirmed complete quiz.
func TestGuidedUseCaseQuizSufficiency(t *testing.T) {
	for _, variant := range []string{"complete", "complete-with-skip", "unconfirmed-skip", "empty", "budget", "skipped", "foreign"} {
		t.Run(variant, func(t *testing.T) {
			state := QuizRuntimeState{SessionID: "quiz", Status: QuizRuntimeStatusDone, ClarificationComplete: true, QuestionsAsked: 1, AnswersRecorded: 1}
			session := AuthoringSession{Guided: &GuidedUseCaseState{Quiz: &GuidedQuizState{SessionID: "quiz", Revision: "revision"}}}

			switch variant {
			case "complete-with-skip", "unconfirmed-skip":
				state.QuestionsAsked, state.SkipsRecorded = 2, 1
				state.ClarificationComplete = variant == "complete-with-skip"
			case "empty":
				state.AnswersRecorded = 0
			case "budget":
				state.ClarificationComplete = false
			case "skipped":
				state.SkipsRecorded = 1
			case "foreign":
				state.SessionID = "other"
			}

			err := guardGuidedQuizCompletion(session, QuizObservation{State: state, Revision: "revision"})
			if (err == nil) != (variant == "complete" || variant == "complete-with-skip") {
				t.Fatalf("unexpected sufficiency: %v", err)
			}
		})
	}
}

// TestGuidedUseCaseQuizLegacy preserves digest inputs when optional guided fields are absent.
func TestGuidedUseCaseQuizLegacy(t *testing.T) {
	scope, err := json.Marshal(QuizScope{ProjectID: "project", CaseSlug: "case"})
	if err != nil || string(scope) != `{"ProjectID":"project","CaseSlug":"case"}` {
		t.Fatalf("legacy scope encoding changed: %s %v", scope, err)
	}

	state, err := json.Marshal(QuizRuntimeState{})
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]any
	if err := json.Unmarshal(state, &fields); err != nil {
		t.Fatal(err)
	}

	if _, exists := fields["ClarificationComplete"]; exists {
		t.Fatal("legacy state gained a null or false guided field")
	}
}

// TestGuidedUseCaseQuizOwner prevents progression writes outside the existing mutation claim.
func TestGuidedUseCaseQuizOwner(t *testing.T) {
	root, session := guidedCaptureFixture(t)
	if err := RetainGuidedUseCaseProgress(t.Context(), root, &session); err == nil || errors.Is(err, ErrWorkflowConflict) {
		t.Fatalf("unexpected owner boundary: %v", err)
	}
}

// TestGuidedUseCaseQuizCheckpoint recovers a retained seal after later working-copy edits.
func TestGuidedUseCaseQuizCheckpoint(t *testing.T) {
	root, command := guidedCheckpointFixture(t)

	session, err := FinishGuidedUseCaseDump(t.Context(), root, command)
	if err != nil {
		t.Fatal(err)
	}

	scope := QuizScope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}
	paths, result := completedGuidedQuizFixture(t, root, scope)
	session.Guided.Quiz = &GuidedQuizState{SessionID: result.State.SessionID, Revision: result.State.Revision}
	session.Guided.PendingOperation = &GuidedUseCaseOperation{ID: "seal-quiz", Action: "answer", Digest: workflowHash([]byte("seal-quiz"))}
	r := &fileAuthoringSessionRepository{projectDir: root}

	var original string

	err = WithAuthoringMutation(t.Context(), root, func(ctx context.Context) error {
		checkpoint, err := r.observeGuidedCheckpoint(ctx, session.ID, *session.Guided.PendingOperation, "quiz", quizImageReferences(paths))
		if err != nil {
			return err
		}

		session.Guided.PendingCheckpoint = &checkpoint
		if err := RetainGuidedUseCaseProgress(ctx, root, &session); err != nil {
			return err
		}

		committed, err := r.commitGuidedCheckpoint(ctx, checkpoint, session.Guided.PendingOperation.Digest)
		original = committed.Commit

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	guidedWriteFile(t, root, paths.CuratedRelativePath, "Later working-copy replacement")

	err = WithAuthoringMutation(t.Context(), root, func(ctx context.Context) error { return SealGuidedQuiz(ctx, root, &session) })
	if err != nil || session.Guided.Checkpoints[1].Commit != original {
		t.Fatalf("retained quiz seal replaced or blocked by working copy: %v", err)
	}
}

func completedGuidedQuizFixture(t *testing.T, root string, scope QuizScope) (QuizArtifactPaths, QuizActionResult) {
	t.Helper()

	repo := NewQuizArtifactRepository(root)

	observed, err := repo.Observe(t.Context(), scope)
	if err != nil {
		t.Fatal(err)
	}

	paths, err := repo.Paths(scope)
	if err != nil {
		t.Fatal(err)
	}

	start := QuizActionRequest{Scope: scope, Action: "start", Binding: QuizActionBinding{OperationID: "quiz-start", ExpectedRevision: observed.Revision}}

	_, err = repo.ApplyAction(t.Context(), start, func(ctx context.Context, before QuizObservation) (QuizTransition, error) {
		next, err := startQuizActionFixture(ctx, before)
		next.State.ProjectID, next.State.CaseSlug = scope.ProjectID, scope.CaseSlug
		next.State.RawPath, next.State.CuratedPath = paths.RawRelativePath, paths.CuratedRelativePath

		return next, err
	})
	if err != nil {
		t.Fatal(err)
	}

	observed, err = repo.Observe(t.Context(), scope)
	if err != nil {
		t.Fatal(err)
	}

	answer := quizAnswerActionFixture(observed)
	answer.Scope = scope

	result, err := repo.ApplyAction(t.Context(), answer, func(ctx context.Context, before QuizObservation) (QuizTransition, error) {
		next, err := answerQuizActionFixture(ctx, before)
		next.State.Status, next.State.ClarificationComplete = QuizRuntimeStatusDone, true
		next.State.Candidates = nil

		return next, err
	})
	if err != nil {
		t.Fatal(err)
	}

	return paths, result
}
