package app

import (
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestServiceWorkflowFrontier keeps the shared service behind canonical delivery
// and receipt finalization, then hydrates its next slice without another PR or
// changes to the original invocation identity.
func TestServiceWorkflowFrontier(t *testing.T) {
	operations, id, provider := serviceWorkflowDeliveryExtentFixture(t, "slice_set")
	if _, err := operations.AdvanceWorkflowFrontier(t.Context(), id); err == nil {
		t.Fatal("service advanced an unfinalized slice")
	}

	if _, err := operations.VerifyWorkflowSliceIntegration(t.Context(), id, "rebase"); err != nil {
		t.Fatal(err)
	}

	receipt, err := operations.FinalizeWorkflowSlice(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	run, err := operations.AdvanceWorkflowFrontier(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	checkServiceWorkflowFrontier(t, run, receipt, provider)

	if run.Selection.Target.Commit == run.ExecutionSelection().Target.Commit || run.Lane != nil || run.PullRequest != nil {
		t.Fatal("service changed admission identity or resurrected prior bindings")
	}
}

func checkServiceWorkflowFrontier(t *testing.T, run core.WorkflowInvocation, receipt core.WorkflowRecord, provider *serviceWorkflowPRFixture) {
	t.Helper()

	if run.Status != "pending" || run.ConfirmedCursor != 3 || len(run.Deliveries) != 1 || run.Deliveries[0].DeliveryRef != receipt.Path || run.ExecutionSelection().Target.Commit != provider.pr.ResultCommit || provider.creates != 1 {
		t.Fatalf("service continuation did not preserve the canonical barrier: %+v", run)
	}
}
