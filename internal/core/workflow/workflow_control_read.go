package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadWorkflowControlReceipt recovers the native journal and reads a completed
// request. Missing receipts never reserve a binding or authorize a new effect.
func (s *Store) ReadWorkflowControlReceipt(ctx context.Context, request workflowstate.ControlMutationRequest) (workflowstate.ControlResult, bool, error) {
	request, digest, err := freezeWorkflowControlMutation(request)
	if err != nil {
		return workflowstate.ControlResult{}, false, err
	}

	var (
		result workflowstate.ControlResult
		found  bool
	)

	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, request.ID)
		if err != nil {
			return err
		}

		result, found, err = invocationLocation.Owner.readWorkflowControlReceipt(request, digest)

		return err
	})

	return result, found, err
}
