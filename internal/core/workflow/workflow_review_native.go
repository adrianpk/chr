package workflow

import (
	"context"
	"slices"

	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
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
	if err := s.bindWorkflowNativeReviewCapability(ctx, run, candidate, provider, input); err != nil {
		return err
	}

	return s.checkWorkflowNativeReviewFence(run, input)
}

func (s *Store) bindWorkflowNativeReviewCapability(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, provider PullRequestProvider, input *workflowstate.ReviewInput) error {
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
	if err := validateWorkflowNativeReviewInput(*input); err != nil {
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

func validateWorkflowNativeReviewInput(input workflowstate.ReviewInput) error {
	if input.Native == nil {
		return nil
	}

	native := input.Native

	binding := native.Candidate.NativeRecordBinding
	if native.ClaimGeneration == 0 || !workflowstate.ValidFingerprint(native.Candidate.CandidateHash) || input.PreparedGit == nil {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeRecordBinding(binding); err != nil {
		return err
	}

	if !workflowNativeReviewOwnerMatches(input) {
		return workflowstate.ErrConflict
	}

	if binding.CapabilityHash != native.Capability.Revision || binding.PolicyHash != native.Capability.PolicyFingerprint {
		return workflowstate.ErrConflict
	}

	preparation := workflowstate.PreparationInput{Source: workflowintegrate.SourceInput(input), Git: *input.PreparedGit}

	return validateWorkflowPreparedNativeCapability(preparation, native.Capability)
}

func workflowNativeReviewOwnerMatches(input workflowstate.ReviewInput) bool {
	binding := input.Native.Candidate.NativeRecordBinding

	return binding.InvocationID == input.InvocationID && binding.Unit == input.Unit && binding.CanonicalPR == input.PullRequest.URL && binding.ContractHash == input.Slice.ContractHash && workflowNativeRecordGitMatches(binding, *input.PreparedGit)
}
