package core

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// ReconcileGuidedResult is an owned lifecycle transition. Only complete native
// receipts can finish a run; pending operations and explicit stops take priority.
func ReconcileGuidedResult(ctx context.Context, root string, scope authoringcore.GuidedUseCaseScope, workflow authoringcore.GuidedWorkflowObserver) (session authoringcore.Session, err error) {
	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, scope.RunID)
		if err != nil {
			return err
		}

		if err := authoringcore.ValidateGuidedScope(session, scope); err != nil {
			return err
		}

		if !guidedResultReady(session) {
			return nil
		}

		result, err := workflow.InspectResult(ctx, session)
		if err != nil || !result.Complete {
			return err
		}

		reference := result.Reference

		session.Guided.Result = &reference
		session.Guided.Stage, session.Guided.Status, session.Guided.Blocker = "results", "completed", ""
		session.Status, session.FinishedAt = authoringcore.SessionStatusFinalized, &result.UpdatedAt

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})

	return session, err
}

func guidedResultReady(session authoringcore.Session) bool {
	state := session.Guided
	if !GuidedDispatchAllowed(session) || state.PendingOperation != nil || state.Stage != "execution" {
		return false
	}

	return state.Control == nil || state.Control.PendingOperation == nil
}
