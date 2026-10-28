package workflow

import (
	"context"
	"encoding/json"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

const (
	workflowNativeAdapterVersion = "chr-native-publication/v1"
)

// NativePolicy contains current host-owned grants and certification.
// Workers cannot supply it. Ownership names one full ref, never a prefix grant.
type NativePolicy struct {
	Repository                    string                   `json:"repository"`
	SourceRef                     string                   `json:"source_ref"`
	TargetRef                     string                   `json:"target_ref"`
	OwnershipRevision             string                   `json:"ownership_revision"`
	SourceRewriteAuthorized       bool                     `json:"source_rewrite_authorized"`
	FastForwardTargetAuthorized   bool                     `json:"fast_forward_target_authorized"`
	HostCandidateReview           bool                     `json:"host_candidate_review"`
	ReviewMode                    workflowstate.ReviewMode `json:"review_mode"`
	CredentialIdentityFingerprint string                   `json:"credential_identity_fingerprint"`
	APIIdentityFingerprint        string                   `json:"api_identity_fingerprint,omitempty"`
	ObservationContract           string                   `json:"observation_contract"`
	CertifiedMethods              []string                 `json:"certified_methods"`
}

// NativePolicyReader rechecks trusted ownership, credential roles and
// provider observation certification before each admission or intent check.
type NativePolicyReader interface {
	ReadWorkflowNativePolicy(context.Context, workflowstate.PullRequestSnapshot, string) (NativePolicy, error)
}

// NativeCapabilityRequest supplies the explicit profile and exact source.
type NativeCapabilityRequest struct {
	Config      workspaceconfig.WorkIntegrationConfig
	PullRequest workflowstate.PullRequestSnapshot
	Publication GitPublication
}

func workflowNativePolicyReasons(request NativeCapabilityRequest, policy NativePolicy) []workflowstate.NativeUnsupported {
	publication := request.Publication
	checks := []struct {
		valid     bool
		condition string
		decision  string
	}{
		{request.Config.PublicationProfile == workspaceconfig.WorkflowPublicationGitAtomicDualRef, "profile_disabled", "Explicitly select git_atomic_dual_ref."},
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

func sealWorkflowNativeCapability(capability workflowstate.NativeCapability) (workflowstate.NativeCapability, error) {
	capability.Revision = ""

	data, err := json.Marshal(capability)
	if err != nil {
		return capability, err
	}

	capability.Revision = contenthash.Sum(data)

	return capability, nil
}
