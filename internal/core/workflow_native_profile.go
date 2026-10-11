package core

import (
	"context"
	"encoding/json"
	"slices"
)

const (
	// WorkflowPublicationDisabled preserves the existing publication policy.
	WorkflowPublicationDisabled = "disabled"
	// WorkflowPublicationGitAtomicDualRef explicitly permits the owned source rewrite.
	WorkflowPublicationGitAtomicDualRef = "git_atomic_dual_ref"
	workflowNativeAdapterVersion        = "chr-native-publication/v1"
)

// WorkIntegrationConfig selects a publication mechanism independently of review.
type WorkIntegrationConfig struct {
	PublicationProfile string
}

// WorkflowNativePolicy contains current host-owned grants and certification.
// Workers cannot supply it. Ownership names one full ref, never a prefix grant.
type WorkflowNativePolicy struct {
	Repository                    string             `json:"repository"`
	SourceRef                     string             `json:"source_ref"`
	TargetRef                     string             `json:"target_ref"`
	OwnershipRevision             string             `json:"ownership_revision"`
	SourceRewriteAuthorized       bool               `json:"source_rewrite_authorized"`
	FastForwardTargetAuthorized   bool               `json:"fast_forward_target_authorized"`
	HostCandidateReview           bool               `json:"host_candidate_review"`
	ReviewMode                    ReviewWorkflowMode `json:"review_mode"`
	CredentialIdentityFingerprint string             `json:"credential_identity_fingerprint"`
	ObservationContract           string             `json:"observation_contract"`
	CertifiedMethods              []string           `json:"certified_methods"`
}

// WorkflowNativePolicyReader rechecks trusted ownership, credential roles and
// provider observation certification before each admission or intent check.
type WorkflowNativePolicyReader interface {
	ReadWorkflowNativePolicy(context.Context, WorkflowPullRequestSnapshot, string) (WorkflowNativePolicy, error)
}

// WorkflowNativeUnsupported identifies a condition and the required decision.
type WorkflowNativeUnsupported struct {
	Condition string `json:"condition"`
	Decision  string `json:"decision"`
}

// WorkflowNativeCapability binds current policy and transport to the original PR.
type WorkflowNativeCapability struct {
	WorkflowPublicationCapability
	SourceRef                     string                      `json:"source_ref"`
	TargetRef                     string                      `json:"target_ref"`
	OwnershipRevision             string                      `json:"ownership_revision"`
	PolicyFingerprint             string                      `json:"policy_fingerprint"`
	ServerCapabilityFingerprint   string                      `json:"server_capability_fingerprint"`
	CredentialIdentityFingerprint string                      `json:"credential_identity_fingerprint"`
	ObservationContract           string                      `json:"observation_contract"`
	AdapterVersion                string                      `json:"adapter_version"`
	OriginalSourceCommit          string                      `json:"original_source_commit"`
	ExpectedBaseCommit            string                      `json:"expected_base_commit"`
	PreparedResultCommit          string                      `json:"prepared_result_commit"`
	Unsupported                   []WorkflowNativeUnsupported `json:"unsupported,omitempty"`
}

// WorkflowNativeCapabilityRequest supplies the explicit profile and exact source.
type WorkflowNativeCapabilityRequest struct {
	Config      WorkIntegrationConfig
	PullRequest WorkflowPullRequestSnapshot
	Publication WorkflowGitPublication
}

func workflowNativePolicyReasons(request WorkflowNativeCapabilityRequest, policy WorkflowNativePolicy) []WorkflowNativeUnsupported {
	publication := request.Publication
	checks := []struct {
		valid     bool
		condition string
		decision  string
	}{
		{request.Config.PublicationProfile == WorkflowPublicationGitAtomicDualRef, "profile_disabled", "Explicitly select git_atomic_dual_ref."},
		{publication.Repository == publication.SourceRepository && publication.Repository == request.PullRequest.HeadRepository, "fork_source", "Select a source owned in the canonical repository."},
		{policy.Repository == publication.Repository && policy.SourceRef == publication.SourceRef && policy.TargetRef == publication.TargetRef, "ownership_binding", "Resolve current grants for these exact full refs."},
		{policy.SourceRewriteAuthorized && publication.SourceRewriteAuthorized && policy.OwnershipRevision != "" && policy.OwnershipRevision == publication.OwnershipRevision, "source_rewrite_unauthorized", "Authorize rewriting the exact owned source ref."},
		{policy.FastForwardTargetAuthorized, "target_push_unauthorized", "Provide an authorized fast-forward target publication role."},
		{policy.HostCandidateReview, "native_approval_required", "Select a policy permitting host-bound review of the prepared result."},
		{workflowConfigChoice(string(policy.ReviewMode), []string{string(ReviewWorkflowModeHumanReview), string(ReviewWorkflowModeAgenticReview), string(ReviewWorkflowModeTrustedAutoMerge)}), "review_policy_unknown", "Resolve the current candidate review policy."},
		{validWorkflowNativeFingerprint(policy.CredentialIdentityFingerprint), "credential_identity_unknown", "Resolve the current publisher credential identity without secrets."},
		{slices.Contains(policy.CertifiedMethods, publication.Git.Method), "method_uncertified", "Certify the configured method for this observation profile."},
	}

	var reasons []WorkflowNativeUnsupported

	for _, check := range checks {
		if !check.valid {
			reasons = append(reasons, WorkflowNativeUnsupported{check.condition, check.decision})
		}
	}

	return reasons
}

func validWorkflowNativeFingerprint(value string) bool {
	return len(value) == 64 && workflowCommitID(value)
}

func sealWorkflowNativeCapability(capability WorkflowNativeCapability) (WorkflowNativeCapability, error) {
	capability.Revision = ""

	data, err := json.Marshal(capability)
	if err != nil {
		return capability, err
	}

	capability.Revision = workflowHash(data)

	return capability, nil
}
