package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// ExportWorkflowTracker delegates derived export and repair to the common owner.
func (s service) ExportWorkflowTracker(ctx context.Context, setID string, opts workflowcore.TrackerOptions) (workflowcore.TrackerStatus, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.TrackerStatus{}, err
	}

	return store.ExportWorkflowTracker(ctx, setID, opts)
}

// InspectWorkflowTracker reports staleness without changing domain or projections.
func (s service) InspectWorkflowTracker(ctx context.Context, setID string) (workflowcore.TrackerStatus, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.TrackerStatus{}, err
	}

	return store.InspectWorkflowTracker(ctx, setID)
}
