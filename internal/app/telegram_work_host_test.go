package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramWorkResponsive runs the actual receiver and a real blocked worker.
// Incoming HTTP status stays responsive and another adapter retains the same owner.
func TestTelegramWorkResponsive(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)
	root := t.TempDir()
	marker, release := filepath.Join(root, "ready"), filepath.Join(root, "release")
	operations.workflowWorker = []string{"sh", "-c", `printf ready > "$1"; while [ ! -f "$2" ]; do sleep 0.02; done; printf 'Delivered\n' > result.txt; git add result.txt; git -c user.name=Test -c user.email=test@example.com commit -qm 'Worker result'`, "worker", marker, release}
	f.receiver.operations = operations

	telegramWorkRunReceiver(t, f.receiver, release)

	telegramWorkWaitFile(t, marker)
	telegramWorkQueueMessage(t, f, "/status")

	telegramWaitNotice(t, f, "Receiver:", 2*time.Second)

	checkTelegramWorkSharedOwner(t, operations, ref)

	if err := os.WriteFile(release, []byte("release"), 0o600); err != nil {
		t.Fatal(err)
	}

	telegramWorkWaitCursor(t, operations, ref.Invocation.ID)
}

func telegramWorkQueueMessage(t *testing.T, f *telegramConversationAppFixture, text string) {
	t.Helper()

	payload, err := json.Marshal([]any{map[string]any{"update_id": f.next, "message": map[string]any{"message_id": f.next, "from": map[string]any{"id": 100001}, "chat": map[string]any{"id": 100001, "type": "private"}, "text": text}}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- payload
}

func telegramWorkWaitFile(t *testing.T, marker string) {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("owned worker did not start")
}

func telegramWorkWaitCursor(t *testing.T, operations service, id string) {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		run, err := operations.ReadWorkflow(t.Context(), id)
		if err != nil {
			t.Fatal(err)
		}

		if run.Cursor == 1 {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("owned worker did not finish its first checkpoint")
}

func telegramWorkRunReceiver(t *testing.T, receiver telegramReceiver, release string) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())

	finished := make(chan error, 1)
	go func() { finished <- receiver.run(ctx) }()

	t.Cleanup(func() {
		if err := os.WriteFile(release, []byte("release"), 0o600); err != nil {
			t.Error(err)
		}

		cancel()

		select {
		case err := <-finished:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(30 * time.Second):
			t.Error("owned receiver activations did not finish")
		}
	})
}

func checkTelegramWorkSharedOwner(t *testing.T, operations service, ref telegramcore.TelegramWorkReference) {
	t.Helper()

	status, err := operations.WorkflowStatus(t.Context(), ref.Invocation.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.TotalAttempts != 1 || status.ValidatedTasks != 0 {
		t.Fatal("blocked worker lost canonical ownership", status, err)
	}

	if progress, err := operations.ResumeRetainedWork(t.Context(), ref.Invocation.ID, status.Revision); (!errors.Is(err, core.ErrWorkflowOwned) && !errors.Is(err, core.ErrWorkflowCapacity)) || progress.Steps != 0 {
		t.Fatal("another adapter duplicated active work", progress, err)
	}
}
