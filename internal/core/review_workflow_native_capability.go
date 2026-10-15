package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// WorkflowNativeIntegrationCapability rechecks current host grants, actual
// receive-pack capability and the existing forge's effective repository policy.
// The legacy merge API remains source-only and is never enabled by this profile.
func (adapter RemoteReviewAdapter) WorkflowNativeIntegrationCapability(ctx context.Context, request WorkflowNativeCapabilityRequest, policies WorkflowNativePolicyReader, git WorkflowGitPublicationInspector) (WorkflowNativeCapability, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	capability := newWorkflowNativeCapability(request)
	if request.Config.PublicationProfile != WorkflowPublicationGitAtomicDualRef {
		capability.Unsupported = []WorkflowNativeUnsupported{{"profile_disabled", "Explicitly select git_atomic_dual_ref."}}

		return sealWorkflowNativeCapability(capability)
	}

	if policies == nil || git == nil {
		return capability, fmt.Errorf("native publication requires current host policy and transport observers")
	}

	pr, err := adapter.ReadWorkflowPullRequest(ctx, request.PullRequest.URL)
	if err != nil {
		return capability, err
	}

	if !sameWorkflowNativeSourcePR(pr, request.PullRequest) || !workflowNativeSourceBinding(request) {
		return capability, fmt.Errorf("%w: canonical PR source binding changed", ErrWorkflowConflict)
	}

	policy, err := policies.ReadWorkflowNativePolicy(ctx, pr, request.Publication.Git.Method)
	if err != nil {
		return capability, err
	}

	capability.OwnershipRevision = policy.OwnershipRevision
	capability.CredentialIdentityFingerprint = policy.CredentialIdentityFingerprint
	capability.ObservationContract = policy.ObservationContract
	capability.Unsupported = workflowNativePolicyReasons(request, policy)

	return adapter.finishWorkflowNativeCapability(ctx, request, policy, git, capability)
}

func sameWorkflowNativeSourcePR(current, retained WorkflowPullRequestSnapshot) bool {
	// GitHub computes a speculative merge_commit_sha while an open PR is still
	// unmerged. It is not a source/base change or evidence of delivered P.
	current.ResultCommit, retained.ResultCommit = "", ""

	return current == retained
}

func newWorkflowNativeCapability(request WorkflowNativeCapabilityRequest) WorkflowNativeCapability {
	return WorkflowNativeCapability{
		WorkflowPublicationCapability: WorkflowPublicationCapability{Repository: request.Publication.Repository, PullRequestURL: request.PullRequest.URL, Method: request.Publication.Git.Method, Mechanism: WorkflowPublicationGitAtomicDualRef},
		SourceRef:                     request.Publication.SourceRef, TargetRef: request.Publication.TargetRef, AdapterVersion: workflowNativeAdapterVersion,
		OriginalSourceCommit: request.Publication.Git.SourceCommit, ExpectedBaseCommit: request.Publication.Git.BaseCommit, PreparedResultCommit: request.Publication.Git.ResultCommit,
	}
}

func workflowNativeSourceBinding(request WorkflowNativeCapabilityRequest) bool {
	pr, publication := request.PullRequest, request.Publication

	return pr.Repository == publication.Repository && pr.HeadRepository == publication.SourceRepository &&
		"refs/heads/"+pr.HeadBranch == publication.SourceRef && "refs/heads/"+pr.BaseBranch == publication.TargetRef &&
		pr.State == "open" && !pr.Merged && pr.HeadCommit == publication.Git.SourceCommit && pr.BaseCommit == publication.Git.BaseCommit
}

func (adapter RemoteReviewAdapter) finishWorkflowNativeCapability(ctx context.Context, request WorkflowNativeCapabilityRequest, policy WorkflowNativePolicy, git WorkflowGitPublicationInspector, capability WorkflowNativeCapability) (WorkflowNativeCapability, error) {
	policyData, reasons, contract, err := adapter.readWorkflowNativePolicy(ctx, request.PullRequest, request.Publication.Git.Method)
	if err != nil {
		return capability, err
	}

	hostData, err := json.Marshal(policy)
	if err != nil {
		return capability, err
	}

	capability.PolicyFingerprint = workflowHash(append(hostData, policyData...))

	identity, err := adapter.readWorkflowNativeIdentity(ctx, request.PullRequest)
	if err != nil {
		return capability, err
	}

	capability.CredentialIdentityFingerprint = workflowHash([]byte(policy.CredentialIdentityFingerprint + "\x00" + identity))

	capability.Unsupported = append(capability.Unsupported, reasons...)
	if !workflowNativeObservationContractMatches(contract, policy.ObservationContract) {
		capability.Unsupported = append(capability.Unsupported, WorkflowNativeUnsupported{"observation_uncertified", "Certify same-original-PR Git merge observation on this repository."})
	}

	if len(capability.Unsupported) != 0 {
		return sealWorkflowNativeCapability(capability)
	}

	return inspectWorkflowNativeCapability(ctx, request, git, capability)
}

func inspectWorkflowNativeCapability(ctx context.Context, request WorkflowNativeCapabilityRequest, git WorkflowGitPublicationInspector, capability WorkflowNativeCapability) (WorkflowNativeCapability, error) {
	advertisement, err := git.InspectWorkflowGitPublication(ctx, request.Publication)
	if errors.Is(err, ErrWorkflowAtomicGuard) {
		capability.Unsupported = append(capability.Unsupported, WorkflowNativeUnsupported{"atomic_transport_unavailable", "Select a server supporting atomic dual-ref transactions."})

		return sealWorkflowNativeCapability(capability)
	}

	if err != nil {
		return capability, err
	}

	if advertisement.SourceCommit != request.Publication.Git.SourceCommit || advertisement.TargetCommit != request.Publication.Git.BaseCommit || !validWorkflowNativeFingerprint(advertisement.Fingerprint) {
		return capability, fmt.Errorf("%w: receive-pack source or target binding changed", ErrWorkflowConflict)
	}

	capability.ServerCapabilityFingerprint = advertisement.Fingerprint
	capability.AtomicSourceBase, capability.ExactResult = true, true

	return sealWorkflowNativeCapability(capability)
}

// RecheckWorkflowNativeCapability invalidates a retained grant on any relevant
// policy, role, ownership, method, transport or canonical PR change before intent.
func (adapter RemoteReviewAdapter) RecheckWorkflowNativeCapability(ctx context.Context, request WorkflowNativeCapabilityRequest, prior WorkflowNativeCapability, policies WorkflowNativePolicyReader, git WorkflowGitPublicationInspector) error {
	current, err := adapter.WorkflowNativeIntegrationCapability(ctx, request, policies, git)
	if err != nil {
		return err
	}

	if !current.AtomicSourceBase || !current.ExactResult || current.Revision != prior.Revision {
		return fmt.Errorf("%w: native capability changed (%s)", ErrWorkflowConflict, changedWorkflowNativeCapability(prior, current))
	}

	return nil
}

func changedWorkflowNativeCapability(prior, current WorkflowNativeCapability) string {
	fields := []struct{ name, prior, current string }{
		{"policy", prior.PolicyFingerprint, current.PolicyFingerprint},
		{"server capability", prior.ServerCapabilityFingerprint, current.ServerCapabilityFingerprint},
		{"publisher identity", prior.CredentialIdentityFingerprint, current.CredentialIdentityFingerprint},
		{"ownership", prior.OwnershipRevision, current.OwnershipRevision},
		{"observation contract", prior.ObservationContract, current.ObservationContract},
	}
	for _, field := range fields {
		if field.prior != field.current {
			return field.name
		}
	}

	return "source, target, method or profile"
}

func (adapter RemoteReviewAdapter) readWorkflowNativePolicy(ctx context.Context, pr WorkflowPullRequestSnapshot, method string) ([]byte, []WorkflowNativeUnsupported, string, error) {
	reference, config, err := adapter.resolve(pr.URL)
	if err != nil {
		return nil, nil, "", err
	}

	paths := []string{workflowRemoteRepositoryPath(reference), workflowRemoteRepositoryPath(reference) + "/branches/" + url.PathEscape(pr.BaseBranch), workflowRemoteRepositoryPath(reference) + "/branches/" + url.PathEscape(pr.HeadBranch)}

	contract := "forgejo_same_pr_git_auto_merge/v1"
	if reference.Provider == ReviewProviderGitHub {
		contract = "github_same_pr_git_merge/v1"

		paths = append(paths, workflowRemoteRepositoryPath(reference)+"/rules/branches/"+url.PathEscape(pr.BaseBranch), workflowRemoteRepositoryPath(reference)+"/rules/branches/"+url.PathEscape(pr.HeadBranch))
	}

	var responses []json.RawMessage

	for _, path := range paths {
		data, err := adapter.workflowProviderRunner().Output(ctx, config.CLI, workflowRemoteAPIArgs(pr.Repository, reference, "GET", path, false)...)
		if err != nil || len(data) > WorkflowFileLimit || !json.Valid(data) {
			return nil, nil, contract, fmt.Errorf("native publication policy could not be completely observed")
		}

		responses = append(responses, data)
	}

	var repository workflowNativeRepositoryPolicy
	if err := json.Unmarshal(responses[0], &repository); err != nil || strings.TrimSuffix(repository.URL, "/") != pr.Repository {
		return nil, nil, contract, ErrWorkflowConflict
	}

	return workflowNativeRemotePolicy(responses, method, contract)
}
