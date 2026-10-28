package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// TestTelegramControlCheckpoint preserves an operator stop when a live worker
// asynchronously exits and journals its fenced checkpoint through the owner.
func TestTelegramControlCheckpoint(t *testing.T) {
	telegramControlParallel(t)

	for _, action := range []string{"pause", "cancel"} {
		t.Run(action, func(t *testing.T) { checkTelegramControlCheckpoint(t, action) })
	}
}

func checkTelegramControlCheckpoint(t *testing.T, action string) {
	t.Helper()

	f, operations, ref := telegramWorkAdmitted(t)
	marker := filepath.Join(t.TempDir(), "ready")
	operations.workflowWorker = []string{"sh", "-c", `printf ready > "$1"; while :; do sleep 0.02; done`, "worker", marker}
	f.receiver.operations = operations
	done := telegramControlStartWorker(t, operations, ref.Invocation)

	telegramWorkWaitFile(t, marker)

	token := telegramControlOffer(t, f, "/"+action+" "+ref.Invocation.ID)
	f.send(t, "/answer "+token+" "+action, 0)

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, workflowcore.ErrWorkflowStop) {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("confirmed stop did not quiesce the native worker")
	}

	run := checkTelegramWorkRunCount(t, operations, 1)
	if run.StopIntent != action || run.Cursor != 0 || len(run.Attempts) != 1 || run.Attempts[0].CheckpointHash == "" {
		t.Fatal("asynchronous checkpoint replaced stop intent or lost evidence", run)
	}
}

func telegramControlStartWorker(t *testing.T, operations service, run telegramcore.InvocationBinding) <-chan error {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	done, finished := make(chan error, 1), make(chan struct{})

	go func() {
		defer close(finished)

		_, err := operations.ResumeRetainedWork(ctx, run.ID, run.Revision)
		done <- err
	}()

	t.Cleanup(func() {
		cancel()

		select {
		case <-finished:
		case <-time.After(30 * time.Second):
			t.Error("owned worker did not exit")
		}
	})

	return done
}
