package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
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

// WorkflowIntegrationCapability is a preliminary mechanism check. Publication
// additionally requires the prepared result's current transport capability.
func (adapter *NativeReviewAdapter) WorkflowIntegrationCapability(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowstate.PublicationCapability, error) {
	policy, err := adapter.ReadWorkflowNativePolicy(ctx, pr, method)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	observation, err := adapter.readWorkflowNativePolicy(ctx, pr, method)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	if len(observation.unsupported) != 0 || !workflowNativeObservationContractMatches(observation.contract, policy.ObservationContract) {
		return workflowstate.PublicationCapability{}, fmt.Errorf("%w: current repository policy does not support the publisher profile", workflowstate.ErrAtomicGuard)
	}

	identity, err := adapter.readWorkflowNativeIdentity(ctx, pr)
	if err != nil || identity != policy.APIIdentityFingerprint {
		return workflowstate.PublicationCapability{}, fmt.Errorf("%w: publisher identity changed", workflowstate.ErrConflict)
	}

	host, err := json.Marshal(policy)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	return workflowstate.PublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: workspaceconfig.WorkflowPublicationGitAtomicDualRef, Revision: contenthash.Sum(append(host, observation.data...)), AtomicSourceBase: true, ExactResult: true}, nil
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

	request.CandidateRevision, err = GitPublicationRevision(request)
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
func (adapter *NativeReviewAdapter) PublishWorkflowGit(ctx context.Context, request GitPublication) (GitPublicationResult, error) {
	nativePublisher2, err := adapter.nativePublisher(ctx)
	run, owner, profile := nativePublisher2.Run, nativePublisher2.Owner, nativePublisher2.Profile

	if err != nil {
		return GitPublicationResult{}, err
	}

	ownership, err := workflowNativeOwnershipRevision(run, profile)
	if err != nil || request.OwnershipRevision != ownership || request.SourceRef != run.Lane.Branch || request.TargetRef != profile.TargetRef || request.Git.SourceCommit != run.PullRequest.SourceCommit {
		return GitPublicationResult{}, workflowstate.ErrConflict
	}

	config, err := workspaceconfig.Load(owner.layout.ProjectDir())
	if err != nil || config.Work.Integration.PublicationProfile != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return GitPublicationResult{}, workflowstate.ErrAtomicGuard
	}

	return NewWorkflowGitPublisher(profile.Repository, profile.Transport).PublishWorkflowGit(ctx, request)
}

// InspectWorkflowGitPublication keeps the configured transport available for
// read-only reconciliation, including after publication authorization stops.
func (adapter *NativeReviewAdapter) InspectWorkflowGitPublication(ctx context.Context, request GitPublication) (GitPublicationAdvertisement, error) {
	profile, err := LoadWorkflowNativePublisher(adapter.configDir, request.Repository, request.TargetRef)
	if err != nil {
		return GitPublicationAdvertisement{}, err
	}

	return NewWorkflowGitPublisher(profile.Repository, profile.Transport).InspectWorkflowGitPublication(ctx, request)
}
