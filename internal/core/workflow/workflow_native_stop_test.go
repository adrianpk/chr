package workflow

import (
	"context"
	"errors"
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeStopPending retains both stop pointers until possible publication is
// reconciled. Neither pause nor cancellation can resume work or replace intent.
func TestNativeStopPending(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) {
			checkNativeStopPending(t, intent)
		})
	}
}

// TestNativeStopBeforeSend requests stop at the trusted publisher boundary.
// The host rechecks intent before commands, then retains quiescent actual output.
func TestNativeStopBeforeSend(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, remote := nativeDispatchPublisher(t, f)
	transport := publisher.publisher
	publisher.publisher = nativeOutcomePublisherFixture{publish: func(ctx context.Context, request GitPublication) (workflowstate.GitPublicationResult, error) {
		if _, err := f.store.StopWorkflow(ctx, f.entry.InvocationID, "cancel"); err != nil {
			t.Fatal(err)
		}

		return transport.PublishWorkflowGit(ctx, request)
	}}

	nativePublication, err := PublishNativeSubmission(t.Context(), f.store, nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)

	submission, result := nativePublication.Submission, nativePublication.Transport
	if err == nil || result.CommandsSent || result.State != "not_sent" || submission.Native.DispatchState != workflowstate.NativeDispatchQuiescent {
		t.Fatal("stopped host sent commands or lost completion", err)
	}

	b := f.candidate.Native
	for ref, want := range map[string]string{b.SourceRef: b.OriginalSource, b.TargetRef: b.ValidatedBase} {
		if actual := workflowIntegrationCommand(t, remote, "rev-parse", ref); actual != want {
			t.Fatal("stop at send boundary changed actual refs")
		}
	}

	nativeSubmissionPendingExcluded(t, f, submission)
}

// TestNativeStopRejected finalizes cancellation only after independent refs and
// completed rejection seal changed-input non-application. Counters stay consumed.
func TestNativeStopRejected(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, remote := nativeDispatchPublisher(t, f)
	b := f.candidate.Native
	transport := NewWorkflowGitPublisher(b.SourceRepository, remote)
	publisher.publisher = nativeOutcomePublisher(t, f, remote, transport, "source race")

	nativePublication2, err := PublishNativeSubmission(t.Context(), f.store, nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)
	submission := nativePublication2.Submission

	if err == nil {
		t.Fatal("source race did not reject the transaction")
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	stopped, err := f.store.StopWorkflow(t.Context(), submission.InvocationID, "cancel")
	if err != nil || stopped.Status == workflowstate.InvocationCancelled {
		t.Fatal("raw rejection finalized stop without independent refs", err)
	}

	submission, err = ReconcileNativeGitSubmission(t.Context(), f.store, nativeBudgetGuard(t, f.store, stopped.ID), transport, NewWorkflowIntegrationGit(b.SourceRepository, remote))
	if err != nil || submission.State != workflowstate.SubmissionUnpublished {
		t.Fatal("definitive changed-input rejection was not sealed", err)
	}

	checkNativeRejectedStop(t, f, stopped, budget)
}

// TestNativeStopReceipt keeps receipt-confirmed work available for required
// fanout. Stop cannot make that unfinished finalization disappear as cancellation.
func TestNativeStopReceipt(t *testing.T) {
	fixture := newNativeCanonicalFixture(t)
	f := fixture.f

	submission, err := f.store.ReadWorkflowSubmission(t.Context(), f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	record := nativeRecordReceipt(t, f.store, f.candidate)
	record.Receipt.Native.TransactionProofRef = *submission.Native.TransportResponseRef
	nativeSubmissionConfirmReceipt(t, f, submission, record)

	confirmed, err := f.store.ReadWorkflowSubmission(t.Context(), f.entry.InvocationID)
	if err != nil || confirmed.State != workflowstate.SubmissionConfirmed {
		t.Fatal("physical receipt was not confirmed", err)
	}

	stopped, err := f.store.StopWorkflow(t.Context(), confirmed.InvocationID, "cancel")
	if err != nil || stopped.Status == workflowstate.InvocationCancelled || stopped.StopIntent != "cancel" {
		t.Fatal("stop discarded receipt finalization", err)
	}

	retained, err := f.store.ReadWorkflowSubmission(t.Context(), confirmed.InvocationID)
	if err != nil || !workflowNativeJSONEqual(retained, confirmed) {
		t.Fatal("stop rewrote the confirmed receipt", err)
	}
}

func checkNativeStopPending(t *testing.T, intent string) {
	t.Helper()

	fixture := newNativeCanonicalFixture(t)
	f := fixture.f

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	stopped, err := f.store.StopWorkflow(t.Context(), f.entry.InvocationID, intent)
	if err != nil || stopped.Status == workflowstate.StopStatus(intent) || stopped.StopIntent != intent {
		t.Fatal("unreconciled native application became terminal stop", stopped.Status, err)
	}

	entry := nativeReviewEntry(t, f.store, stopped.ID)
	if entry.StopIntent != intent || entry.State != "observing" {
		t.Fatal("native stop lost the queue intent", entry)
	}

	again, err := f.store.StopWorkflow(t.Context(), stopped.ID, intent)
	if err != nil || again.Revision != stopped.Revision {
		t.Fatal("repeated pending stop was not idempotent", err)
	}

	if _, err := f.store.ResumeWorkflow(t.Context(), stopped.ID); !errors.Is(err, workflowstate.ErrOwned) {
		t.Fatal("pending publication could resume task execution", err)
	}

	checkNativeStoppedMetadata(t, fixture, stopped, budget, intent)
}

func checkNativeStoppedMetadata(t *testing.T, fixture nativeCanonicalFixture, stopped workflowstate.Invocation, budget IntegrationBudget, intent string) {
	t.Helper()

	f := fixture.f
	pr := fixture.forge.request.PullRequest
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, f.candidate.Native.PreparedResult
	setNativePRFixture(t, fixture.forge, pr)

	submission, err := ObserveNativeSubmission(t.Context(), f.store, nativeBudgetGuard(t, f.store, stopped.ID), fixture.forge.adapter, fixture.git)
	if err != nil || nativeCanonicalLast(t, f.store, submission).State != workflowstate.NativeMerged {
		t.Fatal("stopped intent could not reconcile actual delivery metadata", err)
	}

	nativeCanonicalPending(t, fixture, submission, budget)

	if after := workflowTestRead(t, f.store, stopped.ID); after.Status == workflowstate.StopStatus(intent) || after.StopIntent != intent {
		t.Fatal("merged metadata finalized stop before receipt/fanout")
	}
}

func checkNativeRejectedStop(t *testing.T, f nativeSubmissionFixture, stopped workflowstate.Invocation, budget IntegrationBudget) {
	t.Helper()

	entry := nativeReviewEntry(t, f.store, stopped.ID)
	if after := workflowTestRead(t, f.store, stopped.ID); after.Status != workflowstate.InvocationCancelled || entry.State != "cancelled" || entry.StopIntent != "cancel" {
		t.Fatal("sealed non-application did not finalize retained cancellation")
	}

	if after, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), stopped.ID); err != nil || !reflect.DeepEqual(after, budget) {
		t.Fatal("cancelled rejection refunded consumed budget", err)
	}
}
