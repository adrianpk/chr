package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// ExportWorkflowTracker delegates derived export and repair to the common owner.
func (s service) ExportWorkflowTracker(ctx context.Context, setID string, opts core.WorkflowTrackerOptions) (core.WorkflowTrackerStatus, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowTrackerStatus{}, err
	}

	return store.ExportWorkflowTracker(ctx, setID, opts)
}

// InspectWorkflowTracker reports staleness without changing domain or projections.
func (s service) InspectWorkflowTracker(ctx context.Context, setID string) (core.WorkflowTrackerStatus, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowTrackerStatus{}, err
	}

	return store.InspectWorkflowTracker(ctx, setID)
}
