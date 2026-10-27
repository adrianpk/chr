package core

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowRecoveryBudget replays an interrupted claim, fences it and consumes
// the same per-invocation budget, including attempts that never launched a worker.
func TestWorkflowRecoveryBudget(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	request.Policy.MaxWorkerAttempts = 1
	run := workflowTestStart(t, store, request)
	command := workflowClaimProcess(t, store, run.Invocation.ID, "", "", "0")

	_, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("claim fixture did not crash")
	}

	recovered, err := store.ReconcileWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || recovered.Status != "paused" || len(recovered.Attempts) != 1 || recovered.Attempts[0].Outcome != "recovered" {
		t.Fatalf("claim recovery: %+v, %v", recovered, err)
	}

	if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); err == nil {
		t.Fatal("recovery reset consumed crash budget")
	}
}

// TestWorkflowRecoveryStop replays durable stop/checkpoint writes at every task
// replacement and retains terminal cancellation without launching another attempt.
func TestWorkflowRecoveryStop(t *testing.T) {
	for _, index := range []int{0, 2, 3} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			testWorkflowRecoveryStop(t, index)
		})
	}
}

// TestWorkflowRecoveryForeignLock preserves a replaced operation lock during
// recovery; only the matching invocation, attempt and birth identity may remove it.
func TestWorkflowRecoveryForeignLock(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)

	command := workflowClaimProcess(t, store, run.Invocation.ID, "", "", "0")
	if _, err := command.CombinedOutput(); err == nil {
		t.Fatal("fixture did not crash")
	}

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.layout.Root(), "lock", "agent-operation.lock")

	// This historical coarse file is retained compatibility evidence. New native
	// claims no longer create it; a replacement must still survive recovery.
	lock := agentOperationLockRecord{Operation: "workflow_execution", Owner: retained.ID, AttemptID: "replacement", Process: retained.Attempts[0].OwnerProcess}

	replacement, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}

	if err := atomicWorkflowFile(path, replacement); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReconcileWorkflow(t.Context(), retained.ID); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("foreign lock removed: %v", err)
	}

	observed, err := os.ReadFile(path)
	if err != nil || string(observed) != string(replacement) {
		t.Fatal("recovery overwrote newer lock")
	}
}

// TestWorkflowProcessBirth prevents a reused PID or another boot from becoming
// proof that the old owner is still alive, or from authorizing an unrelated kill.
func TestWorkflowProcessBirth(t *testing.T) {
	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	for _, change := range []func(*WorkflowProcessIdentity){
		func(i *WorkflowProcessIdentity) { i.StartTicks += "0" },
		func(i *WorkflowProcessIdentity) { i.BootID = "00000000-0000-0000-0000-000000000000" },
	} {
		stale := identity
		change(&stale)

		if alive, err := workflowProcessAlive(stale); err != nil || alive {
			t.Fatal("stale birth matched live PID")
		}
	}
}

func expireWorkflowRecoveryLease(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	if err := store.withLock(t.Context(), func() error {
		run, owner, err := store.locateWorkflowInvocation(t.Context(), id)
		if err != nil {
			return err
		}

		run.Attempts[0].LeaseUntil = run.Attempts[0].StartedAt.Add(time.Millisecond)

		return owner.commitWorkflowState(&run)
	}); err != nil {
		t.Fatal(err)
	}
}

func testWorkflowRecoveryStop(t *testing.T, index int) {
	t.Helper()

	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if _, err := store.StopWorkflow(t.Context(), run.Invocation.ID, "cancel"); err != nil {
		t.Fatal(err)
	}

	claim.store.afterReplace = func(replaced int) error {
		if replaced == index {
			return errors.New("owned checkpoint interruption")
		}

		return nil
	}
	if err := store.CheckpointWorkflow(t.Context(), claim); err == nil {
		t.Fatal("checkpoint fault was not exercised")
	}

	claim.store.afterReplace = nil
	if err := claim.releaseReservation(); err != nil {
		t.Fatal(err)
	}

	retained, err := store.ReconcileWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || retained.Status != "cancelled" || retained.Attempts[0].CheckpointRef == "" {
		t.Fatalf("stop replay: %+v, %v", retained, err)
	}

	if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("cancel replay resumed work")
	}
}

// TestWorkflowProcessEdit retains a changed process image without applying a
// worker result. Host-observed birth and command bindings fence finalization.
func TestWorkflowProcessEdit(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "edit process")); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("changed process binding accepted: %v", err)
	}

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || retained.Status != "running" || retained.Cursor != 0 || retained.Attempts[0].Processes[0].State != "running" {
		t.Fatal("host overwrote the incompatible process image")
	}
}

// TestWorkflowProcessSignal rejects a reused PID before sending any group signal.
func TestWorkflowProcessSignal(t *testing.T) {
	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	identity.StartTicks += "0"
	guard := &workflowCommandGuard{identity: &identity}

	command := &exec.Cmd{Process: &os.Process{Pid: os.Getpid()}}
	if err := guard.kill(command); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("stale process was signalled: %v", err)
	}
}
