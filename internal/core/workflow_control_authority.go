package core

import "context"

// StopWorkflowGuarded checks volatile caller authority under native exclusion.
func (s *WorkflowStore) StopWorkflowGuarded(ctx context.Context, id, intent string, revision uint64, check WorkflowAdmissionCheck) (WorkflowInvocation, error) {
	if revision == 0 {
		return WorkflowInvocation{}, ErrWorkflowConflict
	}

	return s.stopWorkflowChecked(ctx, id, intent, check, revision)
}

// ReconcileWorkflowGuarded retains process ownership proof and rechecks caller
// authority before recovery. It never grants execution or additional attempts.
func (s *WorkflowStore) ReconcileWorkflowGuarded(ctx context.Context, id string, revision uint64, check WorkflowAdmissionCheck) (WorkflowInvocation, error) {
	if revision == 0 {
		return WorkflowInvocation{}, ErrWorkflowConflict
	}

	return s.reconcileWorkflowChecked(ctx, id, check, revision)
}
