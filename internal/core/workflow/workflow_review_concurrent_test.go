package workflow

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowReviewConcurrent serializes independent trusted-policy callers
// and creates one immutable decision without consuming worker capacity.
func TestWorkflowReviewConcurrent(t *testing.T) {
	store, _, run, provider := workflowReviewFixture(t, workflowstate.ReviewTrustedAutoMerge)
	start := make(chan struct{})
	results := make(chan error, 2)

	for range 2 {
		go func() {
			<-start

			caller := NewWorkflowStore(store.layout, store.projectID)

			_, err := caller.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil)
			results <- err
		}()
	}

	close(start)

	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}

	retained := workflowTestRead(t, store, run.ID)
	if retained.ReviewCount != 1 || retained.Review.DecisionCount != 1 || len(retained.Attempts) != len(run.Attempts) {
		t.Fatal("concurrent review duplicated evidence or source work")
	}
}
