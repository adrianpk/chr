package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// PrepareWorkflowTicketCandidate retains the independent task's accepted result
// for the same shared PR, review and integration operations used by slices.
func (s service) PrepareWorkflowTicketCandidate(ctx context.Context, id string) (core.WorkflowSliceCandidate, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSliceCandidate{}, err
	}

	return store.PrepareWorkflowTicketCandidate(ctx, id)
}
