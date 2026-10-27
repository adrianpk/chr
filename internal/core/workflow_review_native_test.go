package core

import (
	"context"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeReviewProviderFixture struct {
	*nativePreparationProviderFixture
	change func(*WorkflowNativeCapability)
}

func (p *nativeReviewProviderFixture) WorkflowPreparedNativeCapability(ctx context.Context, pr WorkflowPullRequestSnapshot, proof WorkflowIntegrationProof) (WorkflowNativeCapability, error) {
	capability, err := p.nativePreparationProviderFixture.WorkflowPreparedNativeCapability(ctx, pr, proof)
	if err != nil {
		return capability, err
	}

	p.change(&capability)

	return sealWorkflowNativeCapability(capability)
}

func nativeReviewFixture(t *testing.T, mode workflowstate.ReviewMode) (nativePreparationFixture, WorkflowIntegrationCandidate) {
	t.Helper()

	f := newNativePreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	writeWorkflowTestConfigValue(t, f.f.store.layout.ProjectDir(), "review", "mode", string(mode))

	candidate, err := f.f.store.PrepareWorkflowSliceIntegration(t.Context(), f.f.run.ID, f.provider, f.f.git)
	if err != nil {
		t.Fatal(err)
	}

	return f, candidate
}

// TestNativeReviewPolicy rejects a previously authorized result after its
// trusted publication policy changes, including after store reconstruction.
func TestNativeReviewPolicy(t *testing.T) {
	f, _ := nativeReviewFixture(t, workflowstate.ReviewTrustedAutoMerge)
	store, id := f.f.store, f.f.run.ID

	if _, err := store.ReviewWorkflowSlice(t.Context(), id, f.provider, nil); err != nil {
		t.Fatal(err)
	}

	f.provider.mu.Lock()
	f.provider.policyRevision = "policy-v2"
	f.provider.mu.Unlock()

	store = NewWorkflowStore(store.layout, store.projectID)

	if _, err := store.CheckWorkflowReviewGrant(t.Context(), id, f.provider); err == nil {
		t.Fatal("changed native policy reused publication approval")
	}
}

// TestNativeReviewSourceOnly requires host validation of P before any native
// review grant; source-task validation alone cannot authorize its rewrite.
func TestNativeReviewSourceOnly(t *testing.T) {
	f := newNativePreparationFixture(t, "rebase", workflowPreparationGate("test -f result.txt"))
	if _, err := f.f.store.ReviewWorkflowSlice(t.Context(), f.f.run.ID, f.provider, nil); err == nil {
		t.Fatal("native source-only review produced publication authority")
	}
}

// TestNativeReviewModes retains exact native policy, rewrite and history inputs
// for every review mode. Restart and queue readiness cannot transfer or repeat
// decisions, while publication rechecks the result-bound capability.
func TestNativeReviewModes(t *testing.T) {
	for _, mode := range []workflowstate.ReviewMode{workflowstate.ReviewHuman, workflowstate.ReviewTrustedAutoMerge, workflowstate.ReviewAgentic} {
		t.Run(string(mode), func(t *testing.T) { nativeReviewMode(t, mode) })
	}
}

func nativeReviewMode(t *testing.T, mode workflowstate.ReviewMode) {
	t.Helper()

	f, candidate := nativeReviewFixture(t, mode)
	store, id := f.f.store, f.f.run.ID
	binding := nativeReviewAuthorize(t, f)
	before := workflowTestRead(t, store, id)
	store = NewWorkflowStore(store.layout, store.projectID)

	nativeReviewRetained(t, store, id, f, candidate)

	again, err := store.ReviewWorkflowSlice(t.Context(), id, f.provider, []string{"false"})
	if err != nil || again != binding || workflowTestRead(t, store, id).Revision != before.Revision {
		t.Fatal("restart repeated a decision or host gate", err)
	}

	entry := nativeReviewEntry(t, store, id)
	nativeRoleTransition(t, store, entry, "ready_to_publish")

	guard, err := store.CheckWorkflowPublicationGuard(t.Context(), id, f.provider)
	if err != nil || guard.Git.ResultCommit != candidate.Native.PublishedSource || guard.ReviewHash != binding.InputHash {
		t.Fatal("queue readiness invalidated exact authority", err)
	}

	assertWorkflowPreparationSource(t, f.f)
}

func nativeReviewRetained(t *testing.T, store *WorkflowStore, id string, f nativePreparationFixture, candidate WorkflowIntegrationCandidate) {
	t.Helper()

	input, err := store.CheckWorkflowReviewGrant(t.Context(), id, f.provider)
	if err != nil || input.Native == nil || !workflowNativeJSONEqual(input.Native.Candidate, candidate.Native) || !workflowNativeJSONEqual(input.Native.Capability, candidate.Input.Native.Capability) || input.Native.ClaimGeneration != f.entry.ClaimGeneration {
		t.Fatal("review lost exact native evidence", err)
	}
}

func nativeReviewAuthorize(t *testing.T, f nativePreparationFixture) WorkflowReviewBinding {
	t.Helper()

	binding, err := f.f.store.ReviewWorkflowSlice(t.Context(), f.f.run.ID, f.provider, workflowReviewWorker("pass", "", ""))
	if err != nil {
		t.Fatal(err)
	}

	if binding.State == "waiting_approval" {
		binding, err = f.f.store.ApproveWorkflowSlice(t.Context(), f.f.run.ID, binding.InputHash, f.provider)
	}

	if err != nil || binding.State != "authorized" {
		t.Fatal("native review was not authorized", err)
	}

	return binding
}

func nativeReviewEntry(t *testing.T, store *WorkflowStore, id string) WorkflowIntegrationEntry {
	t.Helper()

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	return entry
}
