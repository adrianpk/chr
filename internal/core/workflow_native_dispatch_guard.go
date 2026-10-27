package core

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) checkWorkflowNativeDispatchGuard(ctx context.Context, run workflowInvocationFile, submission WorkflowSubmission, provider WorkflowPublicationGuardProvider) (WorkflowPublicationGuard, error) {
	if err := s.checkWorkflowNativeDispatchOwner(run, submission); err != nil {
		return WorkflowPublicationGuard{}, err
	}

	observation, input, err := s.workflowReviewSourceInputs(ctx, run, provider)
	if err != nil {
		return WorkflowPublicationGuard{}, err
	}

	candidate, err := s.currentWorkflowPreparation(ctx, run, input)
	if err != nil {
		return WorkflowPublicationGuard{}, err
	}

	input.PreparationRef, input.PreparationHash = run.Preparation.ResultRef, run.Preparation.ResultHash

	input.PreparedGit = &candidate.Input.Git
	if err := s.bindWorkflowNativeReviewCapability(ctx, run, candidate, provider, &input); err != nil {
		return WorkflowPublicationGuard{}, err
	}

	if err := s.checkWorkflowReviewAfterIO(ctx, run, observation, input.Policy); err != nil {
		return WorkflowPublicationGuard{}, err
	}

	hash, err := workflowReviewInputHash(input)
	if err != nil || hash != run.Review.InputHash {
		return WorkflowPublicationGuard{}, workflowstate.ErrConflict
	}

	capability, err := observeWorkflowPreparedCapability(ctx, provider, candidate.Input)
	if err != nil {
		return WorkflowPublicationGuard{}, err
	}

	return WorkflowPublicationGuard{PullRequest: input.PullRequest, Capability: capability, Git: candidate.Input.Git, PreparationRef: input.PreparationRef, PreparationHash: input.PreparationHash, ReviewHash: hash, Policy: input.Policy}, nil
}

func (s *WorkflowStore) checkWorkflowNativeDispatchOwner(run workflowInvocationFile, submission WorkflowSubmission) error {
	if run.Integration == nil || submission.Native == nil || run.Review == nil || run.Review.State != "authorized" {
		return workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if entry.State != "publishing" || entry.StopIntent != "" || !workflowNativeSubmissionEntryMatches(submission, entry) {
		return workflowstate.ErrConflict
	}

	if _, _, err := s.readWorkflowReview(run); err != nil {
		return err
	}

	return s.checkWorkflowIntegrationAuthorization(run, entry)
}
