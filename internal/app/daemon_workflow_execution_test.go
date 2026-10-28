package app

import (
	"bytes"
	"context"
	"errors"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	daemoncore "forge.adrianpk.com/chr/internal/core/harness/daemon"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

// Actual foreground startup executes a retained task using an owned local
// Codex fixture, while every unselected ready task stays outside authorization.
func TestDaemonWorkflowAuthorizedStartup(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	project, root := operations.projectContext.ProjectDir, t.TempDir()
	id := workflowCLIManualStart(t, operations, hierarchy)
	workerDir := t.TempDir()

	worker := "#!/bin/sh\nset -eu\ncat >/dev/null\nprintf 'Delivered\\n' > result.txt\ngit add result.txt\ngit -c user.name=Test -c user.email=test@example.com commit -qm 'Daemon task'\n"
	if err := os.WriteFile(filepath.Join(workerDir, codex.Program), []byte(worker), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", workerDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	addProject(t, root, "authorized", project)

	var output bytes.Buffer
	if err := run(t.Context(), []string{"daemon", "start", "--dev-root", root}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err, output.String())
	}

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Total != 1 || status.ValidatedTasks != 1 || status.ConfirmedTickets != 0 || status.TotalAttempts != 1 || !strings.Contains(output.String(), "workflow_checkpoint_steps: 1") {
		t.Fatal("daemon startup lost exact retained authorization", status, err, output.String())
	}
}

// Pause/cancel interrupts the daemon's real owned group, checkpoints evidence,
// releases capacity only after quiescence, and is preserved on another pass.
func TestDaemonWorkflowActiveStop(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) { checkDaemonWorkflowActiveStop(t, intent) })
	}
}

func checkDaemonWorkflowActiveStop(t *testing.T, intent string) {
	t.Helper()

	operations, _, request, _ := serviceWorkNativeFixture(t)

	start, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	marker := filepath.Join(t.TempDir(), "ready")
	operations.workflowWorker = []string{"sh", "-c", "printf ready > \"$1\"; while :; do sleep 1; done", "worker", marker}

	var output bytes.Buffer

	runner := daemonWorkflowFixture(t, &output)
	result := daemonForegroundLoopResult{}
	done := runDaemonWorkflowTest(t, runner, operations, &result)
	daemonWorkflowMarker(t, marker)

	status, err := operations.WorkflowStatus(t.Context(), start.Invocation.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ControlWorkExtent(t.Context(), WorkflowControlRequest{Action: intent, ID: status.ID, ExpectedRevision: status.Revision}); err != nil {
		t.Fatal(err)
	}

	waitDaemonWorkflowTest(t, done)
	checkDaemonWorkflowStopCheckpoint(t, operations, start.Invocation, intent)

	result = daemonForegroundLoopResult{}
	if err := runner.workflows(t.Context(), operations, &result); err != nil || result.WorkflowSteps != 0 || result.WorkflowsWaiting != 1 {
		t.Fatal("daemon restarted stopped execution", result, err)
	}
}

func checkDaemonWorkflowStopCheckpoint(t *testing.T, operations service, before workflowstate.Invocation, intent string) {
	t.Helper()

	current, err := operations.ReadWorkflow(t.Context(), before.ID)

	expected := "paused"
	if intent == "cancel" {
		expected = "cancelled"
	}

	if err != nil || string(current.Status) != expected || current.Cursor != 0 || current.ConfirmedCursor != 0 || len(current.Attempts) != 1 || current.Attempts[0].CheckpointHash == "" || current.Policy != before.Policy {
		t.Fatal("daemon stop lost checkpoint, budget or ownership", current, err)
	}
}

func runDaemonWorkflowTest(t *testing.T, runner daemonWorkflowRunner, operations service, result *daemonForegroundLoopResult) <-chan error {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan error, 1)
	go func() {
		done <- runner.workflows(ctx, operations, result)

		close(done)
	}()

	t.Cleanup(func() { cancel(); waitDaemonWorkflowTest(t, done) })

	return done
}

func waitDaemonWorkflowTest(t *testing.T, done <-chan error) {
	t.Helper()

	select {
	case err := <-done:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(10 * time.Second):
		t.Error("owned daemon worker did not checkpoint")
	}
}

// Coordinator stop yields after an accepted worker checkpoint. Heartbeats
// continue during that worker, without turning daemon stop into cancellation.
func TestDaemonWorkflowCoordinatorStop(t *testing.T) {
	operations, _, request, _ := serviceWorkNativeFixture(t)

	start, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	barrier := t.TempDir()
	worker := "printf ready > \"$1/ready\"; while [ ! -f \"$1/release\" ]; do sleep 0.05; done; printf 'Delivered\\n' > result.txt; git add result.txt; git -c user.name=Test -c user.email=test@example.com commit -qm result"
	operations.workflowWorker = []string{"sh", "-c", worker, "worker", barrier}

	var output bytes.Buffer

	runner := daemonWorkflowFixture(t, &output)
	ctx, heartbeatCancel := context.WithCancel(t.Context())

	heartbeatWatch := watchDaemonHeartbeat(ctx, runner.repository, heartbeatCancel)

	finish := heartbeatWatch.Finish

	t.Cleanup(func() {
		if err := finish(); err != nil {
			t.Error(err)
		}
	})

	operations.ctx = ctx
	result := daemonForegroundLoopResult{}
	done := runDaemonWorkflowTest(t, runner, operations, &result)
	daemonWorkflowMarker(t, filepath.Join(barrier, "ready"))
	checkDaemonWorkflowHeartbeat(t, runner.repository)

	if err := runner.repository.RequestStop(t.Context(), time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(barrier, "release"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	waitDaemonWorkflowTest(t, done)

	status, err := operations.WorkflowStatus(t.Context(), start.Invocation.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || !result.StopRequested || status.ValidatedTasks != 1 || status.StopIntent != "" || status.ConfirmedTickets != 0 {
		t.Fatal("coordinator stop discarded accepted progress or cancelled authorization", status, result, err)
	}
}

func checkDaemonWorkflowHeartbeat(t *testing.T, repository daemoncore.Repository) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		state, err := repository.Get(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		if state.HeartbeatAt.After(*state.StartedAt) {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("daemon heartbeat stopped during owned work")
}

// A real child exits while holding the manual descriptor. The daemon fences
// that dead owner, preserves its attempt and budget, then validates one retry.
func TestDaemonWorkflowRecovery(t *testing.T) {
	if project := os.Getenv("CHR_DAEMON_TEST_PROJECT"); project != "" {
		daemonWorkflowCrashOwner(t, project)

		return
	}

	for _, intent := range []string{"", "pause", "cancel"} {
		name := intent
		if name == "" {
			name = "retry"
		}

		t.Run(name, func(t *testing.T) { checkDaemonWorkflowRecovery(t, intent) })
	}
}

func checkDaemonWorkflowRecovery(t *testing.T, intent string) {
	t.Helper()

	operations, _, request, _ := serviceWorkNativeFixture(t)

	start, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestDaemonWorkflowRecovery$")

	child.Env = append(os.Environ(), "CHR_DAEMON_TEST_PROJECT="+operations.projectContext.ProjectDir, "CHR_DAEMON_TEST_PROJECT_ID="+operations.projectContext.ProjectID, "CHR_DAEMON_TEST_INVOCATION="+start.Invocation.ID)
	if output, err := child.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}

	if intent != "" {
		if _, err := operations.StopWorkflow(t.Context(), start.Invocation.ID, intent); err != nil {
			t.Fatal(err)
		}
	}

	var output bytes.Buffer

	runner := daemonWorkflowFixture(t, &output)

	result := daemonForegroundLoopResult{}
	if err := runner.workflows(t.Context(), operations, &result); err != nil {
		t.Fatal(err)
	}

	checkDaemonWorkflowRecoveryCheckpoint(t, operations, start.Invocation, intent)
}

func checkDaemonWorkflowRecoveryCheckpoint(t *testing.T, operations service, before workflowstate.Invocation, intent string) {
	t.Helper()

	current, err := operations.ReadWorkflow(t.Context(), before.ID)

	cursor, attempts := 0, 1
	if intent == "" {
		cursor, attempts = 1, 2
	}

	if err != nil || current.Cursor != cursor || len(current.Attempts) != attempts || current.ConfirmedCursor != 0 || current.Policy != before.Policy {
		t.Fatal("dead-owner recovery changed progress or budget", current, err)
	}

	if current.Attempts[0].Outcome != "recovered" || current.Attempts[0].CheckpointHash == "" || current.StopIntent != intent {
		t.Fatal("dead-owner recovery lost checkpoint or operator stop", current)
	}
}

func daemonWorkflowCrashOwner(t *testing.T, project string) {
	t.Helper()

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: project})
	if err != nil {
		t.Fatal(err)
	}

	store := workflowcore.NewWorkflowStore(layout, os.Getenv("CHR_DAEMON_TEST_PROJECT_ID"))

	claim, err := store.ClaimWorkflow(t.Context(), os.Getenv("CHR_DAEMON_TEST_INVOCATION"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.PrepareWorkflowManualTask(t.Context(), claim, nil); err != nil {
		t.Fatal(err)
	}

	os.Exit(0)
}

// Stale daemon continuation guards return current status without writing any
// authorization, recovery, acceptance or execution checkpoint.
func TestDaemonWorkflowRevision(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	paused, err := operations.StopWorkflow(t.Context(), id, "pause", status.Revision)
	if err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	progress, err := operations.ResumeRetainedWork(t.Context(), id, status.Revision)
	if !errors.Is(err, workflowstate.ErrConflict) || progress.Status.Revision != paused.Revision || progress.Attempted || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale daemon revision changed execution or hid current state", progress, err)
	}
}
