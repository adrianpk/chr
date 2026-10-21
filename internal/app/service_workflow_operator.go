package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// GrantWorkflowIntegration records an explicit bounded extension at the common
// invocation owner. It does not start work or alter consumed history.
func (s service) GrantWorkflowIntegration(ctx context.Context, request core.WorkflowIntegrationGrantRequest) (core.WorkflowIntegrationBudget, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowIntegrationBudget{}, err
	}

	return store.GrantWorkflowIntegration(ctx, request)
}

// ReadWorkflowIntegrationDecisions observes the candidate-bound operator wait.
func (s service) ReadWorkflowIntegrationDecisions(ctx context.Context, id string) (core.WorkflowIntegrationDecisions, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowIntegrationDecisions{}, err
	}

	return store.ReadWorkflowIntegrationDecisions(ctx, id)
}

// RecordWorkflowIntegrationDecisions delegates bounded ambiguity to the common
// owner rather than creating an adapter-owned approval record.
func (s service) RecordWorkflowIntegrationDecisions(ctx context.Context, request core.WorkflowIntegrationDecisionRequest) (core.WorkflowIntegrationDecisions, error) {
	store, provider, err := s.workflowReviewDependencies(request.InvocationID)
	if err != nil {
		return core.WorkflowIntegrationDecisions{}, err
	}

	return store.RecordWorkflowIntegrationDecisions(ctx, request, provider)
}

// AnswerWorkflowIntegrationDecision requires the displayed decision and all
// current mutation guards. The publication policy remains a separate authority.
func (s service) AnswerWorkflowIntegrationDecision(ctx context.Context, request core.WorkflowIntegrationAnswerRequest) (core.WorkflowIntegrationDecisions, error) {
	store, provider, err := s.workflowReviewDependencies(request.InvocationID)
	if err != nil {
		return core.WorkflowIntegrationDecisions{}, err
	}

	return store.AnswerWorkflowIntegrationDecision(ctx, request, provider)
}
