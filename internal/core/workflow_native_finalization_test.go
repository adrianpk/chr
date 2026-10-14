package core

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

// TestNativeFinalization reuses the physically confirmed native receipt for all
// unit members. Repeated finalization preserves bytes, task revisions and budget.
func TestNativeFinalization(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)

	f := fixture.f
	nativeReceiptMerged(t, fixture)

	receipt, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
	if err != nil {
		t.Fatal(err)
	}

	entry := nativeReviewEntry(t, f.store, f.entry.InvocationID)
	if entry.State != "finalizing" || entry.ReceiptRef == nil || entry.ReceiptRef.Path != receipt.Path {
		t.Fatal("confirmed receipt did not expose unfinished fanout", entry.State)
	}

	before, err := os.ReadFile(filepath.Join(f.store.layout.Root(), receipt.Path))
	if err != nil {
		t.Fatal(err)
	}

	finalized, err := f.store.FinalizeWorkflowSlice(t.Context(), f.entry.InvocationID, nil, nil)
	if err != nil || finalized.SourceHash != receipt.SourceHash {
		t.Fatal("native receipt could not finalize without another publication", err)
	}

	checkNativeFinalization(t, fixture, receipt, before)
}

func checkNativeFinalization(t *testing.T, fixture nativeCanonicalFixture, receipt WorkflowRecord, original []byte) {
	t.Helper()

	f := fixture.f

	run, err := f.store.readWorkflowInvocation(f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	if run.ConfirmedCursor != run.Cursor || nativeReviewEntry(t, f.store, run.ID).State != "integrated" {
		t.Fatal("queue completed before all members were finalized")
	}

	checkNativeFanoutMembers(t, f.store, run, receipt)

	if err := f.store.checkWorkflowFinalizationImages(t.Context(), run, receipt); err != nil {
		t.Fatal("missing member receipt", err)
	}

	checkNativeFanoutRetry(t, fixture, receipt, original, run)
}

func checkNativeFanoutRetry(t *testing.T, fixture nativeCanonicalFixture, receipt WorkflowRecord, original []byte, run workflowInvocationFile) {
	t.Helper()

	f := fixture.f

	again, err := f.store.FinalizeWorkflowSlice(t.Context(), run.ID, nil, nil)
	if err != nil || again.SourceHash != receipt.SourceHash || workflowTestRead(t, f.store, run.ID).Revision != run.Revision {
		t.Fatal("finalization retry repeated fanout", err)
	}

	after, err := os.ReadFile(filepath.Join(f.store.layout.Root(), receipt.Path))
	if err != nil || string(after) != string(original) {
		t.Fatal("fanout rewrote immutable receipt", err)
	}

	before := workflowTestRead(t, f.store, run.ID)
	if _, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, run.ID)); err != nil || workflowTestRead(t, f.store, run.ID).Revision != before.Revision {
		t.Fatal("receipt confirmation regressed completed fanout", err)
	}
}

func checkNativeFanoutMembers(t *testing.T, store *WorkflowStore, run workflowInvocationFile, receipt WorkflowRecord) {
	t.Helper()

	for _, member := range run.Members {
		record, err := store.Read(t.Context(), member.Admission.Path)
		if err != nil || !workflowFinalizedTaskMatches(record, run.Lane.Branch, receipt) {
			t.Fatal("member did not retain its shared receipt", err)
		}
	}

	checkWorkflowFinalizedReceiptCount(t, store)

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.ID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("fanout consumed another submission", err)
	}
}

// TestNativeFinalizationTicket gives an independent ticket its own receipt and
// fanout without introducing an aggregate set or another implementation call.
func TestNativeFinalizationTicket(t *testing.T) {
	f, _ := workflowTicketPreparationFixture(t, "rebase", "ticket")
	writeWorkflowTestConfigValue(t, f.store.layout.ProjectDir(), "work", "integration.publication_profile", WorkflowPublicationGitAtomicDualRef)

	entry := nativeRoleAdmit(t, f.store, f.run.ID)
	entry = nativeRoleTransition(t, f.store, entry, "preparing")
	nativeRoleTransition(t, f.store, entry, "validating")

	provider := &nativePreparationProviderFixture{f.provider, "policy-v1"}

	candidate, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, provider, f.git)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, provider, nil); err != nil {
		t.Fatal(err)
	}

	entry = nativeRoleTransition(t, f.store, nativeReviewEntry(t, f.store, f.run.ID), "ready_to_publish")
	native := nativeSubmissionFixture{f.store, entry, candidate, provider}
	fixture, receipt := nativeFanoutReceipt(t, nativeCanonicalSubmissionFixture(t, native, false))

	before := workflowTestRead(t, f.store, f.run.ID)
	if _, err := f.store.FinalizeWorkflowSlice(t.Context(), f.run.ID, nil, nil); err != nil {
		t.Fatal(err)
	}

	after := workflowTestRead(t, f.store, f.run.ID)
	if after.Status != "completed" || after.Outcome != "ticket_integrated" || len(after.Attempts) != len(before.Attempts) || receipt.Receipt.Unit.Kind != "ticket" {
		t.Fatal("independent finalization fabricated aggregate or worker work")
	}

	checkNativeFanoutMembers(t, fixture.f.store, nativeFanoutRun(t, fixture), receipt)
}

// TestNativeFinalizationConcurrent serializes competing owners on the existing
// common lock; only one fanout changes member revisions and the queue frontier.
func TestNativeFinalizationConcurrent(t *testing.T) {
	fixture, receipt := nativeFanoutReceipt(t, newNativeCanonicalFixture(t))

	f := fixture.f
	before := workflowTestRead(t, f.store, f.entry.InvocationID)
	ctx := t.Context()

	var group sync.WaitGroup

	failures := make(chan error, 2)

	for range 2 {
		group.Go(func() {
			store := NewWorkflowStore(f.store.layout, f.store.projectID)

			_, err := store.FinalizeWorkflowSlice(ctx, f.entry.InvocationID, nil, nil)
			failures <- err
		})
	}

	group.Wait()
	close(failures)

	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}

	after := workflowTestRead(t, f.store, before.ID)
	if after.Revision != before.Revision+1 || len(after.Attempts) != len(before.Attempts) {
		t.Fatal("competing finalizers repeated task work or commit")
	}

	checkNativeFanoutMembers(t, f.store, nativeFanoutRun(t, fixture), receipt)
}

// TestNativeFinalizationStop retains cancellation at finalizing, then completes
// irreversible delivery fanout before applying terminal stop to the invocation.
func TestNativeFinalizationStop(t *testing.T) {
	fixture, receipt := nativeFanoutReceipt(t, newNativeCanonicalFixture(t))

	f := fixture.f

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	stopped, err := f.store.StopWorkflow(t.Context(), f.entry.InvocationID, "cancel")
	if err != nil || stopped.Status == "cancelled" || nativeReviewEntry(t, f.store, stopped.ID).State != "finalizing" {
		t.Fatal("stop discarded unfinished fanout", err)
	}

	if _, err := f.store.FinalizeWorkflowSlice(t.Context(), stopped.ID, nil, nil); err != nil {
		t.Fatal(err)
	}

	if workflowTestRead(t, f.store, stopped.ID).Status != "cancelled" {
		t.Fatal("completed fanout lost pending stop")
	}

	after, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), stopped.ID)
	if err != nil || !reflect.DeepEqual(after, budget) {
		t.Fatal("finalization refunded or consumed budget", err)
	}

	checkNativeFanoutMembers(t, f.store, nativeFanoutRun(t, fixture), receipt)
}

func nativeFanoutReceipt(t *testing.T, fixture nativeCanonicalFixture) (nativeCanonicalFixture, WorkflowRecord) {
	t.Helper()

	nativeReceiptMerged(t, fixture)

	f := fixture.f

	receipt, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
	if err != nil {
		t.Fatal("native fanout receipt confirmation", err)
	}

	return fixture, receipt
}

func nativeFanoutRun(t *testing.T, fixture nativeCanonicalFixture) workflowInvocationFile {
	t.Helper()

	run, err := fixture.f.store.readWorkflowInvocation(fixture.f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	return run
}
