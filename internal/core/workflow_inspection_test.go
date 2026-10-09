package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowInspectionJournal leaves interrupted transitions byte-identical
// instead of silently recovering them while an adapter displays status.
func TestWorkflowInspectionJournal(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)

	store.afterReplace = func(int) error { return errors.New("interrupted replacement") }
	if _, err := store.StopWorkflow(t.Context(), run.Invocation.ID, "pause"); err == nil {
		t.Fatal("transition interruption not injected")
	}

	journal := filepath.Join(store.layout.Root(), workflowJournalPath)

	before, err := os.ReadFile(journal)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.InspectWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("status replayed a pending transition", err)
	}

	if _, err := store.InspectWorkflowPolicy(t.Context()); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("policy inspection replayed a pending transition", err)
	}

	after, err := os.ReadFile(journal)
	if err != nil || string(before) != string(after) {
		t.Fatal("read-only inspection changed the recovery journal", err)
	}
}

// TestWorkflowClaimClose fences a running execution and releases only host
// descriptors once quiescent, without accepting a checkpoint or task result.
func TestWorkflowClaimClose(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, run.Invocation.ID)
	if err := beginWorkflowClaimExecution(claim); err != nil {
		t.Fatal(err)
	}

	if err := claim.Close(); !errors.Is(err, ErrWorkflowOwned) || claim.reservation == nil {
		t.Fatal("close released a running execution", err)
	}

	endWorkflowClaimExecution(claim)
	checkWorkflowClaimClosed(t, store, claim, run.Invocation.ID)
}

func checkWorkflowClaimClosed(t *testing.T, store *WorkflowStore, claim *WorkflowClaim, id string) {
	t.Helper()

	before, err := store.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if err := claim.Close(); err != nil {
		t.Fatal(err)
	}

	after, err := store.ReadWorkflow(t.Context(), id)
	if err != nil || after.Revision != before.Revision || after.Cursor != 0 || after.Status != "running" || claim.reservation != nil || claim.capacityGuard != nil {
		t.Fatal("close changed authority or retained owned descriptors", err)
	}
}
