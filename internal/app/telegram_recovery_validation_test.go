package app

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramRecoveryValidation starts a real manual host-validation process
// before receiver startup. Receiver shutdown neither fences it nor spends work.
func TestTelegramRecoveryValidation(t *testing.T) {
	f, operations, record := telegramWorkAdmissionFixture(t, "ticket")
	marker, release := filepath.Join(t.TempDir(), "validation"), filepath.Join(t.TempDir(), "release")
	permit := telegramControlSourceSignal(t, release)
	validation := workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"sh", "-c", `printf ready > "$1"; IFS= read -r permit < "$2"; [ "$permit" = release ] && git show HEAD:result.txt`, "validation", marker, release}}}

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	setTelegramRecoveryValidation(t, store, record, validation)

	ref := telegramControlCapacityAdmission(t, f, record.Task.Key)
	operations.workflowWorker = []string{"sh", "-c", `printf 'Delivered\n' > result.txt; git add result.txt; git -c user.name=Test -c user.email=test@example.com commit -qm 'Manual result'`}

	execute, finalize := telegramControlManualCall(t, operations, ref.Invocation.ID)
	if err := execute(t.Context()); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- finalize(t.Context()); close(done) }()

	unblock := sync.OnceFunc(func() {
		if _, err := permit.WriteString("release\n"); err != nil {
			t.Error(err)
		}
	})

	t.Cleanup(func() { unblock(); waitTelegramRecoveryReceiver(t, done) })
	telegramWorkWaitFile(t, marker)

	before, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	cancel, finished := startTelegramRecoveryReceiver(t, f.receiver)
	defer cancel()

	telegramWorkQueueMessage(t, f, "/status")
	telegramWaitNotice(t, f, "Receiver:", 5*time.Second)
	cancel()
	waitTelegramRecoveryReceiver(t, finished)
	checkTelegramRecoveryForeignOwner(t, operations, before)
	checkTelegramControlPool(t, operations, 1, 0)

	select {
	case err := <-done:
		t.Fatal("receiver terminated another adapter's validation", err)
	default:
	}

	unblock()
	waitTelegramRecoveryReceiver(t, done)
	checkTelegramControlPool(t, operations, 1, 1)
}

func setTelegramRecoveryValidation(t *testing.T, store *core.WorkflowStore, record core.WorkflowRecord, validation workflowstate.Validation) {
	t.Helper()

	var contract generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err != nil {
		t.Fatal(err)
	}

	contract.Validation = validation

	encoded, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	record.Domain["execution_contract"] = encoded

	record.Task.ContractHash = core.WorkflowContentHash(append([]byte{0, 0}, encoded...))
	if _, err := store.Write(t.Context(), record, record.Revision); err != nil {
		t.Fatal(err)
	}
}
