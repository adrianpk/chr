package core

import (
	"context"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowNativePreparationProvider observes trusted policy and transport for
// the exact host-prepared P. The preliminary guard cannot authorize publication.
type WorkflowNativePreparationProvider interface {
	WorkflowPublicationGuardProvider
	WorkflowPreparedNativeCapability(context.Context, WorkflowPullRequestSnapshot, WorkflowIntegrationProof) (WorkflowNativeCapability, error)
}

// WorkflowNativePreparationInput freezes the queue owner and full capability
// before the host gate executes. Agent output cannot replace this input.
type WorkflowNativePreparationInput struct {
	Entry           WorkflowIntegrationBinding `json:"entry"`
	ClaimGeneration uint64                     `json:"claim_generation"`
	Capability      WorkflowNativeCapability   `json:"capability"`
}

func (s *WorkflowStore) checkWorkflowNativePreparationAdmission(run workflowInvocationFile, source WorkflowReviewInput, capability WorkflowPublicationCapability, provider WorkflowPublicationGuardProvider) error {
	native, err := workflowNativePreparationProfile(s.layout.ProjectDir(), capability)
	if err != nil || !native {
		return err
	}

	if _, ok := provider.(WorkflowNativePreparationProvider); !ok || run.Integration == nil {
		return ErrWorkflowAtomicGuard
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if !workflowNativePreparationOwnerMatches(entry, source) {
		return workflowstate.ErrConflict
	}

	return workflowNativePreparationEntryReady(run, entry)
}

func workflowNativePreparationProfile(project string, capability WorkflowPublicationCapability) (bool, error) {
	config, err := LoadWorkspaceConfig(project)
	if err != nil {
		return false, err
	}

	native := capability.Mechanism == WorkflowPublicationGitAtomicDualRef
	if native != (config.Work.Integration.PublicationProfile == WorkflowPublicationGitAtomicDualRef) {
		return native, ErrWorkflowAtomicGuard
	}

	return native, nil
}

func workflowNativePreparationOwnerMatches(entry WorkflowIntegrationEntry, source WorkflowReviewInput) bool {
	return entry.Unit == source.Unit && entry.Repository == source.PullRequest.HeadRepository && entry.TargetRef == "refs/heads/"+source.PullRequest.BaseBranch && entry.StopIntent == "" && entry.ClaimGeneration > 0
}

func workflowNativePreparationEntryReady(run workflowInvocationFile, entry WorkflowIntegrationEntry) error {
	if run.Preparation != nil && run.Preparation.State == "validated" {
		if !slices.Contains([]string{"reviewing", "waiting_review", "ready_to_publish"}, entry.State) || !workflowNativePreparedEntryMatches(run, entry) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if entry.State != "validating" {
		return ErrWorkflowOwned
	}

	return nil
}

func (s *WorkflowStore) bindWorkflowNativePreparation(ctx context.Context, run workflowInvocationFile, input *WorkflowPreparationInput, provider WorkflowPublicationGuardProvider) error {
	if input.Capability.Mechanism != WorkflowPublicationGitAtomicDualRef {
		return nil
	}

	observer, ok := provider.(WorkflowNativePreparationProvider)
	if !ok || run.Integration == nil {
		return ErrWorkflowAtomicGuard
	}

	capability, err := observeWorkflowPreparedNativeCapability(ctx, observer, *input)
	if err != nil {
		return err
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	input.Capability = capability.WorkflowPublicationCapability
	input.Native = &WorkflowNativePreparationInput{Entry: *run.Integration, ClaimGeneration: entry.ClaimGeneration, Capability: capability}

	return s.checkWorkflowNativePreparationFence(run, input.Native)
}

func validateWorkflowPreparedNativeCapability(input WorkflowPreparationInput, capability WorkflowNativeCapability) error {
	if err := validateWorkflowNativeCapability(capability); err != nil {
		return err
	}

	if !validWorkflowNativeCapabilityRecord(capability) || !validWorkflowNativeFingerprint(capability.PolicyFingerprint) {
		return workflowstate.ErrConflict
	}

	pr, proof := input.Source.PullRequest, input.Git
	expected := [5]string{"refs/heads/" + pr.HeadBranch, "refs/heads/" + pr.BaseBranch, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit}
	actual := [5]string{capability.SourceRef, capability.TargetRef, capability.OriginalSourceCommit, capability.ExpectedBaseCommit, capability.PreparedResultCommit}

	if expected != actual || proof.ResultCommit == proof.SourceCommit || proof.ResultCommit == proof.BaseCommit {
		return workflowstate.ErrConflict
	}

	return validateWorkflowPublicationCapability(input.Source, capability.WorkflowPublicationCapability)
}

func observeWorkflowPreparedCapability(ctx context.Context, provider WorkflowPublicationGuardProvider, input WorkflowPreparationInput) (WorkflowPublicationCapability, error) {
	if input.Native == nil {
		return observeWorkflowPublicationCapability(ctx, provider, input.Source)
	}

	observer, ok := provider.(WorkflowNativePreparationProvider)
	if !ok {
		return WorkflowPublicationCapability{}, ErrWorkflowAtomicGuard
	}

	capability, err := observeWorkflowPreparedNativeCapability(ctx, observer, input)
	if err != nil {
		return capability.WorkflowPublicationCapability, err
	}

	return capability.WorkflowPublicationCapability, nil
}

func observeWorkflowPreparedNativeCapability(ctx context.Context, provider WorkflowNativePreparationProvider, input WorkflowPreparationInput) (WorkflowNativeCapability, error) {
	head, err := provider.ReadWorkflowBranch(ctx, input.Source.PullRequest.Repository, input.Source.PullRequest.HeadBranch)
	if err != nil {
		return WorkflowNativeCapability{}, err
	}

	if head != input.Git.SourceCommit {
		return WorkflowNativeCapability{}, workflowstate.ErrConflict
	}

	capability, err := provider.WorkflowPreparedNativeCapability(ctx, input.Source.PullRequest, input.Git)
	if err != nil {
		return capability, err
	}

	return capability, validateWorkflowPreparedNativeCapability(input, capability)
}

func workflowNativePreparedEntryMatches(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return entry.CandidateRef != nil && entry.CandidateRef.Path == run.Preparation.ResultRef && entry.CandidateRef.Hash == run.Preparation.ResultHash
}

func (s *WorkflowStore) cachedWorkflowPreparationCapability(ctx context.Context, run workflowInvocationFile, candidate WorkflowIntegrationCandidate, preliminary WorkflowPublicationCapability, provider WorkflowPublicationGuardProvider) (WorkflowPublicationCapability, error) {
	native := candidate.Input.Native
	if native == nil {
		return preliminary, nil
	}

	if run.Integration == nil || run.Integration.EntryID != native.Entry.EntryID {
		return preliminary, workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return preliminary, err
	}

	if entry.ClaimGeneration != native.ClaimGeneration || !workflowNativePreparedEntryMatches(run, entry) {
		return preliminary, workflowstate.ErrConflict
	}

	return observeWorkflowPreparedCapability(ctx, provider, candidate.Input)
}

func (s *WorkflowStore) checkWorkflowNativePreparationFence(run workflowInvocationFile, native *WorkflowNativePreparationInput) error {
	if native == nil {
		return nil
	}

	if run.Integration == nil || *run.Integration != native.Entry {
		return workflowstate.ErrConflict
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	if entry.State != "validating" || entry.ClaimGeneration != native.ClaimGeneration || entry.StopIntent != "" {
		return workflowstate.ErrConflict
	}

	return nil
}
