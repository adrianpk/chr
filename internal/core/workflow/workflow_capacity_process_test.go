package workflow

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// TestWorkflowCapacityProcess is an owned process barrier. It holds the actual
// repository reservation until the parent explicitly completes its assertions.
func TestWorkflowCapacityProcess(t *testing.T) {
	dir := os.Getenv("CHR_CAPACITY_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir, Namespace: os.Getenv("CHR_CAPACITY_NAMESPACE")})
	if err != nil {
		t.Fatal(err)
	}

	slot, guard, err := NewWorkflowStore(layout, "capacity-alias").acquireWorkflowWorkerSlot(t.Context())
	if err != nil && !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal(err)
	}

	value := "waiting"
	if slot != nil {
		value = strconv.Itoa(workflowWorkerSlotNumber(slot.Name()))
	}

	if err := os.WriteFile(os.Getenv("CHR_CAPACITY_RESULT"), []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}

	if slot == nil {
		return
	}

	waitWorkflowClaimFile(t, os.Getenv("CHR_CAPACITY_RELEASE"))

	if err := errors.Join(slot.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowCapacityOverlap proves cross-process overlap through capacity eight,
// serial admission at capacity one, and rejection beyond the shared limit.
func TestWorkflowCapacityOverlap(t *testing.T) {
	for _, capacity := range []int{1, 2, 8} {
		t.Run(strconv.Itoa(capacity), func(t *testing.T) { testWorkflowCapacityOverlap(t, capacity) })
	}
}

func testWorkflowCapacityOverlap(t *testing.T, capacity int) {
	t.Helper()

	store, _, _ := workflowClaimFixture(t)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: "+strconv.Itoa(capacity)+"\n")

	linked := workflowCapacityLinked(t, store, capacity)
	fixture := t.TempDir()
	release := filepath.Join(fixture, "release")
	commands := []*exec.Cmd{}

	slots := map[string]bool{}

	for index := 0; index <= capacity; index++ {
		result := filepath.Join(fixture, strconv.Itoa(index))

		command := workflowCapacityProcessFixture(t, []*Store{store, linked}[index%2], result, release)
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}

		commands = append(commands, command)

		waitWorkflowClaimFile(t, result)

		data, err := os.ReadFile(result)
		if err != nil {
			t.Fatal(err)
		}

		if (string(data) == "waiting") != (index >= capacity) {
			t.Fatal("subprocess admission exceeded or failed capacity", index, string(data))
		}

		if index < capacity && slots[string(data)] {
			t.Fatal("two live processes received the same worker slot", string(data))
		}

		slots[string(data)] = true
	}
	// These barriers are observed before releasing any admitted process. A
	// single in-process mutex or an eventual race-detector pass cannot prove this.
	if _, err := acquireWorkflowCapacity(t.Context(), store.layout.ProjectDir()); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("legacy serial start bypassed live reservations", err)
	}

	finishWorkflowClaimProcesses(t, commands, release)

	slot, guard := workflowTestSlot(t, store)
	if err := errors.Join(slot.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}
}

func workflowCapacityProcessFixture(t *testing.T, store *Store, result, release string) *exec.Cmd {
	t.Helper()

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowCapacityProcess$")

	command.Env = append(os.Environ(), "CHR_CAPACITY_DIR="+store.layout.ProjectDir(), "CHR_CAPACITY_NAMESPACE="+store.layout.Namespace(), "CHR_CAPACITY_RESULT="+result, "CHR_CAPACITY_RELEASE="+release)

	t.Cleanup(func() {
		if command.Process == nil || command.ProcessState != nil {
			return
		}

		if err := command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			t.Error(err)
		}

		var exited *exec.ExitError
		if err := command.Wait(); err != nil && !errors.As(err, &exited) {
			t.Error(err)
		}
	})

	return command
}

func workflowCapacityLinked(t *testing.T, store *Store, workers int) *Store {
	t.Helper()

	directory := filepath.Join(t.TempDir(), "linked")
	workflowSelectionGit(t, store, "worktree", "add", "-qb", "capacity-process-linked", directory, "HEAD")
	writeWorkspaceConfig(t, directory, "work:\n  max_active_workers: "+strconv.Itoa(workers)+"\n")

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: directory, Namespace: "other"})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, "alias")
}

// TestWorkflowCapacityOrphan proves a live group retains only its assigned slot
// after every inherited descriptor is closed, while another independent slot runs.
func TestWorkflowCapacityOrphan(t *testing.T) {
	store, _, _ := workflowClaimFixture(t)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: 2\n")

	slot, guard := workflowTestSlot(t, store)
	fixture := t.TempDir()
	started, release := filepath.Join(fixture, "started"), filepath.Join(fixture, "release")
	worker := workflowCapacityOrphanFixture(t, store, slot, started, release)

	if err := errors.Join(slot.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}

	other, otherGuard := workflowTestSlot(t, store)
	if other.Name() == slot.Name() {
		t.Fatal("orphan's slot was reused")
	}

	if _, _, err := store.acquireWorkflowWorkerSlot(t.Context()); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatal("live orphan was omitted from capacity", err)
	}

	if err := os.WriteFile(release, []byte("finish"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := worker.Wait(); err != nil {
		t.Fatal(err)
	}

	if err := errors.Join(other.Close(), otherGuard.Close()); err != nil {
		t.Fatal(err)
	}

	replacement, replacementGuard := workflowTestSlot(t, store)
	if replacement.Name() != slot.Name() {
		t.Fatal("dead orphan's slot was not reclaimed")
	}

	if err := errors.Join(replacement.Close(), replacementGuard.Close()); err != nil {
		t.Fatal(err)
	}
}

func workflowCapacityOrphanFixture(t *testing.T, store *Store, slot *os.File, started, release string) *exec.Cmd {
	t.Helper()

	worker := exec.CommandContext(t.Context(), "bash", "-c", `printf started > "$1"; while [ ! -e "$2" ]; do sleep 0.05; done`, "worker", started, release)

	worker.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { stopWorkflowCapacityOrphan(t, worker) })
	waitWorkflowClaimFile(t, started)

	identity, err := workflowProcessIdentity(worker.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}

	observation := Process{Identity: identity, GroupID: identity.PID, CapacitySlot: workflowWorkerSlotNumber(slot.Name())}
	if err := registerWorkflowProcessCapacity(t.Context(), &Claim{store: store}, observation); err != nil {
		t.Fatal(err)
	}

	return worker
}

func stopWorkflowCapacityOrphan(t *testing.T, worker *exec.Cmd) {
	t.Helper()

	if worker.ProcessState != nil {
		return
	}

	if err := syscall.Kill(-worker.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		t.Error(err)
	}

	var exited *exec.ExitError
	if err := worker.Wait(); err != nil && !errors.As(err, &exited) {
		t.Error(err)
	}
}
