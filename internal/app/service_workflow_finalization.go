package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// FinalizeWorkflowSlice applies one verified shared receipt to all member tasks.
func (s service) FinalizeWorkflowSlice(ctx context.Context, id string) (workflowcore.Record, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.Record{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowcore.Record{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return workflowcore.Record{}, err
	}

	return store.FinalizeWorkflowSlice(ctx, id, provider, s.workflowIntegrationObserver(run))
}
