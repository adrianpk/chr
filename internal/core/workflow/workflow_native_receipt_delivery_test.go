package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestNativeReceiptDelivery starts with actual atomic publication and retained
// canonical observations. Only the latest exact merged proof may seal delivery.
func TestNativeReceiptDelivery(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	b := fixture.f.candidate.Native.NativeRecordBinding

	cases := []struct {
		name   string
		change func(*workflowstate.PullRequestSnapshot)
	}{
		{"pending", func(_ *workflowstate.PullRequestSnapshot) {}},
		{"closed", func(pr *workflowstate.PullRequestSnapshot) { pr.State = "closed" }},
		{"result", func(pr *workflowstate.PullRequestSnapshot) {
			pr.State, pr.Merged, pr.ResultCommit = "closed", true, b.OriginalSource
		}},
		{"merged", func(pr *workflowstate.PullRequestSnapshot) {
			pr.State, pr.Merged, pr.ResultCommit = "closed", true, b.PreparedResult
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			pr := fixture.forge.request.PullRequest
			test.change(&pr)
			testNativeReceiptObservation(t, fixture, pr, test.name == "merged")
		})
	}
}

func testNativeReceiptObservation(t *testing.T, fixture nativeCanonicalFixture, pr workflowstate.PullRequestSnapshot, merged bool) {
	t.Helper()

	f := fixture.f
	setNativePRFixture(t, fixture.forge, pr)

	submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git)
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
	if merged {
		checkNativeReceiptDelivery(t, f, submission, receipt, err)

		return
	}

	if !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("unmerged or incompatible PR accepted", err)
	}

	run, err := f.store.readWorkflowInvocation(f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	path := workflowNativeReceiptPath(run, submission)
	if _, err := os.Stat(filepath.Join(f.store.layout.Root(), path)); !os.IsNotExist(err) {
		t.Fatal("unproved receipt bytes persisted", err)
	}
}

func checkNativeReceiptDelivery(t *testing.T, f nativeSubmissionFixture, submission workflowstate.Submission, receipt Record, failure error) {
	t.Helper()

	if failure != nil || receipt.Receipt == nil || receipt.Receipt.Native == nil {
		t.Fatal("exact merged publication has no receipt", failure)
	}

	native, intent := receipt.Receipt.Native, submission.Native
	expected := [3]string{intent.CandidateHash, submission.ID, workspaceconfig.WorkflowPublicationGitAtomicDualRef}
	actual := [3]string{native.CandidateHash, receipt.Receipt.SubmissionID, native.Mechanism}

	if native.CandidateRef != intent.CandidateRef || native.TransactionProofRef != *intent.TransportResponseRef || expected != actual {
		t.Fatal("receipt detached from authorized submission")
	}

	if !workflowNativeJSONEqual(native.NativeRecordBinding, intent.NativeRecordBinding) || receipt.Receipt.SourceCommit == native.PublishedSource {
		t.Fatal("receipt lost original source or profile binding")
	}

	checkNativeReceiptRetained(t, f, submission, receipt)
}

func checkNativeReceiptRetained(t *testing.T, f nativeSubmissionFixture, submission workflowstate.Submission, receipt Record) {
	t.Helper()

	confirmed, err := f.store.ReadWorkflowSubmission(t.Context(), submission.InvocationID)
	if err != nil || confirmed.State != "confirmed" || confirmed.Native.ReceiptRef == nil || confirmed.Native.ReceiptRef.Path != receipt.Path {
		t.Fatal("receipt not retained in original journal", err)
	}

	again, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
	if err != nil || again.SourceHash != receipt.SourceHash || again.Revision != 1 {
		t.Fatal("receipt retry rewrote immutable evidence", err)
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), submission.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("receipt replay consumed another submission", err)
	}
}
