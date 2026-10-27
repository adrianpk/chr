package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowClosingStop proves a set gate retains coarse capacity while a
// descendant is alive, then releases its entire group on pause, cancellation or
// the finite deadline. No stop can establish a closure or create task attempts.
func TestWorkflowClosingStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel", "deadline"} {
		t.Run(intent, func(t *testing.T) { testWorkflowClosingStop(t, intent) })
	}
}

func testWorkflowClosingStop(t *testing.T, intent string) {
	t.Helper()

	marker := filepath.Join(t.TempDir(), "running")

	gate := workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"sh", "-c", "printf 'started\\n' >\"$1\"; sleep 20 & wait", "closing-gate", marker}}}
	if intent == "deadline" {
		gate.TimeoutSeconds = 2
	}

	fixture, set, run := workflowClosingEmptyGate(t, gate)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		_, err := fixture.store.CloseWorkflowInvocationSet(ctx, run.ID, "dev", workflowFrontierPorts(fixture))
		done <- err
	}()

	waitWorkflowClosingMarker(t, marker)

	if capacity, err := acquireWorkflowCapacity(t.Context(), fixture.store.layout.ProjectDir()); !errors.Is(err, ErrWorkflowCapacity) {
		if capacity != nil {
			_ = capacity.Close() //nolint:errcheck
		}

		t.Fatal("live closing descendants did not retain repository capacity", err)
	}

	if intent != "deadline" {
		if _, err := fixture.store.StopWorkflow(t.Context(), run.ID, intent); err != nil {
			t.Fatal(err)
		}
	}

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("stopped closing gate established completion")
		}
	case <-time.After(6 * time.Second):
		t.Fatal("closing supervisor did not stop its descendants")
	}

	checkWorkflowClosingStopped(t, fixture.store, set.Path, run.ID, intent)
}

func waitWorkflowClosingMarker(t *testing.T, marker string) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("closing gate did not start")
}

func checkWorkflowClosingStopped(t *testing.T, store *WorkflowStore, path, id, intent string) {
	t.Helper()

	set, err := store.Read(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}

	state, err := store.readWorkflowClosingAttempt(set, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := workflowAttemptQuiescent(WorkflowAttempt{Processes: state.Processes}); err != nil {
		t.Fatal("stopped closing gate retained a descendant", err)
	}

	run := workflowTestRead(t, store, id)
	if set.Set.ClosureRef != "" || run.ClosingSet != "" || len(run.Attempts) != 0 || state.Outcome != "closing_validation_failed" {
		t.Fatal("stopped gate confirmed completion or retained closing ownership")
	}

	if intent != "deadline" && run.Status != workflowStopStatus(intent) {
		t.Fatal("closing finalization lost durable stop intent")
	}
}
