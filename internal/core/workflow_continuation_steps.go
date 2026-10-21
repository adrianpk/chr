package core

import (
	"context"
	"errors"
)

func (s *WorkflowStore) continueWorkflowSteps(ctx context.Context, id string, request WorkflowContinuationRequest, ports WorkflowContinuationPorts, limit int) (int, string, error) {
	for steps := 0; steps < limit; steps++ {
		run, membershipChanged, err := s.readWorkflowContinuation(ctx, id)
		if err != nil {
			return steps, "blocked", err
		}

		if membershipChanged {
			continue
		}

		if workflowContinuationPreparationRetry(run, request, steps) {
			return s.retryWorkflowContinuationPreparation(ctx, id, ports)
		}

		if reason := workflowContinuationBefore(run); reason != "" {
			return steps, reason, nil
		}

		if err := s.advanceWorkflowContinuation(ctx, run, request, ports); err != nil {
			return steps, workflowContinuationFailureReason(err), err
		}

		after, err := s.ReadWorkflow(ctx, id)
		if err != nil {
			return steps, "blocked", err
		}

		accepted, reason := workflowContinuationAfter(run, after)
		if reason != "" {
			return steps + accepted, reason, nil
		}
	}

	return limit, "step_limit", nil
}

func workflowContinuationFailureReason(err error) string {
	if errors.Is(err, ErrWorkflowCapacity) {
		return "waiting_capacity"
	}

	return "blocked"
}

func (s *WorkflowStore) readWorkflowContinuation(ctx context.Context, id string) (WorkflowInvocation, bool, error) {
	run, err := s.ReadWorkflow(ctx, id)
	if err != nil {
		return run, false, err
	}

	refreshed, err := s.ReconcileWorkflowMembership(ctx, id)

	return run, refreshed.Revision != run.Revision, err
}

func workflowContinuationAfter(before, after WorkflowInvocation) (int, string) {
	if after.Revision == before.Revision {
		return 0, "awaiting_authority"
	}

	return 1, workflowContinuationBarrier(after)
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

	if run.Outcome == "closing_validation_failed" || run.Outcome == "interrupted" {
		return run.Outcome
	}

	if run.Status == "blocked" || run.Status == "waiting_capacity" {
		return run.Status
	}

	return ""
}

func (s *WorkflowStore) advanceWorkflowContinuation(ctx context.Context, run WorkflowInvocation, request WorkflowContinuationRequest, ports WorkflowContinuationPorts) error {
	state := workflowInvocationFile{WorkflowInvocation: run}
	restoreWorkflowCapacityWait(&state)

	run = state.WorkflowInvocation

	if workflowContinuationClosing(run) {
		_, err := s.CloseWorkflowInvocationSet(ctx, run.ID, request.PullRequest.BaseBranch, WorkflowFrontierPorts{Provider: ports.Publication.Provider, Observer: ports.Publication.Observer, Target: ports.Target})

		return err
	}

	if run.DeliveryRef != "" {
		return s.advanceWorkflowContinuationDelivery(ctx, run, ports)
	}

	if run.CandidateRef != "" {
		if workflowCandidateNeedsPR(run) {
			return s.ensureWorkflowContinuationPR(ctx, run, request, ports)
		}

		_, err := s.AdvanceWorkflowSliceIntegration(ctx, run.ID, ports.Publication)

		return err
	}

	if run.Lane != nil && run.Lane.Unit.Kind == "ticket" && workflowSliceValidationReady(run) {
		_, err := s.PrepareWorkflowTicketCandidate(ctx, run.ID)

		return err
	}

	if workflowSliceValidationReady(run) {
		_, err := s.ValidateWorkflowSlice(ctx, run.ID)

		return err
	}

	return s.executeWorkflowSliceTask(ctx, run.ID, request.Worker)
}

func (s *WorkflowStore) advanceWorkflowContinuationDelivery(ctx context.Context, run WorkflowInvocation, ports WorkflowContinuationPorts) error {
	if run.Integration != nil && run.ConfirmedCursor < run.Cursor {
		_, err := s.AdvanceWorkflowSliceIntegration(ctx, run.ID, ports.Publication)

		return err
	}

	_, err := s.AdvanceWorkflowFrontier(ctx, run.ID, WorkflowFrontierPorts{Provider: ports.Publication.Provider, Observer: ports.Publication.Observer, Target: ports.Target})

	return err
}

func (s *WorkflowStore) ensureWorkflowContinuationPR(ctx context.Context, run WorkflowInvocation, request WorkflowContinuationRequest, ports WorkflowContinuationPorts) error {
	if ports.Source != nil && (run.PullRequest == nil || run.PullRequest.State == "prepared" || run.PullRequest.State == "bound") {
		if err := s.PublishWorkflowSource(ctx, run.ID, ports.Source); err != nil {
			return err
		}
	}

	_, err := s.EnsureWorkflowSlicePullRequest(ctx, run.ID, request.PullRequest, ports.Publication.Provider)

	return err
}

func workflowContinuationClosing(run WorkflowInvocation) bool {
	return !workflowInvocationTerminal(run.Status) && run.StopIntent == "" && run.Status != "paused" && (run.Outcome == "set_closure_required" || run.Outcome == "closing_validation_failed" || run.Outcome == "interrupted" || run.ClosingSet != "")
}

func workflowContinuationBefore(run WorkflowInvocation) string {
	if run.Status == "waiting_capacity" && run.StopIntent == "" {
		return ""
	}

	if workflowContinuationClosing(run) {
		return ""
	}

	return workflowContinuationBarrier(run)
}

func workflowCandidateNeedsPR(run WorkflowInvocation) bool {
	return run.PullRequest == nil || run.PullRequest.State != "bound" || run.PullRequest.CandidateHash != run.CandidateHash
}
