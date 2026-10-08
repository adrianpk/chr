package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// ExecuteWorkflowSlice executes the remaining ordered tasks and aggregate gate.
func (s service) ExecuteWorkflowSlice(ctx context.Context, id string, worker []string) (core.WorkflowSliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSliceCandidate{}, err
	}

	return store.ExecuteWorkflowSlice(ctx, id, worker)
}

// ValidateWorkflowSlice retries host acceptance while preserving task evidence.
func (s service) ValidateWorkflowSlice(ctx context.Context, id string) (core.WorkflowSliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSliceCandidate{}, err
	}

	return store.ValidateWorkflowSlice(ctx, id)
}

// ReadWorkflowSliceCandidate checks retained aggregate evidence and its lane.
func (s service) ReadWorkflowSliceCandidate(ctx context.Context, id string) (core.WorkflowSliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSliceCandidate{}, err
	}

	return store.ReadWorkflowSliceCandidate(ctx, id)
}

// RepairWorkflowSlice invalidates affected cumulative evidence before a retry.
func (s service) RepairWorkflowSlice(ctx context.Context, id, taskKey string) (core.WorkflowInvocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	return store.RepairWorkflowSlice(ctx, id, taskKey)
}
