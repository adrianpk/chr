package app

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestTelegramRecoveryDeadWork starts the receiver after an actual child leaves
// native ownership behind. Startup fences that dead identity without restoring
// spent attempts or overriding a later operator pause/cancellation.
func TestTelegramRecoveryDeadWork(t *testing.T) {
	for _, intent := range []string{"", "pause", "cancel"} {
		name := intent
		if name == "" {
			name = "retry"
		}

		t.Run(name, func(t *testing.T) { checkTelegramRecoveryDeadWork(t, intent) })
	}
}

func checkTelegramRecoveryDeadWork(t *testing.T, intent string) {
	t.Helper()

	f, operations, ref := telegramWorkAdmitted(t)
	before := checkTelegramWorkRunCount(t, operations, 1)
	child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestDaemonWorkflowRecovery$")

	child.Env = append(os.Environ(), "CHR_DAEMON_TEST_PROJECT="+operations.projectContext.ProjectDir, "CHR_DAEMON_TEST_PROJECT_ID="+operations.projectContext.ProjectID, "CHR_DAEMON_TEST_INVOCATION="+ref.Invocation.ID)
	if output, err := child.CombinedOutput(); err != nil {
		t.Fatal("owned fixture did not abandon its claim", err, string(output))
	}

	if intent != "" {
		if _, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, intent); err != nil {
			t.Fatal(err)
		}
	}

	telegramAuthoringReopen(t, f)

	cancel, finished := startTelegramRecoveryReceiver(t, f.receiver)
	defer cancel()

	waitTelegramRecoveryNative(t, operations, ref.Invocation.ID, intent)
	cancel()
	waitTelegramRecoveryReceiver(t, finished)
	checkDaemonWorkflowRecoveryCheckpoint(t, operations, before, intent)
}

func waitTelegramRecoveryNative(t *testing.T, operations service, id, intent string) {
	t.Helper()

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		current, err := operations.ReadWorkflow(t.Context(), id)
		if err != nil {
			t.Fatal(err)
		}

		if len(current.Attempts) > 0 && current.Attempts[0].Outcome == "recovered" && (intent != "" || current.Cursor == 1) {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("startup did not reconcile the dead native owner")
}
