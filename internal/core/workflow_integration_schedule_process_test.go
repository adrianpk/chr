package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

// TestNativeProcessCapacity launches admitted tasks through the native selector.
// File barriers prove live overlap, isolated checkouts and exclusion of a third
// process; reducing capacity retains live owners until they finish naturally.
func TestNativeProcessCapacity(t *testing.T) {
	for _, capacity := range []int{1, 2} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) { nativeProcessCapacity(t, capacity) })
	}
}

func nativeProcessCapacity(t *testing.T, capacity int) {
	t.Helper()

	stores, runs := workflowParallelRunnerLanes(t, capacity)
	for i, store := range stores {
		run := workflowTestRead(t, store, runs[i].ID)
		nativeEntryAdmit(t, store, WorkflowIntegrationAdmission{InvocationID: run.ID, ExpectedRevision: run.Revision, Unit: WorkflowRef{Kind: "ticket", ID: run.Members[0].Admission.Key}, AdmissionKey: run.RequestKey, TargetRef: "refs/heads/dev"})
	}

	if capacity == 1 {
		// An omitted worker limit exercises the default, rather than an explicit
		// value which could mask a regression in serial configuration.
		for _, store := range stores[:2] {
			writeWorkspaceConfig(t, store.layout.ProjectDir(), "{}\n")
		}
	}

	first := workflowParallelRunnerFixture(t, stores[0], runs[0].ID, "native")
	waitWorkflowClaimFile(t, first.spec.Started)

	second := workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "native")

	if capacity == 1 {
		if result := finishWorkflowParallelRunner(t, second); !result.Capacity {
			t.Fatal("default capacity launched another admitted worker", result)
		}
	} else {
		waitWorkflowClaimFile(t, second.spec.Started)
		checkNativeProcessOverlap(t, stores, runs)

		for _, store := range stores[:2] {
			writeWorkflowTestConfigValue(t, store.layout.ProjectDir(), "work", "max_active_workers", "1")
		}
	}

	checkNativeProcessWait(t, stores[2], runs[2].ID)
	nativeProcessFinish(t, first)

	if capacity == 1 {
		second = workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "native")
		waitWorkflowClaimFile(t, second.spec.Started)
	}

	// Even after one of two owners finishes, the reduced limit must count the
	// remaining live worker rather than reusing the released slot immediately.
	checkNativeProcessWait(t, stores[2], runs[2].ID)
	nativeProcessFinish(t, second)

	third := workflowParallelRunnerFixture(t, stores[2], runs[2].ID, "native")
	waitWorkflowClaimFile(t, third.spec.Started)
	nativeProcessFinish(t, third)

	for i, store := range stores {
		checkNativeProcessProof(t, store, runs[i].ID)
	}
}

func checkNativeProcessOverlap(t *testing.T, stores [3]*WorkflowStore, runs [3]WorkflowInvocation) {
	t.Helper()

	first := workflowTestRead(t, stores[0], runs[0].ID)

	second := workflowTestRead(t, stores[1], runs[1].ID)
	if first.Lane.Worktree == second.Lane.Worktree || len(first.Attempts) != 1 || len(second.Attempts) != 1 {
		t.Fatal("overlapping workers shared a checkout or lost ownership")
	}

	a, b := first.Attempts[0], second.Attempts[0]
	if a.CapacitySlot == b.CapacitySlot || a.OwnerProcess == nil || b.OwnerProcess == nil || *a.OwnerProcess == *b.OwnerProcess {
		t.Fatal("independent processes shared an owner or pool slot")
	}

	for i, store := range stores[:2] {
		assertWorkflowNativeLockAbsent(t, store)
		checkNativeProcessStatus(t, store, runs[i].ID)
	}
}

func checkNativeProcessStatus(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	before := nativeQueueImages(t, store.layout.Root())

	queue, err := store.QueryWorkflowIntegrationQueue(t.Context(), WorkflowIntegrationQueueQuery{InvocationID: id, Limit: 10})
	if err != nil || len(queue.Entries) != 1 {
		t.Fatal("live worker blocked read-only queue status", err)
	}

	if _, err := store.QueryWorkflowIntegrationSchedule(t.Context(), WorkflowIntegrationScheduleQuery{Limit: 10}); err != nil {
		t.Fatal("live worker blocked read-only scheduling", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("live status changed authoritative records")
	}
}

func checkNativeProcessWait(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	claim, err := store.ClaimNextWorkflowIntegration(t.Context(), WorkflowIntegrationScheduleQuery{InvocationID: id, Limit: 1})
	if claim != nil {
		t.Error(claim.Close())
	}

	if !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("native selector exceeded live process capacity", err)
	}

	assertWorkflowCapacityWait(t, store, id)

	legacy, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment})
	if legacy != nil {
		t.Error(legacy.Release())
	}

	if !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("legacy launch bypassed a live native worker", err)
	}
}

func nativeProcessFinish(t *testing.T, runner *workflowParallelRunner) {
	t.Helper()

	if result := finishWorkflowParallelRunner(t, runner); !result.Successful {
		t.Fatal("admitted worker did not commit and validate its source", result)
	}
}

func checkNativeProcessProof(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	run := workflowTestRead(t, store, id)
	if run.Cursor != 1 || run.ConfirmedCursor != 0 || run.Status != "waiting_integration" || len(run.Attempts) != 1 || run.Attempts[0].Outcome != "validated" {
		t.Fatal("worker completion lost source proof or fabricated delivery", run)
	}

	data, err := os.ReadFile(filepath.Join(run.Lane.Worktree, "result.txt"))
	if err != nil || string(data) != "result" || workflowSelectionGit(t, store, "rev-parse", run.Lane.Branch) != run.Lane.HeadCommit {
		t.Fatal("validated source is not the actual isolated Git result", err)
	}
}
