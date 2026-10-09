package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// ReconcileWorkflowMembership preserves the invocation's accepted admission policy.
func (s service) ReconcileWorkflowMembership(ctx context.Context, id string) (core.WorkflowInvocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	return store.ReconcileWorkflowMembership(ctx, id)
}

// AmendWorkflowSet binds an approved changed plan to the observed quiescent set.
// The service supplies native key allocation rather than accepting adapter IO.
func (s service) AmendWorkflowSet(ctx context.Context, request core.WorkflowSetAmendmentRequest) (core.WorkflowMaterializationResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.WorkflowMaterializationResult{}, err
	}

	request.ReserveTicketKey = func(ctx context.Context) (string, error) {
		operations := s
		operations.ctx = ctx
		number, err := operations.reserveNextTicketKeyNumber(config.TicketKeyPrefix)

		return fmt.Sprintf("%s-%d", config.TicketKeyPrefix, number), err
	}

	result, err := store.AmendWorkflowSet(ctx, request)
	if err != nil {
		return result, err
	}

	return result, s.refreshWorkflowProjection(ctx, store)
}
