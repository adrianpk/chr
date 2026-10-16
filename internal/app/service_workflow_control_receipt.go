package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) executeWorkflowControlResult(ctx context.Context, request WorkflowControlRequest) (*core.WorkflowControlResult, error) {
	if request.RequestBinding == nil {
		if request.AuthorizationCheck != nil && workflowDecisionControlAction(request.Action) {
			return nil, core.ErrWorkflowConflict
		}

		return nil, s.executeWorkflowControl(ctx, request)
	}

	store, err := s.workflowStore()
	if err != nil {
		return nil, err
	}

	var provider core.WorkflowPullRequestProvider
	if request.Action != "grant" {
		provider, err = s.workflowIntegrationProvider()
		if err != nil {
			return nil, err
		}
	}

	native := core.WorkflowDecisionControlRequest{Binding: *request.RequestBinding, Integration: request.Integration, Grant: request.BudgetGrant, Answer: request.DecisionAnswer, Action: request.Action, ID: request.ID, ExpectedRevision: request.ExpectedRevision, ReviewHash: request.BindingHash, DecisionRevision: request.DecisionRevision, DecisionHash: request.DecisionHash}

	result, err := store.ControlWorkflowDecision(ctx, native, provider, request.AuthorizationCheck)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func workflowDecisionControlAction(action string) bool {
	switch action {
	case "approve", "grant", "answer_decision":
		return true
	default:
		return false
	}
}
