package app

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowControlReceiptLookup pairs current status with original-receipt
// presence. Status errors after a verified receipt retain Found and that receipt.
type WorkflowControlReceiptLookup struct {
	Status WorkflowWorkStatus
	Found  bool
}

// ReadWorkControlReceipt observes an accepted control without constructing
// execution ports or permitting a fresh mutation through an expired prompt.
func (s service) ReadWorkControlReceipt(ctx context.Context, request WorkflowControlRequest) (WorkflowControlReceiptLookup, error) {
	if request.RequestBinding == nil {
		return WorkflowControlReceiptLookup{}, workflowstate.ErrConflict
	}

	store, err := s.workflowStore()
	if err != nil {
		return WorkflowControlReceiptLookup{}, err
	}

	observed, err := store.ReadWorkflowControlReceipt(ctx, workflowNativeControlRequest(request))
	if err != nil || !observed.Found {
		return WorkflowControlReceiptLookup{}, err
	}

	status, err := s.WorkflowStatus(ctx, request.ID, WorkflowStatusRequest{Limit: 8})
	status.ControlReceipt = &observed.Result

	return WorkflowControlReceiptLookup{Status: status, Found: true}, err
}

func workflowNativeControlRequest(request WorkflowControlRequest) workflowstate.ControlMutationRequest {
	native := workflowstate.ControlMutationRequest{Binding: *request.RequestBinding, Integration: request.Integration, Grant: request.BudgetGrant, Answer: request.DecisionAnswer, Action: request.Action, ID: request.ID, ExpectedRevision: request.ExpectedRevision, ReviewHash: request.BindingHash, DecisionRevision: request.DecisionRevision, DecisionHash: request.DecisionHash}
	if request.Action == "continue" {
		native.StepLimit = request.StepLimit
	}

	return native
}
