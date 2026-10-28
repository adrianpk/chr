package workflow

import (
	"context"
	"fmt"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func observeWorkflowSubmission(ctx context.Context, submission workflowstate.Submission, provider PublicationProvider, git IntegrationGit) (workflowstate.Submission, error) {
	if provider == nil || git == nil {
		return submission, fmt.Errorf("canonical publication provider and Git observer required")
	}

	pr, err := provider.ReadWorkflowPullRequest(ctx, submission.Guard.PullRequest.URL)
	if err != nil {
		return submission, err
	}

	if !workflowforge.MatchesPullRequest(submission, pr) {
		return submission, workflowstate.ErrConflict
	}

	if pr.Merged {
		return observeWorkflowSubmittedDelivery(ctx, submission, pr, workflowDeliveryObservers{Provider: provider, Git: git})
	}

	observation, err := provider.ObserveWorkflowSubmission(ctx, submission)
	if err != nil {
		return submission, err
	}

	if observation.SubmissionID != submission.ID || !observation.Unpublished {
		return submission, workflowstate.ErrObservationRequired
	}

	if err := checkWorkflowUnpublishedObservation(ctx, submission, provider); err != nil {
		return submission, err
	}

	submission.State = workflowstate.SubmissionUnpublished
	submission.Unpublished = &observation

	return submission, nil
}

func checkWorkflowUnpublishedObservation(ctx context.Context, submission workflowstate.Submission, provider PullRequestProvider) error {
	pr, err := provider.ReadWorkflowPullRequest(ctx, submission.Guard.PullRequest.URL)
	if err != nil {
		return err
	}

	if !workflowforge.MatchesPullRequest(submission, pr) || pr.Merged {
		return workflowstate.ErrObservationRequired
	}

	target, err := provider.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)
	if err != nil {
		return err
	}

	if !workflowstate.ValidCommitID(target) {
		return workflowstate.ErrConflict
	}

	return nil
}

func observeWorkflowSubmittedDelivery(ctx context.Context, submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot, observers workflowDeliveryObservers) (workflowstate.Submission, error) {
	query := submission.Guard.Git.IntegrationQuery
	query.ResultCommit, query.SourceDirectory = pr.ResultCommit, submission.PreparedLane.Worktree

	proof, err := observeWorkflowSubmissionGit(ctx, observers.Git, query, submission.Guard.Git)
	if err != nil {
		return submission, err
	}

	if err := workflowintegrate.CheckProof(query, proof, workflowIntegrationRepairs{}); err != nil {
		return submission, err
	}

	submission.State, submission.Proof, submission.ObservedPR = workflowstate.SubmissionConfirmed, &proof, &pr

	submission.Unpublished = nil

	submission.ExternalDelivery = proof.ResultCommit != submission.Guard.Git.ResultCommit || proof.BaseCommit != submission.Guard.Git.BaseCommit || proof.HistoryHash != submission.Guard.Git.HistoryHash
	if err := validateWorkflowSubmissionProof(submission); err != nil {
		return submission, err
	}

	if err := checkWorkflowSubmittedObservation(ctx, submission, observers.Provider); err != nil {
		return submission, err
	}

	return submission, nil
}

func checkWorkflowSubmittedObservation(ctx context.Context, submission workflowstate.Submission, provider PullRequestProvider) error {
	pr, err := provider.ReadWorkflowPullRequest(ctx, submission.Guard.PullRequest.URL)
	if err != nil {
		return err
	}

	if !workflowforge.MatchesPullRequest(submission, pr) || !pr.Merged || pr.ResultCommit != submission.Proof.ResultCommit {
		return workflowstate.ErrConflict
	}

	target, err := provider.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)
	if err != nil {
		return err
	}

	if target != submission.Proof.ObservedTarget {
		return workflowstate.ErrConflict
	}

	return nil
}

func observeWorkflowSubmissionGit(ctx context.Context, git IntegrationGit, query workflowstate.IntegrationQuery, expected workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error) {
	if query.ResultCommit == expected.ResultCommit {
		return workflowforge.ObservePreparedGit(ctx, git, query, expected)
	}

	return git.VerifyWorkflowIntegration(ctx, query)
}
