package core

import "testing"

// TestNativeCandidateReviewRepair keeps accepted source tasks intact when a
// candidate reviewer requests integration repair instead of task execution.
func TestNativeCandidateReviewRepair(t *testing.T) {
	f, _ := nativeReviewFixture(t, ReviewWorkflowModeAgenticReview)
	store, id := f.f.store, f.f.run.ID
	before := workflowTestRead(t, store, id)
	key := before.Members[len(before.Members)-1].Admission.Key

	binding, err := store.ReviewWorkflowSlice(t.Context(), id, f.provider, workflowReviewWorker("needs_changes", key, ""))
	if err != nil || binding.State != "needs_repair" {
		t.Fatal("native review did not retain actionable findings", err)
	}

	after := workflowTestRead(t, store, id)

	entry := nativeReviewEntry(t, store, id)
	if after.Cursor != before.Cursor || after.CandidateHash != before.CandidateHash || entry.State != "merging" {
		t.Fatal("native review rewound task execution instead of routing to the merger")
	}

	assertWorkflowPreparationSource(t, f.f)
}
