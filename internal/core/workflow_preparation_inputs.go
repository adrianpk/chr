package core

import (
	"context"
	"fmt"
	"slices"
)

func observeWorkflowPublicationCapability(ctx context.Context, provider WorkflowPublicationGuardProvider, source WorkflowReviewInput) (WorkflowPublicationCapability, error) {
	if provider == nil {
		return WorkflowPublicationCapability{}, ErrWorkflowAtomicGuard
	}

	head, err := provider.ReadWorkflowBranch(ctx, source.PullRequest.Repository, source.PullRequest.HeadBranch)
	if err != nil {
		return WorkflowPublicationCapability{}, err
	}

	if head != source.PullRequest.HeadCommit {
		return WorkflowPublicationCapability{}, ErrWorkflowConflict
	}

	capability, err := provider.WorkflowIntegrationCapability(ctx, source.PullRequest, source.Policy.Method)
	if err != nil {
		return capability, err
	}

	return capability, validateWorkflowPublicationCapability(source, capability)
}

func validateWorkflowPublicationCapability(source WorkflowReviewInput, capability WorkflowPublicationCapability) error {
	if capability.Repository != source.PullRequest.Repository || capability.PullRequestURL != source.PullRequest.URL || capability.Method != source.Policy.Method || capability.Mechanism == "" || len(capability.Mechanism) > 128 || len(capability.Revision) != 64 {
		return ErrWorkflowConflict
	}

	if !capability.AtomicSourceBase || !capability.ExactResult {
		return ErrWorkflowAtomicGuard
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowPublicationGuard(ctx context.Context, run workflowInvocationFile, provider WorkflowPublicationGuardProvider) (WorkflowPublicationGuard, error) {
	if run.Preparation == nil {
		return WorkflowPublicationGuard{}, fmt.Errorf("publication requires prepared combined-result validation")
	}

	review, err := s.checkWorkflowReviewGrant(ctx, run, provider)
	if err != nil {
		return WorkflowPublicationGuard{}, err
	}

	source := workflowPreparationSource(review)

	candidate, err := s.currentWorkflowPreparation(ctx, run, source)
	if err != nil {
		return WorkflowPublicationGuard{}, err
	}

	capability, err := observeWorkflowPublicationCapability(ctx, provider, source)
	if err != nil {
		return WorkflowPublicationGuard{}, err
	}

	if capability != candidate.Input.Capability || review.PreparationRef != run.Preparation.ResultRef || review.PreparationHash != run.Preparation.ResultHash {
		return WorkflowPublicationGuard{}, ErrWorkflowConflict
	}

	return WorkflowPublicationGuard{PullRequest: review.PullRequest, Capability: capability, Git: candidate.Input.Git, PreparationRef: run.Preparation.ResultRef, PreparationHash: run.Preparation.ResultHash, ReviewHash: run.Review.InputHash, Policy: review.Policy}, nil
}

func workflowPreparationSource(input WorkflowReviewInput) WorkflowReviewInput {
	input.PreparationRef, input.PreparationHash, input.PreparedGit = "", "", nil

	return input
}

func (s *WorkflowStore) currentWorkflowPreparation(ctx context.Context, run workflowInvocationFile, source WorkflowReviewInput) (WorkflowIntegrationCandidate, error) {
	if run.Preparation == nil || run.Preparation.State != "validated" {
		return WorkflowIntegrationCandidate{}, fmt.Errorf("combined-result validation is absent or stale")
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return candidate, err
	}

	hash, err := workflowReviewInputHash(source)
	if err != nil || hash != run.Preparation.SourceHash {
		return candidate, fmt.Errorf("%w: prepared source, base or current policy changed", ErrWorkflowConflict)
	}

	if err := s.checkWorkflowPreparationGit(ctx, candidate); err != nil {
		return candidate, err
	}

	return candidate, nil
}

func (s *WorkflowStore) checkWorkflowPreparationGit(ctx context.Context, candidate WorkflowIntegrationCandidate) error {
	commit, tree, err := s.workflowLaneCandidate(ctx, candidate.Lane)

	proof := candidate.Input.Git
	if err != nil || commit != proof.ResultCommit || tree != proof.PreparedTree {
		return fmt.Errorf("%w: prepared lane changed", ErrWorkflowConflict)
	}

	reader := workflowIntegrationReader{dir: candidate.Lane.Worktree}

	observed := WorkflowIntegrationProof{WorkflowIntegrationQuery: proof.WorkflowIntegrationQuery, ObservedTarget: proof.ObservedTarget}
	if err := reader.verifyMethod(ctx, &observed); err != nil {
		return err
	}

	observed.HistoryHash, err = workflowIntegrationHistoryHash(observed)
	if err != nil || observed.BaseCommit != proof.BaseCommit || observed.HistoryHash != proof.HistoryHash || observed.VerifiedTree != proof.PreparedTree {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowPreparationBinding(run workflowInvocationFile) error {
	if run.PreparationCount < 0 || run.PreparationCount > WorkflowPageLimit {
		return ErrWorkflowConflict
	}

	if run.Preparation == nil {
		if run.PreparationCount != 0 {
			return ErrWorkflowConflict
		}

		return nil
	}

	if !validWorkflowPreparationHeader(run) {
		return ErrWorkflowConflict
	}

	return validateWorkflowPreparationRefs(run.ID, *run.Preparation)
}

func validWorkflowPreparationHeader(run workflowInvocationFile) bool {
	b := run.Preparation

	return run.Lane != nil && b.Unit == run.Lane.Unit && b.Round >= 1 && b.Round == run.PreparationCount && len(b.SourceHash) == 64 && slices.Contains([]string{"preparing", "validating", "validated", "stale"}, b.State)
}

func validateWorkflowPreparationRefs(id string, b WorkflowPreparationBinding) error {
	if b.ResultRef == "" {
		if b.State == "validated" || b.ResultHash != "" {
			return ErrWorkflowConflict
		}

		return nil
	}

	if b.ResultRef != workflowPreparationPath(id, b.Round) || len(b.ResultHash) != 64 {
		return ErrWorkflowConflict
	}

	return nil
}
