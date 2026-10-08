package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// FinalizeWorkflowSlice applies one verified shared receipt to all member tasks.
func (s service) FinalizeWorkflowSlice(ctx context.Context, id string) (core.WorkflowRecord, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowRecord{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return core.WorkflowRecord{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return core.WorkflowRecord{}, err
	}

	return store.FinalizeWorkflowSlice(ctx, id, provider, s.workflowIntegrationObserver(run))
}
