package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowCapacityContinuation retains a distinct wait checkpoint without
// repeated history growth. Capacity release permits the same frozen task to run.
func TestWorkflowCapacityContinuation(t *testing.T) {
	store, _, start := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	run := workflowTestStart(t, store, start).Invocation

	legacy, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment, Owner: "legacy"})
	if err != nil {
		t.Fatal(err)
	}

	request := ContinuationRequest{Worker: workflowPRWorker(), StepLimit: 1}

	assertWorkflowContinuationWait(t, store, run.ID, request)

	if err := legacy.Release(); err != nil {
		t.Fatal(err)
	}

	continued, err := ContinueWorkflow(t.Context(), store, run.ID, request, ContinuationPorts{})
	if err != nil || continued.Steps != 1 || len(continued.Invocation.Attempts) != 1 || continued.Invocation.Cursor != 1 || continued.Invocation.CapacityResumeStatus != workflowstate.InvocationUnspecified {
		t.Fatal("capacity wait could not continue its frozen task", continued, err)
	}
}

func assertWorkflowContinuationWait(t *testing.T, store *Store, id string, request ContinuationRequest) {
	t.Helper()

	waiting, err := ContinueWorkflow(t.Context(), store, id, request, ContinuationPorts{})
	if !errors.Is(err, ErrWorkflowCapacity) || waiting.Reason != "waiting_capacity" || waiting.Steps != 0 {
		t.Fatal("capacity reported as a generic blocker", waiting, err)
	}

	again, err := ContinueWorkflow(t.Context(), store, id, request, ContinuationPorts{})
	if !errors.Is(err, ErrWorkflowCapacity) || again.Invocation.Revision != waiting.Invocation.Revision || again.Invocation.Continuation.Sequence != waiting.Invocation.Continuation.Sequence || len(again.Invocation.Attempts) != 0 {
		t.Fatal("unchanged wait grew checkpoints or attempts", again, err)
	}
}

// TestWorkflowCapacityStoppedWait keeps cancellation terminal and restores a
// paused capacity phase without consuming a new attempt or dropping its extent.
func TestWorkflowCapacityStoppedWait(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { testWorkflowCapacityStoppedWait(t, intent) })
	}
}

func testWorkflowCapacityStoppedWait(t *testing.T, intent string) {
	t.Helper()

	stores, runs := workflowParallelRunnerLanes(t, 1)

	owner := workflowTestClaim(t, stores[0], runs[0].ID)
	if _, err := stores[1].ClaimWorkflow(t.Context(), runs[1].ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal(err)
	}

	if _, err := stores[1].StopWorkflow(t.Context(), runs[1].ID, intent); err != nil {
		t.Fatal(err)
	}

	if err := stores[0].CheckpointWorkflow(t.Context(), owner); err != nil {
		t.Fatal(err)
	}

	if intent == "cancel" {
		if _, err := stores[1].ClaimWorkflow(t.Context(), runs[1].ID); !errors.Is(err, workflowstate.ErrOwned) {
			t.Fatal("cancelled wait resumed", err)
		}

		return
	}

	if _, err := stores[1].ResumeWorkflow(t.Context(), runs[1].ID); err != nil {
		t.Fatal(err)
	}

	claim := workflowTestClaim(t, stores[1], runs[1].ID)
	if err := stores[1].CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowCapacityRetainedLock keeps an unbound legacy file exclusive even
// after its process descriptor disappears; native retries never delete it.
func TestWorkflowCapacityRetainedLock(t *testing.T) {
	store, _, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	run := workflowTestStart(t, store, request).Invocation
	path := filepath.Join(store.layout.Root(), "lock", "agent-operation.lock")

	data := []byte(`{"operation":"development","owner":"legacy-retained","pid":1}`)
	if err := opsfile.WriteAtomic(path, data); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ClaimWorkflow(t.Context(), run.ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("retained legacy file was ignored", err)
	}

	observed, err := os.ReadFile(path)
	if err != nil || string(observed) != string(data) {
		t.Fatal("native retry removed or rewrote a legacy file", err)
	}
}
