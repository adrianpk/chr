package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// The snapshot retains private journals and runs within the host read guard.
type workflowClosingStatusSource struct {
	workflowClosingEvidence
	snapshot workflowIntegrationSnapshot
}

func (source workflowClosingStatusSource) Publications() []workflowstate.SubmissionState {
	states := make([]workflowstate.SubmissionState, 0, len(source.snapshot.Journals))
	for _, journal := range source.snapshot.Journals {
		states = append(states, journal.Attempts[len(journal.Attempts)-1].State)
	}

	return states
}

func (source workflowClosingStatusSource) Capacity(ctx context.Context) (workflowstate.Capacity, error) {
	return source.store.workflowIntegrationCapacity(ctx, source.snapshot.Runs)
}
