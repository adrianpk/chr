package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowManualAlias validates a quiescent human commit through a linked
// checkout, retaining evidence only in the invocation's registered owner store.
func TestWorkflowManualAlias(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	alias := workflowRecoveryAlias(t, store, request.Selection.Target.Commit)
	claim, task := workflowTestManual(t, alias, run.Invocation.ID, nil)

	lane := task.Input.Lane.Worktree

	file := task.Input.TaskKey + ".txt"
	if err := os.WriteFile(filepath.Join(lane, file), []byte(task.Input.TaskKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowSelectionGit(t, store, "-C", lane, "add", file)
	workflowSelectionGit(t, store, "-C", lane, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "Manual candidate")

	candidate := workflowManualTestCandidate(t, store, task)

	proof, err := alias.ReconcileWorkflowTask(t.Context(), claim, candidate)
	if err != nil || !proof.Successful || proof.Commit == request.Selection.Target.Commit {
		t.Fatalf("manual linked candidate: %+v, %v", proof, err)
	}

	checkWorkflowExactTaskPhases(t, store, hierarchy)

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || len(retained.Attempts[0].Processes) != 1 {
		t.Fatal("quiescent manual reconciliation launched a worker")
	}
}

// TestWorkflowManualStop checkpoints a manual host gate rather than accepting
// its candidate when pause/cancel arrives while a validation child is live.
func TestWorkflowManualStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { testWorkflowManualStop(t, intent) })
	}
}

func testWorkflowManualStop(t *testing.T, intent string) {
	t.Helper()

	marker := filepath.Join(t.TempDir(), "started")
	gate := []string{"sh", "-c", "printf 'Retained output\\n'; printf 'started' > \"$1\"; sleep 20", "gate", marker}
	store, hierarchy, request := workflowExecutionFixture(t, gate, 30)
	run := workflowTestStart(t, store, request)
	claim, task := workflowTestManual(t, store, run.Invocation.ID, nil)
	candidate := workflowManualTestCandidate(t, store, task)
	done := make(chan error, 1)

	go func() { _, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate); done <- err }()

	waitWorkflowClaimFile(t, marker)

	if _, err := store.StopWorkflow(t.Context(), run.Invocation.ID, intent); err != nil {
		t.Fatal(err)
	}

	if err := store.CheckpointWorkflow(t.Context(), claim); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("checkpoint released a manual validation child")
	}

	if err := <-done; !errors.Is(err, ErrWorkflowStop) {
		t.Fatalf("manual stop: %v", err)
	}

	checkWorkflowStoppedTask(t, store, hierarchy.Tasks[0].Path, run.Invocation.ID, intent)
}
