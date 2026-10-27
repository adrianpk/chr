package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// TestGuidedUseCaseRecoveryProcess fences a real dead child owner on startup,
// retaining spent attempts and any guided operator stop accepted after the crash.
func TestGuidedUseCaseRecoveryProcess(t *testing.T) {
	for _, action := range []string{"continue", "pause", "cancel"} {
		t.Run(action, func(t *testing.T) { testGuidedProcessRecovery(t, action) })
	}
}

func testGuidedProcessRecovery(t *testing.T, action string) {
	t.Helper()

	operations, session := guidedAdmittedFixture(t)
	before := assertGuidedNativeAdmission(t, operations, session)
	child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestDaemonWorkflowRecovery$")

	child.Env = append(os.Environ(), "CHR_DAEMON_TEST_PROJECT="+operations.projectContext.ProjectDir, "CHR_DAEMON_TEST_PROJECT_ID="+operations.projectContext.ProjectID, "CHR_DAEMON_TEST_INVOCATION="+before.ID)
	if output, err := child.CombinedOutput(); err != nil {
		t.Fatal("owned child did not abandon its native claim", err, string(output))
	}

	if action != "continue" {
		var err error

		session, err = operations.ControlGuidedUseCase(t.Context(), guidedNativeControl(t, operations, session, "stop-after-crash", action))
		if err != nil {
			t.Fatal(err)
		}
	}

	host := startGuidedRecoveryHost(t, operations)
	current := waitGuidedRecovery(t, operations, session, func(current authoringcore.Session) bool {
		if action == "continue" {
			return guidedRecoveryReviewReady(t.Context(), operations, current)
		}

		return current.Guided.Control.AcknowledgedRevision != 0
	})

	host.close()

	after := assertGuidedNativeAdmission(t, operations, current)
	if after.Policy != before.Policy || after.ID != before.ID || after.Attempts[0].Outcome != "recovered" {
		t.Fatal("startup replaced dead ownership or frozen policy")
	}

	expected := 1
	if action == "continue" {
		expected = 2
	}

	if serviceWorkTaskAttempts(after) != expected {
		t.Fatal("recovery refilled attempts or dispatched past a stop")
	}
}

// TestGuidedUseCaseRecoveryShutdown resumes a quiescent interrupted worker after
// graceful host shutdown without treating the shutdown as an operator pause.
func TestGuidedUseCaseRecoveryShutdown(t *testing.T) {
	operations, session := guidedAdmittedFixture(t)
	marker := filepath.Join(t.TempDir(), "worker-started")
	operations.workflowWorker = []string{"sh", "-c", `if [ ! -f "$1" ]; then printf ready > "$1"; while :; do sleep 0.02; done; fi; printf 'Delivered\n' > result.txt; git add result.txt; git -c user.name=Test -c user.email=test@example.com commit -qm 'Recovered task result'`, "worker", marker}
	old := newGuidedWorkHost(t.Context())
	t.Cleanup(old.close)

	if err := old.activate(operations, session.Guided.Work.Invocation.ID); err != nil {
		t.Fatal(err)
	}

	telegramWorkWaitFile(t, marker)
	old.close()

	interrupted := assertGuidedNativeAdmission(t, operations, session)
	if interrupted.Status != "paused" || interrupted.StopIntent != "" || interrupted.Outcome != "interrupted" {
		t.Fatal("shutdown did not retain a quiescent interrupted checkpoint", interrupted.Status, interrupted.Outcome)
	}

	host := startGuidedRecoveryHost(t, operations)
	current := waitGuidedRecovery(t, operations, session, func(current authoringcore.Session) bool {
		return guidedRecoveryReviewReady(t.Context(), operations, current)
	})

	host.close()

	after := assertGuidedNativeAdmission(t, operations, current)
	if after.ID != interrupted.ID || after.Policy != interrupted.Policy || serviceWorkTaskAttempts(after) != 2 {
		t.Fatal("graceful restart lost the original invocation or spent attempt")
	}
}

func startGuidedRecoveryHost(t *testing.T, operations service) *guidedWorkHost {
	t.Helper()

	host := newGuidedWorkHost(t.Context())
	t.Cleanup(host.close)

	operations.guidedWorkHost = host
	host.start(func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy })

	return host
}
