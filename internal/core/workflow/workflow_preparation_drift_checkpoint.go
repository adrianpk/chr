package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowNativePreparationTargetMoved(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, provider PublicationGuardProvider) bool {
	if candidate.Input.Native == nil || s.checkWorkflowNativePreparationFence(run, candidate.Input.Native) != nil || s.checkWorkflowPreparationGit(ctx, candidate) != nil {
		return false
	}

	reviewSource, err := s.workflowReviewSourceInputs(ctx, run, provider)

	source := reviewSource.Input

	return err == nil && source.TargetCommit != candidate.Input.Git.BaseCommit
}

func (s *Store) retainWorkflowNativeTargetDrift(id string) error {
	run, err := s.readWorkflowInvocation(id)
	if err != nil {
		return err
	}

	if run.StopIntent != "" || run.Status != workflowstate.InvocationPaused || run.ResumeStatus != workflowstate.InvocationWaitingIntegration {
		return workflowstate.ErrStop
	}

	run.Status, run.ResumeStatus, run.Outcome = workflowstate.InvocationWaitingIntegration, workflowstate.InvocationUnspecified, "target_changed"

	return s.commitWorkflowState(&run)
}
