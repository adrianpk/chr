package core

import "context"

// WorkflowAdmissionCheck revalidates volatile caller authority before admission
// or a guarded retained activation. Original admission replay bypasses the check.
type WorkflowAdmissionCheck func(context.Context) error

// WorkflowAdmissionGuard retains exactly what a caller displayed before admission.
// It is part of the request identity; replay observes the original invocation.
type WorkflowAdmissionGuard struct {
	ExecutionPolicy WorkflowExecutionPolicy `json:"execution_policy"`
	SnapshotHash    string                  `json:"snapshot_hash"`
	Selected        *WorkflowSelectionGuard `json:"selected,omitempty"`
	Policy          WorkflowReviewPolicy    `json:"publication_policy"`
}

func validateWorkflowAdmissionGuard(guard *WorkflowAdmissionGuard, selection WorkflowPreviewRequest) error {
	if guard == nil {
		return nil
	}

	if !validWorkflowNativeFingerprint(guard.SnapshotHash) || selection.SnapshotHash != "" && selection.SnapshotHash != guard.SnapshotHash {
		return ErrWorkflowConflict
	}

	if guard.Selected != nil && (!validWorkflowRequestKey(guard.Selected.Key) || guard.Selected.Revision == 0) {
		return ErrWorkflowConflict
	}

	if !validWorkflowExecutionPolicy(guard.ExecutionPolicy) || guard.ExecutionPolicy.ConfigRevision != guard.Policy.Revision {
		return ErrWorkflowConflict
	}

	return validateWorkflowReviewPolicy(guard.Policy)
}

func (s *WorkflowStore) checkWorkflowAdmissionGuard(ctx context.Context, request WorkflowStartRequest) error {
	if request.Admission != nil {
		policy, err := s.workflowReviewPolicy(workflowInvocationFile{WorkflowInvocation: WorkflowInvocation{ReviewRestriction: request.ReviewRestriction}})
		if err != nil {
			return err
		}

		if policy != request.Admission.Policy || request.Policy != request.Admission.ExecutionPolicy {
			return ErrWorkflowConflict
		}

		request.Selected = request.Admission.Selected
	}

	return s.checkWorkflowSelectionGuard(ctx, request)
}
