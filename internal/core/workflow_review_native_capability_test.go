package core

import "testing"

// TestNativeReviewCapabilities re-observes ownership, exact refs and all trusted
// capability fingerprints at approval, grant, review and publication entrypoints.
// Old immutable decisions remain readable without approving the changed image.
func TestNativeReviewCapabilities(t *testing.T) {
	cases := []struct {
		name   string
		change func(*WorkflowNativeCapability)
	}{
		{"policy", func(c *WorkflowNativeCapability) { c.PolicyFingerprint = workflowHash([]byte("changed")) }},
		{"ownership", func(c *WorkflowNativeCapability) { c.OwnershipRevision = "changed/v2" }},
		{"server", func(c *WorkflowNativeCapability) { c.ServerCapabilityFingerprint = workflowHash([]byte("changed")) }},
		{"publisher", func(c *WorkflowNativeCapability) { c.CredentialIdentityFingerprint = workflowHash([]byte("changed")) }},
		{"observation", func(c *WorkflowNativeCapability) { c.ObservationContract = "github_same_pr_git_merge/v1" }},
		{"adapter", func(c *WorkflowNativeCapability) { c.AdapterVersion = "unsupported/v2" }},
		{"source ref", func(c *WorkflowNativeCapability) { c.SourceRef += "-changed" }},
		{"source", func(c *WorkflowNativeCapability) { c.OriginalSourceCommit = c.ExpectedBaseCommit }},
		{"base", func(c *WorkflowNativeCapability) { c.ExpectedBaseCommit = c.OriginalSourceCommit }},
		{"result", func(c *WorkflowNativeCapability) { c.PreparedResultCommit = c.OriginalSourceCommit }},
		{"signing", func(c *WorkflowNativeCapability) {
			c.RequiredSignatures, c.SigningProfile = true, workflowHash([]byte("signer"))
		}},
		{"atomic", func(c *WorkflowNativeCapability) { c.AtomicSourceBase = false }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f, _ := nativeReviewFixture(t, ReviewWorkflowModeHumanReview)
			binding := nativeReviewAuthorize(t, f)
			provider := &nativeReviewProviderFixture{f.provider, test.change}
			nativeReviewRejectEntryPoints(t, f, binding, provider)
		})
	}
}

func nativeReviewRejectEntryPoints(t *testing.T, f nativePreparationFixture, binding WorkflowReviewBinding, provider WorkflowPullRequestProvider) {
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

	guarded, ok := provider.(WorkflowPublicationGuardProvider)
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
