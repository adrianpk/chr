package app

import (
	"os"
	"os/exec"
	"testing"
)

// TestTelegramWorkRecovery reopens the receiver after a real child abandons a
// native manual claim. Common recovery retains attempts, budgets and operator stops.
func TestTelegramWorkRecovery(t *testing.T) {
	for _, intent := range []string{"", "pause", "cancel"} {
		name := intent
		if name == "" {
			name = "retry"
		}

		t.Run(name, func(t *testing.T) {
			checkTelegramWorkRecovered(t, intent)
		})
	}
}

func checkTelegramWorkRecovered(t *testing.T, intent string) {
	t.Helper()

	f, operations, ref := telegramWorkAdmitted(t)
	before := checkTelegramWorkRunCount(t, operations, 1)
	child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestDaemonWorkflowRecovery$")

	child.Env = append(os.Environ(), "CHR_DAEMON_TEST_PROJECT="+operations.projectContext.ProjectDir, "CHR_DAEMON_TEST_PROJECT_ID="+operations.projectContext.ProjectID, "CHR_DAEMON_TEST_INVOCATION="+ref.Invocation.ID)
	if output, err := child.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}

	if intent != "" {
		if _, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, intent); err != nil {
			t.Fatal(err)
		}
	}

	telegramAuthoringReopen(t, f)

	result := f.receiver.activateWork(t.Context(), ref)
	if result.Failure != nil {
		t.Fatal(result.Failure)
	}

	if err := f.receiver.completeWorkActivation(t.Context(), result); err != nil {
		t.Fatal(err)
	}

	checkDaemonWorkflowRecoveryCheckpoint(t, operations, before, intent)
}
