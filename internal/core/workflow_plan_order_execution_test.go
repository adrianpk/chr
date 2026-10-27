package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Worker input contains the current canonical order document. Missing required
// maps block input preparation, and native repair restores context without work.
func TestWorkflowPlanOrderContext(t *testing.T) {
	store, hierarchy, status := workflowOrderFixture(t)
	commit, _ := workflowSelectionCommit(t, store)
	run := workflowTestStart(t, store, WorkflowStartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "ordered-context",
		Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "ticket"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
	}})
	claim := workflowTestClaim(t, store, run.Invocation.ID)

	path := filepath.Join(store.layout.Root(), status.PlanRevision.TrackerRef)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if _, err := store.prepareWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker")); !errors.Is(err, ErrWorkflowProjectionStale) {
		t.Fatal("missing map allowed worker input", err)
	}

	if _, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{}); err != nil {
		t.Fatal("native map repair failed", err)
	}

	prepared, err := store.prepareWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker"))
	if err != nil || prepared.input.OrderMap == nil || prepared.input.OrderMap.SourceHash != status.SourceHash {
		t.Fatal("order map absent from worker context", err)
	}

	if prepared.input.OrderMap.Content != string(workflowTrackerRead(t, filepath.Join(store.layout.Root(), status.Path))) {
		t.Fatal("worker received another order map")
	}
}

// Repair requires exactly observed permission for an edited retained map,
// preserving progress and the original structural generation time.
func TestWorkflowPlanOrderRepair(t *testing.T) {
	store, hierarchy, status := workflowOrderFixture(t)
	path := filepath.Join(store.layout.Root(), status.PlanRevision.TrackerRef)
	edited := []byte("Manual retained-map notes.\n")
	workflowTrackerWrite(t, path, edited)

	if _, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{}); !errors.Is(err, ErrWorkflowProjectionStale) {
		t.Fatal("edited retained map overwritten", err)
	}

	repaired, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{ReplaceHash: contenthash.Sum(edited)})
	if err != nil || !repaired.GeneratedAt.Equal(status.GeneratedAt) {
		t.Fatal(repaired, err)
	}

	if _, err := store.requiredWorkflowOrderMap(t.Context(), hierarchy.Tasks[0]); err != nil {
		t.Fatal("repaired context is stale", err)
	}
}
