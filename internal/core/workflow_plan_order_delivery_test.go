package core

import (
	"path/filepath"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Real Git rebase integration and receipt finalization complete three native
// tickets without rewriting order bytes or generation metadata. Approved future
// work then retains those completions, receipt references and historical maps.
func TestWorkflowPlanOrderDelivery(t *testing.T) {
	store, hierarchy, run, provider := workflowPRPlanFixture(t, "slice", workflowstate.ExecutionPolicy{}, workflowOrderPlanFixture)

	status, err := store.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID)
	if err != nil {
		t.Fatal(err)
	}

	output := workflowTrackerRead(t, filepath.Join(store.layout.Root(), status.Path))
	metadata := workflowTrackerRead(t, filepath.Join(store.layout.Root(), workflowTrackerTestMetadata(t, hierarchy.Set.ID)))

	fixture := workflowFinishDeliveryFixture(t, "rebase", store, hierarchy, run, provider)
	if _, err := store.VerifyWorkflowSliceIntegration(t.Context(), run.ID, "rebase", provider, fixture.git); err != nil {
		t.Fatal(err)
	}

	receipt, err := store.FinalizeWorkflowSlice(t.Context(), run.ID, provider, fixture.git)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowOrderUnchanged(t, store, status, output, metadata)

	set, err := store.Read(t.Context(), hierarchy.Set.Path)
	if err != nil {
		t.Fatal(err)
	}

	amended, err := store.AmendWorkflowSet(t.Context(), workflowOrderAmendment(t, store, set))
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowPlanOrderDeliveredHistory(t, store, amended, receipt.Path)

	if string(output) != string(workflowTrackerRead(t, filepath.Join(store.layout.Root(), status.PlanRevision.TrackerRef))) {
		t.Fatal("new revision altered delivered map")
	}
}

func checkWorkflowPlanOrderDeliveredHistory(t *testing.T, store *WorkflowStore, result WorkflowMaterializationResult, receipt string) {
	t.Helper()

	for _, task := range result.Tasks[:3] {
		current, err := store.Read(t.Context(), task.Path)
		if err != nil || current.Task.Status != string(ticketcore.StatusDone) || current.Task.DeliveryRef != receipt {
			t.Fatal("amendment reset delivered work", current, err)
		}
	}

	status, err := store.InspectWorkflowTracker(t.Context(), result.Set.ID)
	if err != nil || status.Status != "current" {
		t.Fatal("amended map is stale", status, err)
	}
}
