package workflow

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeSubmissionApplied retains actual two-ref Git proof while the original
// PR is still pending, then confirms only a physically verified native receipt.
func TestNativeSubmissionApplied(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	submission := beginNativeSubmission(t, f)
	record := nativeRecordReceipt(t, f.store, f.candidate)
	update := nativeSubmissionApplied(t, f, submission, record)

	pending, err := f.store.RetainWorkflowNativeSubmission(t.Context(), update)
	if err != nil || pending.State != "awaiting_observation" || pending.Native.ApplicationState != "git_applied" || pending.Native.ReceiptRef != nil {
		t.Fatal("Git application manufactured delivery", err)
	}

	nativeSubmissionPendingExcluded(t, f, submission)
	nativeSubmissionConfirmReceipt(t, f, pending, record)
}

func nativeSubmissionConfirmReceipt(t *testing.T, f nativeSubmissionFixture, pending workflowstate.Submission, record Record) {
	t.Helper()

	submission := pending
	update := nativeSubmissionUpdate(t, f, pending)

	record.Receipt.SubmissionID = submission.ID
	record.Receipt.Native.CandidateRef = submission.Native.CandidateRef

	run, err := f.store.readWorkflowInvocation(f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	record.ID, record.Path = workflowNativeReceiptID(run, submission), workflowNativeReceiptPath(run, submission)

	created, err := f.store.CreateReceipt(t.Context(), record)
	if err != nil {
		t.Fatal(err)
	}

	update = nativeSubmissionUpdate(t, f, pending)
	update.Facts.ObservationRefs = record.Receipt.Native.ObservationRefs
	update.Facts.ReceiptRef = &workflowstate.EvidenceRef{Path: created.Path, Hash: created.SourceHash}

	confirmed, err := f.store.RetainWorkflowNativeSubmission(t.Context(), update)
	if err != nil || confirmed.State != "confirmed" || confirmed.Native.ReceiptRef == nil || confirmed.Proof.ResultCommit != f.candidate.Native.PreparedResult {
		t.Fatal("exact original PR receipt not confirmed", err)
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("observation consumed another submission", err)
	}
}

func nativeSubmissionApplied(t *testing.T, f nativeSubmissionFixture, submission workflowstate.Submission, record Record) NativeSubmissionUpdate {
	t.Helper()

	update := nativeSubmissionUpdate(t, f, submission)

	identity, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	update.Facts.DispatchState, update.Facts.ApplicationState, update.Facts.DispatchIdentity = "dispatching", "git_applied", &identity
	native := update.Facts

	var capability workflowstate.NativeCapability
	if err := f.store.readWorkflowNativeEvidence(native.InvocationID, native.CapabilityRef, &capability); err != nil {
		t.Fatal(err)
	}

	advertisement := GitPublicationAdvertisement{SourceCommit: native.OriginalSource, TargetCommit: native.ValidatedBase, Fingerprint: capability.ServerCapabilityFingerprint}
	manifest := NativeCommandManifest{SchemaVersion: 1, SubmissionID: submission.ID, Atomic: true, SourceRef: native.SourceRef, ExpectedSource: native.OriginalSource, NewSource: native.PreparedResult, TargetRef: native.TargetRef, ExpectedTarget: native.ValidatedBase, NewTarget: native.PreparedResult}
	advRef := retainNativeRecordEvidence(t, f.store, native.InvocationID, "actual-advertisement", advertisement)
	commandRef := retainNativeRecordEvidence(t, f.store, native.InvocationID, "actual-commands", manifest)
	update.Facts.AdvertisementRef, update.Facts.CommandManifestRef = &advRef, &commandRef
	update.Facts.TransportResponseRef = &record.Receipt.Native.TransactionProofRef

	return update
}

// TestNativeSubmissionNonApplication requires changed inputs and a dead dispatch;
// a live process, unknown exchange or timeout cannot free the common barrier.
func TestNativeSubmissionNonApplication(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	submission := beginNativeSubmission(t, f)
	update := nativeSubmissionUpdate(t, f, submission)
	identity, stop := nativeSubmissionDispatcher(t)

	update.Facts.DispatchState, update.Facts.ApplicationState, update.Facts.DispatchIdentity = "quiescent", "not_applied", &identity
	proof := NativeNonApplicationProof{SchemaVersion: 1, SubmissionID: submission.ID, CannotApply: true, InputChanged: true, ObservedSource: f.candidate.Native.OriginalSource, ObservedTarget: f.candidate.Native.PreparedResult}
	ref := retainNativeRecordEvidence(t, f.store, submission.InvocationID, "no-application", proof)

	update.Facts.NonApplicationRef = &ref
	if _, err := f.store.RetainWorkflowNativeSubmission(t.Context(), update); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("live dispatch released target", err)
	}

	stop()

	sealed, err := f.store.RetainWorkflowNativeSubmission(t.Context(), update)
	if err != nil || sealed.State != "unpublished" {
		t.Fatal("quiescent changed-input proof was not retained", err)
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil || budget.UnitCounters[0].Submission != 1 {
		t.Fatal("sealed attempt was refunded", err)
	}
}

func nativeSubmissionPendingExcluded(t *testing.T, f nativeSubmissionFixture, submission workflowstate.Submission) {
	t.Helper()

	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)

	submissionIntent, err := f.store.BeginWorkflowNativeSubmission(t.Context(), guard, f.provider)

	again, dispatch := submissionIntent.Submission, submissionIntent.Dispatch
	if err != nil || dispatch || again.ID != submission.ID {
		t.Fatal("forge lag permitted another publication", err)
	}

	if err := f.store.guardWorkflowSubmissionTarget(t.Context(), contenthash.Sum([]byte("another publication")), contenthash.Sum([]byte(f.candidate.Native.SourceRepository+"\n"+f.candidate.Native.TargetRef))); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("applied Git released target exclusion", err)
	}
}

func nativeSubmissionDispatcher(t *testing.T) (workflowstate.ProcessIdentity, func()) {
	t.Helper()

	command := exec.Command("cat")

	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := command.Start(); err != nil {
		t.Fatal(err)
	}

	var once sync.Once

	stop := func() {
		once.Do(func() {
			if err := input.Close(); err != nil {
				t.Error(err)
			}

			if err := command.Wait(); err != nil {
				t.Error(err)
			}
		})
	}
	t.Cleanup(stop)

	identity, err := workflowProcessIdentity(command.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}

	return identity, stop
}
