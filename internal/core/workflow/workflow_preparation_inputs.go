package workflow

import (
	"context"
	"fmt"
	"slices"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func observeWorkflowPublicationCapability(ctx context.Context, provider PublicationGuardProvider, source workflowstate.ReviewInput) (workflowstate.PublicationCapability, error) {
	if provider == nil {
		return workflowstate.PublicationCapability{}, ErrWorkflowAtomicGuard
	}

	head, err := provider.ReadWorkflowBranch(ctx, source.PullRequest.Repository, source.PullRequest.HeadBranch)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	if head != source.PullRequest.HeadCommit {
		return workflowstate.PublicationCapability{}, workflowstate.ErrConflict
	}

	capability, err := provider.WorkflowIntegrationCapability(ctx, source.PullRequest, source.Policy.Method)
	if err != nil {
		return capability, err
	}

	return capability, validateWorkflowPublicationCapability(source, capability)
}

func validateWorkflowPublicationCapability(source workflowstate.ReviewInput, capability workflowstate.PublicationCapability) error {
	if capability.Repository != source.PullRequest.Repository || capability.PullRequestURL != source.PullRequest.URL || capability.Method != source.Policy.Method || capability.Mechanism == "" || len(capability.Mechanism) > 128 || len(capability.Revision) != 64 {
		return workflowstate.ErrConflict
	}

	if !capability.AtomicSourceBase || !capability.ExactResult {
		return ErrWorkflowAtomicGuard
	}

	return validateWorkflowSignatureCapability(capability)
}

func validateWorkflowSignatureCapability(capability workflowstate.PublicationCapability) error {
	if capability.RequiredSignatures != (capability.SigningProfile != "") || (capability.RequiredSignatures && !validWorkflowNativeFingerprint(capability.SigningProfile)) {
		return ErrWorkflowAtomicGuard
	}

	return nil
}

func (s *Store) checkWorkflowPublicationGuard(ctx context.Context, run workflowInvocationFile, provider PublicationGuardProvider) (workflowstate.PublicationGuard, error) {
	if run.Preparation == nil {
		return workflowstate.PublicationGuard{}, fmt.Errorf("publication requires prepared combined-result validation")
	}

	review, err := s.checkWorkflowReviewGrant(ctx, run, provider)
	if err != nil {
		return workflowstate.PublicationGuard{}, err
	}

	source := workflowPreparationSource(review)

	candidate, err := s.currentWorkflowPreparation(ctx, run, source)
	if err != nil {
		return workflowstate.PublicationGuard{}, err
	}

	capability, err := observeWorkflowPreparedCapability(ctx, provider, candidate.Input)
	if err != nil {
		return workflowstate.PublicationGuard{}, err
	}

	if capability != candidate.Input.Capability || review.PreparationRef != run.Preparation.ResultRef || review.PreparationHash != run.Preparation.ResultHash {
		return workflowstate.PublicationGuard{}, workflowstate.ErrConflict
	}

	return workflowstate.PublicationGuard{PullRequest: review.PullRequest, Capability: capability, Git: candidate.Input.Git, PreparationRef: run.Preparation.ResultRef, PreparationHash: run.Preparation.ResultHash, ReviewHash: run.Review.InputHash, Policy: review.Policy}, nil
}

func workflowPreparationSource(input workflowstate.ReviewInput) workflowstate.ReviewInput {
	input.PreparationRef, input.PreparationHash, input.PreparedGit = "", "", nil
	input.Native = nil

	return input
}

func (s *Store) currentWorkflowPreparation(ctx context.Context, run workflowInvocationFile, source workflowstate.ReviewInput) (workflowstate.IntegrationCandidate, error) {
	if run.Preparation == nil || run.Preparation.State != "validated" {
		return workflowstate.IntegrationCandidate{}, fmt.Errorf("combined-result validation is absent or stale")
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

func (s *Store) checkWorkflowPreparationGit(ctx context.Context, candidate workflowstate.IntegrationCandidate) error {
	laneImage, err := s.workflowLaneCandidate(ctx, candidate.Lane)

	proof := candidate.Input.Git

	capability := candidate.Input.Capability
	if err := checkWorkflowPreparationSigning(proof, capability.RequiredSignatures, capability.SigningProfile); err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("%w: prepared lane changed: %w", workflowstate.ErrConflict, err)
	}

	if laneImage.Commit != proof.ResultCommit || laneImage.Tree != proof.PreparedTree {
		return fmt.Errorf("%w: prepared lane commit or tree changed", workflowstate.ErrConflict)
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: candidate.Lane.Worktree}}

	observed := workflowstate.IntegrationProof{IntegrationQuery: proof.IntegrationQuery, ObservedTarget: proof.ObservedTarget, Provenance: proof.Provenance}
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

func validateWorkflowPreparationRefs(id string, b PreparationBinding) error {
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
