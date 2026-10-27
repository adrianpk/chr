package core

import (
	"context"
	"path/filepath"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// BindGuidedControlInvocation reconciles uncertain admission without admitting
// new work. Only an existing exact request can become the control target.
func BindGuidedControlInvocation(ctx context.Context, root string, session *authoringcore.Session, workflow authoringcore.GuidedWorkflowObserver) error {
	absolute, err := filepath.Abs(root)
	if err != nil || !holdsAuthoringOwner(ctx, absolute) {
		return workflowstate.ErrConflict
	}

	state := session.Guided
	if state == nil {
		return workflowstate.ErrConflict
	}

	if !guidedControlNeedsInvocation(state) {
		return nil
	}

	if state.Control == nil || state.Control.PendingOperation == nil {
		return workflowstate.ErrConflict
	}

	lookup, err := workflow.FindAdmission(ctx, *session)
	if err != nil || !lookup.Found {
		return err
	}

	run := lookup.Invocation
	state.Work.Invocation = &authoringcore.GuidedInvocationReference{ID: run.ID, RequestKey: run.RequestKey, Revision: run.Revision}
	operation := *state.Control.PendingOperation
	operation.NativeRevision = run.Revision
	state.Control.NativeRequest = guidedNativeControlRequest(*session, operation)

	return RetainGuidedUseCaseProgress(ctx, root, session)
}

func guidedControlNeedsInvocation(state *authoringcore.GuidedUseCaseState) bool {
	return state.Work != nil && state.Work.Admission != nil && state.Work.Invocation == nil
}
