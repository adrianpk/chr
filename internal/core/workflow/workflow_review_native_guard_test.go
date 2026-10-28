package workflow

import (
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestNativeReviewInputs invalidates approval for actual candidate, history,
// contract and queue changes. No rejected check rewrites the retained grant.
func TestNativeReviewInputs(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, nativePreparationFixture, IntegrationCandidate)
	}{
		{"target", func(_ *testing.T, f nativePreparationFixture, _ IntegrationCandidate) {
			f.provider.pr.BaseCommit = contenthash.Sum([]byte("moved"))[:40]
			f.provider.target = f.provider.pr.BaseCommit
		}},
		{"source", func(_ *testing.T, f nativePreparationFixture, _ IntegrationCandidate) {
			f.provider.pr.HeadCommit = contenthash.Sum([]byte("moved"))[:40]
		}},
		{"profile", func(t *testing.T, f nativePreparationFixture, _ IntegrationCandidate) {
			writeWorkflowTestConfigValue(t, f.f.store.layout.ProjectDir(), "work", "integration.publication_profile", workspaceconfig.WorkflowPublicationDisabled)
		}},
		{"method", func(t *testing.T, f nativePreparationFixture, _ IntegrationCandidate) {
			writeWorkflowTestConfigValue(t, f.f.store.layout.ProjectDir(), "work", "merge_strategy", "squash")
		}},
		{"contract", func(t *testing.T, f nativePreparationFixture, c IntegrationCandidate) {
			workflowReviewWrite(t, filepath.Join(f.f.store.layout.Root(), c.Input.Source.Slice.ContractRef), []byte("{}"))
		}},
		{"history", func(t *testing.T, f nativePreparationFixture, c IntegrationCandidate) {
			workflowReviewWrite(t, filepath.Join(f.f.store.layout.Root(), c.Native.HistoryManifestRef.Path), []byte("{}"))
		}},
		{"result", func(t *testing.T, _ nativePreparationFixture, c IntegrationCandidate) {
			workflowIntegrationCommit(t, c.Lane.Worktree, "result.txt", "Changed candidate")
		}},
		{"generation", func(t *testing.T, f nativePreparationFixture, _ IntegrationCandidate) {
			entry := nativeReviewEntry(t, f.f.store, f.f.run.ID)
			for _, state := range []string{"queued", "preparing", "validating", "reviewing"} {
				entry = nativeRoleTransition(t, f.f.store, entry, state)
			}
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f, candidate := nativeReviewFixture(t, workflowstate.ReviewHuman)
			binding := nativeReviewAuthorize(t, f)
			test.change(t, f, candidate)
			nativeReviewRejectEntryPoints(t, f, binding, f.provider)
		})
	}
}

// TestNativeReviewObserver requires the full trusted port even when an otherwise
// valid candidate and current human wait already exist. Source-only observation
// cannot approve the exact rewrite, and refusal consumes no review attempt.
func TestNativeReviewObserver(t *testing.T) {
	f, _ := nativeReviewFixture(t, workflowstate.ReviewHuman)
	store, id := f.f.store, f.f.run.ID

	binding, err := store.ReviewWorkflowSlice(t.Context(), id, f.provider, nil)
	if err != nil {
		t.Fatal(err)
	}

	before := workflowTestRead(t, store, id)
	if _, err := store.ApproveWorkflowSlice(t.Context(), id, binding.InputHash, f.provider.workflowPreparationProviderFixture); err == nil {
		t.Fatal("source-only observer approved native publication")
	}

	if after := workflowTestRead(t, store, id); after.Revision != before.Revision || after.ReviewCount != before.ReviewCount {
		t.Fatal("refused approval consumed a review attempt")
	}
}

// TestNativeReviewPolicyIO changes local review or publication policy inside
// trusted capability IO. The final host reread must refuse the old policy image
// before persisting a wait, decision or role claim.
func TestNativeReviewPolicyIO(t *testing.T) {
	for _, change := range []string{"profile", "review"} {
		t.Run(change, func(t *testing.T) {
			f, _ := nativeReviewFixture(t, workflowstate.ReviewTrustedAutoMerge)
			store, id := f.f.store, f.f.run.ID
			before := workflowTestRead(t, store, id)
			provider := &nativeReviewProviderFixture{f.provider, func(_ *NativeCapability) {
				if change == "profile" {
					writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "work", "integration.publication_profile", workspaceconfig.WorkflowPublicationDisabled)
				} else {
					writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "review", "mode", string(workflowstate.ReviewHuman))
				}
			}}

			if _, err := store.ReviewWorkflowSlice(t.Context(), id, provider, nil); err == nil {
				t.Fatal("capability IO used stale local policy")
			}

			if after := workflowTestRead(t, store, id); after.Revision != before.Revision || after.Review != nil {
				t.Fatal("stale policy IO persisted review authority")
			}
		})
	}
}
