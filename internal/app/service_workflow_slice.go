package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ExecuteWorkflowSlice executes the remaining ordered tasks and aggregate gate.
func (s service) ExecuteWorkflowSlice(ctx context.Context, id string, worker []string) (workflowcore.SliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.SliceCandidate{}, err
	}

	return store.ExecuteWorkflowSlice(ctx, id, worker)
}

// ValidateWorkflowSlice retries host acceptance while preserving task evidence.
func (s service) ValidateWorkflowSlice(ctx context.Context, id string) (workflowcore.SliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.SliceCandidate{}, err
	}

	return store.ValidateWorkflowSlice(ctx, id)
}

// ReadWorkflowSliceCandidate checks retained slice or ticket evidence and its lane.
func (s service) ReadWorkflowSliceCandidate(ctx context.Context, id string) (workflowcore.SliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.SliceCandidate{}, err
	}

	return store.ReadWorkflowSliceCandidate(ctx, id)
}

// RepairWorkflowSlice invalidates affected cumulative evidence before a retry.
func (s service) RepairWorkflowSlice(ctx context.Context, id, taskKey string) (workflowstate.Invocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.Invocation{}, err
	}

	return store.RepairWorkflowSlice(ctx, id, taskKey)
}
