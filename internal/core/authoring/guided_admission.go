package authoring

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GuidedWorkAdmissionPolicy binds existing progression intent to native admission.
const GuidedWorkAdmissionPolicy = "guided-work/v1"

// GuidedPrerequisiteError carries a fixed, user-facing corrective action.
// Underlying provider diagnostics must not be interpolated into Action.
type GuidedPrerequisiteError struct {
	Action string
	Cause  error
}

func (e *GuidedPrerequisiteError) Error() string { return e.Action }
func (e *GuidedPrerequisiteError) Unwrap() error { return e.Cause }

// RetainGuidedHostBlocker reports a stopped activation without changing native facts.
func RetainGuidedHostBlocker(ctx context.Context, root string, scope GuidedUseCaseScope, invocation string) error {
	return WithMutation(ctx, root, func(ctx context.Context) error {
		r := &fileSessionRepository{projectDir: root}

		session, err := r.getGuardedSession(ctx, scope.RunID)
		if err != nil {
			return err
		}

		if err := ValidateGuidedScope(session, scope); err != nil {
			return err
		}

		state := session.Guided
		if state.Work == nil || state.Work.Invocation == nil || state.Work.Invocation.ID != invocation {
			return workflowstate.ErrConflict
		}

		if state.Status == "paused" || state.Status == "cancelled" || state.Status == "completed" {
			return nil
		}

		state.Status = "blocked"
		state.Blocker = "Managed activation stopped. Inspect the retained native work and correct its prerequisite before continuing."

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})
}

// GuidedPreparationCheckpoint returns the sealed canonical preparation input.
func GuidedPreparationCheckpoint(session Session) (GuidedStageCheckpoint, error) {
	if session.Guided == nil {
		return GuidedStageCheckpoint{}, workflowstate.ErrConflict
	}

	return guidedStageCheckpoint(session, "preparation")
}
