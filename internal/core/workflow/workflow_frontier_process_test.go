package workflow

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowFrontierProcessResult struct {
	ID      string
	Claimed bool
	Owned   bool
}

// TestWorkflowFrontierProcess is an owned manual-runner fixture. It performs
// real admission and immutable handoff before reporting its held claim.
func TestWorkflowFrontierProcess(t *testing.T) {
	dir := os.Getenv("CHR_FRONTIER_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")

	var request StartRequest
	if err := json.Unmarshal([]byte(os.Getenv("CHR_FRONTIER_REQUEST")), &request); err != nil {
		t.Fatal(err)
	}

	run, claim, err := workflowFrontierProcessClaim(t, store, request)
	result := workflowFrontierProcessResult{ID: run.ID, Claimed: err == nil, Owned: errors.Is(err, ErrWorkflowOwned) || errors.Is(err, ErrWorkflowCapacity)}
	writeWorkflowFrontierProcess(t, result)

	if err != nil && !result.Owned {
		t.Fatal(err)
	}

	if claim != nil {
		waitWorkflowClaimFile(t, os.Getenv("CHR_FRONTIER_RELEASE"))

		if err := store.CheckpointWorkflow(t.Context(), claim); err != nil {
			t.Fatal(err)
		}
	}
}

func writeWorkflowFrontierProcess(t *testing.T, result workflowFrontierProcessResult) {
	t.Helper()

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(os.Getenv("CHR_FRONTIER_RESULT"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func workflowFrontierProcessClaim(t *testing.T, store *Store, request StartRequest) (Invocation, *Claim, error) {
	t.Helper()

	run, err := store.StartWorkflow(t.Context(), request)
	if err != nil {
		return run.Invocation, nil, err
	}

	claim, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		return run.Invocation, nil, err
	}

	_, err = store.PrepareWorkflowManualTask(t.Context(), claim, nil)

	return run.Invocation, claim, err
}

// TestWorkflowFrontierRace proves competing processes cannot bypass one task,
// slice or ordered set/use-case owner even when repository capacity is two.
func TestWorkflowFrontierRace(t *testing.T) {
	for _, extent := range []string{"ticket", "slice", "slice_set", "use_case"} {
		t.Run(extent, func(t *testing.T) {
			store, hierarchy, request := workflowClaimFixture(t)
			writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: 2\n")

			fixture := t.TempDir()
			release := filepath.Join(fixture, "release")
			paths := []string{filepath.Join(fixture, "first.json"), filepath.Join(fixture, "second.json")}
			first := workflowFrontierProcess(t, store, request, paths[0], release)
			request.RequestKey = "competing"
			request.Selection.Selector = workflowClaimSelector(hierarchy, extent)

			second := workflowFrontierProcess(t, store, request, paths[1], release)
			runWorkflowFrontierProcesses(t, store, []*exec.Cmd{first, second}, paths, release)
		})
	}
}

func runWorkflowFrontierProcesses(t *testing.T, store *Store, commands []*exec.Cmd, paths []string, release string) {
	t.Helper()

	for _, command := range commands {
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
	}

	checkWorkflowFrontierProcesses(t, store, paths)

	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	for _, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatal("owned frontier fixture", err)
		}
	}
}

func workflowFrontierProcess(t *testing.T, store *Store, request StartRequest, result, release string) *exec.Cmd {
	t.Helper()

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowFrontierProcess$")

	command.Env = append(os.Environ(), "CHR_FRONTIER_DIR="+store.layout.ProjectDir(), "CHR_FRONTIER_REQUEST="+string(data), "CHR_FRONTIER_RESULT="+result, "CHR_FRONTIER_RELEASE="+release)

	t.Cleanup(func() {
		if command.Process != nil && command.ProcessState == nil {
			if err := command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				t.Error(err)
			}

			var exited *exec.ExitError
			if err := command.Wait(); err != nil && !errors.As(err, &exited) {
				t.Error(err)
			}
		}
	})

	return command
}

func checkWorkflowFrontierProcesses(t *testing.T, store *Store, paths []string) {
	t.Helper()

	claimed, owned := 0, 0

	for _, path := range paths {
		result := readWorkflowFrontierProcess(t, store, path)

		if result.Claimed {
			claimed++
		}

		if result.Owned {
			owned++
		}
	}

	if claimed != 1 || owned != 1 {
		t.Fatalf("competing frontiers: claimed=%d, owned=%d", claimed, owned)
	}

	if lock, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: store.layout.ProjectDir(), Operation: AgentOperationDevelopment, Owner: "legacy"}); !errors.Is(err, ErrWorkflowCapacity) {
		workflowTestReleaseLegacy(t, lock)

		t.Fatal("legacy start bypassed a held manual claim", err)
	}
}

func readWorkflowFrontierProcess(t *testing.T, store *Store, path string) workflowFrontierProcessResult {
	t.Helper()
	waitWorkflowClaimFile(t, path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var result workflowFrontierProcessResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result.Claimed {
		run := workflowTestRead(t, store, result.ID)
		if len(run.Attempts) != 1 || run.Status != workflowstate.InvocationRunning || run.Attempts[0].InputRef == "" {
			t.Fatal("held manual claim lost its immutable input")
		}
	}

	return result
}
