package workflow

import (
	"context"
	"fmt"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) awaitWorkflowPullRequest(run *workflowInvocationFile, snapshots []workflowstate.PullRequestSnapshot) error {
	run.PullRequest.State, run.PullRequest.Reason = "awaiting_observation", "creation_outcome_unknown"
	run.Status, run.Outcome = workflowstate.InvocationWaitingIntegration, "pr_awaiting_observation"

	change, err := s.workflowPullRequestObservation(run, snapshots)
	if err != nil {
		return err
	}

	if err := s.commitWorkflowPullRequestFiles(run, nil, change); err != nil {
		return err
	}

	return ErrWorkflowAwaitingObservation
}

func (s *Store) acceptWorkflowPullRequest(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, snapshots []workflowstate.PullRequestSnapshot) error {
	reason := workflowforge.PullRequestMismatch(*run.PullRequest, snapshots)
	if reason != "" {
		return s.blockWorkflowPullRequest(run, snapshots, reason)
	}

	if _, err := s.readWorkflowSliceCandidate(ctx, *run, observation); err != nil {
		return s.blockWorkflowPullRequest(run, snapshots, "candidate_changed_during_provider_io")
	}

	pr := snapshots[0]
	run.PullRequest.URL, run.PullRequest.State, run.PullRequest.Reason = pr.URL, "bound", ""

	if run.StopIntent == "" && run.Status != workflowstate.InvocationPaused && !workflowstate.InvocationTerminal(run.Status) {
		run.Status, run.Outcome = workflowstate.InvocationWaitingReview, "slice_pr_bound"
	}

	if pr.Merged {
		workflowRetainDeliveryStatus(run, "slice_pr_merged_unverified", false)
	}

	change, err := s.workflowPullRequestObservation(run, snapshots)
	if err != nil {
		return err
	}

	files, err := s.workflowPullRequestDomainChanges(run, observation)
	if err != nil {
		return err
	}

	return s.commitWorkflowPullRequestFiles(run, files, change)
}

func (s *Store) blockWorkflowPullRequest(run *workflowInvocationFile, snapshots []workflowstate.PullRequestSnapshot, reason string) error {
	run.PullRequest.State, run.PullRequest.Reason = "blocked", reason
	run.Status, run.Outcome = workflowstate.InvocationBlocked, reason
	run.Blocker = &workflowstate.Blocker{Dependency: run.PullRequest.Unit, Reason: reason, NextActionOwner: "workflow_operator"}

	change, err := s.workflowPullRequestObservation(run, snapshots)
	if err != nil {
		return err
	}

	if err := s.commitWorkflowPullRequestFiles(run, nil, change); err != nil {
		return err
	}

	return fmt.Errorf("%w: %s requires explicit reconciliation", workflowstate.ErrConflict, reason)
}

func (s *Store) beginWorkflowPullRequestCreation(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation) error {
	if _, err := s.readWorkflowSliceCandidate(ctx, *run, observation); err != nil {
		return err
	}

	binding := run.PullRequest
	binding.State, binding.Reason = "creating", ""

	return s.commitWorkflowState(run)
}
