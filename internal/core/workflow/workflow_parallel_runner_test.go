package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowParallelExecution proves source-process overlap across namespaces
// and linked checkouts, serial fallback at one, and retained third-lane waits.
func TestWorkflowParallelExecution(t *testing.T) {
	for _, capacity := range []int{1, 2} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) { testWorkflowParallelExecution(t, capacity) })
	}
}

func testWorkflowParallelExecution(t *testing.T, capacity int) {
	t.Helper()

	stores, runs := workflowParallelRunnerLanes(t, capacity)
	first := workflowParallelRunnerFixture(t, stores[0], runs[0].ID, "worker")
	waitWorkflowClaimFile(t, first.spec.Started)

	second := workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "worker")

	if capacity == 1 {
		result := finishWorkflowParallelRunner(t, second)
		if !result.Capacity {
			t.Fatal("capacity one admitted another actual worker", result)
		}
	} else {
		waitWorkflowClaimFile(t, second.spec.Started)
		assertWorkflowNativeLockAbsent(t, stores[0])
		assertWorkflowNativeLockAbsent(t, stores[1])
	}

	if _, err := stores[2].ClaimWorkflow(t.Context(), runs[2].ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("third worker bypassed live processes", err)
	}

	assertWorkflowCapacityWait(t, stores[2], runs[2].ID)

	if legacy, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: stores[1].layout.ProjectDir(), Operation: AgentOperationDevelopment, Owner: "legacy"}); !errors.Is(err, ErrWorkflowCapacity) {
		if legacy != nil {
			t.Error(legacy.Release())
		}

		t.Fatal("legacy start bypassed actual native workers", err)
	}

	if result := finishWorkflowParallelRunner(t, first); !result.Successful {
		t.Fatal("first independent source failed", result)
	}

	if capacity == 1 {
		second = workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "worker")
		waitWorkflowClaimFile(t, second.spec.Started)
	}

	if result := finishWorkflowParallelRunner(t, second); !result.Successful {
		t.Fatal("second independent source failed", result)
	}

	claim := workflowTestClaim(t, stores[2], runs[2].ID)
	if err := stores[2].CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal("retained capacity wait did not resume", err)
	}
}

func assertWorkflowNativeLockAbsent(t *testing.T, store *Store) {
	t.Helper()

	path := filepath.Join(store.layout.Root(), "lock", "agent-operation.lock")
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("native parallel claim created a coarse operation file", err)
	}
}

func assertWorkflowCapacityWait(t *testing.T, store *Store, id string) {
	t.Helper()

	run := workflowTestRead(t, store, id)
	if run.Status != workflowstate.InvocationWaitingCapacity || run.CapacityResumeStatus != workflowstate.InvocationPending || len(run.Attempts) != 0 {
		t.Fatal("capacity wait lost its phase or consumed an attempt", run)
	}
}

// TestWorkflowParallelBlocker proves that a failed source or worker-owned
// coordinator edit cannot stop another live lane or change its validated proof.
func TestWorkflowParallelBlocker(t *testing.T) {
	for _, mode := range []string{"fail", "marker"} {
		t.Run(mode, func(t *testing.T) {
			stores, runs := workflowParallelRunnerLanes(t, 2)
			first := workflowParallelRunnerFixture(t, stores[0], runs[0].ID, mode)
			waitWorkflowClaimFile(t, first.spec.Started)

			second := workflowParallelRunnerFixture(t, stores[1], runs[1].ID, "worker")
			waitWorkflowClaimFile(t, second.spec.Started)

			if result := finishWorkflowParallelRunner(t, first); result.Successful || result.Failure == "" {
				t.Fatal("owned failure was not rejected", result)
			}

			if result := finishWorkflowParallelRunner(t, second); !result.Successful {
				t.Fatal("another lane's blocker stopped an unrelated process", result)
			}

			retained := workflowTestRead(t, stores[1], runs[1].ID)
			if retained.Cursor != 1 || len(retained.Attempts) != 1 || retained.Attempts[0].Outcome != "validated" {
				t.Fatal("unrelated lane lost its accepted proof", retained)
			}
		})
	}
}
