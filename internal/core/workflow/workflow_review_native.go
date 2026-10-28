package workflow

import (
	"context"
	"slices"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s *Store) checkWorkflowSourceReviewProfile() error {
	config, err := workspaceconfig.Load(s.layout.ProjectDir())
	if err != nil {
		return err
	}

	if config.Work.Integration.PublicationProfile == workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return workflowstate.ErrAtomicGuard
	}

	return nil
}

func (s *Store) bindWorkflowNativeReview(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, provider PullRequestProvider, input *workflowstate.ReviewInput) error {
	if err := s.bindWorkflowNativeReviewCapability(ctx, candidate, provider, input); err != nil {
		return err
	}

	return s.checkWorkflowNativeReviewFence(run, input)
}

func (s *Store) bindWorkflowNativeReviewCapability(ctx context.Context, candidate workflowstate.IntegrationCandidate, provider PullRequestProvider, input *workflowstate.ReviewInput) error {
	native, err := workflowNativePreparationProfile(s.layout.ProjectDir(), candidate.Input.Capability)
	if err != nil || !native {
		return err
	}

	observer, ok := provider.(NativePreparationProvider)
	if !ok || candidate.Native == nil || candidate.Input.Native == nil {
		return workflowstate.ErrAtomicGuard
	}

	capability, err := observeWorkflowPreparedNativeCapability(ctx, observer, candidate.Input)
	if err != nil {
		return err
	}

	if !workflowNativeJSONEqual(capability, candidate.Input.Native.Capability) {
		return workflowstate.ErrConflict
	}

	input.Native = &workflowstate.NativeReviewInput{Candidate: *candidate.Native, ClaimGeneration: candidate.Input.Native.ClaimGeneration, Capability: capability}
	if err := workflowreview.CheckNativeInput(*input, workflowNativeReviewProofs{}); err != nil {
		return err
	}

	if _, err := workflowNativePreparationProfile(s.layout.ProjectDir(), candidate.Input.Capability); err != nil {
		return err
	}

	return nil
}

func (s *Store) checkWorkflowNativeReviewFence(run workflowInvocationFile, input *workflowstate.ReviewInput) error {
	if input.Native == nil {
		return nil
	}

	if run.Integration == nil || run.Integration.EntryID != input.Native.Candidate.EntryID || run.Preparation == nil {
		return workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if !workflowNativeReviewEntryMatches(run, entry, input) {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowNativeReviewEntryMatches(run workflowInvocationFile, entry IntegrationEntry, input *workflowstate.ReviewInput) bool {
	return slices.Contains([]IntegrationState{"reviewing", "waiting_review", "ready_to_publish"}, entry.State) && entry.StopIntent == "" && entry.ClaimGeneration == input.Native.ClaimGeneration && workflowNativePreparedEntryMatches(run, entry) && entry.CandidateRef.Path == input.PreparationRef && entry.CandidateRef.Hash == input.PreparationHash
}
