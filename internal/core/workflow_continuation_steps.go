package core

import "context"

func (s *WorkflowStore) continueWorkflowSteps(ctx context.Context, id string, request WorkflowContinuationRequest, ports WorkflowContinuationPorts, limit int) (int, string, error) {
	for steps := 0; steps < limit; steps++ {
		run, err := s.ReadWorkflow(ctx, id)
		if err != nil {
			return steps, "blocked", err
		}

		if reason := workflowContinuationBarrier(run); reason != "" {
			return steps, reason, nil
		}

		if err := s.advanceWorkflowContinuation(ctx, run, request, ports); err != nil {
			return steps, "blocked", err
		}

		after, err := s.ReadWorkflow(ctx, id)
		if err != nil {
			return steps, "blocked", err
		}

		if after.Revision == run.Revision {
			return steps, "awaiting_authority", nil
		}

		if reason := workflowContinuationBarrier(after); reason != "" {
			return steps + 1, reason, nil
		}
	}

	return limit, "step_limit", nil
}

func workflowContinuationBarrier(run WorkflowInvocation) string {
	if workflowInvocationTerminal(run.Status) {
		return run.Status
	}

	if run.StopIntent != "" || run.Status == "paused" {
		return "stopped"
	}

	if run.Outcome == "set_closure_required" {
		return "set_closure_required"
	}

	if run.Status == "blocked" || run.Status == "waiting_capacity" {
		return run.Status
	}

	return ""
}

func (s *WorkflowStore) advanceWorkflowContinuation(ctx context.Context, run WorkflowInvocation, request WorkflowContinuationRequest, ports WorkflowContinuationPorts) error {
	if run.DeliveryRef != "" {
		_, err := s.AdvanceWorkflowFrontier(ctx, run.ID, WorkflowFrontierPorts{Provider: ports.Publication.Provider, Observer: ports.Publication.Observer, Target: ports.Target})

		return err
	}

	if run.CandidateRef != "" {
		if run.PullRequest == nil || run.PullRequest.State != "bound" {
			_, err := s.EnsureWorkflowSlicePullRequest(ctx, run.ID, request.PullRequest, ports.Publication.Provider)

			return err
		}

		_, err := s.AdvanceWorkflowSliceIntegration(ctx, run.ID, ports.Publication)

		return err
	}

	if workflowSliceValidationReady(run) {
		_, err := s.ValidateWorkflowSlice(ctx, run.ID)

		return err
	}

	return s.executeWorkflowSliceTask(ctx, run.ID, request.Worker)
}
