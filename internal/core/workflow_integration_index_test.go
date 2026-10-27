package core

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeQueueIndex proves index deletion/corruption is recoverable without
// changing source records, and an interrupted source update invalidates rows.
func TestNativeQueueIndex(t *testing.T) {
	store, admission := nativeEntryFixture(t)
	entry := nativeEntryAdmit(t, store, admission)

	index, ok := workflowTestIndex(t, filepath.Join(t.TempDir(), "queue.sqlite")).(*workflowTicketIndex)
	if !ok {
		t.Fatal("index lacks SQLite owner")
	}

	before := nativeQueueImages(t, store.layout.Root())
	if err := index.RebuildWorkflowIntegrationIndex(t.Context(), store); err != nil {
		t.Fatal(err)
	}

	q := WorkflowIntegrationQueueQuery{Limit: 10}

	page, err := index.QueryWorkflowIntegrationQueue(t.Context(), store, q)
	if err != nil || len(page.Entries) != 1 || page.Entries[0].Entry.ID != entry.ID {
		t.Fatal(page, err)
	}

	nativeQueueRepairCache(t, index, store, q)

	if !reflect.DeepEqual(before, nativeQueueImages(t, store.layout.Root())) {
		t.Fatal("cache repair changed authority")
	}

	nativeQueueStaleIndex(t, index, store, entry, q)
}

func nativeQueueStaleIndex(t *testing.T, index *workflowTicketIndex, store *WorkflowStore, entry WorkflowIntegrationEntry, q WorkflowIntegrationQueueQuery) {
	t.Helper()

	if _, err := store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, entry, "preparing")); err != nil {
		t.Fatal(err)
	}

	if _, err := index.QueryWorkflowIntegrationQueue(t.Context(), store, q); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("stale projection returned work", err)
	}

	if err := index.RebuildWorkflowIntegrationIndex(t.Context(), store); err != nil {
		t.Fatal(err)
	}

	page, err := index.QueryWorkflowIntegrationQueue(t.Context(), store, q)
	if err != nil || page.Entries[0].Entry.State != "preparing" {
		t.Fatal(page, err)
	}
}

func nativeQueueRepairCache(t *testing.T, index *workflowTicketIndex, store *WorkflowStore, q WorkflowIntegrationQueueQuery) {
	t.Helper()

	for _, query := range []string{"DELETE FROM workflow_integration_entries", "UPDATE workflow_integration_entries SET image_hash = 'corrupt'", "DELETE FROM workflow_integration_snapshots"} {
		if _, err := index.db.ExecContext(t.Context(), query); err != nil {
			t.Fatal(err)
		}

		if _, err := index.QueryWorkflowIntegrationQueue(t.Context(), store, q); !errors.Is(err, workflowstate.ErrConflict) {
			t.Fatal("invalid projection returned work", err)
		}

		if err := index.RebuildWorkflowIntegrationIndex(t.Context(), store); err != nil {
			t.Fatal(err)
		}
	}
}
