package core

import (
	"context"
	"encoding/json"
	"fmt"
)

// WorkflowNativeReviewAdapter connects existing remote observation to the
// invocation's exact source ownership and current host publisher profile.
type WorkflowNativeReviewAdapter struct {
	RemoteReviewAdapter
	store     *WorkflowStore
	id        string
	configDir string
}

// NewWorkflowNativeReviewAdapter stores dependencies without opening resources.
func NewWorkflowNativeReviewAdapter(remote RemoteReviewAdapter, store *WorkflowStore, id, configDir string) *WorkflowNativeReviewAdapter {
	return &WorkflowNativeReviewAdapter{remote, store, id, configDir}
}

// WorkflowIntegrationCapability is a preliminary mechanism check. Publication
// additionally requires the prepared result's current transport capability.
func (adapter *WorkflowNativeReviewAdapter) WorkflowIntegrationCapability(ctx context.Context, pr WorkflowPullRequestSnapshot, method string) (WorkflowPublicationCapability, error) {
	policy, err := adapter.ReadWorkflowNativePolicy(ctx, pr, method)
	if err != nil {
		return WorkflowPublicationCapability{}, err
	}

	data, reasons, contract, err := adapter.readWorkflowNativePolicy(ctx, pr, method)
	if err != nil {
		return WorkflowPublicationCapability{}, err
	}

	if len(reasons) != 0 || !workflowNativeObservationContractMatches(contract, policy.ObservationContract) {
		return WorkflowPublicationCapability{}, fmt.Errorf("%w: current repository policy does not support the publisher profile", ErrWorkflowAtomicGuard)
	}

	identity, err := adapter.readWorkflowNativeIdentity(ctx, pr)
	if err != nil || identity != policy.APIIdentityFingerprint {
		return WorkflowPublicationCapability{}, fmt.Errorf("%w: publisher identity changed", ErrWorkflowConflict)
	}

	host, err := json.Marshal(policy)
	if err != nil {
		return WorkflowPublicationCapability{}, err
	}

	return WorkflowPublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: WorkflowPublicationGitAtomicDualRef, Revision: workflowHash(append(host, data...)), AtomicSourceBase: true, ExactResult: true}, nil
}

// WorkflowPreparedNativeCapability observes the actual receive-pack server for
// the host's exact S/B/P proof, using the same configured publisher as dispatch.
func (adapter *WorkflowNativeReviewAdapter) WorkflowPreparedNativeCapability(ctx context.Context, pr WorkflowPullRequestSnapshot, proof WorkflowIntegrationProof) (WorkflowNativeCapability, error) {
	policy, err := adapter.ReadWorkflowNativePolicy(ctx, pr, proof.Method)
	if err != nil {
		return WorkflowNativeCapability{}, err
	}

	run, owner, profile, err := adapter.nativePublisher(ctx)
	if err != nil {
		return WorkflowNativeCapability{}, err
	}

	request := WorkflowGitPublication{Repository: pr.Repository, SourceRepository: pr.HeadRepository, Directory: owner.layout.ProjectDir(), SourceRef: policy.SourceRef, TargetRef: policy.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: policy.OwnershipRevision, Git: proof}

	request.CandidateRevision, err = WorkflowGitPublicationRevision(request)
	if err != nil {
		return WorkflowNativeCapability{}, err
	}

	config, err := LoadWorkspaceConfig(owner.layout.ProjectDir())
	if err != nil {
		return WorkflowNativeCapability{}, err
	}

	if proof.SourceCommit != run.PullRequest.SourceCommit {
		return WorkflowNativeCapability{}, ErrWorkflowConflict
	}

	query := WorkflowNativeCapabilityRequest{Config: config.Work.Integration, PullRequest: pr, Publication: request}

	return adapter.WorkflowNativeIntegrationCapability(ctx, query, adapter, NewWorkflowGitPublisher(profile.Repository, profile.Transport))
}

// PublishWorkflowGit rechecks the exact local ownership/profile binding before
// entering the existing atomic transport. It never calls a provider merge API.
func (adapter *WorkflowNativeReviewAdapter) PublishWorkflowGit(ctx context.Context, request WorkflowGitPublication) (WorkflowGitPublicationResult, error) {
	run, owner, profile, err := adapter.nativePublisher(ctx)
	if err != nil {
		return WorkflowGitPublicationResult{}, err
	}

	ownership, err := workflowNativeOwnershipRevision(run, profile)
	if err != nil || request.OwnershipRevision != ownership || request.SourceRef != run.Lane.Branch || request.TargetRef != profile.TargetRef || request.Git.SourceCommit != run.PullRequest.SourceCommit {
		return WorkflowGitPublicationResult{}, ErrWorkflowConflict
	}

	config, err := LoadWorkspaceConfig(owner.layout.ProjectDir())
	if err != nil || config.Work.Integration.PublicationProfile != WorkflowPublicationGitAtomicDualRef {
		return WorkflowGitPublicationResult{}, ErrWorkflowAtomicGuard
	}

	return NewWorkflowGitPublisher(profile.Repository, profile.Transport).PublishWorkflowGit(ctx, request)
}
