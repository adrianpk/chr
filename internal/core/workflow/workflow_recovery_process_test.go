package workflow

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowRecoveryProcess is a host fixture that can die while its supervised
// worker closes fd 3 and waits. The surviving supervisor must retain capacity.
func TestWorkflowRecoveryProcess(t *testing.T) {
	dir := os.Getenv("CHR_RECOVERY_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")

	claim, err := store.ClaimWorkflow(t.Context(), os.Getenv("CHR_RECOVERY_ID"))
	if err != nil {
		t.Fatal(err)
	}

	configureWorkflowProcessStartCrash(t, claim)

	worker := []string{"sh", "-c", "exec 3<&-; printf 'Surviving worker\\n'; printf 'started' > \"$1\"; while [ ! -e \"$2\" ]; do sleep 0.05; done", "worker", os.Getenv("CHR_RECOVERY_STARTED"), os.Getenv("CHR_RECOVERY_RELEASE")}
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, worker); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowRecoveryChild proves host loss and lease expiry cannot free a live
// worker, across linked aliases and legacy starts. Dead replacement is fenced.
func TestWorkflowRecoveryChild(t *testing.T) {
	for _, mode := range []string{"host loss", "supervisor loss"} {
		t.Run(mode, func(t *testing.T) { testWorkflowRecoveryChild(t, mode) })
	}
}

func testWorkflowRecoveryChild(t *testing.T, mode string) {
	t.Helper()

	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Policy.WorkerTimeoutSeconds = 3
	run := workflowTestStart(t, store, request)
	fixture := t.TempDir()
	started, release := filepath.Join(fixture, "started"), filepath.Join(fixture, "release")
	host := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowRecoveryProcess$")

	host.Env = append(os.Environ(), "CHR_RECOVERY_DIR="+store.layout.ProjectDir(), "CHR_RECOVERY_ID="+run.Invocation.ID, "CHR_RECOVERY_STARTED="+started, "CHR_RECOVERY_RELEASE="+release)
	if err := host.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { stopWorkflowRecoveryHost(t, host, release, store, run.Invocation.ID) })
	waitWorkflowClaimFile(t, started)

	if err := host.Process.Kill(); err != nil {
		t.Fatal(err)
	}

	var exited *exec.ExitError
	if err := host.Wait(); !errors.As(err, &exited) {
		t.Fatalf("host crash: %v", err)
	}

	if mode == "supervisor loss" {
		killWorkflowFixtureSupervisor(t, store, run.Invocation.ID)
	}

	alias := workflowRecoveryAlias(t, store, request.Selection.Target.Commit)
	checkWorkflowRecoveryReplacement(t, alias, run.Invocation.ID, release)
}

func assertWorkflowRecoveryCapacity(t *testing.T, store *Store, id string) {
	t.Helper()

	if _, err := store.ReconcileWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatalf("live worker replaced: %v", err)
	}

	if _, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment}); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatalf("legacy bypassed surviving child: %v", err)
	}
}

func waitWorkflowRecovery(t *testing.T, store *Store, id string) Invocation {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		run, err := store.ReconcileWorkflow(t.Context(), id)
		if err == nil {
			return run
		}

		if !errors.Is(err, ErrWorkflowCapacity) {
			t.Fatal(err)
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("owned worker did not become quiescent")

	return Invocation{}
}

func stopWorkflowRecoveryHost(t *testing.T, host *exec.Cmd, release string, store *Store, id string) {
	t.Helper()

	if err := os.WriteFile(release, []byte("cleanup"), 0o600); err != nil {
		t.Error(err)
	}

	if host.Process != nil && host.ProcessState == nil {
		if err := host.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			t.Error(err)
		}

		var exited *exec.ExitError
		if err := host.Wait(); err != nil && !errors.As(err, &exited) {
			t.Error(err)
		}
	}

	stopWorkflowRecoveryChildren(t, store, id, release)
}

func checkWorkflowRecoveryReplacement(t *testing.T, store *Store, id, release string) {
	t.Helper()

	retained, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || len(retained.Attempts[0].Processes) != 1 {
		t.Fatal("host lost process identity")
	}

	expireWorkflowRecoveryLease(t, store, id)
	assertWorkflowRecoveryCapacity(t, store, id)

	if err := os.WriteFile(release, []byte("release"), 0o600); err != nil {
		t.Fatal(err)
	}

	recovered := waitWorkflowRecovery(t, store, id)
	if recovered.Status != "paused" || recovered.Attempts[0].Outcome != "recovered" || recovered.Attempts[0].CheckpointRef == "" || recovered.Cursor != 0 {
		t.Fatal("dead owner was not checkpointed")
	}

	checkWorkflowReplacementFence(t, store, id, retained.Attempts[0])
}

func checkWorkflowReplacementFence(t *testing.T, store *Store, id string, previous Attempt) {
	t.Helper()

	replacement := workflowTestClaim(t, store, id)
	if replacement.Generation != previous.Generation+1 {
		t.Fatal("replacement lost generation")
	}

	next, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || next.Attempts[1].Replaces != previous.ID {
		t.Fatal("replacement lost history")
	}
}

func workflowRecoveryAlias(t *testing.T, store *Store, commit string) *Store {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "linked")
	workflowSelectionGit(t, store, "worktree", "add", "-qb", "recovery-alias", dir, commit)

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, "alias")
}

func configureWorkflowProcessStartCrash(t *testing.T, claim *Claim) {
	t.Helper()

	if os.Getenv("CHR_RECOVERY_CRASH_START") == "" {
		return
	}

	claim.store.afterReplace = func(index int) error {
		if index != 0 {
			return nil
		}

		run, err := claim.store.readWorkflowInvocation(claim.InvocationID)
		if err != nil {
			return err
		}

		if len(run.Attempts[len(run.Attempts)-1].Processes) > 0 {
			os.Exit(17)
		}

		return nil
	}
}

// TestWorkflowRecoveryStart kills the host after journaling process birth but
// before launch permission. Replay retains input and never executes the worker.
func TestWorkflowRecoveryStart(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	fixture := t.TempDir()
	started, release := filepath.Join(fixture, "started"), filepath.Join(fixture, "release")
	host := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowRecoveryProcess$")

	host.Env = append(os.Environ(), "CHR_RECOVERY_DIR="+store.layout.ProjectDir(), "CHR_RECOVERY_ID="+run.Invocation.ID, "CHR_RECOVERY_STARTED="+started, "CHR_RECOVERY_RELEASE="+release, "CHR_RECOVERY_CRASH_START=1")
	output, err := host.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 17 {
		t.Fatalf("startup crash: %v: %s", err, output)
	}

	recovered := waitWorkflowRecovery(t, store, run.Invocation.ID)
	if len(recovered.Attempts) != 1 || recovered.Attempts[0].InputRef == "" || recovered.Attempts[0].CheckpointRef == "" {
		t.Fatal("startup replay lost retained input")
	}

	if _, err := os.Stat(started); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("worker ran before journaled start permission")
	}
}

// TestWorkflowRecoveryTimeout enforces the original finite worker lifetime even
// when the application host disappears before it can cancel or checkpoint.
func TestWorkflowRecoveryTimeout(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Policy.WorkerTimeoutSeconds = 1
	run := workflowTestStart(t, store, request)
	fixture := t.TempDir()
	started, release := filepath.Join(fixture, "started"), filepath.Join(fixture, "release")
	host := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowRecoveryProcess$")

	host.Env = append(os.Environ(), "CHR_RECOVERY_DIR="+store.layout.ProjectDir(), "CHR_RECOVERY_ID="+run.Invocation.ID, "CHR_RECOVERY_STARTED="+started, "CHR_RECOVERY_RELEASE="+release)
	if err := host.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { stopWorkflowRecoveryHost(t, host, release, store, run.Invocation.ID) })
	waitWorkflowClaimFile(t, started)

	if err := host.Process.Kill(); err != nil {
		t.Fatal(err)
	}

	var exited *exec.ExitError
	if err := host.Wait(); !errors.As(err, &exited) {
		t.Fatalf("host crash: %v", err)
	}

	recovered := waitWorkflowRecovery(t, store, run.Invocation.ID)
	if recovered.Status != "paused" || recovered.Attempts[0].CheckpointRef == "" {
		t.Fatal("timeout discarded evidence")
	}
}

func killWorkflowFixtureSupervisor(t *testing.T, store *Store, id string) {
	t.Helper()

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	identity := run.Attempts[0].Processes[0].Identity

	alive, err := workflowProcessAlive(identity)
	if err != nil || !alive {
		t.Fatal("owned supervisor is not alive")
	}

	if err := syscall.Kill(identity.PID, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
}

func stopWorkflowRecoveryChildren(t *testing.T, store *Store, id, release string) {
	t.Helper()
	// Test contexts are cancelled before cleanup. Cleanup owns only these fixtures.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		t.Error(err)

		return
	}

	for _, attempt := range run.Attempts {
		for _, process := range attempt.Processes {
			if err := stopWorkflowRecoveryGroup(process, release); err != nil {
				t.Error(err)
			}
		}
	}
}

func stopWorkflowRecoveryGroup(process Process, release string) error {
	alive, err := workflowProcessGroupAlive(process)
	if err != nil || !alive {
		return err
	}

	owned, err := workflowProcessAlive(process.Identity)
	if err != nil {
		return err
	}

	if !owned {
		owned, err = workflowFixtureGroupMatches(process.GroupID, release)
	}

	if err != nil {
		return err
	}

	if !owned {
		return workflowstate.ErrConflict
	}

	err = syscall.Kill(-process.GroupID, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}

	return err
}

func workflowFixtureGroupMatches(group int, release string) (bool, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		matches, err := workflowFixturePIDMatches(pid, group, release)
		if err != nil || matches {
			return matches, err
		}
	}

	return false, nil
}

func workflowFixturePIDMatches(pid, group int, release string) (bool, error) {
	stat, err := workflowReadProcStat(pid)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if stat.group != group {
		return false, nil
	}

	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return strings.Contains(string(data), release), err
}
