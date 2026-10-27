package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// QuizActionBinding guards a retained question and makes retries idempotent.
type QuizActionBinding struct {
	SessionID, QuestionID, ExpectedRevision, OperationID string
}

// QuizActionRequest binds one shared quiz transition to its observed scope.
type QuizActionRequest struct {
	Scope   quizcore.Scope
	Binding QuizActionBinding
	Action  string
	Answer  string
	Mode    quizcore.Mode
}

// QuizObservation is a bounded, read-only snapshot. An absent session still has
// a revision so concurrent starts cannot overwrite one another.
type QuizObservation struct {
	Exists   bool
	State    quizcore.RuntimeState
	Entries  []quizcore.Entry
	Revision string
	images   []workflowFileChange
}

// QuizTransition is prepared under the shared authoring owner before publication.
type QuizTransition struct {
	State   quizcore.RuntimeState
	Entry   *quizcore.Entry
	Message string
}

// QuizTransitionFunc prepares model-dependent state with a bounded context.
// An answer followed by a provider failure must return a blocked transition,
// allowing the confirmed entry to publish independently of continuation.
type QuizTransitionFunc func(context.Context, QuizObservation) (QuizTransition, error)

// QuizActionResult retains the exact committed acknowledgement for a retry.
type QuizActionResult struct {
	OperationID string
	State       quizcore.RuntimeState
	Message     string
}

func validateQuizAction(request QuizActionRequest) error {
	if err := quizcore.ValidateScope(request.Scope); err != nil {
		return err
	}

	if err := authoringstate.ValidateSessionID(request.Binding.OperationID); err != nil {
		return err
	}

	if request.Binding.ExpectedRevision == "" {
		return errors.New("quiz revision is required")
	}

	if request.Action == "answer" && strings.TrimSpace(request.Answer) == "" {
		return errors.New("quiz answer is required")
	}

	switch request.Action {
	case "start", "resume", "answer", "skip", "done", "cancel", "mode":
		return nil
	default:
		return fmt.Errorf("unsupported quiz action: %s", request.Action)
	}
}

func validateQuizGenerationState(state quizcore.RuntimeState) error {
	if state.GenerationOperationID != "" && !validGenerationOperationIdentity(state.GenerationOperationID) {
		return errors.New("quiz generation operation identity exceeds its bound")
	}

	return quizcore.ValidateRuntimeState(state)
}

func validateRetainedQuizState(state quizcore.RuntimeState) error {
	if state.SchemaVersion != 0 && state.SchemaVersion != 2 {
		return errors.New("unsupported quiz runtime schema")
	}

	if err := validateQuizGenerationState(state); err != nil {
		return err
	}

	if state.SchemaVersion == 2 {
		if err := authoringstate.ValidateSessionID(state.SessionID); err != nil {
			return err
		}
	}

	if len(state.FastQueue) > quizcore.DefaultQuestionBudget || len(state.FastConsumed) > quizcore.DefaultQuestionBudget {
		return errors.New("quiz Fast batch capacity exceeded")
	}

	for _, candidate := range state.FastQueue {
		if err := quizcore.ValidateCandidate(candidate); err != nil {
			return err
		}
	}

	return nil
}

func prepareQuizTransition(before QuizObservation, next QuizTransition, request QuizActionRequest) (QuizTransition, error) {
	state := &next.State
	state.SchemaVersion = 2
	state.Revision = ""

	if state.ProjectID != request.Scope.ProjectID || state.CaseSlug != request.Scope.CaseSlug {
		return next, workflowstate.ErrConflict
	}

	if request.Action != "start" && state.SessionID != before.State.SessionID {
		return next, workflowstate.ErrConflict
	}

	if err := validateRetainedQuizState(*state); err != nil {
		return next, err
	}

	if next.Entry != nil {
		if err := quizcore.ValidateEntry(*next.Entry); err != nil {
			return next, err
		}
	}

	return next, nil
}

const quizProviderTimeout = 90 * time.Second
