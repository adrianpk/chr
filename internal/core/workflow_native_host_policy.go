package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

// ReadWorkflowNativePolicy combines a private target certificate with retained
// host ownership of one exact source. A branch prefix never supplies authority.
func (adapter *WorkflowNativeReviewAdapter) ReadWorkflowNativePolicy(ctx context.Context, pr WorkflowPullRequestSnapshot, method string) (WorkflowNativePolicy, error) {
	run, owner, profile, err := adapter.nativePublisher(ctx)
	if err != nil {
		return WorkflowNativePolicy{}, err
	}

	if !workflowNativeOwnedPR(run, pr) || !slices.Contains(profile.CertifiedMethods, method) {
		return WorkflowNativePolicy{}, fmt.Errorf("%w: source ownership or method is outside the publisher profile", ErrWorkflowAtomicGuard)
	}

	policy, err := owner.workflowReviewPolicy(run)
	if err != nil || policy.Method != method {
		return WorkflowNativePolicy{}, ErrWorkflowConflict
	}

	ownership, err := workflowNativeOwnershipRevision(run, profile)
	if err != nil {
		return WorkflowNativePolicy{}, err
	}

	image, err := json.Marshal(profile)
	if err != nil {
		return WorkflowNativePolicy{}, err
	}

	return WorkflowNativePolicy{Repository: profile.Repository, SourceRef: run.Lane.Branch, TargetRef: profile.TargetRef, OwnershipRevision: ownership, SourceRewriteAuthorized: true, FastForwardTargetAuthorized: true, HostCandidateReview: true, ReviewMode: policy.Mode, CredentialIdentityFingerprint: workflowHash(image), APIIdentityFingerprint: profile.APIIdentityFingerprint, ObservationContract: profile.ObservationContract, CertifiedMethods: profile.CertifiedMethods}, nil
}

func (adapter *WorkflowNativeReviewAdapter) nativePublisher(ctx context.Context) (workflowInvocationFile, *WorkflowStore, WorkflowNativePublisherProfile, error) {
	// Capability checks also run inside the repository mutation lock. Read the
	// already prepared owner images without acquiring that lock recursively.
	run, owner, err := adapter.store.locateWorkflowInvocationPrepared(ctx, adapter.id, func(*WorkflowStore) error { return nil })
	if err != nil {
		return run, owner, WorkflowNativePublisherProfile{}, err
	}

	if !workflowNativePublicationOwner(run) {
		return run, owner, WorkflowNativePublisherProfile{}, ErrWorkflowConflict
	}

	config, err := LoadWorkspaceConfig(owner.layout.ProjectDir())
	if err != nil || config.Work.Integration.PublicationProfile != WorkflowPublicationGitAtomicDualRef {
		return run, owner, WorkflowNativePublisherProfile{}, ErrWorkflowAtomicGuard
	}

	if err := owner.checkWorkflowLaneRegistration(ctx, *run.Lane); err != nil {
		return run, owner, WorkflowNativePublisherProfile{}, err
	}

	profile, err := LoadWorkflowNativePublisher(adapter.configDir, run.PullRequest.Request.Repository, "refs/heads/"+run.PullRequest.Request.BaseBranch)

	return run, owner, profile, err
}

func workflowNativePublicationOwner(run workflowInvocationFile) bool {
	return run.Lane != nil && run.PullRequest != nil && run.PullRequest.State == "bound" && run.PullRequest.CandidateHash == run.CandidateHash && run.StopIntent == "" && run.Status != "paused" && !workflowInvocationTerminal(run.Status)
}

func workflowNativeOwnedPR(run workflowInvocationFile, pr WorkflowPullRequestSnapshot) bool {
	binding := run.PullRequest

	return binding.URL == pr.URL && binding.Request.Repository == pr.Repository && pr.HeadRepository == pr.Repository && binding.Request.HeadBranch == pr.HeadBranch && binding.Request.BaseBranch == pr.BaseBranch && run.Lane.Branch == "refs/heads/"+pr.HeadBranch && binding.Unit == run.Lane.Unit && binding.SourceCommit == pr.HeadCommit && pr.State == "open" && !pr.Merged
}

func workflowNativeOwnershipRevision(run workflowInvocationFile, profile WorkflowNativePublisherProfile) (string, error) {
	image, err := json.Marshal(struct {
		Project, Namespace, Invocation, Owner, Request, PullRequest, Candidate string
		Lane                                                                   WorkflowLane
		Publisher                                                              WorkflowNativePublisherProfile
	}{run.ProjectID, run.Namespace, run.ID, run.OwnerRoot, run.RequestKey, run.PullRequest.ID, run.CandidateHash, *run.Lane, profile})

	return workflowHash(image), err
}
