package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowManualReuse requires the retained prerequisite records, evidence
// bytes and checkout to remain unchanged even when replaying an accepted result.
func TestWorkflowManualReuse(t *testing.T) {
	for _, image := range []string{"prerequisite", "evidence", "checkout"} {
		t.Run(image, func(t *testing.T) {
			store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
			first := workflowTestStart(t, store, request)
			workflowTestExecute(t, store, first.Invocation.ID, "worker")

			request.RequestKey = "manual-reuse"
			request.Selection.Selector.Identity = hierarchy.Tasks[1].Task.Key
			next := workflowTestStart(t, store, request)

			claim, task := workflowTestManual(t, store, next.Invocation.ID, workflowTaskCommand(t, "worker"))
			if _, err := ExecuteManualTask(t.Context(), store, claim); err != nil {
				t.Fatal(err)
			}

			candidate := workflowManualTestCandidate(t, store, task)
			if _, err := ReconcileTask(t.Context(), store, claim, candidate); err != nil {
				t.Fatal(err)
			}

			changeWorkflowManualReuseImage(t, store, task, image)

			if _, err := ReconcileTask(t.Context(), store, claim, candidate); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatalf("changed retained %s accepted: %v", image, err)
			}

			retained, err := store.ReadWorkflow(t.Context(), next.Invocation.ID)
			if err != nil || retained.Cursor != 1 || len(retained.Attempts) != 1 || len(retained.Attempts[0].Processes) != 2 {
				t.Fatal("rejected replay executed or advanced another task")
			}
		})
	}
}

func changeWorkflowManualReuseImage(t *testing.T, store *Store, task ManualTask, image string) {
	t.Helper()

	path := filepath.Join(store.layout.Root(), filepath.Dir(task.Input.Prerequisites[0].Path), "artifact.md")
	if image == "evidence" {
		for ref := range task.Input.EvidenceHashes {
			path = filepath.Join(store.layout.Root(), ref)

			break
		}
	}

	if image == "checkout" {
		path = filepath.Join(task.Input.Lane.Worktree, "uncommitted.txt")
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, writeErr := file.WriteString("\nChanged manual replay image\n")
	if err := errors.Join(writeErr, file.Close()); err != nil {
		t.Fatal(err)
	}
}
