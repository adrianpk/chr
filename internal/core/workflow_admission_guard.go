package core

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowAdmissionCheck revalidates volatile caller authority before admission
// or a guarded retained activation. Original admission replay bypasses the check.
type WorkflowAdmissionCheck func(context.Context) error

func validateWorkflowAdmissionGuard(guard *workflowstate.WorkflowAdmissionGuard, selection workflowstate.WorkflowPreviewRequest) error {
	if guard == nil {
		return nil
	}

	if !validWorkflowNativeFingerprint(guard.SnapshotHash) || selection.SnapshotHash != "" && selection.SnapshotHash != guard.SnapshotHash {
		return workflowstate.ErrConflict
	}

	if guard.Selected != nil && (!workflowstate.ValidRequestKey(guard.Selected.Key) || guard.Selected.Revision == 0) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowTicketAdmission(guard, selection); err != nil {
		return err
	}

	return validateWorkflowAdmissionPolicy(guard)
}

func validateWorkflowAdmissionPolicy(guard *workflowstate.WorkflowAdmissionGuard) error {
	if !validWorkflowExecutionPolicy(guard.ExecutionPolicy) || guard.ExecutionPolicy.ConfigRevision != guard.Policy.Revision {
		return workflowstate.ErrConflict
	}

	return validateWorkflowReviewPolicy(guard.Policy)
}

func (s *WorkflowStore) checkWorkflowAdmissionGuard(ctx context.Context, request WorkflowStartRequest) error {
	if err := s.checkGuidedWorkflowAdmission(ctx, request); err != nil {
		return err
	}

	if request.Admission != nil {
		policy, err := s.workflowReviewPolicy(workflowInvocationFile{WorkflowInvocation: WorkflowInvocation{WorkflowInvocationReview: WorkflowInvocationReview{ReviewRestriction: request.ReviewRestriction}}})
		if err != nil {
			return err
		}

		if policy != request.Admission.Policy || request.Policy != request.Admission.ExecutionPolicy {
			return workflowstate.ErrConflict
		}

		request.Selected = request.Admission.Selected
	}

	return s.checkWorkflowSelectionGuard(ctx, request)
}
