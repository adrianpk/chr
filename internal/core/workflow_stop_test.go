package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowStop retains live intent, synced output and a fenced task checkpoint.
func TestWorkflowStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { testWorkflowStop(t, intent) })
	}
}

func testWorkflowStop(t *testing.T, intent string) {
	t.Helper()

	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)
	marker := filepath.Join(t.TempDir(), "started")
	done := make(chan error, 1)

	worker := []string{"sh", "-c", "printf 'Retained output\\n'; printf 'started' > \"$1\"; sleep 20", "worker", marker}
	go func() { _, err := store.ExecuteWorkflowTask(t.Context(), claim, worker); done <- err }()

	waitWorkflowClaimFile(t, marker)

	stopped, err := store.StopWorkflow(t.Context(), run.Invocation.ID, intent)
	if err != nil || stopped.Status != "running" || stopped.StopIntent != intent {
		t.Fatalf("stop intent: %+v, %v", stopped, err)
	}

	if err := store.CheckpointWorkflow(t.Context(), claim); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("checkpoint freed a running worker")
	}

	if err := <-done; !errors.Is(err, ErrWorkflowStop) {
		t.Fatalf("stopped result: %v", err)
	}

	checkWorkflowStoppedTask(t, store, hierarchy.Tasks[0].Path, run.Invocation.ID, intent)

	checkWorkflowStoppedContinuation(t, store, request, run.Invocation.ID, claim, intent)
}

func checkWorkflowStopEvidence(t *testing.T, store *WorkflowStore, taskPath string, retained WorkflowInvocation) {
	t.Helper()

	attempt := retained.Attempts[0]

	data, err := opsfile.Read(filepath.Join(store.layout.Root(), attempt.CheckpointRef))
	if err != nil || contenthash.Sum(data) != attempt.CheckpointHash {
		t.Fatal("stop lost checkpoint evidence")
	}

	output, err := os.ReadFile(filepath.Join(store.layout.Root(), attempt.Processes[0].StdoutRef))
	if err != nil || string(output) != "Retained output\n" || contenthash.Sum(output) != attempt.Processes[0].StdoutHash {
		t.Fatalf("stop lost durable output: %q, %v", output, err)
	}

	record, err := store.Read(t.Context(), taskPath)
	if err != nil || record.Task.Phase != "pending" || record.Task.AttemptID != "" {
		t.Fatal("stop fabricated validation")
	}
}

// TestWorkflowStopBarrier resumes the original integration barrier without workers.
func TestWorkflowStopBarrier(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice")

	run := workflowTestStart(t, store, request)
	for i := 0; i < 3; i++ {
		workflowTestExecute(t, store, run.Invocation.ID, "worker")
	}

	if _, err := store.StopWorkflow(t.Context(), run.Invocation.ID, "pause"); err != nil {
		t.Fatal(err)
	}

	resumed, err := store.ResumeWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || resumed.Status != "waiting_integration" || resumed.Cursor != 3 || len(resumed.Attempts) != 3 {
		t.Fatal("pause lost integration barrier")
	}
}

// TestWorkflowOwnerExpiry rejects stale results and keeps a held reservation counted.
func TestWorkflowOwnerExpiry(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	request.Policy = workflowstate.WorkflowExecutionPolicy{WorkerTimeoutSeconds: 1, MaxWorkerAttempts: 1}
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)

	time.Sleep(1100 * time.Millisecond)

	if err := store.CheckWorkflowClaim(t.Context(), claim); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("expired owner accepted")
	}

	if _, err := store.ReconcileWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatalf("replaced a held reservation: %v", err)
	}

	if err := store.CheckpointWorkflow(t.Context(), claim); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("expired token applied a checkpoint")
	}
}

func checkWorkflowStoppedContinuation(t *testing.T, store *WorkflowStore, request WorkflowStartRequest, id string, claim *WorkflowClaim, intent string) {
	t.Helper()

	if intent == "pause" {
		checkWorkflowPauseResume(t, store, id, claim)

		return
	}

	attached := workflowTestStart(t, store, request)
	if !attached.Attached || attached.Invocation.Status != "cancelled" {
		t.Fatal("terminal retry restarted cancelled invocation")
	}

	if _, err := store.ResumeWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("cancelled invocation resumed")
	}
}

func checkWorkflowPauseResume(t *testing.T, store *WorkflowStore, id string, claim *WorkflowClaim) {
	t.Helper()

	if _, err := store.ResumeWorkflow(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	replacement := workflowTestClaim(t, store, id)
	if replacement.Generation != claim.Generation+1 {
		t.Fatal("resume lost the fence")
	}

	if err := store.CheckpointWorkflow(t.Context(), claim); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("old credential released the resumed claim")
	}
}

func checkWorkflowStoppedTask(t *testing.T, store *WorkflowStore, path, id, intent string) {
	t.Helper()

	retained, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || retained.Status != workflowStopStatus(intent) || retained.Cursor != 0 || retained.Lane == nil {
		t.Fatalf("retained checkpoint: %+v, %v", retained, err)
	}

	checkWorkflowStopEvidence(t, store, path, retained)

	if _, err := store.ClaimWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("stop intent allowed continuation")
	}
}
