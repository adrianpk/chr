package core

import (
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// TestNativeReceiptApplication refuses missing or substituted application proof
// after a real lost receive-pack response, before any receipt bytes are written.
func TestNativeReceiptApplication(t *testing.T) {
	fixture := nativeCanonicalPublicationFixture(t, true)
	f := fixture.f
	submission := nativeReceiptMerged(t, fixture)
	original := nativeReceiptRecord(t, f, submission)

	cases := []struct {
		name   string
		change func(*WorkflowNativeGitObservation)
	}{
		{"missing", nil},
		{"submission", func(o *WorkflowNativeGitObservation) { o.SubmissionID = "another-submission" }},
		{"revision", func(o *WorkflowNativeGitObservation) {
			o.CandidateRevision = contenthash.Sum([]byte("another-revision"))
		}},
		{"source", func(o *WorkflowNativeGitObservation) { o.SourceCommit = submission.Native.OriginalSource }},
		{"history", func(o *WorkflowNativeGitObservation) { o.Git = nil }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			record := original
			receipt, native := *original.Receipt, *original.Receipt.Native
			receipt.Native, record.Receipt = &native, &receipt

			if test.change == nil {
				native.GitApplicationRef = nil
			} else {
				ref := nativeReceiptAlteredApplication(t, f.store, submission, *native.GitApplicationRef, test.change)
				native.GitApplicationRef = &ref
			}

			if _, err := f.store.CreateReceipt(t.Context(), record); err == nil {
				t.Fatal("unattributed application created a receipt")
			}

			if _, err := os.Stat(filepath.Join(f.store.layout.Root(), record.Path)); !os.IsNotExist(err) {
				t.Fatal("invalid application wrote receipt", err)
			}
		})
	}

	receipt, err := f.store.ConfirmWorkflowNativeReceipt(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID))
	checkNativeReceiptDelivery(t, f, submission, receipt, err)
}

func nativeReceiptRecord(t *testing.T, f nativeSubmissionFixture, submission WorkflowSubmission) WorkflowRecord {
	t.Helper()

	run, err := f.store.readWorkflowInvocation(submission.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	ref := submission.Native.ObservationRefs[len(submission.Native.ObservationRefs)-1]
	observation := nativeCanonicalLast(t, f.store, submission)

	record, err := f.store.workflowNativeReceipt(run, submission, observation, ref)
	if err != nil {
		t.Fatal(err)
	}

	return record
}

func nativeReceiptAlteredApplication(t *testing.T, store *WorkflowStore, submission WorkflowSubmission, ref WorkflowEvidenceRef, change func(*WorkflowNativeGitObservation)) WorkflowEvidenceRef {
	t.Helper()

	var observation WorkflowNativeGitObservation
	if err := store.readWorkflowNativeEvidence(submission.InvocationID, ref, &observation); err != nil {
		t.Fatal(err)
	}

	change(&observation)

	return retainNativeRecordEvidence(t, store, submission.InvocationID, "receipt-application", observation)
}
