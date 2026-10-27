package app

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadWorkControlReceipt observes an accepted control without constructing
// execution ports or permitting a fresh mutation through an expired prompt.
func (s service) ReadWorkControlReceipt(ctx context.Context, request WorkflowControlRequest) (WorkflowWorkStatus, bool, error) {
	if request.RequestBinding == nil {
		return WorkflowWorkStatus{}, false, workflowstate.ErrConflict
	}

	store, err := s.workflowStore()
	if err != nil {
		return WorkflowWorkStatus{}, false, err
	}

	receipt, found, err := store.ReadWorkflowControlReceipt(ctx, workflowNativeControlRequest(request))
	if err != nil || !found {
		return WorkflowWorkStatus{}, false, err
	}

	status, err := s.WorkflowStatus(ctx, request.ID, WorkflowStatusRequest{Limit: 8})
	status.ControlReceipt = &receipt

	return status, true, err
}

func workflowNativeControlRequest(request WorkflowControlRequest) workflowstate.ControlMutationRequest {
	native := workflowstate.ControlMutationRequest{Binding: *request.RequestBinding, Integration: request.Integration, Grant: request.BudgetGrant, Answer: request.DecisionAnswer, Action: request.Action, ID: request.ID, ExpectedRevision: request.ExpectedRevision, ReviewHash: request.BindingHash, DecisionRevision: request.DecisionRevision, DecisionHash: request.DecisionHash}
	if request.Action == "continue" {
		native.StepLimit = request.StepLimit
	}

	return native
}
