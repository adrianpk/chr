package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// QuizActionBinding guards a retained question and makes retries idempotent.
type QuizActionBinding struct {
	SessionID, QuestionID, ExpectedRevision, OperationID string
}

// QuizActionRequest binds one shared quiz transition to its observed scope.
type QuizActionRequest struct {
	Scope   QuizScope
	Binding QuizActionBinding
	Action  string
	Answer  string
	Mode    QuizMode
}

// QuizObservation is a bounded, read-only snapshot. An absent session still has
// a revision so concurrent starts cannot overwrite one another.
type QuizObservation struct {
	Exists   bool
	State    QuizRuntimeState
	Entries  []QuizEntry
	Revision string
	images   []workflowFileChange
}

// QuizTransition is prepared under the shared authoring owner before publication.
type QuizTransition struct {
	State   QuizRuntimeState
	Entry   *QuizEntry
	Message string
}

// QuizTransitionFunc prepares model-dependent state with a bounded context.
// An answer followed by a provider failure must return a blocked transition,
// allowing the confirmed entry to publish independently of continuation.
type QuizTransitionFunc func(context.Context, QuizObservation) (QuizTransition, error)

// QuizActionResult retains the exact committed acknowledgement for a retry.
type QuizActionResult struct {
	OperationID string
	State       QuizRuntimeState
	Message     string
}

func validateQuizAction(request QuizActionRequest) error {
	if err := ValidateQuizScope(request.Scope); err != nil {
		return err
	}

	if err := validateAuthoringSessionID(request.Binding.OperationID); err != nil {
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

func validateRetainedQuizState(state QuizRuntimeState) error {
	if state.SchemaVersion != 0 && state.SchemaVersion != 2 {
		return errors.New("unsupported quiz runtime schema")
	}

	if err := ValidateQuizRuntimeState(state); err != nil {
		return err
	}

	if state.SchemaVersion == 2 {
		if err := validateAuthoringSessionID(state.SessionID); err != nil {
			return err
		}
	}

	if len(state.FastQueue) > DefaultQuizQuestionBudget || len(state.FastConsumed) > DefaultQuizQuestionBudget {
		return errors.New("quiz Fast batch capacity exceeded")
	}

	for _, candidate := range state.FastQueue {
		if err := ValidateQuizCandidate(candidate); err != nil {
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
		return next, ErrWorkflowConflict
	}

	if request.Action != "start" && state.SessionID != before.State.SessionID {
		return next, ErrWorkflowConflict
	}

	if err := validateRetainedQuizState(*state); err != nil {
		return next, err
	}

	if next.Entry != nil {
		if err := ValidateQuizEntry(*next.Entry); err != nil {
			return next, err
		}
	}

	return next, nil
}

const quizProviderTimeout = 90 * time.Second
