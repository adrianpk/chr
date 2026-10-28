package workflow

import "context"

func (s *Store) workflowNativePreparationTargetMoved(ctx context.Context, run workflowInvocationFile, candidate IntegrationCandidate, provider PublicationGuardProvider) bool {
	if candidate.Input.Native == nil || s.checkWorkflowNativePreparationFence(run, candidate.Input.Native) != nil || s.checkWorkflowPreparationGit(ctx, candidate) != nil {
		return false
	}

	_, source, err := s.workflowReviewSourceInputs(ctx, run, provider)

	return err == nil && source.TargetCommit != candidate.Input.Git.BaseCommit
}

func (s *Store) retainWorkflowNativeTargetDrift(id string) error {
	run, err := s.readWorkflowInvocation(id)
	if err != nil {
		return err
	}

	if run.StopIntent != "" || run.Status != "paused" || run.ResumeStatus != "waiting_integration" {
		return ErrWorkflowStop
	}

	run.Status, run.ResumeStatus, run.Outcome = "waiting_integration", "", "target_changed"

	return s.commitWorkflowState(&run)
}
