package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowContinuationPreparationRetry(run workflowstate.Invocation, request ContinuationRequest, steps int) bool {
	return steps == 0 && request.ExpectedRevision != 0 && run.Status == workflowstate.InvocationBlocked && run.Outcome == "integration_preparation_failed" && run.StopIntent == "" && run.Integration != nil && workflowInterruptedPreparation(run)
}

func workflowInterruptedPreparation(run workflowstate.Invocation) bool {
	if run.Preparation == nil || len(run.Attempts) == 0 {
		return false
	}

	last := run.Attempts[len(run.Attempts)-1]

	return run.Preparation.State == "stale" && run.Preparation.ResultRef == "" && last.Purpose == workflowIntegrationValidationPurpose && last.Outcome == "interrupted" && last.InputRef == ""
}

// An accepted operator continuation can retry preparation that failed before
// its gate started. The common preparer retains counts and rechecks every guard.
// Failed gates, unbound calls and receipt replay cannot take this path.
func (s *Store) retryWorkflowContinuationPreparation(ctx context.Context, id string, ports ContinuationPorts) (workflowContinuationStep, error) {
	provider, ok := ports.Publication.Provider.(PublicationGuardProvider)
	if !ok || ports.Publication.Preparer == nil {
		return workflowContinuationStep{Steps: 0, Reason: "blocked"}, workflowstate.ErrAtomicGuard
	}

	if _, err := PrepareIntegration(ctx, s, id, provider, ports.Publication.Preparer); err != nil {
		return workflowContinuationStep{Steps: 0, Reason: "blocked"}, err
	}

	return workflowContinuationStep{Steps: 1, Reason: "step_limit"}, nil
}
