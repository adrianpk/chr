package workflow

import (
	"errors"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeReceiptUncertainty retains explicit reconciliation when fetch fails
// or merged metadata disappears. Neither case proves a destructive target
// rewrite, invalidates old receipt bytes or permits another publication.
func TestNativeReceiptUncertainty(t *testing.T) {
	for _, condition := range []string{"transport", "metadata"} {
		t.Run(condition, func(t *testing.T) {
			fixture, receipt := nativeReceiptObservedFixture(t, "rebase", "slice")

			before := nativeReceiptPreservedFiles(t, fixture, receipt)

			git, reason := fixture.git, "canonical_target_unverified"
			if condition == "transport" {
				git = NewWorkflowIntegrationGit(fixture.f.candidate.Native.SourceRepository, filepath.Join(t.TempDir(), "missing.git"))
			} else {
				setNativePRFixture(t, fixture.forge, fixture.forge.request.PullRequest)

				reason = "canonical_receipt_pr_unmerged"
			}

			f := fixture.f

			submission, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, git)
			if errors.Is(err, workflowstate.ErrHistoryDivergence) || submission.State != workflowstate.SubmissionConfirmed {
				t.Fatal("uncertain observation invented removed history", err)
			}

			entry := nativeReviewEntry(t, f.store, f.entry.InvocationID)
			if entry.State != "blocked" || entry.Blocker.Reason != reason {
				t.Fatal("receipt uncertainty was not visible", entry)
			}

			checkNativeReceiptPreserved(t, fixture, receipt, before)
		})
	}
}

// TestNativeReceiptRecovery repairs the owner projection after retained common
// observation IO fails. Recovery invalidates the old callback revision before
// any new IO; the original receipt and consumed reservation remain unchanged.
func TestNativeReceiptRecovery(t *testing.T) {
	fixture, receipt := nativeReceiptObservedFixture(t, "rebase", "slice")
	f, b := fixture.f, fixture.f.candidate.Native.NativeRecordBinding

	before := nativeReceiptPreservedFiles(t, fixture, receipt)
	guard := nativeBudgetGuard(t, f.store, f.entry.InvocationID)
	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	cut := errors.New("receipt observation checkpoint interrupted")

	f.store.afterReplace = func(index int) error {
		if index == -1 {
			return cut
		}

		return nil
	}

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), guard, fixture.forge.adapter, fixture.git); !errors.Is(err, cut) {
		t.Fatal("common observation checkpoint was not interrupted", err)
	}

	f.store.afterReplace = nil
	f.store = NewWorkflowStore(f.store.layout, f.store.projectID)

	fixture.f.store = f.store

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), guard, fixture.forge.adapter, fixture.git); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("recovered owner did not fence its old callback", err)
	}

	if entry := nativeReviewEntry(t, f.store, f.entry.InvocationID); entry.State != "blocked" || entry.Blocker.Reason != "canonical_target_history_diverged" {
		t.Fatal("retained divergence lost its owner projection")
	}

	checkNativeReceiptPreserved(t, fixture, receipt, before)

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git); !errors.Is(err, workflowstate.ErrHistoryDivergence) {
		t.Fatal("fresh owner could not observe retained divergence", err)
	}

	checkNativeReceiptPreserved(t, fixture, receipt, before)
}

// TestNativeReceiptTransition prevents operator callbacks from clearing known
// divergence by advancing queue state without fresh canonical receipt proof.
func TestNativeReceiptTransition(t *testing.T) {
	fixture, _ := nativeReceiptObservedFixture(t, "rebase", "slice")
	f, b := fixture.f, fixture.f.candidate.Native.NativeRecordBinding
	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git); !errors.Is(err, workflowstate.ErrHistoryDivergence) {
		t.Fatal(err)
	}

	run := nativeFanoutRun(t, fixture)
	entry := nativeReviewEntry(t, f.store, run.ID)

	request := IntegrationTransition{InvocationID: run.ID, InvocationRevision: run.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: entry.CandidateRef.Hash, State: "finalizing"}
	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), request); !errors.Is(err, ErrWorkflowObservationRequired) {
		t.Fatal("queue callback skipped receipt reconciliation", err)
	}

	if workflowTestRead(t, f.store, run.ID).Revision != run.Revision {
		t.Fatal("rejected callback changed divergence")
	}
}

// TestNativeReceiptBeforeFanout keeps a confirmed receipt finalizing after
// target restoration, until the original member transaction actually completes.
func TestNativeReceiptBeforeFanout(t *testing.T) {
	fixture, receipt := nativeFanoutReceipt(t, newNativeCanonicalFixture(t))
	f, b := fixture.f, fixture.f.candidate.Native.NativeRecordBinding
	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git); !errors.Is(err, workflowstate.ErrHistoryDivergence) {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.PreparedResult, b.ValidatedBase)

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git); err != nil {
		t.Fatal(err)
	}

	if nativeReviewEntry(t, f.store, f.entry.InvocationID).State != "finalizing" {
		t.Fatal("re-observation completed missing fanout")
	}

	if _, err := f.store.FinalizeWorkflowSlice(t.Context(), f.entry.InvocationID, nil, nil); err != nil {
		t.Fatal(err)
	}

	checkNativeFanoutMembers(t, f.store, nativeFanoutRun(t, fixture), receipt)
}

// TestNativeReceiptTargetEvidence rejects a later observation whose displayed
// target is malformed or differs from its verified Git proof, before changing
// the common journal or owner revision.
func TestNativeReceiptTargetEvidence(t *testing.T) {
	fixture, receipt := nativeReceiptObservedFixture(t, "rebase", "slice")
	f := fixture.f

	submission, err := f.store.ReadWorkflowSubmission(t.Context(), f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	original := nativeCanonicalLast(t, f.store, submission)

	before := workflowTestRead(t, f.store, f.entry.InvocationID)
	for _, test := range []struct{ name, target string }{
		{"malformed", "invalid-target"},
		{"mismatch", f.candidate.Native.ValidatedBase},
	} {
		t.Run(test.name, func(t *testing.T) {
			observation := original
			observation.ObservedTargetCommit = test.target
			ref := retainNativeRecordEvidence(t, f.store, submission.InvocationID, "target-observation", observation)
			update := nativeSubmissionUpdate(t, f, submission)

			update.Facts.ObservationRefs = append(update.Facts.ObservationRefs, ref)
			if _, err := f.store.RetainWorkflowNativeSubmission(t.Context(), update); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("invalid target observation changed publication", err)
			}

			if workflowTestRead(t, f.store, f.entry.InvocationID).Revision != before.Revision {
				t.Fatal("rejected observation changed owner")
			}
		})
	}

	if record, err := f.store.Read(t.Context(), receipt.Path); err != nil || record.SourceHash != receipt.SourceHash {
		t.Fatal("rejected observation rewrote receipt", err)
	}
}

// TestNativeReceiptStopped preserves the stop intent of a delivered prefix when
// later history diverges. Reconciliation remains visible on the retained queue
// entry while cancellation, completed members and receipt bytes stay intact.
func TestNativeReceiptStopped(t *testing.T) {
	fixture, receipt := nativeReceiptObservedFixture(t, "rebase", "slice_set")
	f, b := fixture.f, fixture.f.candidate.Native.NativeRecordBinding

	before := nativeReceiptPreservedFiles(t, fixture, receipt)
	if _, err := f.store.StopWorkflow(t.Context(), f.entry.InvocationID, "cancel"); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, fixture.remote, "update-ref", b.TargetRef, b.ValidatedBase, b.PreparedResult)

	if _, err := f.store.ObserveWorkflowNativeSubmission(t.Context(), nativeBudgetGuard(t, f.store, f.entry.InvocationID), fixture.forge.adapter, fixture.git); !errors.Is(err, workflowstate.ErrHistoryDivergence) {
		t.Fatal(err)
	}

	run := workflowTestRead(t, f.store, f.entry.InvocationID)
	if run.Status != workflowstate.InvocationCancelled || run.StopIntent != "cancel" || nativeReviewEntry(t, f.store, run.ID).State != "blocked" {
		t.Fatal("receipt observation erased stop or reconciliation")
	}

	checkNativeReceiptPreserved(t, fixture, receipt, before)
}
