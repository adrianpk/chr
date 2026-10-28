package workflow

import (
	"errors"
	"strconv"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeRepairedTargetDrift cannot import a previous generation's accepted
// semantic repair into a moved base. The unchanged gate fails again and retains
// fresh bounded findings while earlier merger consumption remains charged.
func TestNativeRepairedTargetDrift(t *testing.T) {
	f, request, harness := nativeCandidateRepairFixture(t, "rebase", "combined_gate_regression")
	store, id := f.f.store, f.f.run.ID

	attempt, err := ExecuteMergerRepair(t.Context(), store, request, harness, MergerRepairPorts{Provider: f.provider, Git: f.f.git})
	if err != nil || attempt.Acceptance == nil {
		t.Fatal("repair was not accepted", err)
	}

	first, err := PrepareIntegration(t.Context(), store, id, f.provider, f.f.git)
	if err != nil {
		t.Fatal(err)
	}

	nativeReviewAuthorize(t, f)
	nativeDriftMoveTarget(t, &f, "New semantic target")

	store = NewWorkflowStore(store.layout, store.projectID)

	fresh, err := PrepareIntegration(t.Context(), store, id, f.provider, f.f.git)
	if err == nil || fresh.Input.Git.BaseCommit != f.f.base || fresh.Input.Native.ClaimGeneration != first.Input.Native.ClaimGeneration+1 || len(fresh.Input.MergerAttemptRefs) != 0 {
		t.Fatal("replacement reused old semantic repair or missed the unchanged gate", err)
	}

	nativeRepairedDriftFacts(t, f, store, fresh)
}

func nativeRepairedDriftFacts(t *testing.T, f nativePreparationFixture, store *Store, fresh workflowstate.IntegrationCandidate) {
	t.Helper()

	id := f.f.run.ID
	if _, err := store.CheckWorkflowPublicationGuard(t.Context(), id, f.provider); err == nil {
		t.Fatal("failed new-base gate retained publication authority")
	}

	run := workflowTestRead(t, store, id)
	if run.CandidateRepair == nil || run.CandidateRepair.ClaimGeneration != fresh.Input.Native.ClaimGeneration {
		t.Fatal("new-base failure lost fresh bounded findings")
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || budget.UnitCounters[0].Merger != 1 || budget.UnitCounters[0].Drift != 1 {
		t.Fatal("replacement reset accepted merger consumption", err)
	}

	f.f.store = store
	assertWorkflowPreparationSource(t, f.f)
}

// TestNativeDriftLimit retains the invocation-wide bound through reconstruction.
// Exhaustion cannot launch a sixth replacement or refill earlier consumption.
func TestNativeDriftLimit(t *testing.T) {
	f, _ := nativeReviewFixture(t, workflowstate.ReviewTrustedAutoMerge)

	store, id := f.f.store, f.f.run.ID
	for round := range 5 {
		nativeDriftMoveTarget(t, &f, "Bounded replacement "+strconv.Itoa(round+1))

		store = NewWorkflowStore(store.layout, store.projectID)
		if _, err := PrepareIntegration(t.Context(), store, id, f.provider, f.f.git); err != nil {
			t.Fatal("admitted drift failed", err)
		}
	}

	before := workflowTestRead(t, store, id)
	nativeDriftMoveTarget(t, &f, "Exhausted replacement")

	if _, err := PrepareIntegration(t.Context(), store, id, f.provider, f.f.git); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("exhausted invocation admitted another replacement", err)
	}

	if after := workflowTestRead(t, store, id); after.Revision != before.Revision || after.PreparationCount != before.PreparationCount {
		t.Fatal("rejected drift changed its candidate or execution count")
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || budget.UnitCounters[0].Drift != 5 {
		t.Fatal("restart reset the drift ceiling", err)
	}
}
