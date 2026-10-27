package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowSliceFixture(t *testing.T, gate workflowstate.Validation) (*WorkflowStore, WorkflowMaterializationResult, WorkflowStartRequest) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)

	opts.Plan.Slices[0].Validation = gate
	for i := range opts.Plan.Slices[0].Tasks {
		opts.Plan.Slices[0].Tasks[i].Validation = workflowstate.Validation{Required: true, Commands: [][]string{workflowTaskCommand(t, "gate")}, TimeoutSeconds: 30}
	}

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	commit, _ := workflowSelectionCommit(t, store)
	request := WorkflowStartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "slice-1",
		Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "slice"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}}

	return store, hierarchy, request
}

func workflowSliceGate(t *testing.T, mode string) workflowstate.Validation {
	t.Helper()

	return workflowstate.Validation{Required: true, Commands: [][]string{workflowTaskCommand(t, mode)}, TimeoutSeconds: 30}
}

func workflowTestExecuteSlice(t *testing.T, store *WorkflowStore, id, worker string) WorkflowSliceCandidate {
	t.Helper()

	candidate, err := store.ExecuteWorkflowSlice(t.Context(), id, workflowTaskCommand(t, worker))
	if err != nil {
		t.Fatalf("slice candidate %s (%s, %s): %v", candidate.ID, candidate.Result.Outcome, candidate.Result.Failure, err)
	}

	return candidate
}

func workflowTestValidatedSlice(t *testing.T, store *WorkflowStore, hierarchy WorkflowMaterializationResult, id string) WorkflowInvocation {
	t.Helper()

	run, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || run.Cursor != 3 {
		t.Fatalf("aggregate cursor: %+v, %v", run, err)
	}

	for _, task := range hierarchy.Tasks[:3] {
		record, err := store.Read(t.Context(), task.Path)
		if err != nil || record.Task.Phase != "validated" || record.Task.Status != ticketcore.StatusInProgress || len(record.Task.Evidence) != 1 || record.Task.AttemptID != "" {
			t.Fatal("aggregate validation changed task completion or evidence")
		}
	}

	return run
}

// TestWorkflowSliceAggregate validates all remaining tasks and a distinct frozen
// aggregate contract, then reuses its candidate without rerunning any command.
func TestWorkflowSliceAggregate(t *testing.T) {
	store, hierarchy, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
	request.Policy.MaxWorkerAttempts = 1
	run := workflowTestStart(t, store, request)
	candidate := workflowTestExecuteSlice(t, store, run.Invocation.ID, "worker")
	retained := workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)

	checkWorkflowAggregateAttempt(t, retained)

	if !candidate.Result.Successful || len(candidate.Result.Validation) != 1 {
		t.Fatal("slice lacks distinct aggregate host proof")
	}

	replayed := workflowTestExecuteSlice(t, store, run.Invocation.ID, "fail")
	if replayed.ID != candidate.ID || replayed.Result.Commit != candidate.Result.Commit {
		t.Fatal("unchanged slice retry created another candidate")
	}

	retained = workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
	if len(retained.Attempts) != 4 || workflowSelectionGit(t, store, "rev-parse", "HEAD") != request.Selection.Target.Commit {
		t.Fatal("aggregate retry reran work or changed the primary checkout")
	}
}

func checkWorkflowAggregateAttempt(t *testing.T, run WorkflowInvocation) {
	t.Helper()

	if run.Status != "waiting_integration" || len(run.Attempts) != 4 || run.Attempts[3].Purpose != workflowSliceValidationPurpose || run.CandidateRef == "" {
		t.Fatal("slice lacks a separately owned aggregate attempt")
	}
}

// TestWorkflowSlicePrefix reuses a grouped exact task without creating a partial
// aggregate candidate and executes only its remaining siblings in the slice run.
func TestWorkflowSlicePrefix(t *testing.T) {
	store, hierarchy, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
	request.Selection.Selector = workflowClaimSelector(hierarchy, "ticket")
	first := workflowTestStart(t, store, request)
	proof := workflowTestExecute(t, store, first.Invocation.ID, "worker")

	if _, err := store.ValidateWorkflowSlice(t.Context(), first.Invocation.ID); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("exact grouped task created an aggregate candidate")
	}

	request.RequestKey = "slice-after-exact"
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice")
	run := workflowTestStart(t, store, request)
	candidate := workflowTestExecuteSlice(t, store, run.Invocation.ID, "worker")
	retained := workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)

	if len(retained.Attempts) != 3 || candidate.Lane.Worktree != firstLanePath(t, store, first.Invocation.ID) || candidate.Slice.Tasks[0].EvidenceRef != proof.InputRef[:len(proof.InputRef)-len("input.json")]+"result.json" {
		t.Fatal("slice did not retain exact task evidence and cumulative lane")
	}
}

// TestWorkflowSliceRetry retains validated tasks after aggregate failure and
// retries only the aggregate command against the same clean commit and tree.
func TestWorkflowSliceRetry(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "aggregate-ready")
	gate := workflowstate.Validation{Required: true, Commands: [][]string{{"sh", "-c", "test -f \"$1\"", "gate", marker}}, TimeoutSeconds: 30}
	store, hierarchy, request := workflowSliceFixture(t, gate)
	run := workflowTestStart(t, store, request)

	failed, err := store.ExecuteWorkflowSlice(t.Context(), run.Invocation.ID, workflowTaskCommand(t, "worker"))
	if err == nil || failed.Result.Successful {
		t.Fatal("failed aggregate gate accepted")
	}

	retained := workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
	checkWorkflowAggregateFailure(t, retained)

	if err := os.WriteFile(marker, []byte("Ready\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	candidate := workflowTestExecuteSlice(t, store, run.Invocation.ID, "fail")

	retained = workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
	if candidate.Result.Commit != failed.Result.Commit || candidate.Result.Tree != failed.Result.Tree || len(retained.Attempts) != 5 || len(retained.Attempts[4].Processes) != 1 {
		t.Fatal("aggregate retry reran workers or changed its immutable image")
	}
}

func checkWorkflowAggregateFailure(t *testing.T, run WorkflowInvocation) {
	t.Helper()

	if run.Status != "paused" || run.CandidateRef != "" || len(run.Attempts) != 4 || run.Attempts[3].ResultRef == "" {
		t.Fatal("aggregate failure lost reusable task results")
	}
}

// TestWorkflowSliceRepair invalidates affected ordered evidence and aggregate
// acceptance while preserving the unchanged prefix and historical result files.
func TestWorkflowSliceRepair(t *testing.T) {
	store, hierarchy, request := workflowSliceFixture(t, workflowSliceGate(t, "gate aggregate"))
	run := workflowTestStart(t, store, request)
	first := workflowTestExecuteSlice(t, store, run.Invocation.ID, "worker")

	repaired, err := store.RepairWorkflowSlice(t.Context(), run.Invocation.ID, hierarchy.Tasks[1].Task.Key)
	if err != nil || repaired.Cursor != 1 || repaired.Status != "pending" || repaired.CandidateRef != "" || len(repaired.Attempts) != 5 {
		t.Fatalf("slice repair: %+v, %v", repaired, err)
	}

	checkWorkflowSliceRepair(t, store, hierarchy, first)

	second := workflowTestExecuteSlice(t, store, run.Invocation.ID, "repair worker")
	if second.ID == first.ID || second.Result.Commit == first.Result.Commit || second.Slice.Tasks[0] != first.Slice.Tasks[0] || second.Slice.Tasks[1] == first.Slice.Tasks[1] {
		t.Fatal("repair reused affected evidence or discarded the unchanged prefix")
	}

	workflowTestValidatedSlice(t, store, hierarchy, run.Invocation.ID)
}

func checkWorkflowSliceRepair(t *testing.T, store *WorkflowStore, hierarchy WorkflowMaterializationResult, first WorkflowSliceCandidate) {
	t.Helper()

	for i, task := range hierarchy.Tasks[:3] {
		record, err := store.Read(t.Context(), task.Path)
		if err != nil {
			t.Fatal(err)
		}

		if i == 0 && record.Task.Phase != "validated" || i > 0 && (record.Task.Phase != "pending" || len(record.Task.Evidence) != 0) {
			t.Fatal("repair invalidated the wrong task evidence")
		}

		if _, err := os.Stat(filepath.Join(store.layout.Root(), first.Slice.Tasks[i].EvidenceRef)); err != nil {
			t.Fatal("repair deleted immutable historical evidence")
		}
	}
}
