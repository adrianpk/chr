package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowExecutionFixture(t *testing.T, gate []string, timeout int) (*Store, MaterializationResult, StartRequest) {
	t.Helper()

	return workflowExecutionGateFixture(t, workflowstate.Validation{Required: true, Commands: [][]string{gate}, TimeoutSeconds: timeout})
}

func workflowExecutionGateFixture(t *testing.T, gate workflowstate.Validation) (*Store, MaterializationResult, StartRequest) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)
	for i := range opts.Plan.Slices {
		for j := range opts.Plan.Slices[i].Tasks {
			opts.Plan.Slices[i].Tasks[j].Validation = gate
		}
	}

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, store)
	request := StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "execution-1",
		Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "ticket"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}}

	return store, hierarchy, request
}

func workflowTestExecute(t *testing.T, store *Store, id, mode string) workflowstate.TaskResult {
	t.Helper()

	claim := workflowTestClaim(t, store, id)

	result, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, mode))
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// TestWorkflowTaskExact validates only the requested task and preserves retries.
func TestWorkflowTaskExact(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)

	proof := workflowTestExecute(t, store, run.Invocation.ID, "instructed worker")
	if !proof.Successful || proof.Commit == request.Selection.Target.Commit || len(proof.Validation) != 1 || !proof.IntegrationPending || proof.Outcome != "task_validated" {
		t.Fatal("host did not validate the worker commit")
	}

	checkWorkflowTaskExact(t, store, hierarchy, request, run, proof)
}

func checkWorkflowTaskExact(t *testing.T, store *Store, hierarchy MaterializationResult, request StartRequest, run StartResult, proof workflowstate.TaskResult) {
	t.Helper()

	retry := workflowTestStart(t, store, request)
	if !retry.Attached || retry.Invocation.ID != run.Invocation.ID || retry.Invocation.Status != workflowstate.InvocationCompleted || len(retry.Invocation.Attempts) != 1 || retry.Invocation.Lane.HeadCommit != proof.Commit {
		t.Fatal("unchanged retry did not retain evidence and worktree")
	}

	checkWorkflowExactTaskPhases(t, store, hierarchy)

	set, err := store.Read(t.Context(), hierarchy.Set.Path)
	if err != nil || set.Set.Slices[0].PullRequest != "" || set.Set.Slices[0].CandidateRef != "" {
		t.Fatal("exact grouped task produced a slice candidate")
	}

	if workflowSelectionGit(t, store, "rev-parse", "HEAD") != request.Selection.Target.Commit {
		t.Fatal("worker changed the primary checkout")
	}
}

func checkWorkflowExactTaskPhases(t *testing.T, store *Store, hierarchy MaterializationResult) {
	t.Helper()

	for i, task := range hierarchy.Tasks {
		record, err := store.Read(t.Context(), task.Path)
		if err != nil {
			t.Fatal(err)
		}

		if i == 0 && record.Task.Phase != "validated" || i != 0 && record.Task.Phase != "pending" {
			t.Fatal("exact execution changed another task")
		}
	}
}

// TestWorkflowTaskCumulative reuses a completed exact task in a later slice run.
func TestWorkflowTaskCumulative(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	first := workflowTestStart(t, store, request)
	proof := workflowTestExecute(t, store, first.Invocation.ID, "worker")
	request.RequestKey = "execution-2"
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice")
	next := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, next.Invocation.ID)

	second, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker"))
	if err != nil {
		t.Fatal(err)
	}

	retained, err := store.ReadWorkflow(t.Context(), next.Invocation.ID)
	if err != nil || !workflowCumulativeMatches(retained, firstLanePath(t, store, first.Invocation.ID), proof.Commit, second.Commit) {
		t.Fatal("slice did not reuse its original lane and validated prefix")
	}

	workflowTestExecute(t, store, next.Invocation.ID, "worker")

	retained, err = store.ReadWorkflow(t.Context(), next.Invocation.ID)
	if err != nil || retained.Status != workflowstate.InvocationWaitingIntegration || retained.Cursor != 3 || len(retained.Attempts) != 2 {
		t.Fatal("slice did not stop at the integration barrier")
	}

	if _, err := store.ClaimWorkflow(t.Context(), next.Invocation.ID); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("integration barrier admitted another claim")
	}
}

func workflowCumulativeMatches(run Invocation, path, first, second string) bool {
	return run.Cursor == 2 && len(run.Attempts) == 1 && run.Lane.Worktree == path && run.Attempts[0].Lane.HeadCommit == first && run.Lane.HeadCommit == second
}

func firstLanePath(t *testing.T, store *Store, id string) string {
	t.Helper()

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	return run.Lane.Worktree
}

// TestWorkflowTaskFailures retains failed attempts without accepting task success.
func TestWorkflowTaskFailures(t *testing.T) {
	cases := []struct {
		name, worker, gate string
		timeout            int
		conflict           bool
	}{
		{"worker failure", "fail", "gate", 30, false},
		{"dirty worker", "dirty", "gate", 30, false},
		{"gate failure", "worker", "fail", 30, false},
		{"gate mutation", "worker", "gate dirty", 30, false},
		{"gate commit", "worker", "gate commit", 30, false},
		{"gate timeout", "worker", "gate sleep", 1, false},
		{"input mutation", "edit input", "gate", 30, true},
		{"checkpoint mutation", "edit checkpoint", "gate", 30, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, tc.gate), tc.timeout)
			run := workflowTestStart(t, store, request)
			claim := workflowTestClaim(t, store, run.Invocation.ID)

			proof, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, tc.worker))
			if err == nil || proof.Successful {
				t.Fatal("invalid worker or host gate accepted")
			}

			checkWorkflowTaskFailure(t, store, hierarchy.Tasks[0], run.Invocation.ID, tc.conflict)
		})
	}
}

func checkWorkflowTaskFailure(t *testing.T, store *Store, task workflowstate.Record, id string, conflict bool) {
	t.Helper()

	record, err := store.Read(t.Context(), task.Path)
	if err != nil || len(record.Task.Evidence) != 0 || record.Task.Phase == "validated" {
		t.Fatal("failure produced accepted task evidence")
	}

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || run.Cursor != 0 || len(run.Attempts) != 1 {
		t.Fatal("failure advanced or lost history")
	}

	if conflict {
		return
	}

	checkWorkflowFailedAttempt(t, store, run)
}

func checkWorkflowFailedAttempt(t *testing.T, store *Store, run Invocation) {
	t.Helper()

	if run.Status != workflowstate.InvocationPaused || run.Attempts[0].Outcome != "failed" || run.Attempts[0].ResultRef == "" {
		t.Fatal("failed attempt was not retained")
	}

	if _, err := os.Stat(filepath.Join(store.layout.Root(), run.Attempts[0].ResultRef)); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowTaskOutputBounds retains truncated subprocess output explicitly.
func TestWorkflowTaskOutputBounds(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	run := workflowTestStart(t, store, request)

	proof := workflowTestExecute(t, store, run.Invocation.ID, "flood")
	if !proof.Worker.Truncated || len(proof.Worker.Stdout) != opsfile.CommandOutputLimit {
		t.Fatal("worker output was unbounded or silently discarded")
	}
}
