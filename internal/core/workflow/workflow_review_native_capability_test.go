package workflow

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeReviewCapabilities re-observes ownership, exact refs and all trusted
// capability fingerprints at approval, grant, review and publication entrypoints.
// Old immutable decisions remain readable without approving the changed image.
func TestNativeReviewCapabilities(t *testing.T) {
	cases := []struct {
		name   string
		change func(*NativeCapability)
	}{
		{"policy", func(c *NativeCapability) { c.PolicyFingerprint = contenthash.Sum([]byte("changed")) }},
		{"ownership", func(c *NativeCapability) { c.OwnershipRevision = "changed/v2" }},
		{"server", func(c *NativeCapability) { c.ServerCapabilityFingerprint = contenthash.Sum([]byte("changed")) }},
		{"publisher", func(c *NativeCapability) {
			c.CredentialIdentityFingerprint = contenthash.Sum([]byte("changed"))
		}},
		{"observation", func(c *NativeCapability) { c.ObservationContract = "github_same_pr_git_merge/v1" }},
		{"adapter", func(c *NativeCapability) { c.AdapterVersion = "unsupported/v2" }},
		{"source ref", func(c *NativeCapability) { c.SourceRef += "-changed" }},
		{"source", func(c *NativeCapability) { c.OriginalSourceCommit = c.ExpectedBaseCommit }},
		{"base", func(c *NativeCapability) { c.ExpectedBaseCommit = c.OriginalSourceCommit }},
		{"result", func(c *NativeCapability) { c.PreparedResultCommit = c.OriginalSourceCommit }},
		{"signing", func(c *NativeCapability) {
			c.RequiredSignatures, c.SigningProfile = true, contenthash.Sum([]byte("signer"))
		}},
		{"atomic", func(c *NativeCapability) { c.AtomicSourceBase = false }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f, _ := nativeReviewFixture(t, workflowstate.ReviewHuman)
			binding := nativeReviewAuthorize(t, f)
			provider := &nativeReviewProviderFixture{f.provider, test.change}
			nativeReviewRejectEntryPoints(t, f, binding, provider)
		})
	}
}

func nativeReviewRejectEntryPoints(t *testing.T, f nativePreparationFixture, binding ReviewBinding, provider PullRequestProvider) {
	t.Helper()

	store, id := f.f.store, f.f.run.ID

	before := workflowTestRead(t, store, id)
	if _, err := store.ApproveWorkflowSlice(t.Context(), id, binding.InputHash, provider); err == nil {
		t.Fatal("changed native inputs reused human approval")
	}

	if _, err := store.CheckWorkflowReviewGrant(t.Context(), id, provider); err == nil {
		t.Fatal("changed native inputs reused a grant")
	}

	if _, err := store.ReviewWorkflowSlice(t.Context(), id, provider, nil); err == nil {
		t.Fatal("changed native inputs reused review")
	}

	guarded, ok := provider.(PublicationGuardProvider)
	if !ok {
		guarded = f.provider
	}

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, guarded); err == nil {
		t.Fatal("changed native inputs authorized publication")
	}

	after := workflowTestRead(t, store, id)
	if after.Revision != before.Revision || after.Review.ResultHash != binding.ResultHash {
		t.Fatal("rejected observation overwrote historical decisions")
	}
}
