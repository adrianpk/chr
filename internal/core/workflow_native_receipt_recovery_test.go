package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeReceiptMerged(t *testing.T, fixture nativeCanonicalFixture) WorkflowSubmission {
	t.Helper()

	f, b := fixture.f, fixture.f.candidate.Native.WorkflowNativeRecordBinding
	pr := fixture.forge.request.PullRequest
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, b.PreparedResult
	setNativePRFixture(t, fixture.forge, pr)

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if err != nil {
		t.Fatal("native merged observation", err)
	}

	return submission
}

// TestNativeReceiptLost proves a real lost SSH response remains unknown in the
// receipt. Earlier exact Git application still attributes P after source moves.
func TestNativeReceiptLost(t *testing.T) {
	fixture := nativeCanonicalPublicationFixture(t, true)
	f, b := fixture.f, fixture.f.candidate.Native.WorkflowNativeRecordBinding
	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.SourceRef, b.OriginalSource, b.PreparedResult)

	submission := nativeReceiptMerged(t, fixture)
	receipt, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
	checkNativeReceiptDelivery(t, f, submission, receipt, err)

	if receipt.Receipt.Native.GitApplicationRef == nil {
		t.Fatal("lost exchange lacks independent application proof")
	}

	var response WorkflowGitPublicationResult
	if err := f.store.readWorkflowNativeEvidence(submission.InvocationID, receipt.Receipt.Native.TransactionProofRef, &response); err != nil || response.State != "unknown" || response.PublishedSourceCommit != "" {
		t.Fatal("receipt rewrote lost transport response", err)
	}
}

// TestNativeReceiptFences rejects stale owner, budget and generation callbacks
// without creating immutable bytes or changing the retained publication journal.
func TestNativeReceiptFences(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	f := fixture.f
	submission := nativeReceiptMerged(t, fixture)
	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)

	cases := []struct {
		name   string
		change func(*WorkflowIntegrationBudgetGuard)
	}{
		{"invocation", func(g *WorkflowIntegrationBudgetGuard) { g.InvocationRevision++ }},
		{"entry", func(g *WorkflowIntegrationBudgetGuard) { g.EntryRevision++ }},
		{"budget", func(g *WorkflowIntegrationBudgetGuard) { g.BudgetRevision++ }},
		{"generation", func(g *WorkflowIntegrationBudgetGuard) { g.ClaimGeneration++ }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			stale := guard
			test.change(&stale)

			if _, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), stale); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("stale receipt callback accepted", err)
			}

			checkNativeReceiptAbsent(t, f, submission)
		})
	}
}

func checkNativeReceiptAbsent(t *testing.T, f nativeSubmissionFixture, submission WorkflowSubmission) {
	t.Helper()

	run, err := f.store.readWorkflowInvocation(submission.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(f.store.layout.Root(), workflowNativeReceiptPath(run, submission))); !os.IsNotExist(err) {
		t.Fatal("stale callback wrote receipt", err)
	}

	current, err := f.store.ReadWorkflowSubmission(t.Context(), submission.InvocationID)
	if err != nil || !workflowNativeJSONEqual(current, submission) {
		t.Fatal("stale callback changed publication", err)
	}
}

// TestNativeReceiptRestart exercises interruption after receipt introduction or
// confirmed journal replacement. Restart adopts the same bytes without new IO.
func TestNativeReceiptRestart(t *testing.T) {
	for _, stage := range []string{"receipt", "journal"} {
		t.Run(stage, func(t *testing.T) {
			fixture := newNativeCanonicalFixture(t)
			f := fixture.f
			submission := nativeReceiptMerged(t, fixture)
			cut := errors.New("receipt checkpoint interrupted")
			f.store.afterReplace = func(index int) error {
				if stage == "receipt" && index == 0 || stage == "journal" && index == -1 {
					return cut
				}

				return nil
			}

			_, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
			if !errors.Is(err, cut) {
				t.Fatal("receipt checkpoint was not interrupted", err)
			}

			f.store.afterReplace = nil
			f.store = NewWorkflowStore(f.store.layout, f.store.projectID)
			receipt, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
			checkNativeReceiptDelivery(t, f, submission, receipt, err)
		})
	}
}
