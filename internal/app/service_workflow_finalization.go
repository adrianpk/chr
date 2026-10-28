package app

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// FinalizeWorkflowSlice applies one verified shared receipt to all member tasks.
func (s service) FinalizeWorkflowSlice(ctx context.Context, id string) (workflowstate.Record, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.Record{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowstate.Record{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return workflowstate.Record{}, err
	}

	return store.FinalizeWorkflowSlice(ctx, id, provider, s.workflowIntegrationObserver(run))
}
