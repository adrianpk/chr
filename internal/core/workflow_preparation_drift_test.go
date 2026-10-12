package core

import "testing"

// TestNativeTargetDrift rebuilds each Git method against an external target
// writer, retaining source task proofs while rejecting the old review grant.
func TestNativeTargetDrift(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) { nativeTargetDrift(t, method) })
	}
}

func nativeTargetDrift(t *testing.T, method string) {
	t.Helper()

	f := newNativePreparationFixture(t, method, workflowPreparationGate("test -f result.txt && if test -f target.txt; then cat target.txt; fi"))
	store, id := f.f.store, f.f.run.ID

	first, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, f.provider, f.f.git)
	if err != nil {
		t.Fatal(err)
	}

	old := nativeReviewAuthorize(t, f)
	before := workflowTestRead(t, store, id)
	nativeDriftMoveTarget(t, &f, "Replacement target content")

	store = NewWorkflowStore(store.layout, store.projectID)
	f.f.store = store

	fresh, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, f.provider, f.f.git)
	if err != nil {
		t.Fatal("changed target was not revalidated", err)
	}

	if fresh.Input.Git.BaseCommit != f.f.base || fresh.Input.Git.ResultCommit == first.Input.Git.ResultCommit || fresh.Validation.AttemptID == first.Validation.AttemptID || fresh.Input.Native.ClaimGeneration != first.Input.Native.ClaimGeneration+1 || fresh.Validation.Validation[0].Stdout != "Replacement target content\n" {
		t.Fatalf("replacement reused old evidence: base=%s result=%s generation=%d stdout=%q", fresh.Input.Git.BaseCommit, fresh.Input.Git.ResultCommit, fresh.Input.Native.ClaimGeneration, fresh.Validation.Validation[0].Stdout)
	}

	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, f.provider); err == nil {
		t.Fatal("replacement inherited the old review")
	}

	grant := nativeReviewAuthorize(t, f)
	if grant.InputHash == old.InputHash {
		t.Fatal("replacement did not receive an exact fresh review")
	}

	nativeDriftRetained(t, f, before, fresh)
}

func nativeDriftMoveTarget(t *testing.T, f *nativePreparationFixture, content string) {
	t.Helper()

	base := workflowIntegrationCommit(t, f.f.actor, "target.txt", content)
	workflowIntegrationCommand(t, f.f.actor, "push", "origin", "dev")

	f.f.base, f.provider.pr.BaseCommit, f.provider.target = base, base, base
}

func nativeDriftRetained(t *testing.T, f nativePreparationFixture, before WorkflowInvocation, fresh WorkflowIntegrationCandidate) {
	t.Helper()

	store, id := f.f.store, f.f.run.ID
	assertWorkflowPreparationSource(t, f.f)

	after := workflowTestRead(t, store, id)
	if after.Cursor != before.Cursor || after.CandidateHash != before.CandidateHash || after.PreparationCount != before.PreparationCount+1 {
		t.Fatal("drift reran source work or lost its preparation count")
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || budget.UnitCounters[0].Drift != 1 {
		t.Fatal("replacement base was not charged once", err)
	}

	again, err := store.PrepareWorkflowSliceIntegration(t.Context(), id, f.provider, f.f.git)
	if err != nil || again.Validation.AttemptID != fresh.Validation.AttemptID {
		t.Fatal("unchanged replacement repeated its gate", err)
	}

	retained, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || !workflowNativeJSONEqual(budget, retained) {
		t.Fatal("retry changed retained drift accounting", err)
	}
}
