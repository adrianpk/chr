package authoring

import (
	"context"
)

// ReconcileGuidedStop acknowledges quiescence only from the original native owner.
func ReconcileGuidedStop(ctx context.Context, root string, scope GuidedUseCaseScope, workflow GuidedWorkflowObserver) (session Session, err error) {
	r := &fileSessionRepository{projectDir: root}
	err = WithMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, scope.RunID)
		if err != nil {
			return err
		}

		if err := ValidateGuidedScope(session, scope); err != nil {
			return err
		}

		control := session.Guided.Control
		if !guidedStopNeedsAcknowledgement(control) {
			return nil
		}

		run, err := workflow.ObserveStop(ctx, session, control.StopIntent)
		if err != nil {
			return err
		}

		if !run.Acknowledged {
			return nil
		}

		control.AcknowledgedRevision = run.Revision
		session.Guided.Status, session.Guided.Blocker = run.Status, ""

		return RetainGuidedUseCaseProgress(ctx, root, &session)
	})

	return session, err
}

func guidedStopNeedsAcknowledgement(control *GuidedControlState) bool {
	return control != nil && control.PendingOperation == nil && control.StopIntent != "" && control.NativeReceipt != nil && control.AcknowledgedRevision == 0
}
