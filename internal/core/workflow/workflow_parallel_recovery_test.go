package workflow

import (
	"errors"
	"testing"
)

// TestWorkflowParallelExpiredWorker keeps an expired actual process counted,
// rejects its result, then recovers it while an unrelated process remains live.
func TestWorkflowParallelExpiredWorker(t *testing.T) {
	stores, runs := workflowParallelRunnerLanes(t, 2)
	first := workflowParallelRunnerFixture(t, stores[0], runs[0].ID, "worker")
	waitWorkflowClaimFile(t, first.spec.Started)
	expireWorkflowRecoveryLease(t, stores[0], runs[0].ID)

	second := workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "worker")
	waitWorkflowClaimFile(t, second.spec.Started)

	assertWorkflowLiveRecoveryBlocked(t, stores[0], runs[0].ID)

	if _, err := stores[2].ClaimWorkflow(t.Context(), runs[2].ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("expired live group freed its counted slot", err)
	}

	if result := finishWorkflowParallelRunner(t, first); result.Successful || result.Failure == "" {
		t.Fatal("expired worker applied its result", result)
	}

	recovered, err := stores[0].ReconcileWorkflow(t.Context(), runs[0].ID)
	if err != nil || recovered.Status != "paused" || recovered.Attempts[0].Outcome != "recovered" || recovered.Cursor != 0 {
		t.Fatal("unrelated live worker blocked quiescent recovery", recovered, err)
	}

	if result := finishWorkflowParallelRunner(t, second); !result.Successful {
		t.Fatal("recovery stopped unrelated work", result)
	}
}

func assertWorkflowLiveRecoveryBlocked(t *testing.T, store *Store, id string) {
	t.Helper()

	if _, err := store.ReconcileWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowOwned) && !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("expired live group was recovered", err)
	}
}

// TestWorkflowParallelReplacement fences old credentials and closes only their
// own descriptors. Repeated old cleanup cannot release a newer slot occupant.
func TestWorkflowParallelReplacement(t *testing.T) {
	stores, runs := workflowParallelRunnerLanes(t, 2)
	old := workflowTestClaim(t, stores[0], runs[0].ID)
	second := workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "worker")
	waitWorkflowClaimFile(t, second.spec.Started)
	expireWorkflowRecoveryLease(t, stores[0], runs[0].ID)

	if _, err := stores[0].ReconcileWorkflow(t.Context(), runs[0].ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("still-open old descriptor permitted recovery", err)
	}

	if err := old.releaseReservation(); err != nil {
		t.Fatal(err)
	}

	if _, err := stores[0].ReconcileWorkflow(t.Context(), runs[0].ID); err != nil {
		t.Fatal("quiescent manual owner could not be fenced", err)
	}

	replacement := workflowTestClaim(t, stores[0], runs[0].ID)
	if err := old.releaseReservation(); err != nil {
		t.Fatal(err)
	}

	if err := stores[0].CheckpointWorkflow(t.Context(), old); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("old owner changed the replacement checkpoint", err)
	}

	assertWorkflowReplacementCounted(t, stores[2], runs[2].ID, stores[0], replacement)

	if err := stores[0].CheckpointWorkflow(t.Context(), replacement); err != nil {
		t.Fatal(err)
	}

	if result := finishWorkflowParallelRunner(t, second); !result.Successful {
		t.Fatal("replacement disturbed another lane", result)
	}
}

func assertWorkflowReplacementCounted(t *testing.T, third *Store, id string, owner *Store, claim *Claim) {
	t.Helper()

	if _, err := third.ClaimWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("old cleanup released the newer reservation", err)
	}

	if err := owner.CheckWorkflowClaim(t.Context(), claim); err != nil {
		t.Fatal("old checkpoint replaced newer credentials", err)
	}

	run := workflowTestRead(t, owner, claim.InvocationID)
	if len(run.Attempts) != 2 || run.Attempts[1].Replaces != run.Attempts[0].ID || run.Attempts[1].Generation <= run.Attempts[0].Generation {
		t.Fatal("replacement reset history or generation", run)
	}
}

// TestWorkflowParallelClosingRecovery preserves interrupted closure history and
// its budget while an independent native reservation remains occupied.
func TestWorkflowParallelClosingRecovery(t *testing.T) {
	fixture, set, run := workflowClosingEmptyOwner(t)
	writeWorkspaceConfig(t, fixture.store.layout.ProjectDir(), "work:\n  max_active_workers: 2\n")

	claim := workflowClosingBeginTest(t, fixture, run, workflowFrontierPorts(fixture))
	slot, guard := workflowTestSlot(t, fixture.store)
	expireWorkflowClosingTest(t, fixture.store, claim)

	if _, err := fixture.store.ReconcileWorkflow(t.Context(), run.ID); err != nil {
		t.Fatal("another native reservation blocked closing recovery", err)
	}

	checkWorkflowClosingRecovered(t, fixture.store, set.Path, workflowTestRead(t, fixture.store, run.ID), 1)

	if err := errors.Join(slot.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}
}
