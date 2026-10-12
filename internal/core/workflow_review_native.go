package core

import (
	"context"
	"slices"
)

// WorkflowNativeReviewInput binds source rewrite, history and publication
// policy to one host-validated candidate and durable queue generation.
type WorkflowNativeReviewInput struct {
	Candidate       WorkflowNativeCandidate  `json:"candidate"`
	ClaimGeneration uint64                   `json:"claim_generation"`
	Capability      WorkflowNativeCapability `json:"capability"`
}

func (s *WorkflowStore) checkWorkflowSourceReviewProfile() error {
	config, err := LoadWorkspaceConfig(s.layout.ProjectDir())
	if err != nil {
		return err
	}

	if config.Work.Integration.PublicationProfile == WorkflowPublicationGitAtomicDualRef {
		return ErrWorkflowAtomicGuard
	}

	return nil
}

func (s *WorkflowStore) bindWorkflowNativeReview(ctx context.Context, run workflowInvocationFile, candidate WorkflowIntegrationCandidate, provider WorkflowPullRequestProvider, input *WorkflowReviewInput) error {
	native, err := workflowNativePreparationProfile(s.layout.ProjectDir(), candidate.Input.Capability)
	if err != nil || !native {
		return err
	}

	observer, ok := provider.(WorkflowNativePreparationProvider)
	if !ok || candidate.Native == nil || candidate.Input.Native == nil {
		return ErrWorkflowAtomicGuard
	}

	capability, err := observeWorkflowPreparedNativeCapability(ctx, observer, candidate.Input)
	if err != nil {
		return err
	}

	if !workflowNativeJSONEqual(capability, candidate.Input.Native.Capability) {
		return ErrWorkflowConflict
	}

	input.Native = &WorkflowNativeReviewInput{Candidate: *candidate.Native, ClaimGeneration: candidate.Input.Native.ClaimGeneration, Capability: capability}
	if err := validateWorkflowNativeReviewInput(*input); err != nil {
		return err
	}

	if _, err := workflowNativePreparationProfile(s.layout.ProjectDir(), candidate.Input.Capability); err != nil {
		return err
	}

	return s.checkWorkflowNativeReviewFence(run, input)
}

func (s *WorkflowStore) checkWorkflowNativeReviewFence(run workflowInvocationFile, input *WorkflowReviewInput) error {
	if input.Native == nil {
		return nil
	}

	if run.Integration == nil || run.Integration.EntryID != input.Native.Candidate.EntryID || run.Preparation == nil {
		return ErrWorkflowConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if !workflowNativeReviewEntryMatches(run, entry, input) {
		return ErrWorkflowConflict
	}

	return nil
}

func workflowNativeReviewEntryMatches(run workflowInvocationFile, entry WorkflowIntegrationEntry, input *WorkflowReviewInput) bool {
	return slices.Contains([]string{"reviewing", "waiting_review", "ready_to_publish"}, entry.State) && entry.StopIntent == "" && entry.ClaimGeneration == input.Native.ClaimGeneration && workflowNativePreparedEntryMatches(run, entry) && entry.CandidateRef.Path == input.PreparationRef && entry.CandidateRef.Hash == input.PreparationHash
}

func validateWorkflowNativeReviewInput(input WorkflowReviewInput) error {
	if input.Native == nil {
		return nil
	}

	native := input.Native

	binding := native.Candidate.WorkflowNativeRecordBinding
	if native.ClaimGeneration == 0 || !validWorkflowNativeFingerprint(native.Candidate.CandidateHash) || input.PreparedGit == nil {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowNativeRecordBinding(binding); err != nil {
		return err
	}

	if !workflowNativeReviewOwnerMatches(input) {
		return ErrWorkflowConflict
	}

	if binding.CapabilityHash != native.Capability.Revision || binding.PolicyHash != native.Capability.PolicyFingerprint {
		return ErrWorkflowConflict
	}

	preparation := WorkflowPreparationInput{Source: workflowPreparationSource(input), Git: *input.PreparedGit}

	return validateWorkflowPreparedNativeCapability(preparation, native.Capability)
}

func workflowNativeReviewOwnerMatches(input WorkflowReviewInput) bool {
	binding := input.Native.Candidate.WorkflowNativeRecordBinding

	return binding.InvocationID == input.InvocationID && binding.Unit == input.Unit && binding.CanonicalPR == input.PullRequest.URL && binding.ContractHash == input.Slice.ContractHash && workflowNativeRecordGitMatches(binding, *input.PreparedGit)
}
