package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowSelectionFixture(t *testing.T) (*WorkflowStore, WorkflowMaterializationResult) {
	t.Helper()

	store, opts := workflowMaterializationFixture(t)
	opts.Plan.UseCase = "example"

	result, err := store.MaterializePlan(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	return store, result
}

func workflowTestPreview(t *testing.T, store *WorkflowStore, extent, identity string) WorkflowPreview {
	t.Helper()

	result, err := store.Preview(t.Context(), workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: extent, Identity: identity}, Limit: 128})
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// TestWorkflowSelectionExtents proves exact scope never expands into the backlog.
func TestWorkflowSelectionExtents(t *testing.T) {
	store, result := workflowSelectionFixture(t)

	cases := []struct {
		name, extent, identity string
		count                  int
	}{
		{"ticket", "ticket", result.Tasks[0].Task.Key, 1},
		{"slice", "slice", result.Set.Set.Slices[0].ID, 3},
		{"set", "slice_set", result.Set.ID, 6},
		{"use case", "use_case", "example", 6},
		{"loose excludes owned", "loose", "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			preview := workflowTestPreview(t, store, tc.extent, tc.identity)
			if preview.Total != tc.count || len(preview.Members) != tc.count {
				t.Fatalf("wrong exact extent: %+v", preview)
			}
		})
	}

	for _, selector := range []workflowstate.Selector{{}, {Extent: "ticket"}, {Extent: "ticket", Identity: "absent"}, {Extent: "slice", Identity: "absent"}, {Extent: "slice_set", Identity: "absent"}, {Extent: "use_case", Identity: "absent"}, {Extent: "loose", Identity: "unexpected"}} {
		if _, err := store.Preview(t.Context(), workflowstate.PreviewRequest{Selector: selector, Limit: 1}); err == nil {
			t.Fatalf("invalid selector accepted: %+v", selector)
		}
	}
}

// TestWorkflowSelectionPages rejects stale continuation and oversized pages.
func TestWorkflowSelectionPages(t *testing.T) {
	store, result := workflowSelectionFixture(t)
	request := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "slice_set", Identity: result.Set.ID}, Limit: 2}

	first, err := store.Preview(t.Context(), request)
	if err != nil || first.Total != 6 || len(first.Members) != 2 || first.NextOffset != 2 {
		t.Fatalf("first page = %+v, %v", first, err)
	}

	request.Offset, request.SnapshotHash = first.NextOffset, first.SnapshotHash

	next, err := store.Preview(t.Context(), request)
	if err != nil || next.Members[0].Key == first.Members[0].Key {
		t.Fatalf("next page = %+v, %v", next, err)
	}

	task := result.Tasks[0]
	task.Task.Status = string(ticketcore.StatusBacklog)
	workflowTestWrite(t, store, task)

	if _, err := store.Preview(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("stale page accepted: %v", err)
	}

	for _, limit := range []int{0, 129} {
		request.Limit = limit
		if _, err := store.Preview(t.Context(), request); err == nil {
			t.Fatal("invalid page accepted")
		}
	}
}

// TestWorkflowSelectionExclusions counts unprepared and unsliced records without adoption.
func TestWorkflowSelectionExclusions(t *testing.T) {
	store, _ := workflowSelectionFixture(t)
	independent := workflowTestTask("PRJ-9")

	independent.Task.UseCase = "example"
	if _, err := store.Create(t.Context(), "Independent", independent); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.layout.Root(), "ticket", "legacy.md")

	content := []byte("---\nkey: PRJ-10\ntitle: Legacy\nstatus: ready_for_dev\nuse_case: example\n---\nOriginal\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	preview := workflowTestPreview(t, store, "use_case", "example")
	if preview.Total != 6 || preview.Excluded["unsliced"] != 1 || preview.Excluded["unprepared"] != 1 {
		t.Fatalf("lost exclusion reasons: %+v", preview)
	}

	exact := workflowTestPreview(t, store, "ticket", "PRJ-10")
	if exact.Total != 0 || exact.Excluded["unprepared"] != 1 {
		t.Fatal("legacy ticket became executable")
	}

	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(content) {
		t.Fatal("preview changed legacy content")
	}
}

// TestWorkflowSelectionPendingJournal proves read-only preview cannot repair or overwrite an interrupted transition.
func TestWorkflowSelectionPendingJournal(t *testing.T) {
	store := workflowTestStore(t)
	record := workflowTestCreate(t, store, "PRJ-1")
	store.afterReplace = func(int) error { return errors.New("interrupted replacement") }

	record.Task.Status = string(ticketcore.StatusBacklog)
	if _, err := store.Write(t.Context(), record, record.Revision); err == nil {
		t.Fatal("interruption not injected")
	}

	path := filepath.Join(store.layout.Root(), workflowstate.JournalPath)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	request := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: "PRJ-1"}, Limit: 1}
	if _, err := store.Preview(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("pending journal preview = %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatal("preview changed recovery journal")
	}

	store.afterReplace = nil
	if _, err := store.Snapshot(t.Context()); err != nil {
		t.Fatal(err)
	}

	workflowTestPreview(t, store, "ticket", "PRJ-1")
}
