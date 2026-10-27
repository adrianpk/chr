package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowOrderFixture(t *testing.T) (*WorkflowStore, WorkflowMaterializationResult, WorkflowTrackerStatus) {
	t.Helper()

	store, opts := workflowOrderPlanFixture(t)

	hierarchy, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	projection, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{Schema: "chr-tracker/v2"})
	if err != nil {
		t.Fatal(err)
	}

	return store, hierarchy, projection
}

// A real claim and host-validated worker result change authoritative processing
// state, while v2 output, metadata, structural hash and generation time stay fixed.
func TestWorkflowOrderProgress(t *testing.T) {
	store, hierarchy, projection := workflowOrderFixture(t)
	output := workflowTrackerRead(t, filepath.Join(store.layout.Root(), projection.Path))
	metadata := workflowTrackerRead(t, filepath.Join(store.layout.Root(), workflowTrackerTestMetadata(t, hierarchy.Set.ID)))
	commit, _ := workflowSelectionCommit(t, store)
	request := WorkflowStartRequest{RequestKey: "ordered-progress", Selection: workflowstate.WorkflowPreviewRequest{Selector: workflowClaimSelector(hierarchy, "ticket"), Target: workflowstate.WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}}}
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)
	checkWorkflowOrderUnchanged(t, store, projection, output, metadata)

	proof, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker"))
	if err != nil || !proof.Successful {
		t.Fatal(proof, err)
	}

	task, err := store.Read(t.Context(), hierarchy.Tasks[0].Path)
	if err != nil || task.Task.Phase != "validated" {
		t.Fatal(task, err)
	}

	checkWorkflowOrderUnchanged(t, store, projection, output, metadata)

	if bytes.Contains(output, []byte("validated")) || !bytes.Contains(output, []byte(hierarchy.Tasks[0].Task.Key)) {
		t.Fatal("order map lost membership or acquired processing state")
	}
}

func checkWorkflowOrderUnchanged(t *testing.T, store *WorkflowStore, projection WorkflowTrackerStatus, output, metadata []byte) {
	t.Helper()

	status, err := store.InspectWorkflowTracker(t.Context(), projection.SetID)
	if err != nil || status.Status != "current" || status.SourceHash != projection.SourceHash {
		t.Fatal("progress invalidated order map", status, err)
	}

	refreshed, err := store.ExportWorkflowTracker(t.Context(), projection.SetID, WorkflowTrackerOptions{})
	if err != nil || !refreshed.GeneratedAt.Equal(projection.GeneratedAt) {
		t.Fatal("progress changed generation time", refreshed, err)
	}

	actual := workflowTrackerRead(t, filepath.Join(store.layout.Root(), projection.Path))

	state := workflowTrackerRead(t, filepath.Join(store.layout.Root(), workflowTrackerTestMetadata(t, projection.SetID)))
	if !bytes.Equal(output, actual) || !bytes.Equal(metadata, state) {
		t.Fatal("processing rewrote v2 projection bytes")
	}
}

// Structural order/title changes alter the projection hash, but writing them
// requires approved amendment. Rejection preserves the existing current map.
func TestWorkflowOrderStructure(t *testing.T) {
	store, hierarchy, first := workflowOrderFixture(t)

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	var changed WorkflowRecord

	for i := range snapshot.Records {
		if snapshot.Records[i].ID == hierarchy.Set.ID {
			snapshot.Records[i].Set.Slices[0].Title = "Different integration title"
			changed = snapshot.Records[i]
		}
	}

	source, output, err := store.workflowOrderTrackerImage(snapshot, hierarchy.Set.ID)
	if err != nil || source.SourceHash == first.SourceHash || !strings.Contains(string(output), "Different integration title") {
		t.Fatal("structural title missing from map", source, err)
	}

	if _, err := store.Write(t.Context(), changed, changed.Revision); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("unapproved structural write accepted", err)
	}

	status, err := store.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID)
	if err != nil || status.Status != "current" {
		t.Fatal("rejected amendment changed existing map", status, err)
	}
}

// Edited v2 output requires the exact observed hash. Repair of unchanged
// structure keeps its original generation time and never imports progress cells.
func TestWorkflowOrderEdited(t *testing.T) {
	store, hierarchy, first := workflowOrderFixture(t)
	path := filepath.Join(store.layout.Root(), first.Path)
	edited := []byte("Operator notes with unsupported processed checkboxes.\n")
	workflowTrackerWrite(t, path, edited)

	if _, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{}); !errors.Is(err, ErrWorkflowProjectionStale) {
		t.Fatal("edited output replaced", err)
	}

	repaired, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{ReplaceHash: contenthash.Sum(edited)})
	if err != nil || !repaired.GeneratedAt.Equal(first.GeneratedAt) {
		t.Fatal("repair lost structural generation time", repaired, err)
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && record.Task.Phase != "pending" {
			t.Fatal("projection edit imported into authority")
		}
	}
}

// Old registration bytes keep their v1 schema. Requesting v2 for an unbound
// native set fails explicitly instead of silently changing the legacy contract.
func TestWorkflowOrderLegacy(t *testing.T) {
	store, hierarchy, old := workflowTrackerFixture(t)

	encoded := workflowTrackerRead(t, filepath.Join(store.layout.Root(), workflowTrackerTestMetadata(t, hierarchy.Set.ID)))
	if _, err := store.InspectWorkflowTracker(t.Context(), hierarchy.Set.ID); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encoded, workflowTrackerRead(t, filepath.Join(store.layout.Root(), workflowTrackerTestMetadata(t, hierarchy.Set.ID)))) {
		t.Fatal("inspection migrated legacy metadata")
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}

	if old.Schema != "chr-tracker/v1" || fields["plan_revision"] != nil {
		t.Fatal("new metadata changed legacy checksum shape")
	}

	if _, err := store.ExportWorkflowTracker(t.Context(), hierarchy.Set.ID, WorkflowTrackerOptions{Schema: "chr-tracker/v2"}); err == nil {
		t.Fatal("unbound set acquired v2")
	}
}

// Receipt payloads do not participate in v2 source discovery or provenance.
// An unavailable receipt image cannot invalidate unchanged order and references.
func TestWorkflowOrderReceiptIndependence(t *testing.T) {
	store, _, projection := workflowOrderFixture(t)
	output := workflowTrackerRead(t, filepath.Join(store.layout.Root(), projection.Path))
	metadata := workflowTrackerRead(t, filepath.Join(store.layout.Root(), workflowTrackerTestMetadata(t, projection.SetID)))
	writeTestFile(t, filepath.Join(store.layout.Root(), "run", "retained-delivery", "evidence", "receipt.json"), "{interrupted receipt image")
	checkWorkflowOrderUnchanged(t, store, projection, output, metadata)
}

// Corrupting immutable planned order cannot supply new provenance during
// inspection or repair, even if mutable native records themselves are valid.
func TestWorkflowOrderSourceIntegrity(t *testing.T) {
	store, _, projection := workflowOrderFixture(t)
	reference := filepath.Join(filepath.Dir(projection.PlanRevision.OrderRef), "planned-order.json")
	workflowTrackerWrite(t, filepath.Join(store.layout.Root(), reference), []byte(`{"schema":"chr-tracker/v2"}`))

	if _, err := store.InspectWorkflowTracker(t.Context(), projection.SetID); !errors.Is(err, ErrWorkflowProjectionStale) {
		t.Fatal("edited immutable source accepted", err)
	}
}

// Lost v2 registration can be rebuilt using exact observed-output permission
// without making receipt payloads an implicit dependency of source discovery.
func TestWorkflowOrderRegistrationRecovery(t *testing.T) {
	store, _, projection := workflowOrderFixture(t)
	writeTestFile(t, filepath.Join(store.layout.Root(), "run", "retained-delivery", "evidence", "receipt.json"), "{unavailable receipt image")

	metadata := filepath.Join(store.layout.Root(), workflowTrackerTestMetadata(t, projection.SetID))
	if err := os.Remove(metadata); err != nil {
		t.Fatal(err)
	}

	repaired, err := store.ExportWorkflowTracker(t.Context(), projection.SetID, WorkflowTrackerOptions{ReplaceHash: projection.OutputHash})
	if err != nil || repaired.Schema != "chr-tracker/v2" || repaired.SourceHash != projection.SourceHash {
		t.Fatal("registration recovery depended on progress receipts", repaired, err)
	}

	if !repaired.GeneratedAt.Equal(projection.GeneratedAt) {
		t.Fatal("registration repair changed structural generation time")
	}

	status, err := store.InspectWorkflowTracker(t.Context(), projection.SetID)
	if err != nil || status.Status != "current" {
		t.Fatal("registration was not restored", status, err)
	}
}

func workflowOrderPlanFixture(t *testing.T) (*WorkflowStore, WorkflowMaterializationOptions) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)
	for i := range opts.Plan.Slices {
		for j := range opts.Plan.Slices[i].Tasks {
			opts.Plan.Slices[i].Tasks[j].Validation = workflowstate.Validation{Required: true, Commands: [][]string{workflowTaskCommand(t, "gate")}, TimeoutSeconds: 30}
		}
	}

	generated := orderedPlanCandidate()
	generated.Scope = quizcore.Scope{ProjectID: store.projectID}

	generated.Order = generationcore.PlannedWorkflowOrder(opts.Plan)
	if _, err := WriteGeneratedPlan(t.Context(), store.layout.ProjectDir(), generated); err != nil {
		t.Fatal(err)
	}

	binding := readTestPlanBinding(t, store.layout.ProjectDir(), generated.Scope)
	opts.Plan.PlanRef, opts.Plan.PlanHash = binding.BodyRef, binding.BodyHash
	opts.Plan.PlanRevision = &binding

	return store, opts
}
