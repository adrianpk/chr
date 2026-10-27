package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowMaterializationProcess exits its owned process after a selected
// journal replacement. This exercises lost deferred cleanup as well as replay.
func TestWorkflowMaterializationProcess(t *testing.T) {
	dir := os.Getenv("CHR_MATERIALIZATION_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveOpsLayout(opscore.ResolveOpsLayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	stop, err := strconv.Atoi(os.Getenv("CHR_MATERIALIZATION_STOP"))
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	store.afterReplace = func(index int) error {
		if index == stop {
			os.Exit(17)
		}

		return nil
	}
	next := 0

	_, err = store.MaterializePlan(t.Context(), WorkflowMaterializationOptions{Plan: workflowPlanFixture(), ReserveTicketKey: func(context.Context) (string, error) { next++; return fmt.Sprintf("PRJ-%d", next), nil }})
	if err != nil {
		t.Fatal(err)
	}

	t.Fatal("crash replacement was not reached")
}

func crashWorkflowMaterialization(t *testing.T, store *WorkflowStore, index int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^TestWorkflowMaterializationProcess$")

	cmd.Env = append(os.Environ(), "CHR_MATERIALIZATION_DIR="+store.layout.ProjectDir(), "CHR_MATERIALIZATION_STOP="+strconv.Itoa(index))
	output, err := cmd.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("owned crash: %v: %s", err, output)
	}
}

// TestWorkflowMaterializationRecovery retries after every created ticket and
// proves exact keys, a single immutable receipt and allocator cleanup on restart.
func TestWorkflowMaterializationRecovery(t *testing.T) {
	for _, index := range []int{0, 3, 5, 7, 9, 11, 13, 16} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			store, opts := workflowMaterializationFixture(t)
			crashWorkflowMaterialization(t, store, index)

			opts.ReserveTicketKey = func(context.Context) (string, error) { return "", fmt.Errorf("retry allocated another key") }

			result, err := store.MaterializePlan(t.Context(), opts)
			if err != nil {
				t.Fatal(err)
			}

			checkMaterializationRetry(t, store, opts, result)

			if _, err := os.Stat(filepath.Join(store.layout.Root(), "lock", "id.lock")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("allocation guard remains: %v", err)
			}
		})
	}
}

func checkMaterializationRetry(t *testing.T, store *WorkflowStore, opts WorkflowMaterializationOptions, result WorkflowMaterializationResult) {
	t.Helper()

	if len(result.Tasks) != 6 || result.Set.SourceHash == "" {
		t.Fatal("missing recovered records")
	}

	path := filepath.Join(store.layout.Root(), result.ReceiptRef)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	again, err := store.MaterializePlan(t.Context(), opts)
	if err != nil || again.Set.ID != result.Set.ID {
		t.Fatalf("retry changed set: %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatal("retry changed immutable receipt")
	}

	checkRetriedWorkflowIdentities(t, again, result)
}

// TestWorkflowMaterializationConflict refuses receipt-input mismatches and
// incompatible manual images instead of creating another set or replacing edits.
func TestWorkflowMaterializationConflict(t *testing.T) {
	t.Run("changed request", func(t *testing.T) {
		store, opts := workflowMaterializationFixture(t)
		if _, err := store.MaterializePlan(t.Context(), opts); err != nil {
			t.Fatal(err)
		}

		opts.Plan.Slices[0].Tasks[0].Scope = "Changed scope"
		if _, err := store.MaterializePlan(t.Context(), opts); !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatalf("changed request: %v", err)
		}
	})
	t.Run("manual image", func(t *testing.T) {
		store, opts := workflowMaterializationFixture(t)
		crashWorkflowMaterialization(t, store, 3)

		paths, err := filepath.Glob(filepath.Join(store.layout.Root(), "ticket", "*", "artifact.md"))
		if err != nil || len(paths) != 1 {
			t.Fatalf("partial tasks: %v", err)
		}

		if err := os.WriteFile(paths[0], []byte("Manual edit\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		if _, err := store.MaterializePlan(t.Context(), opts); !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatalf("manual edit overwritten: %v", err)
		}
	})
}

// TestWorkflowMaterializationConcurrent proves two callers of the same approved
// request allocate only six keys and return the same set and task identities.
func TestWorkflowMaterializationConcurrent(t *testing.T) {
	store, opts := workflowMaterializationFixture(t)

	var count atomic.Int32

	opts.ReserveTicketKey = func(context.Context) (string, error) { return fmt.Sprintf("PRJ-%d", count.Add(1)), nil }
	results := make([]WorkflowMaterializationResult, 2)
	failures := make([]error, 2)

	var group sync.WaitGroup
	for i := range results {
		group.Go(func() { results[i], failures[i] = store.MaterializePlan(t.Context(), opts) })
	}

	group.Wait()

	if failures[0] != nil || failures[1] != nil {
		t.Fatalf("competing callers: %v", failures)
	}

	if count.Load() != 6 || results[0].Set.ID != results[1].Set.ID {
		t.Fatal("competing callers duplicated materialization")
	}
}

func checkRetriedWorkflowIdentities(t *testing.T, again, result WorkflowMaterializationResult) {
	t.Helper()

	for i, task := range again.Tasks {
		if task.Task.Key != result.Tasks[i].Task.Key || task.ID != result.Tasks[i].ID {
			t.Fatal("retry changed task identity")
		}
	}
}
