package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ReadWorkflowNativePolicy combines a private target certificate with retained
// host ownership of one exact source. A branch prefix never supplies authority.
func (adapter *NativeReviewAdapter) ReadWorkflowNativePolicy(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowstate.NativePolicy, error) {
	nativePublisher, err := adapter.nativePublisher(ctx)
	run, owner, profile := nativePublisher.Run, nativePublisher.Owner, nativePublisher.Profile

	if err != nil {
		return workflowstate.NativePolicy{}, err
	}

	if !workflowNativeOwnedPR(run, pr) || !slices.Contains(profile.CertifiedMethods, method) {
		return workflowstate.NativePolicy{}, fmt.Errorf("%w: source ownership or method is outside the publisher profile", workflowstate.ErrAtomicGuard)
	}

	policy, err := owner.workflowReviewPolicy(run)
	if err != nil || policy.Method != method {
		return workflowstate.NativePolicy{}, workflowstate.ErrConflict
	}

	ownership, err := workflowNativeOwnershipRevision(run, profile)
	if err != nil {
		return workflowstate.NativePolicy{}, err
	}

	image, err := json.Marshal(profile)
	if err != nil {
		return workflowstate.NativePolicy{}, err
	}

	return workflowstate.NativePolicy{Repository: profile.Repository, SourceRef: run.Lane.Branch, TargetRef: profile.TargetRef, OwnershipRevision: ownership, SourceRewriteAuthorized: true, FastForwardTargetAuthorized: true, HostCandidateReview: true, ReviewMode: policy.Mode, CredentialIdentityFingerprint: contenthash.Sum(image), APIIdentityFingerprint: profile.APIIdentityFingerprint, ObservationContract: profile.ObservationContract, CertifiedMethods: profile.CertifiedMethods}, nil
}

func (adapter *NativeReviewAdapter) nativePublisher(ctx context.Context) (workflowNativePublisher, error) {
	// Capability checks also run inside the repository mutation lock. Read the
	// already prepared owner images without acquiring that lock recursively.
	invocationLocation, err := adapter.store.locateWorkflowInvocationPrepared(ctx, adapter.id, func(*Store) error { return nil })
	if err != nil {
		return workflowNativePublisher{Run: invocationLocation.Run, Owner: invocationLocation.Owner, Profile: NativePublisherProfile{}}, err
	}

	if !workflowNativePublicationOwner(invocationLocation.Run) {
		return workflowNativePublisher{Run: invocationLocation.Run, Owner: invocationLocation.Owner, Profile: NativePublisherProfile{}}, workflowstate.ErrConflict
	}

	config, err := workspaceconfig.Load(invocationLocation.Owner.layout.ProjectDir())
	if err != nil || config.Work.Integration.PublicationProfile != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return workflowNativePublisher{Run: invocationLocation.Run, Owner: invocationLocation.Owner, Profile: NativePublisherProfile{}}, workflowstate.ErrAtomicGuard
	}

	if err := invocationLocation.Owner.checkWorkflowLaneRegistration(ctx, *invocationLocation.Run.Lane); err != nil {
		return workflowNativePublisher{Run: invocationLocation.Run, Owner: invocationLocation.Owner, Profile: NativePublisherProfile{}}, err
	}

	profile, err := LoadWorkflowNativePublisher(adapter.configDir, invocationLocation.Run.PullRequest.Request.Repository, "refs/heads/"+invocationLocation.Run.PullRequest.Request.BaseBranch)

	return workflowNativePublisher{Run: invocationLocation.Run, Owner: invocationLocation.Owner, Profile: profile}, err
}

func workflowNativePublicationOwner(run workflowInvocationFile) bool {
	return run.Lane != nil && run.PullRequest != nil && run.PullRequest.State == "bound" && run.PullRequest.CandidateHash == run.CandidateHash && run.StopIntent == "" && run.Status != workflowstate.InvocationPaused && !workflowstate.InvocationTerminal(run.Status)
}

func workflowNativeOwnedPR(run workflowInvocationFile, pr workflowstate.PullRequestSnapshot) bool {
	binding := run.PullRequest

	return binding.URL == pr.URL && binding.Request.Repository == pr.Repository && pr.HeadRepository == pr.Repository && binding.Request.HeadBranch == pr.HeadBranch && binding.Request.BaseBranch == pr.BaseBranch && run.Lane.Branch == "refs/heads/"+pr.HeadBranch && binding.Unit == run.Lane.Unit && binding.SourceCommit == pr.HeadCommit && pr.State == "open" && !pr.Merged
}

func workflowNativeOwnershipRevision(run workflowInvocationFile, profile NativePublisherProfile) (string, error) {
	image, err := json.Marshal(struct {
		Project, Namespace, Invocation, Owner, Request, PullRequest, Candidate string
		Lane                                                                   workflowstate.Lane
		Publisher                                                              NativePublisherProfile
	}{run.ProjectID, run.Namespace, run.ID, run.OwnerRoot, run.RequestKey, run.PullRequest.ID, run.CandidateHash, *run.Lane, profile})

	return contenthash.Sum(image), err
}
