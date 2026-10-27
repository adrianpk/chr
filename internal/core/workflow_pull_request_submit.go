package core

import (
	"context"
	"errors"
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) ensureWorkflowPullRequest(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, provider WorkflowPullRequestProvider) (WorkflowPullRequestSnapshot, error) {
	if run.PullRequest.URL != "" {
		pr, err := provider.ReadWorkflowPullRequest(ctx, run.PullRequest.URL)
		if err != nil {
			return pr, err
		}

		return pr, s.acceptWorkflowPullRequest(ctx, run, observation, []WorkflowPullRequestSnapshot{pr})
	}

	found, err := provider.FindWorkflowPullRequests(ctx, run.PullRequest.Request.WorkflowPullRequestQuery)
	if err != nil {
		return WorkflowPullRequestSnapshot{}, err
	}

	if len(found) > 0 {
		err := s.acceptWorkflowPullRequest(ctx, run, observation, found)

		return found[0], err
	}

	if run.PullRequest.State != "prepared" {
		return WorkflowPullRequestSnapshot{}, s.awaitWorkflowPullRequest(run, nil)
	}

	return s.createWorkflowPullRequest(ctx, run, observation, provider)
}

func (s *WorkflowStore) createWorkflowPullRequest(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, provider WorkflowPullRequestProvider) (WorkflowPullRequestSnapshot, error) {
	binding := run.PullRequest

	commit, err := provider.ReadWorkflowBranch(ctx, binding.Request.Repository, binding.Request.HeadBranch)
	if err != nil {
		return WorkflowPullRequestSnapshot{}, err
	}

	if commit != binding.SourceCommit {
		return WorkflowPullRequestSnapshot{}, fmt.Errorf("%w: canonical source branch differs from validated candidate", workflowstate.ErrConflict)
	}

	if _, err := s.readWorkflowSliceCandidate(ctx, *run, observation); err != nil {
		return WorkflowPullRequestSnapshot{}, err
	}

	binding.State, binding.Reason = "creating", ""

	if err := s.commitWorkflowState(run); err != nil {
		return WorkflowPullRequestSnapshot{}, err
	}

	pr, err := provider.CreateWorkflowPullRequest(ctx, binding.Request)
	if err != nil {
		return pr, errors.Join(err, s.awaitWorkflowPullRequest(run, []WorkflowPullRequestSnapshot{pr}))
	}

	return pr, s.acceptWorkflowPullRequest(ctx, run, observation, []WorkflowPullRequestSnapshot{pr})
}

func (s *WorkflowStore) awaitWorkflowPullRequest(run *workflowInvocationFile, snapshots []WorkflowPullRequestSnapshot) error {
	run.PullRequest.State, run.PullRequest.Reason = "awaiting_observation", "creation_outcome_unknown"
	run.Status, run.Outcome = "waiting_integration", "pr_awaiting_observation"

	change, err := s.workflowPullRequestObservation(run, snapshots)
	if err != nil {
		return err
	}

	if err := s.commitWorkflowPullRequestFiles(run, nil, change); err != nil {
		return err
	}

	return ErrWorkflowAwaitingObservation
}

func (s *WorkflowStore) acceptWorkflowPullRequest(ctx context.Context, run *workflowInvocationFile, observation workflowSliceObservation, snapshots []WorkflowPullRequestSnapshot) error {
	reason := workflowPullRequestMismatch(*run.PullRequest, snapshots)
	if reason != "" {
		return s.blockWorkflowPullRequest(run, snapshots, reason)
	}

	if _, err := s.readWorkflowSliceCandidate(ctx, *run, observation); err != nil {
		return s.blockWorkflowPullRequest(run, snapshots, "candidate_changed_during_provider_io")
	}

	pr := snapshots[0]
	run.PullRequest.URL, run.PullRequest.State, run.PullRequest.Reason = pr.URL, "bound", ""

	if run.StopIntent == "" && run.Status != "paused" && !workflowInvocationTerminal(run.Status) {
		run.Status, run.Outcome = "waiting_review", "slice_pr_bound"
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

func (s *WorkflowStore) blockWorkflowPullRequest(run *workflowInvocationFile, snapshots []WorkflowPullRequestSnapshot, reason string) error {
	run.PullRequest.State, run.PullRequest.Reason = "blocked", reason
	run.Status, run.Outcome = "blocked", reason
	run.Blocker = &WorkflowBlocker{Dependency: run.PullRequest.Unit, Reason: reason, NextActionOwner: "workflow_operator"}

	change, err := s.workflowPullRequestObservation(run, snapshots)
	if err != nil {
		return err
	}

	if err := s.commitWorkflowPullRequestFiles(run, nil, change); err != nil {
		return err
	}

	return fmt.Errorf("%w: %s requires explicit reconciliation", workflowstate.ErrConflict, reason)
}

func workflowPullRequestMismatch(binding WorkflowPullRequestBinding, snapshots []WorkflowPullRequestSnapshot) string {
	if len(snapshots) != 1 {
		return "canonical_pr_ambiguous"
	}

	pr := snapshots[0]
	if !workflowPullRequestIdentityMatches(binding, pr) {
		return "canonical_pr_mismatch"
	}

	if pr.State == "closed" && !pr.Merged {
		return "canonical_pr_closed_unmerged"
	}

	if pr.State != "open" && pr.State != "closed" {
		return "canonical_pr_state_unknown"
	}

	if pr.Merged && (pr.State != "closed" || !workflowCommitID(pr.ResultCommit)) {
		return "canonical_pr_merge_unverifiable"
	}

	return ""
}

func workflowPullRequestIdentityMatches(binding WorkflowPullRequestBinding, pr WorkflowPullRequestSnapshot) bool {
	if pr.Repository != binding.Request.Repository || pr.HeadRepository != pr.Repository || pr.HeadBranch != binding.Request.HeadBranch || pr.BaseBranch != binding.Request.BaseBranch || pr.HeadCommit != binding.SourceCommit || !workflowCommitID(pr.BaseCommit) {
		return false
	}

	if binding.URL != "" && binding.URL != pr.URL {
		return false
	}

	repository, err := workflowPullRequestRepository(pr.URL)

	return err == nil && repository == binding.Request.Repository
}

func workflowPullRequestRepository(raw string) (string, error) {
	reference, err := DetectReviewProviderReference(raw)
	if err != nil {
		return "", err
	}

	marker := "/pulls/"
	if reference.Provider == ReviewProviderGitHub {
		marker = "/pull/"
	}

	pos := len(raw) - len(marker) - len(fmt.Sprint(reference.Number))
	if pos < 0 || raw[pos:] != marker+fmt.Sprint(reference.Number) {
		return "", workflowstate.ErrConflict
	}

	return workflowCanonicalRepository(raw[:pos])
}
