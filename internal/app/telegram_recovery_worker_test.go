package app

import (
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramRecoveryOwnedStop cancels only the receiver's actual source
// group. A manual adapter keeps its live claim, worker and finite policy.
func TestTelegramRecoveryOwnedStop(t *testing.T) {
	f, operations, record := telegramWorkAdmissionFixture(t, "ticket")
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n  max_active_workers: 2\n")

	first := telegramControlCapacityAdmission(t, f, record.Task.Key)

	runs, waiting := telegramControlCapacityRuns(t, f, operations, first, 2)
	if _, err := operations.StopWorkflow(t.Context(), waiting.Invocation.ID, "pause"); err != nil {
		t.Fatal(err)
	}

	external := telegramControlSharedWorker(t, f.receiver, operations, runs[0].ID, nil, false)

	before, err := operations.ReadWorkflow(t.Context(), runs[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	marker := filepath.Join(t.TempDir(), "receiver-source")
	operations.workflowWorker = []string{"sh", "-c", `printf ready > "$1"; while :; do sleep 1; done`, "worker", marker}
	f.receiver.operations = operations

	cancel, finished := startTelegramRecoveryReceiver(t, f.receiver)
	defer cancel()

	telegramWorkWaitFile(t, marker)
	checkTelegramControlPool(t, operations, 2, 0)
	cancel()
	waitTelegramRecoveryReceiver(t, finished)

	owned, err := operations.ReadWorkflow(t.Context(), first.Invocation.ID)
	if err != nil || len(owned.Attempts) != 1 || owned.Attempts[0].CheckpointHash == "" || owned.StopIntent != "" || owned.Cursor != 0 {
		t.Fatal("shutdown lost its checkpoint or fabricated operator intent", owned, err)
	}

	checkTelegramRecoveryForeignOwner(t, operations, before)
	checkTelegramControlPool(t, operations, 2, 1)

	select {
	case err := <-external.done:
		t.Fatal("receiver terminated the other adapter's source", err)
	default:
	}

	finishTelegramControlWorker(t, external)
	checkTelegramControlPool(t, operations, 2, 2)
}

func checkTelegramRecoveryForeignOwner(t *testing.T, operations service, before workflowstate.Invocation) {
	t.Helper()

	after, err := operations.ReadWorkflow(t.Context(), before.ID)
	if err != nil || len(after.Attempts) != 1 || after.Status != before.Status || after.Cursor != before.Cursor || after.StopIntent != before.StopIntent || after.Policy != before.Policy {
		t.Fatal("receiver changed another adapter's work", after, err)
	}

	checkTelegramRecoveryAttempt(t, before.Attempts[0], after.Attempts[0])
}

func checkTelegramRecoveryAttempt(t *testing.T, first, current workflowstate.Attempt) {
	t.Helper()

	if first.ID != current.ID || first.Generation != current.Generation || first.InputHash != current.InputHash || first.Outcome != current.Outcome {
		t.Fatal("receiver replaced another adapter's active attempt", current.ID)
	}
}
