package forge

import (
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SameNativeSourcePR compares the retained PR source while ignoring speculative merge metadata.
func SameNativeSourcePR(current, retained workflowstate.PullRequestSnapshot) bool {
	// GitHub computes a speculative merge_commit_sha while an open PR is still
	// unmerged. It is not a source/base change or evidence of delivered P.
	current.ResultCommit, retained.ResultCommit = "", ""

	return current == retained
}

// NewNativeCapability starts an observation bound to the exact source, base and prepared result.
func NewNativeCapability(request NativeCapabilityInput) workflowstate.NativeCapability {
	return workflowstate.NativeCapability{
		PublicationCapability: workflowstate.PublicationCapability{Repository: request.Publication.Repository, PullRequestURL: request.PullRequest.URL, Method: request.Publication.Git.Method, Mechanism: nativePublicationMechanism},
		SourceRef:             request.Publication.SourceRef, TargetRef: request.Publication.TargetRef, AdapterVersion: NativeAdapterVersion,
		OriginalSourceCommit: request.Publication.Git.SourceCommit, ExpectedBaseCommit: request.Publication.Git.BaseCommit, PreparedResultCommit: request.Publication.Git.ResultCommit,
	}
}

// NativeSourceBinding checks the exact open canonical PR against the immutable publication identity.
func NativeSourceBinding(request NativeCapabilityInput) bool {
	pr, publication := request.PullRequest, request.Publication

	return pr.Repository == publication.Repository && pr.HeadRepository == publication.SourceRepository &&
		"refs/heads/"+pr.HeadBranch == publication.SourceRef && "refs/heads/"+pr.BaseBranch == publication.TargetRef &&
		pr.State == "open" && !pr.Merged && pr.HeadCommit == publication.Git.SourceCommit && pr.BaseCommit == publication.Git.BaseCommit
}

// ChangedNativeCapability identifies the first changed policy, transport, publisher or ownership fact.
func ChangedNativeCapability(prior, current workflowstate.NativeCapability) string {
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

const (
	// NativeAdapterVersion identifies the retained native capability protocol.
	NativeAdapterVersion = "chr-native-publication/v1"
)

// NativePolicyReasons retains every denial from current host grants for these exact refs.
func NativePolicyReasons(request NativeCapabilityInput, policy workflowstate.NativePolicy) []workflowstate.NativeUnsupported {
	publication := request.Publication
	checks := []struct {
		valid     bool
		condition string
		decision  string
	}{
		{request.Profile == nativePublicationMechanism, "profile_disabled", "Explicitly select git_atomic_dual_ref."},
		{publication.Repository == publication.SourceRepository && publication.Repository == request.PullRequest.HeadRepository, "fork_source", "Select a source owned in the canonical repository."},
		{policy.Repository == publication.Repository && policy.SourceRef == publication.SourceRef && policy.TargetRef == publication.TargetRef, "ownership_binding", "Resolve current grants for these exact full refs."},
		{policy.SourceRewriteAuthorized && publication.SourceRewriteAuthorized && policy.OwnershipRevision != "" && policy.OwnershipRevision == publication.OwnershipRevision, "source_rewrite_unauthorized", "Authorize rewriting the exact owned source ref."},
		{policy.FastForwardTargetAuthorized, "target_push_unauthorized", "Provide an authorized fast-forward target publication role."},
		{policy.HostCandidateReview, "native_approval_required", "Select a policy permitting host-bound review of the prepared result."},
		{slices.Contains([]string{string(workflowstate.ReviewHuman), string(workflowstate.ReviewAgentic), string(workflowstate.ReviewTrustedAutoMerge)}, string(policy.ReviewMode)), "review_policy_unknown", "Resolve the current candidate review policy."},
		{workflowstate.ValidFingerprint(policy.CredentialIdentityFingerprint), "credential_identity_unknown", "Resolve the current publisher credential identity without secrets."},
		{slices.Contains(policy.CertifiedMethods, publication.Git.Method), "method_uncertified", "Certify the configured method for this observation profile."},
	}

	var reasons []workflowstate.NativeUnsupported

	for _, check := range checks {
		if !check.valid {
			reasons = append(reasons, workflowstate.NativeUnsupported{Condition: check.condition, Decision: check.decision})
		}
	}

	return reasons
}

// SealNativeCapability hashes the complete capability after clearing its retained revision.
func SealNativeCapability(capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	capability.Revision = ""

	data, err := json.Marshal(capability)
	if err != nil {
		return capability, err
	}

	capability.Revision = contenthash.Sum(data)

	return capability, nil
}

const nativePublicationMechanism = "git_atomic_dual_ref"
