package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowTaskRetry retains the failed attempt and reuses its lane on repair.
func TestWorkflowTaskRetry(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "fail")); err == nil {
		t.Fatal("failed worker accepted")
	}

	retry := workflowTestStart(t, store, request)
	if !retry.Attached || retry.Invocation.Status != workflowstate.InvocationPaused || len(retry.Invocation.Attempts) != 1 {
		t.Fatal("retry reset consumed attempts")
	}

	proof := workflowTestExecute(t, store, run.Invocation.ID, "worker")

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !workflowRepairHistoryMatches(retained, claim.Generation) {
		t.Fatal("repair replaced earlier evidence or fence")
	}

	if retained.Lane.Worktree != retry.Invocation.Lane.Worktree {
		t.Fatal("repair discarded the retained lane")
	}

	if retained.Lane.HeadCommit != proof.Commit {
		t.Fatal("repair lost the validated head")
	}
}

func workflowRepairHistoryMatches(run Invocation, generation uint64) bool {
	return len(run.Attempts) == 2 && run.Attempts[0].Outcome == "failed" && run.Attempts[1].Generation == generation+1
}

// TestWorkflowTaskLinked executes through a linked alias against the original owner.
func TestWorkflowTaskLinked(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	dir := filepath.Join(t.TempDir(), "linked")
	workflowSelectionGit(t, store, "worktree", "add", "-qb", "linked", dir, request.Selection.Target.Commit)

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	alias := NewWorkflowStore(layout, "linked-alias")
	claim := workflowTestClaim(t, alias, run.Invocation.ID)

	proof, err := alias.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker"))
	if err != nil || !proof.Successful {
		t.Fatalf("linked execution: %v", err)
	}

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if err != nil || retained.Status != workflowstate.InvocationCompleted || retained.Lane.Worktree == dir {
		t.Fatal("alias forked durable execution ownership")
	}
}

// TestWorkflowTaskJournal replays a partial result write without executing again.
func TestWorkflowTaskJournal(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)
	interrupted := false

	claim.store.afterReplace = func(index int) error {
		if index != 0 || interrupted {
			return nil
		}

		record, err := store.readRecord(hierarchy.Tasks[0].Path)
		if err != nil {
			return err
		}

		if record.Task.Phase == "validated" {
			interrupted = true

			return fmt.Errorf("fixture interrupted result journal")
		}

		return nil
	}
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker")); err == nil || !interrupted {
		t.Fatal("result journal interruption was not exercised")
	}

	claim.store.afterReplace = nil

	retry := workflowTestStart(t, store, request)
	if !retry.Attached || retry.Invocation.Status != workflowstate.InvocationCompleted || len(retry.Invocation.Attempts) != 1 {
		t.Fatal("partial result replay reran or lost the task")
	}

	checkWorkflowExactTaskPhases(t, store, hierarchy)
}

// TestWorkflowTaskPrerequisiteEdit fences a prerequisite changed after preparation.
func TestWorkflowTaskPrerequisiteEdit(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	first := workflowTestStart(t, store, request)
	workflowTestExecute(t, store, first.Invocation.ID, "worker")

	request.RequestKey = "execution-2"
	request.Selection.Selector.Identity = hierarchy.Tasks[1].Task.Key
	next := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, next.Invocation.ID)
	path := filepath.Join(store.layout.Root(), filepath.Dir(hierarchy.Tasks[0].Path), "artifact.md")

	worker := []string{"sh", "-c", "printf 'Changed prerequisite\\n' >> \"$1\"", "worker", path}
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, worker); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("changed prerequisite accepted: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "Changed prerequisite") {
		t.Fatal("prerequisite image was overwritten")
	}
}
