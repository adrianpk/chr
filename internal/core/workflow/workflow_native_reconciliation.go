package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReconcileWorkflowNativeIntegration explicitly observes the retained
// submission, including after its polling deadline. It cannot publish again or
// change a consumed budget, and confirms only canonical Git and PR evidence.
func ReconcileWorkflowNativeIntegration(ctx context.Context, store *Store, id string, ports PublicationPorts) (workflowstate.Submission, error) {
	nativeContinuation, err := store.workflowNativeContinuationOwner(ctx, id)
	guard := nativeContinuation.Guard

	if err != nil {
		return workflowstate.Submission{}, err
	}

	if err := observeWorkflowNativeIntegration(ctx, store, guard, ports, workflowNativeObservationOnce); err != nil {
		return workflowstate.Submission{}, err
	}

	return store.ReadWorkflowSubmission(ctx, id)
}
