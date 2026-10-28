package workflow

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowFinalizationSet stops at the current slice receipt and leaves
// later frozen set members untouched rather than completing the whole invocation.
func TestWorkflowFinalizationSet(t *testing.T) {
	fixture := workflowDeliveryExtentFixture(t, "merge", "slice_set")
	if _, err := fixture.store.VerifyWorkflowSliceIntegration(t.Context(), fixture.run.ID, "merge", fixture.provider, fixture.git); err != nil {
		t.Fatal(err)
	}

	receipt, err := fixture.store.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, fixture.provider, fixture.git)
	if err != nil {
		t.Fatal(err)
	}

	run := workflowTestRead(t, fixture.store, fixture.run.ID)
	if run.Cursor != 3 || run.Cursor == len(run.Members) || run.Status != workflowstate.InvocationWaitingIntegration || run.Outcome != "slice_integrated" || run.DeliveryRef != receipt.Path {
		t.Fatal("current slice receipt advanced or completed later selected set members")
	}

	checkWorkflowUnfinalizedTasks(t, fixture)
	checkWorkflowFinalizedTasks(t, fixture, receipt)
	checkWorkflowFinalizedSet(t, fixture, receipt)
}

func checkWorkflowUnfinalizedTasks(t *testing.T, fixture workflowIntegrationFixture) {
	t.Helper()

	for _, task := range fixture.hierarchy.Tasks[3:] {
		record, err := fixture.store.Read(t.Context(), task.Path)
		if err != nil || record.Task.Phase != "pending" || record.Task.DeliveryRef != "" {
			t.Fatal("current slice receipt finalized a later task")
		}
	}
}
