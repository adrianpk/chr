package workflow

import (
	"context"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// NativeReviewAdapter connects existing remote observation to the
// invocation's exact source ownership and current host publisher profile.
type NativeReviewAdapter struct {
	RemoteReviewAdapter
	store     *Store
	id        string
	configDir string
}

// NewWorkflowNativeReviewAdapter stores dependencies without opening resources.
func NewWorkflowNativeReviewAdapter(remote RemoteReviewAdapter, store *Store, id, configDir string) *NativeReviewAdapter {
	return &NativeReviewAdapter{remote, store, id, configDir}
}

// WorkflowIntegrationCapability composes current host grants with configured provider observations.
func (adapter *NativeReviewAdapter) WorkflowIntegrationCapability(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowstate.PublicationCapability, error) {
	return workflowforge.ObserveNativePublicationCapability(ctx, pr, method, workflowforge.NativeCapabilityReaders{Provider: adapter.RemoteReviewAdapter, Policy: adapter})
}

// WorkflowPreparedNativeCapability observes the actual receive-pack server for
// the host's exact S/B/P proof, using the same configured publisher as dispatch.
func (adapter *NativeReviewAdapter) WorkflowPreparedNativeCapability(ctx context.Context, pr workflowstate.PullRequestSnapshot, proof workflowstate.IntegrationProof) (workflowstate.NativeCapability, error) {
	policy, err := adapter.ReadWorkflowNativePolicy(ctx, pr, proof.Method)
	if err != nil {
		return workflowstate.NativeCapability{}, err
	}

	nativePublisher, err := adapter.nativePublisher(ctx)
	run, owner, profile := nativePublisher.Run, nativePublisher.Owner, nativePublisher.Profile

	if err != nil {
		return workflowstate.NativeCapability{}, err
	}

	request := GitPublication{Repository: pr.Repository, SourceRepository: pr.HeadRepository, Directory: owner.layout.ProjectDir(), SourceRef: policy.SourceRef, TargetRef: policy.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: policy.OwnershipRevision, Git: proof}

	request.CandidateRevision, err = workflowforge.PublicationRevision(workflowPublicationIdentity(request))
	if err != nil {
		return workflowstate.NativeCapability{}, err
	}

	config, err := workspaceconfig.Load(owner.layout.ProjectDir())
	if err != nil {
		return workflowstate.NativeCapability{}, err
	}

	if proof.SourceCommit != run.PullRequest.SourceCommit {
		return workflowstate.NativeCapability{}, workflowstate.ErrConflict
	}

	query := NativeCapabilityRequest{Config: config.Work.Integration, PullRequest: pr, Publication: request}

	return adapter.WorkflowNativeIntegrationCapability(ctx, query, adapter, NewWorkflowGitPublisher(profile.Repository, profile.Transport))
}

// PublishWorkflowGit rechecks the exact local ownership/profile binding before
// entering the existing atomic transport. It never calls a provider merge API.
func (adapter *NativeReviewAdapter) PublishWorkflowGit(ctx context.Context, request GitPublication) (workflowstate.GitPublicationResult, error) {
	nativePublisher2, err := adapter.nativePublisher(ctx)
	run, owner, profile := nativePublisher2.Run, nativePublisher2.Owner, nativePublisher2.Profile

	if err != nil {
		return workflowstate.GitPublicationResult{}, err
	}

	ownership, err := workflowNativeOwnershipRevision(run, profile)
	if err != nil || request.OwnershipRevision != ownership || request.SourceRef != run.Lane.Branch || request.TargetRef != profile.TargetRef || request.Git.SourceCommit != run.PullRequest.SourceCommit {
		return workflowstate.GitPublicationResult{}, workflowstate.ErrConflict
	}

	config, err := workspaceconfig.Load(owner.layout.ProjectDir())
	if err != nil || config.Work.Integration.PublicationProfile != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return workflowstate.GitPublicationResult{}, workflowstate.ErrAtomicGuard
	}

	return NewWorkflowGitPublisher(profile.Repository, profile.Transport).PublishWorkflowGit(ctx, request)
}

// InspectWorkflowGitPublication keeps the configured transport available for
// read-only reconciliation, including after publication authorization stops.
func (adapter *NativeReviewAdapter) InspectWorkflowGitPublication(ctx context.Context, request GitPublication) (workflowstate.GitPublicationAdvertisement, error) {
	profile, err := LoadWorkflowNativePublisher(adapter.configDir, request.Repository, request.TargetRef)
	if err != nil {
		return workflowstate.GitPublicationAdvertisement{}, err
	}

	return NewWorkflowGitPublisher(profile.Repository, profile.Transport).InspectWorkflowGitPublication(ctx, request)
}
