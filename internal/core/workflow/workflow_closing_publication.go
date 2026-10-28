package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) checkWorkflowClosingPublications(ctx context.Context) error {
	journals, err := s.workflowIntegrationJournals(ctx)
	if err != nil {
		return err
	}

	for _, journal := range journals {
		if journal.Attempts[len(journal.Attempts)-1].State == workflowstate.SubmissionAwaitingObservation {
			return ErrWorkflowObservationRequired
		}
	}

	return nil
}
