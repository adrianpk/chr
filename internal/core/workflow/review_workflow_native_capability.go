package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// WorkflowNativeIntegrationCapability rechecks current host grants, actual
// receive-pack capability and the existing forge's effective repository policy.
// The legacy merge API remains source-only and is never enabled by this profile.
func (adapter RemoteReviewAdapter) WorkflowNativeIntegrationCapability(ctx context.Context, request NativeCapabilityRequest, policies NativePolicyReader, git GitPublicationInspector) (workflowstate.NativeCapability, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	capability := workflowforge.NewNativeCapability(workflowNativeCapabilityInput(request))
	if request.Config.PublicationProfile != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		capability.Unsupported = []workflowstate.NativeUnsupported{{Condition: "profile_disabled", Decision: "Explicitly select git_atomic_dual_ref."}}

		return workflowforge.SealNativeCapability(capability)
	}

	if policies == nil || git == nil {
		return capability, fmt.Errorf("native publication requires current host policy and transport observers")
	}

	pr, err := adapter.ReadWorkflowPullRequest(ctx, request.PullRequest.URL)
	if err != nil {
		return capability, err
	}

	if !workflowforge.SameNativeSourcePR(pr, request.PullRequest) || !workflowforge.NativeSourceBinding(workflowNativeCapabilityInput(request)) {
		return capability, fmt.Errorf("%w: canonical PR source binding changed", workflowstate.ErrConflict)
	}

	policy, err := policies.ReadWorkflowNativePolicy(ctx, pr, request.Publication.Git.Method)
	if err != nil {
		return capability, err
	}

	capability.OwnershipRevision = policy.OwnershipRevision
	capability.CredentialIdentityFingerprint = policy.CredentialIdentityFingerprint
	capability.ObservationContract = policy.ObservationContract
	capability.Unsupported = workflowforge.NativePolicyReasons(workflowNativeCapabilityInput(request), policy)

	return adapter.finishWorkflowNativeCapability(ctx, request, policy, git, capability)
}

func (adapter RemoteReviewAdapter) finishWorkflowNativeCapability(ctx context.Context, request NativeCapabilityRequest, policy workflowstate.NativePolicy, git GitPublicationInspector, capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	observation, err := adapter.readWorkflowNativePolicy(ctx, request.PullRequest, request.Publication.Git.Method)
	if err != nil {
		return capability, err
	}

	hostData, err := json.Marshal(policy)
	if err != nil {
		return capability, err
	}

	capability.PolicyFingerprint = contenthash.Sum(append(hostData, observation.Data...))

	identity, err := adapter.readWorkflowNativeIdentity(ctx, request.PullRequest)
	if err != nil {
		return capability, err
	}

	if policy.APIIdentityFingerprint != "" && identity != policy.APIIdentityFingerprint {
		return capability, fmt.Errorf("%w: publisher identity changed", workflowstate.ErrConflict)
	}

	capability.CredentialIdentityFingerprint = contenthash.Sum([]byte(policy.CredentialIdentityFingerprint + "\x00" + identity))

	capability.Unsupported = append(capability.Unsupported, observation.Unsupported...)
	if !workflowforge.NativeObservationContractMatches(observation.Contract, policy.ObservationContract) {
		capability.Unsupported = append(capability.Unsupported, workflowstate.NativeUnsupported{Condition: "observation_uncertified", Decision: "Certify same-original-PR Git merge observation on this repository."})
	}

	if len(capability.Unsupported) != 0 {
		return workflowforge.SealNativeCapability(capability)
	}

	return inspectWorkflowNativeCapability(ctx, request, git, capability)
}

func inspectWorkflowNativeCapability(ctx context.Context, request NativeCapabilityRequest, git GitPublicationInspector, capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	advertisement, err := git.InspectWorkflowGitPublication(ctx, request.Publication)
	if errors.Is(err, workflowstate.ErrAtomicGuard) {
		capability.Unsupported = append(capability.Unsupported, workflowstate.NativeUnsupported{Condition: "atomic_transport_unavailable", Decision: "Select a server supporting atomic dual-ref transactions."})

		return workflowforge.SealNativeCapability(capability)
	}

	if err != nil {
		return capability, err
	}

	if advertisement.SourceCommit != request.Publication.Git.SourceCommit || advertisement.TargetCommit != request.Publication.Git.BaseCommit || !workflowstate.ValidFingerprint(advertisement.Fingerprint) {
		return capability, fmt.Errorf("%w: receive-pack source or target binding changed", workflowstate.ErrConflict)
	}

	capability.ServerCapabilityFingerprint = advertisement.Fingerprint
	capability.AtomicSourceBase, capability.ExactResult = true, true

	return workflowforge.SealNativeCapability(capability)
}

// RecheckWorkflowNativeCapability invalidates a retained grant on any relevant
// policy, role, ownership, method, transport or canonical PR change before intent.
func (adapter RemoteReviewAdapter) RecheckWorkflowNativeCapability(ctx context.Context, request NativeCapabilityRequest, prior workflowstate.NativeCapability, policies NativePolicyReader, git GitPublicationInspector) error {
	current, err := adapter.WorkflowNativeIntegrationCapability(ctx, request, policies, git)
	if err != nil {
		return err
	}

	if !current.AtomicSourceBase || !current.ExactResult || current.Revision != prior.Revision {
		return fmt.Errorf("%w: native capability changed (%s)", workflowstate.ErrConflict, workflowforge.ChangedNativeCapability(prior, current))
	}

	return nil
}

func (adapter RemoteReviewAdapter) readWorkflowNativePolicy(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string) (workflowforge.NativePolicyObservation, error) {
	resolved, err := adapter.resolve(pr.URL)
	if err != nil {
		return workflowforge.NativePolicyObservation{}, err
	}

	client := adapter.workflowForgeClient(pr.Repository, resolved)

	return client.NativePolicy(ctx, pr, method)
}
