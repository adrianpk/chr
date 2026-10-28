package workflow

import (
	"context"
	"slices"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// NativePreparationProvider observes trusted policy and transport for
// the exact host-prepared P. The preliminary guard cannot authorize publication.
type NativePreparationProvider interface {
	PublicationGuardProvider
	WorkflowPreparedNativeCapability(context.Context, workflowstate.PullRequestSnapshot, workflowstate.IntegrationProof) (workflowstate.NativeCapability, error)
}

func (s *Store) checkWorkflowNativePreparationAdmission(run workflowInvocationFile, source workflowstate.ReviewInput, capability workflowstate.PublicationCapability, provider PublicationGuardProvider) error {
	native, err := workflowNativePreparationProfile(s.layout.ProjectDir(), capability)
	if err != nil || !native {
		return err
	}

	if _, ok := provider.(NativePreparationProvider); !ok || run.Integration == nil {
		return workflowstate.ErrAtomicGuard
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

func workflowNativePreparationProfile(project string, capability workflowstate.PublicationCapability) (bool, error) {
	config, err := workspaceconfig.Load(project)
	if err != nil {
		return false, err
	}

	native := capability.Mechanism == workspaceconfig.WorkflowPublicationGitAtomicDualRef
	if native != (config.Work.Integration.PublicationProfile == workspaceconfig.WorkflowPublicationGitAtomicDualRef) {
		return native, workflowstate.ErrAtomicGuard
	}

	return native, nil
}

func workflowNativePreparationOwnerMatches(entry IntegrationEntry, source workflowstate.ReviewInput) bool {
	return entry.Unit == source.Unit && entry.Repository == source.PullRequest.HeadRepository && entry.TargetRef == "refs/heads/"+source.PullRequest.BaseBranch && entry.StopIntent == "" && entry.ClaimGeneration > 0
}

func workflowNativePreparationEntryReady(run workflowInvocationFile, entry IntegrationEntry) error {
	if run.Preparation != nil && run.Preparation.State == "validated" {
		if !slices.Contains([]IntegrationState{"reviewing", "waiting_review", "ready_to_publish"}, entry.State) || !workflowNativePreparedEntryMatches(run, entry) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if entry.State != "validating" {
		return workflowstate.ErrOwned
	}

	return nil
}

func (s *Store) bindWorkflowNativePreparation(ctx context.Context, run workflowInvocationFile, input *workflowstate.PreparationInput, provider PublicationGuardProvider) error {
	if input.Capability.Mechanism != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return nil
	}

	observer, ok := provider.(NativePreparationProvider)
	if !ok || run.Integration == nil {
		return workflowstate.ErrAtomicGuard
	}

	capability, err := observeWorkflowPreparedNativeCapability(ctx, observer, *input)
	if err != nil {
		return err
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return err
	}

	input.Capability = capability.PublicationCapability
	input.Native = &workflowstate.NativePreparationInput{Entry: *run.Integration, ClaimGeneration: entry.ClaimGeneration, Capability: capability}

	return s.checkWorkflowNativePreparationFence(run, input.Native)
}

func validateWorkflowPreparedNativeCapability(input workflowstate.PreparationInput, capability workflowstate.NativeCapability) error {
	if err := workflowforge.CheckNativeCapability(capability); err != nil {
		return err
	}

	if !validWorkflowNativeCapabilityRecord(capability) || !workflowstate.ValidFingerprint(capability.PolicyFingerprint) {
		return workflowstate.ErrConflict
	}

	pr, proof := input.Source.PullRequest, input.Git
	expected := [5]string{"refs/heads/" + pr.HeadBranch, "refs/heads/" + pr.BaseBranch, proof.SourceCommit, proof.BaseCommit, proof.ResultCommit}
	actual := [5]string{capability.SourceRef, capability.TargetRef, capability.OriginalSourceCommit, capability.ExpectedBaseCommit, capability.PreparedResultCommit}

	if expected != actual || proof.ResultCommit == proof.SourceCommit || proof.ResultCommit == proof.BaseCommit {
		return workflowstate.ErrConflict
	}

	return workflowintegrate.CheckCapability(input.Source, capability.PublicationCapability)
}

func observeWorkflowPreparedCapability(ctx context.Context, provider PublicationGuardProvider, input workflowstate.PreparationInput) (workflowstate.PublicationCapability, error) {
	if input.Native == nil {
		return observeWorkflowPublicationCapability(ctx, provider, input.Source)
	}

	observer, ok := provider.(NativePreparationProvider)
	if !ok {
		return workflowstate.PublicationCapability{}, workflowstate.ErrAtomicGuard
	}

	capability, err := observeWorkflowPreparedNativeCapability(ctx, observer, input)
	if err != nil {
		return capability.PublicationCapability, err
	}

	return capability.PublicationCapability, nil
}

func observeWorkflowPreparedNativeCapability(ctx context.Context, provider NativePreparationProvider, input workflowstate.PreparationInput) (workflowstate.NativeCapability, error) {
	head, err := provider.ReadWorkflowBranch(ctx, input.Source.PullRequest.Repository, input.Source.PullRequest.HeadBranch)
	if err != nil {
		return workflowstate.NativeCapability{}, err
	}

	if head != input.Git.SourceCommit {
		return workflowstate.NativeCapability{}, workflowstate.ErrConflict
	}

	capability, err := provider.WorkflowPreparedNativeCapability(ctx, input.Source.PullRequest, input.Git)
	if err != nil {
		return capability, err
	}

	return capability, validateWorkflowPreparedNativeCapability(input, capability)
}

func workflowNativePreparedEntryMatches(run workflowInvocationFile, entry IntegrationEntry) bool {
	return entry.CandidateRef != nil && entry.CandidateRef.Path == run.Preparation.ResultRef && entry.CandidateRef.Hash == run.Preparation.ResultHash
}

func (s *Store) cachedWorkflowPreparationCapability(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, preliminary workflowstate.PublicationCapability, provider PublicationGuardProvider) (workflowstate.PublicationCapability, error) {
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

func (s *Store) checkWorkflowNativePreparationFence(run workflowInvocationFile, native *workflowstate.NativePreparationInput) error {
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
