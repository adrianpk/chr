package workflow

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowReopening preserves prior closure bytes and cancelled invocation
// scope. A terminal retry neither attaches later work nor alters its checkpoint.
func TestWorkflowReopening(t *testing.T) {
	store, hierarchy, request := workflowClaimFixture(t)
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice_set")
	request.Policy.IncludeLaterTickets = true

	run := workflowTestStart(t, store, request)
	if _, err := store.StopWorkflow(t.Context(), run.Invocation.ID, "cancel"); err != nil {
		t.Fatal(err)
	}

	terminal := workflowTestRead(t, store, run.Invocation.ID)
	before := workflowReadTestBytes(t, store, workflowInvocationPath(terminal.ID))
	closed := workflowRetainPriorClosure(t, store, hierarchy.Set)
	closure := workflowReadTestBytes(t, store, closed.Set.ClosureRef)
	plan := workflowReadTestBytes(t, store, closed.Set.PlanRef)
	amend := workflowApprovedAmendment(t, store, closed)
	workflowAppendPlanSlice(&amend.Plan)

	if _, err := store.AmendWorkflowSet(t.Context(), amend); err == nil {
		t.Fatal("completed set reopened implicitly")
	}

	amend.Reopen = true

	changed, err := store.AmendWorkflowSet(t.Context(), amend)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowReopeningClosure(t, store, changed, closed, closure, plan)
	checkWorkflowReopeningTerminal(t, store, request, terminal, before)

	if err := os.WriteFile(filepath.Join(store.layout.Root(), closed.Set.ClosureRef), []byte("changed prior closure"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := store.checkWorkflowSetAmendmentEvidence(changed.Set); err == nil {
		t.Fatal("reopening accepted changed prior closure bytes")
	}
}

func checkWorkflowReopeningClosure(t *testing.T, store *Store, changed MaterializationResult, closed Record, closure, plan []byte) {
	t.Helper()

	if changed.Set.Set.Status != "ready" || changed.Set.Set.ClosureRef != "" || len(changed.Set.Set.ClosureHistory) != 1 || changed.Set.Set.ClosureHistory[0] != closed.Set.ClosureRef {
		t.Fatal("reopening replaced the current receipt or lost prior closure")
	}

	if !bytes.Equal(closure, workflowReadTestBytes(t, store, closed.Set.ClosureRef)) {
		t.Fatal("reopening rewrote immutable closure evidence")
	}

	if !bytes.Equal(plan, workflowReadTestBytes(t, store, closed.Set.PlanRef)) {
		t.Fatal("amendment rewrote the original approved plan")
	}
}

func checkWorkflowReopeningTerminal(t *testing.T, store *Store, request StartRequest, terminal Invocation, before []byte) {
	t.Helper()

	if current, err := store.ReconcileWorkflowMembership(t.Context(), terminal.ID); err != nil || current.Revision != terminal.Revision || len(current.Members) != 6 {
		t.Fatal("terminal membership changed on reopening", err)
	}

	attached, err := store.StartWorkflow(t.Context(), request)
	if err != nil || !attached.Attached || attached.Invocation.Status != "cancelled" || !bytes.Equal(before, workflowReadTestBytes(t, store, workflowInvocationPath(terminal.ID))) {
		t.Fatal("terminal retry rewrote the previous invocation", err)
	}
}

func workflowRetainPriorClosure(t *testing.T, store *Store, set Record) Record {
	t.Helper()

	// This fixture represents a prior immutable closure. The closing gate and
	// integration-proof producer are exercised separately by closure tests.
	ref := fmt.Sprintf("plan/%s/extension/chr/prior-closure.json", set.ID)
	set.Set.Status, set.Set.ClosureRef = "completed", ref

	if err := store.withLock(t.Context(), func() error {
		files, err := store.encodeChanges(set)
		if err != nil {
			return err
		}

		return store.commitFiles(append(files, workflowstate.FileChange{Path: ref, After: []byte("{\"schema\":\"prior-closure-fixture\"}\n")}))
	}); err != nil {
		t.Fatal(err)
	}

	closed, err := store.Read(t.Context(), set.Path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Write(t.Context(), closed, closed.Revision); err == nil {
		t.Fatal("generic write bypassed completed-set immutability")
	}

	return closed
}

func workflowReadTestBytes(t *testing.T, store *Store, relative string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(store.layout.Root(), relative))
	if err != nil {
		t.Fatal(err)
	}

	return data
}
