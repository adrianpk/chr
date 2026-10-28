package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadWorkflowControlReceipt recovers the native journal and reads a completed
// request. Missing receipts never reserve a binding or authorize a new effect.
func (s *Store) ReadWorkflowControlReceipt(ctx context.Context, request workflowstate.ControlMutationRequest) (workflowstate.ControlReceiptLookup, error) {
	frozen, err := freezeWorkflowControlMutation(request)
	if err != nil {
		return workflowstate.ControlReceiptLookup{}, err
	}

	var result workflowstate.ControlReceiptLookup

	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, frozen.Request.ID)
		if err != nil {
			return err
		}

		result, err = invocationLocation.Owner.readWorkflowControlReceipt(frozen.Request, frozen.Digest)

		return err
	})

	return result, err
}
