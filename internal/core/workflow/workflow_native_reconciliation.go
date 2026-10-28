package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReconcileWorkflowNativeIntegration explicitly observes the retained
// submission, including after its polling deadline. It cannot publish again or
// change a consumed budget, and confirms only canonical Git and PR evidence.
func (s *Store) ReconcileWorkflowNativeIntegration(ctx context.Context, id string, ports PublicationPorts) (workflowstate.Submission, error) {
	nativeContinuation, err := s.workflowNativeContinuationOwner(ctx, id)
	guard := nativeContinuation.Guard

	if err != nil {
		return workflowstate.Submission{}, err
	}

	if err := s.observeWorkflowNativeIntegration(ctx, guard, ports, workflowNativeObservationOnce); err != nil {
		return workflowstate.Submission{}, err
	}

	return s.ReadWorkflowSubmission(ctx, id)
}
