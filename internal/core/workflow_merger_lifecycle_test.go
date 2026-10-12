package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNativeMergerDescendants holds the shared slot after the driver reports a
// proposal and exits. Its orphaned descendant must stop before host completion.
func TestNativeMergerDescendants(t *testing.T) {
	f, request, harness := nativeMergerExecution(t, "merger-orphan")
	finished := make(chan error, 1)

	go func() {
		_, err := f.store.ExecuteWorkflowMerger(t.Context(), request, harness, f.provider)
		finished <- err
	}()

	nativeMergerWaitMarker(t, f, "child.pid")

	file, guard, err := f.store.acquireWorkflowWorkerSlot(t.Context())
	if err == nil {
		_ = file.Close()  //nolint:errcheck
		_ = guard.Close() //nolint:errcheck

		t.Fatal("live descendant lost the shared slot")
	}

	if !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal(err)
	}

	select {
	case err := <-finished:
		t.Fatal("role finished while child was alive", err)
	default:
	}

	if err := <-finished; err != nil {
		t.Fatal(err)
	}

	file, guard, err = f.store.acquireWorkflowWorkerSlot(t.Context())
	if err != nil {
		t.Fatal("stopped descendant retained capacity", err)
	}

	if err := errors.Join(file.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}
}

func nativeMergerWaitMarker(t *testing.T, f workflowPreparationFixture, marker string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		run := workflowTestRead(t, f.store, f.run.ID)

		attempt := run.Attempts[len(run.Attempts)-1]
		if attempt.Purpose == workflowMergerPurpose && attempt.Lane != nil {
			if _, err := os.Stat(filepath.Join(filepath.Dir(attempt.Lane.Worktree), marker)); err == nil {
				return
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("owned child did not start")
}

// TestNativeMergerTimeout verifies the native deadline and whole process group.
// Timeout is retained as failure and never refunds the automatic call counter.
func TestNativeMergerTimeout(t *testing.T) {
	f, request, harness := nativeMergerExecution(t, "merger-timeout")
	request.Input.Deadline = time.Now().UTC().Add(3 * time.Second)

	var err error

	request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	attempt, err := f.store.ExecuteWorkflowMerger(t.Context(), request, harness, f.provider)
	if err == nil || attempt.Proposal != nil {
		t.Fatal("timeout produced a proposal", err)
	}

	run := workflowTestRead(t, f.store, f.run.ID)

	current := run.Attempts[len(run.Attempts)-1]
	if current.Purpose != workflowMergerPurpose || current.Outcome != "failed" || len(current.Processes) != 1 || workflowAttemptQuiescent(current) != nil {
		t.Fatal("timeout did not retain stopped ownership")
	}

	checkNativeMergerBudget(t, f)
}

// TestNativeMergerStop uses the existing owner controls while a real driver is
// running. Pause retains the integration frontier; cancellation stays terminal.
func TestNativeMergerStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { nativeMergerStop(t, intent) })
	}
}

func nativeMergerStop(t *testing.T, intent string) {
	t.Helper()

	f, request, harness := nativeMergerExecution(t, "merger-timeout")
	finished := make(chan error, 1)

	go func() {
		_, err := f.store.ExecuteWorkflowMerger(t.Context(), request, harness, f.provider)
		finished <- err
	}()

	nativeMergerWaitMarker(t, f, "driver-started")

	if _, err := f.store.StopWorkflow(t.Context(), f.run.ID, intent); err != nil {
		t.Fatal(err)
	}

	if err := <-finished; !errors.Is(err, ErrWorkflowStop) {
		t.Fatal("owned stop was not retained", err)
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Status != workflowStopStatus(intent) || workflowAttemptQuiescent(run.Attempts[len(run.Attempts)-1]) != nil {
		t.Fatal("stop discarded process ownership")
	}

	checkNativeMergerBudget(t, f)

	if intent == "pause" {
		if run.ResumeStatus != "waiting_integration" {
			t.Fatal("pause lost its integration continuation")
		}

		if _, err := f.store.ResumeWorkflow(t.Context(), f.run.ID); err != nil {
			t.Fatal("pause was not resumable", err)
		}
	}
}
