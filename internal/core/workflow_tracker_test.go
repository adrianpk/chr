package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTrackerFixture(t *testing.T) (*WorkflowStore, WorkflowMaterializationResult, WorkflowTrackerStatus) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	projection, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if err != nil {
		t.Fatal(err)
	}

	return store, hierarchy, projection
}

// TestWorkflowTrackerExport binds actual membership and bytes without admitting
// work; unchanged retries and read-only status retain the exact generated image.
func TestWorkflowTrackerExport(t *testing.T) {
	store, hierarchy, first := workflowTrackerFixture(t)
	before := workflowTrackerRead(t, filepath.Join(store.layout.Root(), first.Path))
	metadata := workflowTrackerTestMetadata(t, hierarchy.Set.ID)
	state := workflowTrackerRead(t, filepath.Join(store.layout.Root(), metadata))

	second, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if err != nil || first.SourceHash != second.SourceHash || first.GeneratedAt != second.GeneratedAt || len(second.Sources) != 7 {
		t.Fatal("export retry changed provenance", second, err)
	}

	status, err := store.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID)
	if err != nil || status.Status != "current" || status.OutputHash != contenthash.Sum(before) {
		t.Fatal("status omitted exact output", status, err)
	}

	checkWorkflowTrackerExport(t, store, metadata, state, before, hierarchy)
}

func checkWorkflowTrackerExport(t *testing.T, store *WorkflowStore, metadata string, state, before []byte, hierarchy WorkflowMaterializationResult) {
	t.Helper()

	if string(state) != string(workflowTrackerRead(t, filepath.Join(store.layout.Root(), metadata))) {
		t.Fatal("read-only status wrote metadata")
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil || len(snapshot.Records) != 7 || !strings.Contains(string(before), "state_authority: tickets") || !strings.Contains(string(before), hierarchy.Tasks[5].Task.Key) {
		t.Fatal("export changed native authority or lost membership", err)
	}
}

// TestWorkflowTrackerEdited preserves existing and subsequently edited views
// until replacement names the exact current bytes, without importing edits.
func TestWorkflowTrackerEdited(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "registered", true: "unregistered"}[existing], func(t *testing.T) {
			checkWorkflowTrackerEdited(t, existing)
		})
	}
}

func checkWorkflowTrackerEdited(t *testing.T, existing bool) {
	t.Helper()

	store, hierarchy, projection := workflowTrackerFixture(t)
	if existing {
		metadata := workflowTrackerTestMetadata(t, hierarchy.Set.ID)
		if err := os.Remove(filepath.Join(store.layout.Root(), metadata)); err != nil {
			t.Fatal(err)
		}
	}

	edited := []byte("# Keep operator notes\n")
	path := filepath.Join(store.layout.Root(), projection.Path)
	workflowTrackerWrite(t, path, edited)

	for _, hash := range []string{"", contenthash.Sum([]byte("old image"))} {
		stale, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{ReplaceHash: hash})
		if !errors.Is(err, workflowstate.ErrProjectionStale) || stale.Status != "stale" || string(workflowTrackerRead(t, path)) != string(edited) {
			t.Fatal("edited view was overwritten", stale, err)
		}
	}

	repaired, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{ReplaceHash: contenthash.Sum(edited)})
	if err != nil || repaired.Status != "current" || repaired.ObservedHash == contenthash.Sum(edited) {
		t.Fatal("explicit replacement failed", repaired, err)
	}
}

func workflowTrackerTestMetadata(t *testing.T, id string) string {
	t.Helper()

	path, err := workflowstate.TrackerMetadataPath(id)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

func workflowTrackerWrite(t *testing.T, path string, content []byte) {
	t.Helper()

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func workflowTrackerRead(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// TestWorkflowTrackerReceipt binds a real finalized slice's shared immutable
// receipt once, and distinguishes integrated tasks from the pending next slice.
func TestWorkflowTrackerReceipt(t *testing.T) {
	fixture := workflowVerifiedDelivery(t, "rebase")

	receipt, err := fixture.store.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, fixture.provider, fixture.git)
	if err != nil {
		t.Fatal(err)
	}

	projection, err := fixture.store.ExportWorkflowTracker(t.Context(), fixture.hierarchy.Set.ID, WorkflowTrackerOptions{})
	if err != nil || len(projection.Sources) != 8 {
		t.Fatal("tracker did not retain one shared receipt", projection, err)
	}

	count := 0

	for _, source := range projection.Sources {
		if source.Path == receipt.Path && source.Hash == contenthash.Sum(workflowTrackerRead(t, filepath.Join(fixture.store.layout.Root(), receipt.Path))) {
			count++
		}
	}

	output := string(workflowTrackerRead(t, filepath.Join(fixture.store.layout.Root(), projection.Path)))
	if count != 1 || strings.Count(output, "| done | integrated |") != 3 || !strings.Contains(output, fixture.provider.pr.URL) {
		t.Fatal("view lost receipt-finalized delivery or completed unrelated tasks", output)
	}
}

// TestWorkflowTrackerProgress retains a real host-validated task when an edited
// projection cannot refresh; export repair never claims another task or worker.
func TestWorkflowTrackerProgress(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)

	projection, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if err != nil {
		t.Fatal(err)
	}

	run := workflowTestStart(t, store, request)
	edited := []byte("Operator notes\n")
	workflowTrackerWrite(t, filepath.Join(store.layout.Root(), projection.Path), edited)

	proof := workflowTestExecute(t, store, run.Invocation.ID, "worker")
	if !proof.Successful {
		t.Fatal("projection prevented direct native execution")
	}

	status, err := store.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID)
	if err != nil || status.Status != "stale" || status.Reason != "projection_stale" {
		t.Fatal("domain transition did not report stale projection", status, err)
	}

	_, err = store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{})
	if !errors.Is(err, workflowstate.ErrProjectionStale) {
		t.Fatal("edited projection silently refreshed", err)
	}

	checkWorkflowTrackerProgress(t, store, hierarchy, run.Invocation.ID, edited)
}

func checkWorkflowTrackerProgress(t *testing.T, store *WorkflowStore, hierarchy WorkflowMaterializationResult, id string, edited []byte) {
	t.Helper()

	before := workflowTestRead(t, store, id)

	task, err := store.Read(t.Context(), hierarchy.Tasks[0].Path)
	if err != nil || task.Task.Phase != "validated" || task.Task.Status == ticketcore.StatusDone {
		t.Fatal("tracker error rolled back or integrated task progress", err)
	}

	repaired, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{ReplaceHash: contenthash.Sum(edited)})
	if err != nil || repaired.Status != "current" || workflowTestRead(t, store, id).Revision != before.Revision {
		t.Fatal("repair changed execution instead of its view", repaired, err)
	}

	output := workflowTrackerRead(t, filepath.Join(store.layout.Root(), repaired.Path))
	if !strings.Contains(string(output), "validated") || strings.Contains(string(output), "| done |") {
		t.Fatal("view confused validation with integration")
	}
}
