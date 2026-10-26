package core

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"
)

// GuidedUseCaseStep composes canonical owners under the existing authoring exclusion.
// Effects must retain their own request identity before dispatch, and reconcile it on retry.
type GuidedUseCaseStep func(context.Context, *AuthoringSession) error

// ApplyGuidedUseCaseStep reserves and replays one revision-bound progression command.
func ApplyGuidedUseCaseStep(ctx context.Context, root string, command GuidedUseCaseCommand, apply GuidedUseCaseStep) (session AuthoringSession, err error) {
	if apply == nil {
		return session, errors.New("guided step owner is required")
	}

	operation, err := guidedOperation(command)
	if err != nil {
		return session, err
	}

	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, command.RunID)
		if err != nil {
			return err
		}

		if err := guardGuidedScope(session, command.GuidedUseCaseScope); err != nil {
			return err
		}

		replay, err := guardGuidedCommand(session, command, operation)
		if err != nil || replay {
			return err
		}

		return r.applyGuidedStep(ctx, &session, operation, apply)
	})

	return session, err
}

func (r *fileAuthoringSessionRepository) applyGuidedStep(ctx context.Context, session *AuthoringSession, operation GuidedUseCaseOperation, apply GuidedUseCaseStep) error {
	if session.Guided.PendingOperation == nil {
		if err := guardGuidedStepInput(*session, operation); err != nil {
			return err
		}

		session.Guided.PendingOperation = &operation

		session.Guided.Blocker = ""
		if err := RetainGuidedUseCaseProgress(ctx, r.projectDir, session); err != nil {
			return err
		}
	}

	if err := apply(ctx, session); err != nil {
		session.Guided.Status = "blocked"
		session.Guided.Blocker = "Progression needs recovery or corrected setup. Retry the retained operation."

		var prerequisite *GuidedPrerequisiteError
		if errors.As(err, &prerequisite) {
			session.Guided.Blocker = prerequisite.Action
		}

		return errors.Join(err, RetainGuidedUseCaseProgress(ctx, r.projectDir, session))
	}

	session.Guided.CompletedOperations = append(session.Guided.CompletedOperations, GuidedUseCaseOperation{ID: operation.ID, Action: operation.Action, Digest: operation.Digest})

	session.Guided.PendingOperation = nil
	if session.Guided.Status != "blocked" {
		session.Guided.Blocker = ""
	}

	return RetainGuidedUseCaseProgress(ctx, r.projectDir, session)
}

func guardGuidedStepInput(session AuthoringSession, operation GuidedUseCaseOperation) error {
	switch operation.Action {
	case "quiz":
		if session.Guided.Stage != "quiz" {
			return ErrWorkflowConflict
		}

		return nil
	case "prepare":
		return guardGuidedPreparationInput(session)
	case "admit":
		if session.Guided.Stage != "preparation" {
			return ErrWorkflowConflict
		}

		_, err := guidedStageCheckpoint(session, "preparation")

		return err
	case "spec":
		if session.Guided.Stage != "specification" {
			return ErrWorkflowConflict
		}

		return nil
	case "answer", "quiz-mode", "quiz-skip", "quiz-done":
		return guardGuidedQuizInput(session, operation)
	default:
		return ErrWorkflowConflict
	}
}

func guardGuidedQuizInput(session AuthoringSession, operation GuidedUseCaseOperation) error {
	if err := guardGuidedQuizContent(operation); err != nil {
		return err
	}

	quiz := session.Guided.Quiz
	if session.Guided.Stage != "quiz" || quiz == nil || (operation.Action == "answer" || operation.Action == "quiz-skip") && quiz.QuestionID == "" {
		return ErrWorkflowConflict
	}

	if operation.QuestionID != quiz.QuestionID || operation.InputDigest != quiz.InputDigest {
		return ErrWorkflowConflict
	}

	return nil
}

// RetainGuidedUseCaseProgress is available only inside the canonical mutation claim.
func RetainGuidedUseCaseProgress(ctx context.Context, root string, session *AuthoringSession) error {
	if session == nil || session.Guided == nil {
		return ErrWorkflowConflict
	}

	absolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	if !holdsAuthoringOwner(ctx, absolute) {
		return errors.New("guided progress requires its authoring owner")
	}

	r := &fileAuthoringSessionRepository{projectDir: root}
	if err := r.recoverAuthoringMutation(); err != nil {
		return err
	}

	current, err := r.getGuardedSession(ctx, session.ID)
	if err != nil {
		return err
	}

	if !sameGuidedRevision(current, *session) {
		return ErrWorkflowConflict
	}

	session.Revision++
	session.UpdatedAt = time.Now().UTC().Truncate(time.Second)

	return r.saveGuidedSession(*session, false)
}

// GuidedCheckpointDigest identifies the complete retained committed input binding.
func GuidedCheckpointDigest(checkpoint GuidedStageCheckpoint) (string, error) {
	encoded, err := json.Marshal(checkpoint)

	return workflowHash(encoded), err
}

// SealGuidedQuiz checks canonical sufficiency and seals the answered owner artifacts.
func SealGuidedQuiz(ctx context.Context, root string, session *AuthoringSession) error {
	if len(session.Guided.Checkpoints) == 0 || session.Guided.PendingOperation == nil {
		return ErrWorkflowConflict
	}

	references, err := guidedQuizSealReferences(ctx, root, *session)
	if err != nil {
		return err
	}

	if err := sealGuidedStage(ctx, root, session, guidedStageSeal{Stage: "quiz", Paths: references, Source: session.Guided.Checkpoints[0]}); err != nil {
		return err
	}

	session.Guided.Stage, session.Guided.Status = "specification", "blocked"
	session.Guided.Blocker = "Specification generation is not available yet."

	return nil
}

func guardGuidedQuizCompletion(session AuthoringSession, observed QuizObservation) error {
	quiz := session.Guided.Quiz

	state := observed.State
	if quiz == nil || quiz.SessionID != state.SessionID || quiz.Revision != observed.Revision {
		return ErrWorkflowConflict
	}

	if state.Status != QuizRuntimeStatusDone || !state.ClarificationComplete || state.AnswersRecorded < 1 || state.AnswersRecorded+state.SkipsRecorded != state.QuestionsAsked || state.ActiveQuestion != nil {
		return errors.New("quiz clarification is incomplete")
	}

	return nil
}

type guidedStageSeal struct {
	Stage  string
	Paths  []string
	Source GuidedStageCheckpoint
}

func sealGuidedStage(ctx context.Context, root string, session *AuthoringSession, seal guidedStageSeal) error {
	r := &fileAuthoringSessionRepository{projectDir: root}

	operation := session.Guided.PendingOperation
	if operation == nil {
		return ErrWorkflowConflict
	}

	if session.Guided.PendingCheckpoint == nil {
		checkpoint, err := r.observeGuidedCheckpoint(ctx, session.ID, *operation, seal.Stage, seal.Paths)
		if err != nil {
			return err
		}

		checkpoint.Sources = append(checkpoint.Sources, seal.Source.Sources...)
		for _, artifact := range seal.Source.Artifacts {
			artifact.Repository, artifact.Commit = seal.Source.Repository, seal.Source.Commit
			checkpoint.Sources = append(checkpoint.Sources, artifact)
		}

		session.Guided.PendingCheckpoint = &checkpoint
		if err := RetainGuidedUseCaseProgress(ctx, root, session); err != nil {
			return err
		}
	}

	pending := session.Guided.PendingCheckpoint
	if pending.OperationID != operation.ID || pending.Stage != seal.Stage || pending.Ref != guidedCheckpointRef(session.ID, operation.Digest) {
		return ErrWorkflowConflict
	}

	checkpoint, err := r.commitGuidedCheckpoint(ctx, *pending, operation.Digest)
	if err != nil {
		return err
	}

	session.Guided.Checkpoints = append(session.Guided.Checkpoints, checkpoint)
	session.Guided.PendingCheckpoint = nil

	return nil
}

func sameGuidedRevision(current, next AuthoringSession) bool {
	return current.Guided != nil && current.ProjectID == next.ProjectID && current.Namespace == next.Namespace && current.Guided.Actor == next.Guided.Actor && current.Revision == next.Revision
}

func guidedQuizSealReferences(ctx context.Context, root string, session AuthoringSession) ([]string, error) {
	if session.Guided.PendingCheckpoint != nil {
		return nil, nil
	}

	scope := QuizScope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}

	observed, err := NewQuizArtifactRepository(root).Observe(ctx, scope)
	if err != nil {
		return nil, err
	}

	if err := guardGuidedQuizCompletion(session, observed); err != nil {
		return nil, err
	}

	paths, err := QuizArtifactLayout(root, scope)
	if err != nil {
		return nil, err
	}

	return quizImageReferences(paths), nil
}

func guardGuidedPreparationInput(session AuthoringSession) error {
	if session.Guided.Stage != "preparation" {
		return ErrWorkflowConflict
	}

	if _, err := guidedStageCheckpoint(session, "preparation"); err == nil {
		return ErrWorkflowConflict
	}

	return nil
}

func guardGuidedQuizContent(operation GuidedUseCaseOperation) error {
	if operation.Action == "answer" && strings.TrimSpace(operation.Content) == "" {
		return errors.New("quiz answer is required")
	}

	if operation.Action == "quiz-mode" && !QuizMode(operation.Content).Valid() {
		return errors.New("quiz mode is invalid")
	}

	return nil
}
