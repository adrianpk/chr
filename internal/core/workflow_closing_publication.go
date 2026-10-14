package core

import "context"

func (s *WorkflowStore) checkWorkflowClosingPublications(ctx context.Context) error {
	journals, err := s.workflowIntegrationJournals(ctx)
	if err != nil {
		return err
	}

	for _, journal := range journals {
		if journal.Attempts[len(journal.Attempts)-1].State == "awaiting_observation" {
			return ErrWorkflowObservationRequired
		}
	}

	return nil
}
