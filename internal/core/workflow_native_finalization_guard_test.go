package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNativeFanoutEvidence refuses corrupted receipt, prepared candidate and
// delivery proof before changing any member or confirmed cursor. Restoring the
// original immutable image permits the same receipt to finish exactly once.
func TestNativeFanoutEvidence(t *testing.T) {
	fixture, receipt := nativeFanoutReceipt(t, newNativeCanonicalFixture(t))
	f := fixture.f

	run := nativeFanoutRun(t, fixture)
	for _, test := range []struct{ name, path string }{
		{"receipt", receipt.Path},
		{"candidate", f.entry.CandidateRef.Path},
		{"proof", run.DeliveryProofRef},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(f.store.layout.Root(), test.path)

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
				t.Fatal(err)
			}

			_, failure := f.store.FinalizeWorkflowSlice(t.Context(), run.ID, nil, nil)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}

			after := nativeFanoutRun(t, fixture)
			if failure == nil || after.Revision != run.Revision || after.ConfirmedCursor != run.ConfirmedCursor || nativeReviewEntry(t, f.store, run.ID).State != "finalizing" {
				t.Fatal("invalid proof changed unfinished fanout", failure)
			}
		})
	}

	if _, err := f.store.FinalizeWorkflowSlice(t.Context(), run.ID, nil, nil); err != nil {
		t.Fatal(err)
	}

	checkNativeFanoutMembers(t, f.store, nativeFanoutRun(t, fixture), receipt)
}

// TestNativeFanoutBarrier prevents an optimistic queue transition from declaring
// integration before the existing member/set transaction actually completes.
func TestNativeFanoutBarrier(t *testing.T) {
	fixture, receipt := nativeFanoutReceipt(t, newNativeCanonicalFixture(t))
	f := fixture.f
	run := nativeFanoutRun(t, fixture)
	entry := nativeReviewEntry(t, f.store, run.ID)

	request := WorkflowIntegrationTransition{InvocationID: run.ID, InvocationRevision: run.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, CandidateHash: entry.CandidateRef.Hash, State: "integrated"}
	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), request); err == nil {
		t.Fatal("queue became integrated with unfinished members")
	}

	if workflowTestRead(t, f.store, run.ID).Revision != run.Revision || nativeReviewEntry(t, f.store, run.ID).State != "finalizing" {
		t.Fatal("rejected transition changed queue owner")
	}

	if _, err := f.store.FinalizeWorkflowSlice(t.Context(), run.ID, nil, nil); err != nil {
		t.Fatal(err)
	}

	checkNativeFanoutMembers(t, f.store, nativeFanoutRun(t, fixture), receipt)
}
