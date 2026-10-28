package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ReconcileWorkflowMembership preserves the invocation's accepted admission policy.
func (s service) ReconcileWorkflowMembership(ctx context.Context, id string) (workflowstate.Invocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.Invocation{}, err
	}

	return store.ReconcileWorkflowMembership(ctx, id)
}

// AmendWorkflowSet binds an approved changed plan to the observed quiescent set.
// The service supplies native key allocation rather than accepting adapter IO.
func (s service) AmendWorkflowSet(ctx context.Context, request workflowcore.SetAmendmentRequest) (workflowcore.MaterializationResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.MaterializationResult{}, err
	}

	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return workflowcore.MaterializationResult{}, err
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
