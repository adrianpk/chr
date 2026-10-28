package workflow

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeDispatchEvidence requires durable transport facts after real IO.
// A response alone still cannot create a delivery receipt or release the target.
func TestNativeDispatchEvidence(t *testing.T) {
	f := newNativeSubmissionFixture(t)
	publisher, _ := nativeDispatchPublisher(t, f)

	nativePublication, err := PublishNativeSubmission(t.Context(), f.store, nativeBudgetGuard(t, f.store, f.entry.InvocationID), f.provider, publisher)
	submission := nativePublication.Submission

	if err != nil {
		t.Fatal(err)
	}

	native := submission.Native
	if native.DispatchState != workflowstate.NativeDispatchQuiescent || native.TransportResponseRef == nil || native.AdvertisementRef == nil || native.CommandManifestRef == nil {
		t.Fatal("completed transport facts were not retained")
	}

	if submission.State != workflowstate.SubmissionAwaitingObservation || native.ReceiptRef != nil {
		t.Fatal("transport response manufactured delivery")
	}

	nativeSubmissionPendingExcluded(t, f, submission)
}
