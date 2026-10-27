package core

import (
	"context"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// ReconcileGuidedStop acknowledges quiescence only from the original native owner.
func ReconcileGuidedStop(ctx context.Context, root string, scope authoringcore.GuidedUseCaseScope, workflow authoringcore.GuidedWorkflowObserver) (session authoringcore.Session, err error) {
	r := &fileAuthoringSessionRepository{projectDir: root}
	err = WithAuthoringMutation(ctx, root, func(ctx context.Context) error {
		session, err = r.getGuardedSession(ctx, scope.RunID)
		if err != nil {
			return err
		}

		if err := authoringcore.ValidateGuidedScope(session, scope); err != nil {
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

func guidedStopNeedsAcknowledgement(control *authoringcore.GuidedControlState) bool {
	return control != nil && control.PendingOperation == nil && control.StopIntent != "" && control.NativeReceipt != nil && control.AcknowledgedRevision == 0
}
