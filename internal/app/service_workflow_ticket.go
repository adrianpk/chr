package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// PrepareWorkflowTicketCandidate retains the independent task's accepted result
// for the same shared PR, review and integration operations used by slices.
func (s service) PrepareWorkflowTicketCandidate(ctx context.Context, id string) (workflowcore.SliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.SliceCandidate{}, err
	}

	return store.PrepareWorkflowTicketCandidate(ctx, id)
}
