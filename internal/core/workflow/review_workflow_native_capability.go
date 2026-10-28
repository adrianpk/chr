package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// WorkflowNativeIntegrationCapability rechecks current host grants, actual
// receive-pack capability and the existing forge's effective repository policy.
// The legacy merge API remains source-only and is never enabled by this profile.
func (adapter RemoteReviewAdapter) WorkflowNativeIntegrationCapability(ctx context.Context, request NativeCapabilityRequest, policies NativePolicyReader, git GitPublicationInspector) (workflowstate.NativeCapability, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	capability := newWorkflowNativeCapability(request)
	if request.Config.PublicationProfile != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		capability.Unsupported = []workflowstate.NativeUnsupported{{Condition: "profile_disabled", Decision: "Explicitly select git_atomic_dual_ref."}}

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
		return capability, fmt.Errorf("%w: canonical PR source binding changed", workflowstate.ErrConflict)
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

func sameWorkflowNativeSourcePR(current, retained workflowstate.PullRequestSnapshot) bool {
	// GitHub computes a speculative merge_commit_sha while an open PR is still
	// unmerged. It is not a source/base change or evidence of delivered P.
	current.ResultCommit, retained.ResultCommit = "", ""

	return current == retained
}

func newWorkflowNativeCapability(request NativeCapabilityRequest) workflowstate.NativeCapability {
	return workflowstate.NativeCapability{
		PublicationCapability: workflowstate.PublicationCapability{Repository: request.Publication.Repository, PullRequestURL: request.PullRequest.URL, Method: request.Publication.Git.Method, Mechanism: workspaceconfig.WorkflowPublicationGitAtomicDualRef},
		SourceRef:             request.Publication.SourceRef, TargetRef: request.Publication.TargetRef, AdapterVersion: workflowNativeAdapterVersion,
		OriginalSourceCommit: request.Publication.Git.SourceCommit, ExpectedBaseCommit: request.Publication.Git.BaseCommit, PreparedResultCommit: request.Publication.Git.ResultCommit,
	}
}

func workflowNativeSourceBinding(request NativeCapabilityRequest) bool {
	pr, publication := request.PullRequest, request.Publication

	return pr.Repository == publication.Repository && pr.HeadRepository == publication.SourceRepository &&
		"refs/heads/"+pr.HeadBranch == publication.SourceRef && "refs/heads/"+pr.BaseBranch == publication.TargetRef &&
		pr.State == "open" && !pr.Merged && pr.HeadCommit == publication.Git.SourceCommit && pr.BaseCommit == publication.Git.BaseCommit
}

func (adapter RemoteReviewAdapter) finishWorkflowNativeCapability(ctx context.Context, request NativeCapabilityRequest, policy NativePolicy, git GitPublicationInspector, capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	observation, err := adapter.readWorkflowNativePolicy(ctx, request.PullRequest, request.Publication.Git.Method)
	if err != nil {
		return capability, err
	}

	hostData, err := json.Marshal(policy)
	if err != nil {
		return capability, err
	}

	capability.PolicyFingerprint = contenthash.Sum(append(hostData, observation.data...))

	identity, err := adapter.readWorkflowNativeIdentity(ctx, request.PullRequest)
	if err != nil {
		return capability, err
	}

	if policy.APIIdentityFingerprint != "" && identity != policy.APIIdentityFingerprint {
		return capability, fmt.Errorf("%w: publisher identity changed", workflowstate.ErrConflict)
	}

	capability.CredentialIdentityFingerprint = contenthash.Sum([]byte(policy.CredentialIdentityFingerprint + "\x00" + identity))

	capability.Unsupported = append(capability.Unsupported, observation.unsupported...)
	if !workflowNativeObservationContractMatches(observation.contract, policy.ObservationContract) {
		capability.Unsupported = append(capability.Unsupported, workflowstate.NativeUnsupported{Condition: "observation_uncertified", Decision: "Certify same-original-PR Git merge observation on this repository."})
	}

	if len(capability.Unsupported) != 0 {
		return sealWorkflowNativeCapability(capability)
	}

	return inspectWorkflowNativeCapability(ctx, request, git, capability)
}

func inspectWorkflowNativeCapability(ctx context.Context, request NativeCapabilityRequest, git GitPublicationInspector, capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	advertisement, err := git.InspectWorkflowGitPublication(ctx, request.Publication)
	if errors.Is(err, ErrWorkflowAtomicGuard) {
		capability.Unsupported = append(capability.Unsupported, workflowstate.NativeUnsupported{Condition: "atomic_transport_unavailable", Decision: "Select a server supporting atomic dual-ref transactions."})

		return sealWorkflowNativeCapability(capability)
	}

	if err != nil {
		return capability, err
	}

	if advertisement.SourceCommit != request.Publication.Git.SourceCommit || advertisement.TargetCommit != request.Publication.Git.BaseCommit || !workflowstate.ValidFingerprint(advertisement.Fingerprint) {
		return capability, fmt.Errorf("%w: receive-pack source or target binding changed", workflowstate.ErrConflict)
	}

	capability.ServerCapabilityFingerprint = advertisement.Fingerprint
	capability.AtomicSourceBase, capability.ExactResult = true, true

	return sealWorkflowNativeCapability(capability)
}

// RecheckWorkflowNativeCapability invalidates a retained grant on any relevant
// policy, role, ownership, method, transport or canonical PR change before intent.
func (adapter RemoteReviewAdapter) RecheckWorkflowNativeCapability(ctx context.Context, request NativeCapabilityRequest, prior workflowstate.NativeCapability, policies NativePolicyReader, git GitPublicationInspector) error {
	current, err := adapter.WorkflowNativeIntegrationCapability(ctx, request, policies, git)
	if err != nil {
		return err
	}

	if !current.AtomicSourceBase || !current.ExactResult || current.Revision != prior.Revision {
		return fmt.Errorf("%w: native capability changed (%s)", workflowstate.ErrConflict, changedWorkflowNativeCapability(prior, current))
	}

	return nil
}

func changedWorkflowNativeCapability(prior, current workflowstate.NativeCapability) string {
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

func (adapter RemoteReviewAdapter) readWorkflowNativePolicy(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowNativePolicyObservation, error) {
	resolvedReview, err := adapter.resolve(pr.URL)
	if err != nil {
		return workflowNativePolicyObservation{}, err
	}

	reference, config := resolvedReview.reference, resolvedReview.config

	paths := []string{workflowRemoteRepositoryPath(reference), workflowRemoteRepositoryPath(reference) + "/branches/" + url.PathEscape(pr.BaseBranch), workflowRemoteRepositoryPath(reference) + "/branches/" + url.PathEscape(pr.HeadBranch)}

	contract := "forgejo_same_pr_git_auto_merge/v1"
	if reference.Provider == ReviewProviderGitHub {
		contract = "github_same_pr_git_merge/v1"

		paths = append(paths, workflowRemoteRepositoryPath(reference)+"/rules/branches/"+url.PathEscape(pr.BaseBranch), workflowRemoteRepositoryPath(reference)+"/rules/branches/"+url.PathEscape(pr.HeadBranch))
	}

	var responses []json.RawMessage

	for _, path := range paths {
		data, err := adapter.workflowProviderRunner().Output(ctx, config.CLI, workflowRemoteAPIArgs(pr.Repository, reference, "GET", path, false)...)
		if err != nil || len(data) > workflowstate.FileLimit || !json.Valid(data) {
			return workflowNativePolicyObservation{contract: contract}, fmt.Errorf("native publication policy could not be completely observed")
		}

		responses = append(responses, data)
	}

	var repository workflowNativeRepositoryPolicy
	if err := json.Unmarshal(responses[0], &repository); err != nil || strings.TrimSuffix(repository.URL, "/") != pr.Repository {
		return workflowNativePolicyObservation{contract: contract}, workflowstate.ErrConflict
	}

	return workflowNativeRemotePolicy(responses, method, contract)
}
