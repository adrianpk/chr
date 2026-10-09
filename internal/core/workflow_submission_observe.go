package core

import (
	"context"
	"fmt"
)

func observeWorkflowSubmission(ctx context.Context, submission WorkflowSubmission, provider WorkflowPublicationProvider, git WorkflowIntegrationGit) (WorkflowSubmission, error) {
	if provider == nil || git == nil {
		return submission, fmt.Errorf("canonical publication provider and Git observer required")
	}

	pr, err := provider.ReadWorkflowPullRequest(ctx, submission.Guard.PullRequest.URL)
	if err != nil {
		return submission, err
	}

	if !workflowSubmissionPRMatches(submission, pr) {
		return submission, ErrWorkflowConflict
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

func workflowSubmissionPRMatches(submission WorkflowSubmission, pr WorkflowPullRequestSnapshot) bool {
	expected := submission.Guard.PullRequest
	if pr.URL != expected.URL || pr.Repository != expected.Repository || pr.HeadRepository != expected.HeadRepository || pr.HeadBranch != expected.HeadBranch || pr.BaseBranch != expected.BaseBranch || pr.HeadCommit != expected.HeadCommit {
		return false
	}

	if pr.Merged {
		return pr.State == "closed" && workflowCommitID(pr.ResultCommit)
	}

	return pr.State == "open" || pr.State == "closed"
}

func checkWorkflowUnpublishedObservation(ctx context.Context, submission WorkflowSubmission, provider WorkflowPublicationProvider) error {
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

	if !workflowCommitID(target) {
		return ErrWorkflowConflict
	}

	return nil
}

func observeWorkflowSubmittedDelivery(ctx context.Context, submission WorkflowSubmission, pr WorkflowPullRequestSnapshot, provider WorkflowPublicationProvider, git WorkflowIntegrationGit) (WorkflowSubmission, error) {
	query := submission.Guard.Git.WorkflowIntegrationQuery
	query.ResultCommit, query.SourceDirectory = pr.ResultCommit, submission.PreparedLane.Worktree

	proof, err := git.VerifyWorkflowIntegration(ctx, query)
	if err != nil {
		return submission, err
	}

	if err := validateWorkflowIntegrationProof(query, proof); err != nil {
		return submission, err
	}

	submission.State, submission.Proof, submission.ObservedPR = "confirmed", &proof, &pr

	submission.Unpublished = nil
	if err := validateWorkflowSubmissionProof(submission); err != nil {
		return submission, err
	}

	if err := checkWorkflowSubmittedObservation(ctx, submission, provider); err != nil {
		return submission, err
	}

	return submission, nil
}

func checkWorkflowSubmittedObservation(ctx context.Context, submission WorkflowSubmission, provider WorkflowPublicationProvider) error {
	pr, err := provider.ReadWorkflowPullRequest(ctx, submission.Guard.PullRequest.URL)
	if err != nil {
		return err
	}

	if !workflowSubmissionPRMatches(submission, pr) || !pr.Merged || pr.ResultCommit != submission.Proof.ResultCommit {
		return ErrWorkflowConflict
	}

	target, err := provider.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)
	if err != nil {
		return err
	}

	if target != submission.Proof.ObservedTarget {
		return ErrWorkflowConflict
	}

	return nil
}
