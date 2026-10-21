package core

import "context"

// ReconcileWorkflowNativeIntegration explicitly observes the retained
// submission, including after its polling deadline. It cannot publish again or
// change a consumed budget, and confirms only canonical Git and PR evidence.
func (s *WorkflowStore) ReconcileWorkflowNativeIntegration(ctx context.Context, id string, ports WorkflowPublicationPorts) (WorkflowSubmission, error) {
	_, guard, err := s.workflowNativeContinuationOwner(ctx, id)
	if err != nil {
		return WorkflowSubmission{}, err
	}

	if err := s.observeWorkflowNativeIntegration(ctx, guard, ports, false); err != nil {
		return WorkflowSubmission{}, err
	}

	return s.ReadWorkflowSubmission(ctx, id)
}
