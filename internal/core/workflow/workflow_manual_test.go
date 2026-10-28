package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTestManual(t *testing.T, store *Store, id string, worker []string) (*Claim, ManualTask) {
	t.Helper()

	claim := workflowTestClaim(t, store, id)

	task, err := store.PrepareWorkflowManualTask(t.Context(), claim, worker)
	if err != nil {
		t.Fatal(err)
	}

	if task.Input.ExecutionMode != "manual" || len(task.InputHash) != 64 {
		t.Fatal("manual input binding missing")
	}

	return claim, task
}

func workflowManualTestCandidate(t *testing.T, store *Store, task ManualTask) ManualCandidate {
	t.Helper()

	commit, tree, err := store.workflowLaneCandidate(t.Context(), task.Input.Lane)
	if err != nil {
		t.Fatal(err)
	}

	return ManualCandidate{InputHash: task.InputHash, Branch: task.Input.Lane.Branch, Commit: commit, Tree: tree}
}

// TestWorkflowManualExact uses a real supervised harness and the shared host
// validator; neither preparation nor harness success advances the task cursor.
func TestWorkflowManualExact(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)

	claim, task := workflowTestManual(t, store, run.Invocation.ID, workflowTaskCommand(t, "worker"))
	if _, err := store.ExecuteWorkflowManualTask(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	checkWorkflowManualPending(t, store, run.Invocation.ID)

	candidate := workflowManualTestCandidate(t, store, task)

	result, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate)
	if err != nil || !result.Successful || result.ExecutionMode != "manual" || !result.IntegrationPending {
		t.Fatalf("manual validation: %+v, %v", result, err)
	}

	checkWorkflowExactTaskPhases(t, store, hierarchy)
	checkWorkflowManualReplay(t, store, claim, candidate, result)
}

func checkWorkflowManualPending(t *testing.T, store *Store, id string) {
	t.Helper()

	before, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || before.Cursor != 0 || before.Status != "running" || before.Attempts[0].ResultRef != "" {
		t.Fatal("manual harness completed a task without host evidence")
	}
}

func checkWorkflowManualReplay(t *testing.T, store *Store, claim *Claim, candidate ManualCandidate, result TaskResult) {
	t.Helper()

	replay, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate)
	if err != nil || replay.Commit != result.Commit {
		t.Fatal("accepted manual result did not replay")
	}

	retained, err := store.ReadWorkflow(t.Context(), claim.InvocationID)
	if err != nil || retained.Status != "completed" || len(retained.Attempts) != 1 || len(retained.Attempts[0].Processes) != 2 {
		t.Fatal("manual retry reran host validation")
	}
}

// TestWorkflowManualCandidate rejects caller assertions that disagree with the
// frozen input or actual Git checkout before it can run a gate or mutate status.
func TestWorkflowManualCandidate(t *testing.T) {
	cases := []struct {
		name   string
		change func(*ManualCandidate)
	}{
		{"input", func(c *ManualCandidate) { c.InputHash = "" }},
		{"branch", func(c *ManualCandidate) { c.Branch = "refs/heads/other" }},
		{"commit", func(c *ManualCandidate) { c.Commit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" }},
		{"tree", func(c *ManualCandidate) { c.Tree = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" }},
		{"short commit", func(c *ManualCandidate) { c.Commit = c.Commit[:8] }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
			run := workflowTestStart(t, store, request)
			claim, task := workflowTestManual(t, store, run.Invocation.ID, nil)
			candidate := workflowManualTestCandidate(t, store, task)
			test.change(&candidate)

			if _, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate); err == nil {
				t.Fatal("reported assertion was accepted")
			}

			retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
			if err != nil || retained.Cursor != 0 || retained.Status != "running" || len(retained.Attempts[0].Processes) != 0 {
				t.Fatal("invalid candidate advanced work or ran validation")
			}
		})
	}
}

// TestWorkflowManualSnapshot fences direct ticket/set/input edits. A manual
// status edit alone cannot advance the cursor, even with a valid Git candidate.
func TestWorkflowManualSnapshot(t *testing.T) {
	for _, image := range []string{"ticket", "set", "input", "status"} {
		t.Run(image, func(t *testing.T) {
			store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
			run := workflowTestStart(t, store, request)
			claim, task := workflowTestManual(t, store, run.Invocation.ID, nil)
			candidate := workflowManualTestCandidate(t, store, task)

			path, changed := changeWorkflowManualImage(t, store, hierarchy, task, image)
			if _, err := store.ReconcileWorkflowTask(t.Context(), claim, candidate); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatalf("changed image accepted: %v", err)
			}

			observed, err := os.ReadFile(path)
			if err != nil || string(observed) != string(changed) {
				t.Fatal("manual image was overwritten")
			}

			retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
			if err != nil || retained.Cursor != 0 {
				t.Fatal("manual edit advanced cursor")
			}
		})
	}
}

func changeWorkflowManualImage(t *testing.T, store *Store, hierarchy MaterializationResult, task ManualTask, image string) (string, []byte) {
	t.Helper()

	path := filepath.Join(store.layout.Root(), filepath.Dir(hierarchy.Tasks[0].Path), "artifact.md")
	if image == "set" {
		path = filepath.Join(store.layout.Root(), hierarchy.Set.Path)
	}

	if image == "input" {
		path = filepath.Join(store.layout.Root(), task.InputRef)
	}

	if image == "status" {
		path = filepath.Join(store.layout.Root(), hierarchy.Tasks[0].Path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	changed := append(data, []byte("\nChanged manual image\n")...)
	if image == "set" {
		changed = replaceWorkflowManualField(t, data, "manual_note", "changed image")
	}

	if image == "status" {
		changed = replaceWorkflowManualStatus(t, data)
	}

	if err := os.WriteFile(path, changed, 0o600); err != nil {
		t.Fatal(err)
	}

	return path, changed
}
