package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// GrantWorkflowIntegration records an explicit bounded extension at the common
// invocation owner. It does not start work or alter consumed history.
func (s service) GrantWorkflowIntegration(ctx context.Context, request workflowcore.IntegrationGrantRequest) (workflowcore.IntegrationBudget, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.IntegrationBudget{}, err
	}

	return store.GrantWorkflowIntegration(ctx, request)
}

// ReadWorkflowIntegrationDecisions observes the candidate-bound operator wait.
func (s service) ReadWorkflowIntegrationDecisions(ctx context.Context, id string) (workflowcore.IntegrationDecisions, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.IntegrationDecisions{}, err
	}

	return store.ReadWorkflowIntegrationDecisions(ctx, id)
}

// RecordWorkflowIntegrationDecisions delegates bounded ambiguity to the common
// owner rather than creating an adapter-owned approval record.
func (s service) RecordWorkflowIntegrationDecisions(ctx context.Context, request workflowcore.IntegrationDecisionRequest) (workflowcore.IntegrationDecisions, error) {
	reviewDependencies, err := s.workflowReviewDependencies(request.InvocationID)
	store, provider := reviewDependencies.Store, reviewDependencies.Provider

	if err != nil {
		return workflowcore.IntegrationDecisions{}, err
	}

	return store.RecordWorkflowIntegrationDecisions(ctx, request, provider)
}

// AnswerWorkflowIntegrationDecision requires the displayed decision and all
// current mutation guards. The publication policy remains a separate authority.
func (s service) AnswerWorkflowIntegrationDecision(ctx context.Context, request workflowcore.IntegrationAnswerRequest) (workflowcore.IntegrationDecisions, error) {
	reviewDependencies2, err := s.workflowReviewDependencies(request.InvocationID)
	store, provider := reviewDependencies2.Store, reviewDependencies2.Provider

	if err != nil {
		return workflowcore.IntegrationDecisions{}, err
	}

	return store.AnswerWorkflowIntegrationDecision(ctx, request, provider)
}
