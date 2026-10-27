package core

import (
	"context"
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
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
		return WorkflowPublicationCapability{}, workflowstate.ErrConflict
	}

	capability, err := provider.WorkflowIntegrationCapability(ctx, source.PullRequest, source.Policy.Method)
	if err != nil {
		return capability, err
	}

	return capability, validateWorkflowPublicationCapability(source, capability)
}

func validateWorkflowPublicationCapability(source WorkflowReviewInput, capability WorkflowPublicationCapability) error {
	if capability.Repository != source.PullRequest.Repository || capability.PullRequestURL != source.PullRequest.URL || capability.Method != source.Policy.Method || capability.Mechanism == "" || len(capability.Mechanism) > 128 || len(capability.Revision) != 64 {
		return workflowstate.ErrConflict
	}

	if !capability.AtomicSourceBase || !capability.ExactResult {
		return ErrWorkflowAtomicGuard
	}

	return validateWorkflowSignatureCapability(capability)
}

func validateWorkflowSignatureCapability(capability WorkflowPublicationCapability) error {
	if capability.RequiredSignatures != (capability.SigningProfile != "") || (capability.RequiredSignatures && !validWorkflowNativeFingerprint(capability.SigningProfile)) {
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

	capability, err := observeWorkflowPreparedCapability(ctx, provider, candidate.Input)
	if err != nil {
		return WorkflowPublicationGuard{}, err
	}

	if capability != candidate.Input.Capability || review.PreparationRef != run.Preparation.ResultRef || review.PreparationHash != run.Preparation.ResultHash {
		return WorkflowPublicationGuard{}, workflowstate.ErrConflict
	}

	return WorkflowPublicationGuard{PullRequest: review.PullRequest, Capability: capability, Git: candidate.Input.Git, PreparationRef: run.Preparation.ResultRef, PreparationHash: run.Preparation.ResultHash, ReviewHash: run.Review.InputHash, Policy: review.Policy}, nil
}

func workflowPreparationSource(input WorkflowReviewInput) WorkflowReviewInput {
	input.PreparationRef, input.PreparationHash, input.PreparedGit = "", "", nil
	input.Native = nil

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
		return candidate, fmt.Errorf("%w: prepared source, base or current policy changed", workflowstate.ErrConflict)
	}

	if err := s.checkWorkflowPreparationGit(ctx, candidate); err != nil {
		return candidate, err
	}

	return candidate, nil
}

func (s *WorkflowStore) checkWorkflowPreparationGit(ctx context.Context, candidate WorkflowIntegrationCandidate) error {
	commit, tree, err := s.workflowLaneCandidate(ctx, candidate.Lane)

	proof := candidate.Input.Git

	capability := candidate.Input.Capability
	if err := checkWorkflowPreparationSigning(proof, capability.RequiredSignatures, capability.SigningProfile); err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("%w: prepared lane changed: %w", workflowstate.ErrConflict, err)
	}

	if commit != proof.ResultCommit || tree != proof.PreparedTree {
		return fmt.Errorf("%w: prepared lane commit or tree changed", workflowstate.ErrConflict)
	}

	reader := workflowIntegrationReader{dir: candidate.Lane.Worktree}

	observed := WorkflowIntegrationProof{WorkflowIntegrationQuery: proof.WorkflowIntegrationQuery, ObservedTarget: proof.ObservedTarget, Provenance: proof.Provenance}
	if err := reader.verifyMethod(ctx, &observed); err != nil {
		return err
	}

	observed.HistoryHash, err = workflowIntegrationHistoryHash(observed)
	if err != nil || observed.BaseCommit != proof.BaseCommit || observed.HistoryHash != proof.HistoryHash || observed.VerifiedTree != proof.PreparedTree {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowPreparationBinding(run workflowInvocationFile) error {
	if err := validateWorkflowCandidateRepairBinding(run); err != nil {
		return err
	}

	if run.PreparationCount < 0 || run.PreparationCount > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	if run.Preparation == nil {
		retained, _ := workflowRetainedRoleCounts(run)
		if run.PreparationCount != retained {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if !validWorkflowPreparationHeader(run) {
		return workflowstate.ErrConflict
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
			return workflowstate.ErrConflict
		}

		return nil
	}

	if b.ResultRef != workflowPreparationPath(id, b.Round) || len(b.ResultHash) != 64 {
		return workflowstate.ErrConflict
	}

	return nil
}
