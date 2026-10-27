package quiz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ActionBinding guards a retained question and makes retries idempotent.
type ActionBinding struct {
	SessionID, QuestionID, ExpectedRevision, OperationID string
}

// ActionRequest binds one shared quiz transition to its observed scope.
type ActionRequest struct {
	Scope   Scope
	Binding ActionBinding
	Action  string
	Answer  string
	Mode    Mode
}

// Observation is a bounded, read-only snapshot. An absent session still has
// a revision so concurrent starts cannot overwrite one another.
type Observation struct {
	Exists   bool
	State    RuntimeState
	Entries  []Entry
	Revision string
}

// Transition is prepared under the shared authoring owner before publication.
type Transition struct {
	State   RuntimeState
	Entry   *Entry
	Message string
}

// TransitionFunc prepares model-dependent state with a bounded context.
// An answer followed by a provider failure must return a blocked transition,
// allowing the confirmed entry to publish independently of continuation.
type TransitionFunc func(context.Context, Observation) (Transition, error)

// ActionResult retains the exact committed acknowledgement for a retry.
type ActionResult struct {
	OperationID string
	State       RuntimeState
	Message     string
}

// ValidateAction checks the original scope, operation identity and action binding.
func ValidateAction(request ActionRequest) error {
	if err := ValidateScope(request.Scope); err != nil {
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

// ValidateGenerationState retains the generation identity and runtime validation bounds.
func ValidateGenerationState(state RuntimeState) error {
	if state.GenerationOperationID != "" && !authoringstate.ValidOperationIdentity(state.GenerationOperationID) {
		return errors.New("quiz generation operation identity exceeds its bound")
	}

	return ValidateRuntimeState(state)
}

// ValidateRetainedState checks supported schemas and retained Fast question capacity.
func ValidateRetainedState(state RuntimeState) error {
	if state.SchemaVersion != 0 && state.SchemaVersion != 2 {
		return errors.New("unsupported quiz runtime schema")
	}

	if err := ValidateGenerationState(state); err != nil {
		return err
	}

	if state.SchemaVersion == 2 {
		if err := authoringstate.ValidateSessionID(state.SessionID); err != nil {
			return err
		}
	}

	if len(state.FastQueue) > DefaultQuestionBudget || len(state.FastConsumed) > DefaultQuestionBudget {
		return errors.New("quiz Fast batch capacity exceeded")
	}

	for _, candidate := range state.FastQueue {
		if err := ValidateCandidate(candidate); err != nil {
			return err
		}
	}

	return nil
}

// PrepareTransition validates a transition against the original observed binding.
func PrepareTransition(before Observation, next Transition, request ActionRequest) (Transition, error) {
	state := &next.State
	state.SchemaVersion = 2
	state.Revision = ""

	if state.ProjectID != request.Scope.ProjectID || state.CaseSlug != request.Scope.CaseSlug {
		return next, workflowstate.ErrConflict
	}

	if request.Action != "start" && state.SessionID != before.State.SessionID {
		return next, workflowstate.ErrConflict
	}

	if err := ValidateRetainedState(*state); err != nil {
		return next, err
	}

	if next.Entry != nil {
		if err := ValidateEntry(*next.Entry); err != nil {
			return next, err
		}
	}

	return next, nil
}
