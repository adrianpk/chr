package core

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func telegramLifecycleJournalFixture(t *testing.T, kind string) (*WorkflowStore, WorkflowControlMutationRequest) {
	t.Helper()

	action := kind

	var (
		store *WorkflowStore
		id    string
	)

	switch kind {
	case "closing":
		f, _, run := workflowClosingEmptyOwner(t)
		claim := workflowClosingBeginTest(t, f, run, workflowFrontierPorts(f))
		expireWorkflowClosingTest(t, f.store, claim)

		store, id, action = f.store, run.ID, "reconcile"
	case "submitted", "unprojected":
		fixture := newNativeCanonicalFixture(t)

		store, id, action = fixture.f.store, fixture.f.entry.InvocationID, "pause"
		if kind == "unprojected" {
			clearTelegramNativeProjection(t, store, id)
		}
	default:
		current, _, start := workflowClaimFixture(t)
		store, id = current, workflowTestStart(t, current, start).Invocation.ID
		prepareTelegramLifecycleJournal(t, store, id, kind)

		if kind == "recovered" {
			action = "reconcile"
		}
	}

	run := workflowTestRead(t, store, id)

	request := WorkflowControlMutationRequest{Binding: WorkflowMutationBinding{RequestID: "lifecycle-receipt", Operator: "operator"}, Action: action, ID: id, ExpectedRevision: run.Revision}
	if action == "continue" {
		request.StepLimit = 1
	}

	return store, request
}

func prepareTelegramLifecycleJournal(t *testing.T, store *WorkflowStore, id, kind string) {
	t.Helper()

	if kind == "resume" {
		if _, err := store.StopWorkflow(t.Context(), id, "pause"); err != nil {
			t.Fatal(err)
		}
	}

	if kind == "recovered" {
		command := workflowClaimProcess(t, store, id, "", "", "0")
		if _, err := command.CombinedOutput(); err == nil {
			t.Fatal("owner did not exit at retained claim")
		}
	}
}

// TestTelegramControlLifecycleJournal recovers every lifecycle receipt with its
// effect, including task checkpoints, closing recovery and a submitted native
// stop. Replaying a recovered acceptance never launches continuation work.
func TestTelegramControlLifecycleJournal(t *testing.T) {
	for _, kind := range []string{"pause", "cancel", "resume", "continue", "reconcile", "recovered", "closing", "submitted", "unprojected"} {
		t.Run(kind, func(t *testing.T) { checkTelegramLifecycleJournal(t, kind) })
	}
}

func checkTelegramLifecycleJournal(t *testing.T, kind string) {
	t.Helper()

	store, request := telegramLifecycleJournalFixture(t, kind)
	fault := errors.New("lifecycle commit interrupted")

	store.afterReplace = func(index int) error {
		if index == 0 {
			return fault
		}

		return nil
	}
	if result, err := store.ControlWorkflowLifecycle(t.Context(), request, WorkflowLifecycleControlPorts{}, nil); !errors.Is(err, fault) || result.RequestID != "" {
		t.Fatal("partial commit was reported as confirmed", result, err)
	}

	reopened := NewWorkflowStore(store.layout, store.projectID)

	first := readTelegramLifecycleReceipt(t, reopened, request)

	before := workflowTestRead(t, reopened, request.ID)
	checkTelegramLifecycleRecovery(t, kind, before)

	second, err := reopened.ControlWorkflowLifecycle(t.Context(), request, WorkflowLifecycleControlPorts{}, func(context.Context) error { return ErrWorkflowConflict })
	if err != nil || second != first || !reflect.DeepEqual(before, workflowTestRead(t, reopened, request.ID)) {
		t.Fatal("original receipt replay mutated current authority", second, err)
	}

	checkTelegramLifecycleActor(t, reopened, request)
}

func readTelegramLifecycleReceipt(t *testing.T, store *WorkflowStore, request WorkflowControlMutationRequest) WorkflowControlResult {
	t.Helper()

	result, found, err := store.ReadWorkflowControlReceipt(t.Context(), request)
	if err != nil || !found || result.RequestID != request.Binding.RequestID || result.Revision != request.ExpectedRevision+1 {
		t.Fatal("lifecycle receipt did not recover", result, err)
	}

	return result
}

func checkTelegramLifecycleActor(t *testing.T, reopened *WorkflowStore, request WorkflowControlMutationRequest) {
	t.Helper()

	request.Binding.Operator = "another-operator"
	if _, found, err := reopened.ReadWorkflowControlReceipt(t.Context(), request); found || !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("receipt read accepted another actor", found, err)
	}

	if _, err := reopened.ControlWorkflowLifecycle(t.Context(), request, WorkflowLifecycleControlPorts{}, nil); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("receipt was rebound to another actor", err)
	}
}

func checkTelegramLifecycleRecovery(t *testing.T, kind string, run WorkflowInvocation) {
	t.Helper()

	if !validTelegramLifecycleRecovery(kind, run) {
		t.Fatal("lifecycle recovery changed its native contract", kind, run)
	}
}

func validTelegramLifecycleRecovery(kind string, run WorkflowInvocation) bool {
	valid := false

	switch kind {
	case "pause", "cancel":
		valid = run.StopIntent == kind && run.Status == workflowStopStatus(kind)
	case "resume":
		valid = run.StopIntent == "" && run.Status == "pending"
	case "submitted", "unprojected":
		valid = run.StopIntent == "pause" && run.Status != "paused"
	case "continue", "recovered", "closing":
		valid = validTelegramLifecycleCheckpoint(kind, run)
	case "reconcile":
		valid = run.Status == "pending" && len(run.Attempts) == 0
	}

	return valid
}

func validTelegramLifecycleCheckpoint(kind string, run WorkflowInvocation) bool {
	switch kind {
	case "continue":
		return len(run.Attempts) == 0 && run.Continuation != nil && run.Continuation.Reason == "accepted"
	case "recovered":
		return len(run.Attempts) == 1 && run.Attempts[0].Outcome == "recovered" && run.Attempts[0].CheckpointRef != ""
	default:
		return run.ClosingSet == "" && run.Outcome == "interrupted"
	}
}

// TestTelegramControlLifecycleFence checks late authorization under native
// exclusion before any lifecycle effect or receipt can be committed.
func TestTelegramControlLifecycleFence(t *testing.T) {
	for _, action := range []string{"pause", "cancel", "resume", "continue", "reconcile"} {
		t.Run(action, func(t *testing.T) {
			store, request := telegramLifecycleJournalFixture(t, action)
			before := workflowTestRead(t, store, request.ID)
			calls := 0

			check := func(context.Context) error {
				calls++
				if calls == 2 {
					return ErrWorkflowConflict
				}

				return nil
			}
			if result, err := store.ControlWorkflowLifecycle(t.Context(), request, WorkflowLifecycleControlPorts{}, check); !errors.Is(err, ErrWorkflowConflict) || result.RequestID != "" {
				t.Fatal("late authority loss was accepted", result, err)
			}

			if !reflect.DeepEqual(before, workflowTestRead(t, store, request.ID)) {
				t.Fatal("late rejection changed native state")
			}
		})
	}
}

func clearTelegramNativeProjection(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	if err := store.withLock(t.Context(), func() error {
		run, err := store.readWorkflowInvocation(id)
		if err != nil {
			return err
		}

		run.Submission = nil

		return store.commitWorkflowState(&run)
	}); err != nil {
		t.Fatal(err)
	}
}
