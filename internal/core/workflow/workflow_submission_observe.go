package workflow

import (
	"context"
	"fmt"

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

	if !workflowSubmissionPRMatches(submission, pr) {
		return submission, workflowstate.ErrConflict
	}

	if pr.Merged {
		return observeWorkflowSubmittedDelivery(ctx, submission, pr, provider, git)
	}

	observation, err := provider.ObserveWorkflowSubmission(ctx, submission)
	if err != nil {
		return submission, err
	}

	if observation.SubmissionID != submission.ID || !observation.Unpublished {
		return submission, ErrWorkflowObservationRequired
	}

	if err := checkWorkflowUnpublishedObservation(ctx, submission, provider); err != nil {
		return submission, err
	}

	submission.State = "unpublished"
	submission.Unpublished = &observation

	return submission, nil
}

func workflowSubmissionPRMatches(submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot) bool {
	if submission.Native != nil {
		return workflowNativeSubmissionPRMatches(submission, pr)
	}

	if !workflowLegacySubmissionPRIdentityMatches(submission, pr) {
		return false
	}

	if pr.Merged {
		return pr.State == "closed" && workflowstate.ValidCommitID(pr.ResultCommit)
	}

	return pr.State == "open" || pr.State == "closed"
}

func checkWorkflowUnpublishedObservation(ctx context.Context, submission workflowstate.Submission, provider PullRequestProvider) error {
	pr, err := provider.ReadWorkflowPullRequest(ctx, submission.Guard.PullRequest.URL)
	if err != nil {
		return err
	}

	if !workflowSubmissionPRMatches(submission, pr) || pr.Merged {
		return ErrWorkflowObservationRequired
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

func observeWorkflowSubmittedDelivery(ctx context.Context, submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot, provider PullRequestProvider, git IntegrationGit) (workflowstate.Submission, error) {
	query := submission.Guard.Git.IntegrationQuery
	query.ResultCommit, query.SourceDirectory = pr.ResultCommit, submission.PreparedLane.Worktree

	proof, err := observeWorkflowSubmissionGit(ctx, git, query, submission.Guard.Git)
	if err != nil {
		return submission, err
	}

	if err := validateWorkflowIntegrationProof(query, proof); err != nil {
		return submission, err
	}

	submission.State, submission.Proof, submission.ObservedPR = "confirmed", &proof, &pr

	submission.Unpublished = nil

	submission.ExternalDelivery = proof.ResultCommit != submission.Guard.Git.ResultCommit || proof.BaseCommit != submission.Guard.Git.BaseCommit || proof.HistoryHash != submission.Guard.Git.HistoryHash
	if err := validateWorkflowSubmissionProof(submission); err != nil {
		return submission, err
	}

	if err := checkWorkflowSubmittedObservation(ctx, submission, provider); err != nil {
		return submission, err
	}

	return submission, nil
}

func checkWorkflowSubmittedObservation(ctx context.Context, submission workflowstate.Submission, provider PullRequestProvider) error {
	pr, err := provider.ReadWorkflowPullRequest(ctx, submission.Guard.PullRequest.URL)
	if err != nil {
		return err
	}

	if !workflowSubmissionPRMatches(submission, pr) || !pr.Merged || pr.ResultCommit != submission.Proof.ResultCommit {
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

func workflowNativeSubmissionPRMatches(submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot) bool {
	native := submission.Native
	expected := [5]string{native.CanonicalPR, native.SourceRepository, native.SourceRepository, native.SourceRef, native.TargetRef}
	actual := [5]string{pr.URL, pr.Repository, pr.HeadRepository, "refs/heads/" + pr.HeadBranch, "refs/heads/" + pr.BaseBranch}

	// PR head metadata may lag at S or advance after P. The immutable receipt's
	// transaction/application evidence establishes published P independently.
	return expected == actual && workflowstate.ValidCommitID(pr.HeadCommit) && pr.Merged && pr.State == "closed" && pr.ResultCommit == native.PreparedResult
}

func workflowLegacySubmissionPRIdentityMatches(submission workflowstate.Submission, pr workflowstate.PullRequestSnapshot) bool {
	expected := submission.Guard.PullRequest

	return pr.URL == expected.URL && pr.Repository == expected.Repository && pr.HeadRepository == expected.HeadRepository && pr.HeadBranch == expected.HeadBranch && pr.BaseBranch == expected.BaseBranch && pr.HeadCommit == expected.HeadCommit
}

func observeWorkflowSubmissionGit(ctx context.Context, git IntegrationGit, query workflowstate.IntegrationQuery, expected workflowstate.IntegrationProof) (workflowstate.IntegrationProof, error) {
	if query.ResultCommit == expected.ResultCommit {
		return observeWorkflowPreparedGit(ctx, git, query, expected)
	}

	return git.VerifyWorkflowIntegration(ctx, query)
}
