package core

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// TestWorkflowLooseFailure keeps each failed attempt and its finite budget.
// Neither continuation nor direct claims can skip to the next independent task.
func TestWorkflowLooseFailure(t *testing.T) {
	cases := []struct {
		name   string
		gate   string
		worker []string
	}{
		{"worker", "test -f result.txt", []string{"sh", "-c", "exit 17"}},
		{"gate", "false", workflowPRWorker()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testWorkflowLooseFailure(t, tc.gate, tc.worker)
		})
	}
}

func testWorkflowLooseFailure(t *testing.T, gate string, worker []string) {
	t.Helper()

	store, first, request := workflowTicketFixture(t, "loose", workflowPreparationGate(gate))
	workflowLooseCreate(t, store, first, "PRJ-91", "normal", nil)

	request.Policy.MaxWorkerAttempts = 2

	run := workflowTestStart(t, store, request).Invocation
	for attempt := 1; attempt <= 2; attempt++ {
		claim := workflowTestClaim(t, store, run.ID)

		result, err := store.ExecuteWorkflowTask(t.Context(), claim, worker)
		if err == nil || result.Successful {
			t.Fatal("failed worker/gate produced success")
		}

		checkWorkflowLooseFailure(t, store, run.ID, attempt)

		if _, err := store.ResumeWorkflow(t.Context(), run.ID); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := store.ClaimWorkflow(t.Context(), run.ID); err == nil {
		t.Fatal("resume reset the consumed worker budget")
	}

	if retained := workflowTestRead(t, store, run.ID); len(retained.Attempts) != 2 || retained.Cursor != 0 || retained.ConfirmedCursor != 0 {
		t.Fatal("budget exhaustion admitted a successor")
	}
}

func checkWorkflowLooseFailure(t *testing.T, store *WorkflowStore, id string, attempts int) {
	t.Helper()

	before := workflowTestRead(t, store, id)

	progress, err := store.ContinueWorkflow(t.Context(), id, WorkflowContinuationRequest{Worker: workflowPRWorker()}, WorkflowContinuationPorts{})
	if err != nil || progress.Reason != "stopped" || progress.Steps != 0 || progress.Invocation.Status != "paused" || progress.Invocation.Cursor != 0 || len(progress.Invocation.Attempts) != attempts || progress.Invocation.Lane.Unit.ID != "PRJ-90" {
		t.Fatal("failed ticket did not stop its invocation", err)
	}

	if _, err := store.ClaimWorkflow(t.Context(), id); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("failed ticket resumed without an explicit resume", err)
	}

	if !reflect.DeepEqual(progress.Invocation.Members, before.Members) {
		t.Fatal("failed continuation changed scope")
	}
}

// TestWorkflowLooseBounded admits every page once, then checkpoints before PR
// publication. Equivalent requests attach; changed inputs cannot reset scope.
func TestWorkflowLooseBounded(t *testing.T) {
	store, first, request := workflowTicketFixture(t, "loose", workflowPreparationGate("test -f result.txt"))
	for i := 0; i < WorkflowPageLimit; i++ {
		workflowLooseCreate(t, store, first, fmt.Sprintf("PRJ-%d", 91+i), "normal", nil)
	}

	request.Policy.CheckpointStepLimit = 1

	run := workflowTestStart(t, store, request).Invocation
	if len(run.Members) != WorkflowPageLimit+1 || len(run.LooseQueue.Admitted) != len(run.Members) {
		t.Fatal("initial selection truncated at the query page bound")
	}

	for step := 1; step <= 2; step++ {
		restarted := NewWorkflowStore(store.layout, store.projectID)

		progress := workflowContinueOne(t, restarted, run.ID, WorkflowContinuationPorts{})
		checkWorkflowLooseBoundedStep(t, progress, step)

		request.RequestKey = "equivalent-loose"

		attached := workflowTestStart(t, restarted, request)
		if !attached.Attached || attached.Invocation.ID != run.ID || !reflect.DeepEqual(attached.Invocation.LooseQueue.Admitted, run.LooseQueue.Admitted) {
			t.Fatal("equivalent loose request launched a second queue")
		}
	}

	request.Policy.IncludeLaterTickets = true
	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("same key changed its admission policy", err)
	}
}

func checkWorkflowLooseBoundedStep(t *testing.T, progress WorkflowContinuationProgress, step int) {
	t.Helper()

	if progress.Steps != 1 || progress.Reason != "step_limit" || progress.Invocation.Continuation.Sequence != step || len(progress.Invocation.Attempts) != 1 || progress.Invocation.ConfirmedCursor != 0 {
		t.Fatal("bounded restart repeated implementation or confirmed an unmerged ticket")
	}
}

// TestWorkflowLooseTerminal keeps accepted later admission dormant after finite
// exhaustion. Retry and continuation do not subscribe the terminal invocation.
func TestWorkflowLooseTerminal(t *testing.T) {
	f := workflowLooseQueueFixture(t, true)
	for _, key := range []string{"PRJ-91", "PRJ-92"} {
		workflowLooseClassify(t, f.store, workflowLooseRecord(t, f.store, key), "normal", TicketStatusBacklog)
	}

	workflowFinalizeFrontier(t, f)
	workflowAdvanceFrontier(t, f)

	run, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID)
	if err != nil || run.Status != "completed" {
		t.Fatal("queue did not exhaust", err)
	}

	workflowLooseCreate(t, f.store, workflowLooseRecord(t, f.store, "PRJ-91"), "PRJ-93", "critical", nil)

	request := WorkflowStartRequest{RequestKey: run.RequestKey, Selection: run.Selection, Policy: run.Policy, ReviewRestriction: run.ReviewRestriction}
	retry := workflowTestStart(t, f.store, request)

	progress, err := f.store.ContinueWorkflow(t.Context(), run.ID, WorkflowContinuationRequest{Worker: workflowPRWorker()}, WorkflowContinuationPorts{})
	if err != nil || !retry.Attached || progress.Reason != "completed" || progress.Steps != 0 || !reflect.DeepEqual(progress.Invocation, run) {
		t.Fatal("terminal retry admitted or executed a later ticket", err)
	}
}
