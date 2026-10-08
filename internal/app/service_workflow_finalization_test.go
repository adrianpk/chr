package app

import "testing"

// TestServiceWorkflowFinalization applies and reuses one real remote receipt
// through the central service without allocating another PR or worker.
func TestServiceWorkflowFinalization(t *testing.T) {
	operations, id, provider := serviceWorkflowDeliveryFixture(t)
	if _, err := operations.VerifyWorkflowSliceIntegration(t.Context(), id, "rebase"); err != nil {
		t.Fatal(err)
	}

	receipt, err := operations.FinalizeWorkflowSlice(t.Context(), id)
	if err != nil || receipt.Receipt.ResultCommit != provider.pr.ResultCommit {
		t.Fatalf("service finalization: %+v: %v", receipt, err)
	}

	again, err := operations.FinalizeWorkflowSlice(t.Context(), id)
	if err != nil || again.SourceHash != receipt.SourceHash || provider.creates != 1 {
		t.Fatal("service finalization repeated canonical delivery")
	}

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || run.Status != "completed" || run.DeliveryRef != receipt.Path {
		t.Fatal("service finalization lost the shared receipt")
	}
}
