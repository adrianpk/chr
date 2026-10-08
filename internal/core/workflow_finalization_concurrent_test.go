package core

import "testing"

// TestWorkflowFinalizationConcurrent serializes two independent host callers
// and returns the identical receipt without repeated task fan-out or provider creation.
func TestWorkflowFinalizationConcurrent(t *testing.T) {
	fixture := workflowVerifiedDelivery(t, "merge")
	results := make(chan WorkflowRecord, 2)
	errors := make(chan error, 2)
	start := make(chan struct{})

	for range 2 {
		go func() {
			<-start

			store := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

			receipt, err := store.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, fixture.provider, fixture.git)
			results <- receipt

			errors <- err
		}()
	}

	close(start)

	for range 2 {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}

	first, second := <-results, <-results
	if first.Path != second.Path || first.SourceHash != second.SourceHash {
		t.Fatal("concurrent callers created different shared receipts")
	}

	checkWorkflowFinalizedSlice(t, fixture, first)
}

// TestWorkflowFinalizationLaterTarget preserves a verified receipt when later
// canonical target commits arrive between proof capture and task finalization.
func TestWorkflowFinalizationLaterTarget(t *testing.T) {
	fixture := workflowVerifiedDelivery(t, "squash")
	fixture.provider.pr.BaseCommit = workflowIntegrationCommit(t, fixture.actor, "after-proof.txt", "Target advanced after proof")
	workflowIntegrationCommand(t, fixture.actor, "push", "origin", "dev")

	receipt, err := fixture.store.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, fixture.provider, fixture.git)
	if err != nil || receipt.Receipt.ObservedTarget != fixture.target {
		t.Fatalf("later target invalidated reachable immutable proof: %v", err)
	}

	checkWorkflowFinalizedSlice(t, fixture, receipt)
}
