package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) executeWorkflowControlResult(ctx context.Context, request WorkflowControlRequest) (*workflowstate.ControlResult, error) {
	if request.RequestBinding == nil {
		if request.AuthorizationCheck != nil && workflowDecisionControlAction(request.Action) {
			return nil, workflowstate.ErrConflict
		}

		return nil, s.executeWorkflowControl(ctx, request)
	}

	store, err := s.workflowStore()
	if err != nil {
		return nil, err
	}

	if !workflowDecisionControlAction(request.Action) {
		return s.executeWorkflowLifecycleResult(ctx, store, request)
	}

	var provider workflowcore.PullRequestProvider

	if request.Action != "grant" {
		reviewDependencies, dependencyErr := s.workflowReviewDependencies(request.ID)

		provider, err = reviewDependencies.Provider, dependencyErr
		if err != nil {
			return nil, err
		}
	}

	native := workflowNativeControlRequest(request)

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
