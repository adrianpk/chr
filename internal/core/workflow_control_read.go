package core

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadWorkflowControlReceipt recovers the native journal and reads a completed
// request. Missing receipts never reserve a binding or authorize a new effect.
func (s *WorkflowStore) ReadWorkflowControlReceipt(ctx context.Context, request workflowstate.WorkflowControlMutationRequest) (workflowstate.WorkflowControlResult, bool, error) {
	request, digest, err := freezeWorkflowControlMutation(request)
	if err != nil {
		return workflowstate.WorkflowControlResult{}, false, err
	}

	var (
		result workflowstate.WorkflowControlResult
		found  bool
	)

	err = s.withLock(ctx, func() error {
		_, owner, err := s.locateWorkflowInvocation(ctx, request.ID)
		if err != nil {
			return err
		}

		result, found, err = owner.readWorkflowControlReceipt(request, digest)

		return err
	})

	return result, found, err
}
