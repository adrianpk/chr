package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) checkWorkflowNativeDispatchGuard(ctx context.Context, run workflowInvocationFile, submission Submission, provider PublicationGuardProvider) (PublicationGuard, error) {
	if err := s.checkWorkflowNativeDispatchOwner(run, submission); err != nil {
		return PublicationGuard{}, err
	}

	observation, input, err := s.workflowReviewSourceInputs(ctx, run, provider)
	if err != nil {
		return PublicationGuard{}, err
	}

	candidate, err := s.currentWorkflowPreparation(ctx, run, input)
	if err != nil {
		return PublicationGuard{}, err
	}

	input.PreparationRef, input.PreparationHash = run.Preparation.ResultRef, run.Preparation.ResultHash

	input.PreparedGit = &candidate.Input.Git
	if err := s.bindWorkflowNativeReviewCapability(ctx, run, candidate, provider, &input); err != nil {
		return PublicationGuard{}, err
	}

	if err := s.checkWorkflowReviewAfterIO(ctx, run, observation, input.Policy); err != nil {
		return PublicationGuard{}, err
	}

	hash, err := workflowReviewInputHash(input)
	if err != nil || hash != run.Review.InputHash {
		return PublicationGuard{}, workflowstate.ErrConflict
	}

	capability, err := observeWorkflowPreparedCapability(ctx, provider, candidate.Input)
	if err != nil {
		return PublicationGuard{}, err
	}

	return PublicationGuard{PullRequest: input.PullRequest, Capability: capability, Git: candidate.Input.Git, PreparationRef: input.PreparationRef, PreparationHash: input.PreparationHash, ReviewHash: hash, Policy: input.Policy}, nil
}

func (s *Store) checkWorkflowNativeDispatchOwner(run workflowInvocationFile, submission Submission) error {
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
