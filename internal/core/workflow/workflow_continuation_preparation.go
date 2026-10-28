package workflow

import "context"

func workflowContinuationPreparationRetry(run Invocation, request ContinuationRequest, steps int) bool {
	return steps == 0 && request.ExpectedRevision != 0 && run.Status == "blocked" && run.Outcome == "integration_preparation_failed" && run.StopIntent == "" && run.Integration != nil && workflowInterruptedPreparation(run)
}

func workflowInterruptedPreparation(run Invocation) bool {
	if run.Preparation == nil || len(run.Attempts) == 0 {
		return false
	}

	last := run.Attempts[len(run.Attempts)-1]

	return run.Preparation.State == "stale" && run.Preparation.ResultRef == "" && last.Purpose == workflowIntegrationValidationPurpose && last.Outcome == "interrupted" && last.InputRef == ""
}

// An accepted operator continuation can retry preparation that failed before
// its gate started. The common preparer retains counts and rechecks every guard.
// Failed gates, unbound calls and receipt replay cannot take this path.
func (s *Store) retryWorkflowContinuationPreparation(ctx context.Context, id string, ports ContinuationPorts) (int, string, error) {
	provider, ok := ports.Publication.Provider.(PublicationGuardProvider)
	if !ok || ports.Publication.Preparer == nil {
		return 0, "blocked", ErrWorkflowAtomicGuard
	}

	if _, err := s.PrepareWorkflowSliceIntegration(ctx, id, provider, ports.Publication.Preparer); err != nil {
		return 0, "blocked", err
	}

	return 1, "step_limit", nil
}
