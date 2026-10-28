package workflow

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestNativeParallelDrift admits two independent queues on a physical target.
// An external writer integrates the first result; only a fresh gate and review
// of the second combined tree can regain publication authority.
func TestNativeParallelDrift(t *testing.T) {
	lanes := workflowParallelPublicationFixture(t, false)
	first, one := nativeDriftParallelLane(t, lanes[0])
	second, old := nativeDriftParallelLane(t, lanes[1])
	store, id := second.f.store, second.f.run.ID
	provider := nativeDriftProviderFixture{second.provider, second.f.remote}
	before := workflowTestRead(t, store, id)

	workflowIntegrationCommand(t, one.Lane.Worktree, "push", first.f.remote, one.Input.Git.ResultCommit+":refs/heads/dev")

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, provider); err == nil {
		t.Fatal("external queue publication retained a stale grant")
	}

	entry := nativeReviewEntry(t, store, id)
	nativeRoleTransition(t, store, entry, "ready_to_publish")

	if _, dispatch, err := store.BeginWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, store, id), provider); err == nil || dispatch {
		t.Fatal("changed target reached native dispatch", err)
	}

	fresh, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, provider, second.f.git)
	if err != nil || fresh.Input.Git.BaseCommit != one.Input.Git.ResultCommit || fresh.Validation.Validation[0].Stdout != "combined" || fresh.Input.Native.ClaimGeneration != old.Input.Native.ClaimGeneration+1 {
		t.Fatal("parallel target was not revalidated as a combined result", err)
	}

	nativeDriftParallelReview(t, store, id, provider, fresh)

	second.f.base = one.Input.Git.ResultCommit
	second.provider.pr.BaseCommit, second.provider.target = second.f.base, second.f.base
	nativeDriftRetained(t, second, before, fresh)
}

func nativeDriftParallelReview(t *testing.T, store *Store, id string, provider nativeDriftProviderFixture, fresh workflowstate.IntegrationCandidate) {
	t.Helper()

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, provider); err == nil {
		t.Fatal("fresh parallel gate inherited old review")
	}

	if _, err := store.ReviewWorkflowSlice(t.Context(), id, provider, nil); err != nil {
		t.Fatal(err)
	}

	if guard, err := store.CheckWorkflowPublicationGuard(t.Context(), id, provider); err != nil || guard.Git.ResultCommit != fresh.Input.Git.ResultCommit {
		t.Fatal("fresh parallel review did not authorize exact P", err)
	}
}

func nativeDriftParallelLane(t *testing.T, lane workflowParallelPublicationLane) (nativePreparationFixture, workflowstate.IntegrationCandidate) {
	t.Helper()

	f := lane.fixture
	writeWorkflowTestConfigValue(t, f.store.layout.ProjectDir(), "work", "integration.publication_profile", workspaceconfig.WorkflowPublicationGitAtomicDualRef)

	entry := nativeRoleAdmit(t, f.store, f.run.ID)
	entry = nativeRoleTransition(t, f.store, entry, "preparing")
	entry = nativeRoleTransition(t, f.store, entry, "validating")
	provider := &nativePreparationProviderFixture{f.provider, "policy-v1"}
	observed := nativeDriftProviderFixture{provider, f.remote}

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	run.Preparation.State, run.Review.State = "stale", "stale"
	if err := f.store.commitWorkflowState(&run); err != nil {
		t.Fatal(err)
	}

	candidate, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, observed, f.git)
	if err != nil {
		t.Fatal("native parallel preparation", err)
	}

	if _, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, observed, nil); err != nil {
		t.Fatal(err)
	}

	return nativePreparationFixture{f, provider, entry}, candidate
}
