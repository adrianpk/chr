package core

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
)

// TestWorkflowCoordinatorClone rejects copied authority before an invocation
// exists, including legacy starts and a copied host-owned registration file.
func TestWorkflowCoordinatorClone(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	workflowSelectionGit(t, store, "add", "ops")
	workflowSelectionCommit(t, store)

	clone := workflowCoordinatorClone(t, store)

	checkWorkflowCoordinatorClone(t, clone, request)

	common, err := workflowRepositoryDirectory(t.Context(), store.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(common, "chr-workflow-coordinator.json"))
	if err != nil {
		t.Fatal(err)
	}

	copyDir, err := workflowRepositoryDirectory(t.Context(), clone.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(copyDir, "chr-workflow-coordinator.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	checkWorkflowCoordinatorClone(t, clone, request)

	if err := os.Remove(filepath.Join(clone.layout.Root(), workflowCoordinatorPath)); err != nil {
		t.Fatal(err)
	}

	checkWorkflowCoordinatorClone(t, clone, request)
}

func workflowCoordinatorClone(t *testing.T, store *WorkflowStore) *WorkflowStore {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "clone")

	command := exec.CommandContext(t.Context(), "git", "clone", "-q", store.layout.ProjectDir(), dir)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("clone: %v: %s", err, output)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir, Namespace: store.layout.Namespace()})
	if err != nil {
		t.Fatal(err)
	}

	return NewWorkflowStore(layout, store.projectID)
}

func checkWorkflowCoordinatorClone(t *testing.T, store *WorkflowStore, request WorkflowStartRequest) {
	t.Helper()

	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, ErrWorkflowCoordinator) {
		t.Fatal("clone admitted a native invocation", err)
	}

	if lock, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment, Owner: "legacy"}); !errors.Is(err, ErrWorkflowCoordinator) {
		workflowTestReleaseLegacy(t, lock)

		t.Fatal("clone admitted a legacy runner", err)
	}

	if _, err := store.Create(t.Context(), "Foreign task", workflowTestTask("PRJ-99")); !errors.Is(err, ErrWorkflowCoordinator) {
		t.Fatal("clone silently registered copied authority", err)
	}

	if _, err := os.Stat(filepath.Join(store.layout.Root(), "run")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("rejected clone retained invocation files", err)
	}
}

// TestWorkflowCoordinatorTarget requires the canonical input to retain the
// registered owner. A different namespace cannot register a second target owner.
func TestWorkflowCoordinatorTarget(t *testing.T) {
	cases := []struct {
		name   string
		marker bool
	}{{"missing", false}, {"foreign namespace", true}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _, request := workflowClaimFixture(t)
			request.Selection.Target.Commit = changeWorkflowCoordinatorTarget(t, store, tc.marker)

			run := workflowTestStart(t, store, request)
			if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowCoordinator) {
				t.Fatal("unregistered target admitted a worker", err)
			}

			if current := workflowTestRead(t, store, run.Invocation.ID); len(current.Attempts) != 0 || current.Cursor != 0 {
				t.Fatal("unregistered target allocated an attempt or advanced work")
			}
		})
	}
}

func changeWorkflowCoordinatorTarget(t *testing.T, store *WorkflowStore, foreign bool) string {
	t.Helper()

	markerPath := filepath.Join(store.layout.Root(), workflowCoordinatorPath)

	original, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}

	workflowSelectionGit(t, store, "rm", filepath.Join("ops", "default", workflowCoordinatorPath))

	path := filepath.Join(store.layout.ProjectDir(), "ops", "other", workflowCoordinatorPath)

	if foreign {
		marker := workflowCoordinator{Schema: "chr-coordinator/v1", ID: contenthash.Sum([]byte("foreign"))}

		data, err := json.Marshal(marker)
		if err != nil {
			t.Fatal(err)
		}

		if err := opsfile.WriteAtomic(path, data); err != nil {
			t.Fatal(err)
		}

		workflowSelectionGit(t, store, "add", "ops/other/coordinator.json")
	}

	workflowSelectionGit(t, store, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "Changed coordinator")

	commit := workflowSelectionGit(t, store, "rev-parse", "HEAD")
	if err := opsfile.WriteAtomic(markerPath, original); err != nil {
		t.Fatal(err)
	}

	if foreign {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}

	return commit
}

func workflowTestReleaseLegacy(t *testing.T, lock *AgentOperationLock) {
	t.Helper()

	if lock == nil {
		return
	}

	t.Cleanup(func() {
		if err := lock.Release(); err != nil {
			t.Error(err)
		}
	})
}

// TestWorkflowCoordinatorMove preserves authority when the actual common Git
// directory moves, while TestWorkflowCoordinatorClone covers copied directories.
func TestWorkflowCoordinatorMove(t *testing.T) {
	store, _, request := workflowClaimFixture(t)

	destination := filepath.Join(t.TempDir(), "moved")
	if err := os.Rename(store.layout.ProjectDir(), destination); err != nil {
		t.Fatal(err)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: destination, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	moved := NewWorkflowStore(layout, store.projectID)
	run := workflowTestStart(t, moved, request)

	claim := workflowTestClaim(t, moved, run.Invocation.ID)
	if err := moved.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowCoordinatorLoss prevents a lost registration from creating a new
// pool or admitting a result from an otherwise valid retained credential.
func TestWorkflowCoordinatorLoss(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)

	common, err := workflowRepositoryDirectory(t.Context(), store.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(common, "chr-workflow-coordinator.json")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if err := store.CheckpointWorkflow(t.Context(), claim); !errors.Is(err, ErrWorkflowCoordinator) {
		t.Fatal("unregistered owner changed domain state", err)
	}

	if _, err := store.Create(t.Context(), "Replacement pool", workflowTestTask("PRJ-99")); !errors.Is(err, ErrWorkflowCoordinator) {
		t.Fatal("lost registration silently rebound authority", err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("a replacement registration was created", err)
	}

	retained := workflowTestRead(t, store, run.Invocation.ID)
	if retained.Status != "running" || len(retained.Attempts) != 1 {
		t.Fatal("registration loss changed the live claim")
	}
}

// TestWorkflowCoordinatorCandidate prevents a worker from removing the owner
// registration from an otherwise clean candidate that passes its task gate.
func TestWorkflowCoordinatorCandidate(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)
	worker := []string{"sh", "-c", "git rm ops/default/coordinator.json && git -c user.name=Test -c user.email=test@example.com commit -qm 'Remove coordinator'"}

	result, err := store.ExecuteWorkflowTask(t.Context(), claim, worker)
	if err == nil || result.Successful {
		t.Fatal("a coordinator removal became accepted task proof", err)
	}

	if current := workflowTestRead(t, store, run.Invocation.ID); current.Cursor != 0 || current.Status == "completed" {
		t.Fatal("coordinator removal advanced the frontier")
	}
}

// TestWorkflowLegacyCapacity rejects invalid shared policy before a legacy
// caller obtains capacity, so changing adapters cannot create another pool.
func TestWorkflowLegacyCapacity(t *testing.T) {
	store, _, _ := workflowClaimFixture(t)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: 9\n")

	lock, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment})
	if err == nil {
		workflowTestReleaseLegacy(t, lock)
		t.Fatal("legacy execution admitted invalid repository capacity")
	}

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: 2\n")

	lock, err = AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment})
	if err != nil {
		t.Fatal("rejected legacy start retained capacity", err)
	}

	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
}
