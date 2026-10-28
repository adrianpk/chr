package forge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativePolicyReader rechecks trusted ownership, credential roles and
// provider observation certification before each admission or intent check.
type NativePolicyReader interface {
	ReadWorkflowNativePolicy(context.Context, workflowstate.PullRequestSnapshot, string) (workflowstate.NativePolicy, error)
}

// NativeCapabilityProvider observes enabled endpoint policy and publisher identity.
type NativeCapabilityProvider interface {
	ReadWorkflowPullRequest(context.Context, string) (workflowstate.PullRequestSnapshot, error)
	ReadWorkflowNativeProviderPolicy(context.Context, workflowstate.PullRequestSnapshot, string) (NativePolicyObservation, error)
	ReadWorkflowPublisherIdentity(context.Context, workflowstate.PullRequestSnapshot) (string, error)
}

// NativeCapabilityTransport inspects the exact host-bound publication request.
// The callback-bearing request and local transport mapping remain in the host.
type NativeCapabilityTransport interface {
	InspectNativePublication(context.Context) (workflowstate.GitPublicationAdvertisement, error)
}

// NativeCapabilityReaders supplies current grants and independent read-only observations.
type NativeCapabilityReaders struct {
	Provider  NativeCapabilityProvider
	Policy    NativePolicyReader
	Transport NativeCapabilityTransport
}

// ObserveNativeCapability rechecks current host grants, actual
// receive-pack capability and the existing forge's effective repository policy.
// The legacy merge API remains source-only and is never enabled by this profile.
func ObserveNativeCapability(ctx context.Context, request NativeCapabilityInput, readers NativeCapabilityReaders) (workflowstate.NativeCapability, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	capability := NewNativeCapability(request)
	if request.Profile != nativePublicationMechanism {
		capability.Unsupported = []workflowstate.NativeUnsupported{{Condition: "profile_disabled", Decision: "Explicitly select git_atomic_dual_ref."}}

		return SealNativeCapability(capability)
	}

	if readers.Policy == nil || readers.Transport == nil {
		return capability, fmt.Errorf("native publication requires current host policy and transport observers")
	}

	pr, err := readers.Provider.ReadWorkflowPullRequest(ctx, request.PullRequest.URL)
	if err != nil {
		return capability, err
	}

	if !SameNativeSourcePR(pr, request.PullRequest) || !NativeSourceBinding(request) {
		return capability, fmt.Errorf("%w: canonical PR source binding changed", workflowstate.ErrConflict)
	}

	policy, err := readers.Policy.ReadWorkflowNativePolicy(ctx, pr, request.Publication.Git.Method)
	if err != nil {
		return capability, err
	}

	capability.OwnershipRevision = policy.OwnershipRevision
	capability.CredentialIdentityFingerprint = policy.CredentialIdentityFingerprint
	capability.ObservationContract = policy.ObservationContract
	capability.Unsupported = NativePolicyReasons(request, policy)

	return finishNativeCapability(ctx, request, policy, readers, capability)
}

func finishNativeCapability(ctx context.Context, request NativeCapabilityInput, policy workflowstate.NativePolicy, readers NativeCapabilityReaders, capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	observation, err := readers.Provider.ReadWorkflowNativeProviderPolicy(ctx, request.PullRequest, request.Publication.Git.Method)
	if err != nil {
		return capability, err
	}

	hostData, err := json.Marshal(policy)
	if err != nil {
		return capability, err
	}

	capability.PolicyFingerprint = contenthash.Sum(append(hostData, observation.Data...))

	identity, err := readers.Provider.ReadWorkflowPublisherIdentity(ctx, request.PullRequest)
	if err != nil {
		return capability, err
	}

	if policy.APIIdentityFingerprint != "" && identity != policy.APIIdentityFingerprint {
		return capability, fmt.Errorf("%w: publisher identity changed", workflowstate.ErrConflict)
	}

	capability.CredentialIdentityFingerprint = contenthash.Sum([]byte(policy.CredentialIdentityFingerprint + "\x00" + identity))

	capability.Unsupported = append(capability.Unsupported, observation.Unsupported...)
	if !NativeObservationContractMatches(observation.Contract, policy.ObservationContract) {
		capability.Unsupported = append(capability.Unsupported, workflowstate.NativeUnsupported{Condition: "observation_uncertified", Decision: "Certify same-original-PR Git merge observation on this repository."})
	}

	if len(capability.Unsupported) != 0 {
		return SealNativeCapability(capability)
	}

	return inspectNativeCapability(ctx, request, readers.Transport, capability)
}

func inspectNativeCapability(ctx context.Context, request NativeCapabilityInput, transport NativeCapabilityTransport, capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	advertisement, err := transport.InspectNativePublication(ctx)
	if errors.Is(err, workflowstate.ErrAtomicGuard) {
		capability.Unsupported = append(capability.Unsupported, workflowstate.NativeUnsupported{Condition: "atomic_transport_unavailable", Decision: "Select a server supporting atomic dual-ref transactions."})

		return SealNativeCapability(capability)
	}

	if err != nil {
		return capability, err
	}

	if advertisement.SourceCommit != request.Publication.Git.SourceCommit || advertisement.TargetCommit != request.Publication.Git.BaseCommit || !workflowstate.ValidFingerprint(advertisement.Fingerprint) {
		return capability, fmt.Errorf("%w: receive-pack source or target binding changed", workflowstate.ErrConflict)
	}

	capability.ServerCapabilityFingerprint = advertisement.Fingerprint
	capability.AtomicSourceBase, capability.ExactResult = true, true

	return SealNativeCapability(capability)
}

// RecheckNativeCapability invalidates a retained grant on any relevant
// policy, role, ownership, method, transport or canonical PR change before intent.
func RecheckNativeCapability(ctx context.Context, request NativeCapabilityInput, prior workflowstate.NativeCapability, readers NativeCapabilityReaders) error {
	current, err := ObserveNativeCapability(ctx, request, readers)
	if err != nil {
		return err
	}

	if !current.AtomicSourceBase || !current.ExactResult || current.Revision != prior.Revision {
		return fmt.Errorf("%w: native capability changed (%s)", workflowstate.ErrConflict, ChangedNativeCapability(prior, current))
	}

	return nil
}

// ObserveNativePublicationCapability is a preliminary mechanism check. Publication
// additionally requires the prepared result's current transport capability.
func ObserveNativePublicationCapability(ctx context.Context, pr workflowstate.PullRequestSnapshot, method string, readers NativeCapabilityReaders) (workflowstate.PublicationCapability, error) {
	policy, err := readers.Policy.ReadWorkflowNativePolicy(ctx, pr, method)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	observation, err := readers.Provider.ReadWorkflowNativeProviderPolicy(ctx, pr, method)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	if len(observation.Unsupported) != 0 || !NativeObservationContractMatches(observation.Contract, policy.ObservationContract) {
		return workflowstate.PublicationCapability{}, fmt.Errorf("%w: current repository policy does not support the publisher profile", workflowstate.ErrAtomicGuard)
	}

	identity, err := readers.Provider.ReadWorkflowPublisherIdentity(ctx, pr)
	if err != nil || identity != policy.APIIdentityFingerprint {
		return workflowstate.PublicationCapability{}, fmt.Errorf("%w: publisher identity changed", workflowstate.ErrConflict)
	}

	host, err := json.Marshal(policy)
	if err != nil {
		return workflowstate.PublicationCapability{}, err
	}

	return workflowstate.PublicationCapability{Repository: pr.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: nativePublicationMechanism, Revision: contenthash.Sum(append(host, observation.Data...)), AtomicSourceBase: true, ExactResult: true}, nil
}
