package core

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTestIndex(t *testing.T, path string) WorkflowIndex {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}

	db.SetMaxOpenConns(1)

	registry, err := projectcore.NewRegistryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatal(err)
	}

	addTestProject(t, registry, projectcore.Project{ID: "project-1", Name: "example", Path: "/tmp/example", Enabled: true, RegisteredAt: time.Now().UTC()})

	index, err := NewTicketIndexFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := index.Close(); err != nil {
			t.Error(err)
		}
	})

	workflow, ok := index.(WorkflowIndex)
	if !ok {
		t.Fatal("ticket index lacks workflow projection")
	}

	return workflow
}

func workflowTestGroupedTask(t *testing.T, store *WorkflowStore, setID, key string, order int) WorkflowRecord {
	t.Helper()

	r := workflowTestTask(key)
	r.Task.SetID, r.Task.SliceID, r.Task.Order = setID, "slice-1", order

	created, err := store.Create(context.Background(), key, r)
	if err != nil {
		t.Fatal(err)
	}

	return created
}

func workflowTestQuery(t *testing.T, index WorkflowIndex, store *WorkflowStore, q WorkflowQuery) WorkflowPage {
	t.Helper()

	page, err := index.QueryWorkflow(context.Background(), store, q)
	if err != nil {
		t.Fatal(err)
	}

	return page
}

// TestWorkflowIndex checks durable ordering, bounded pages and stale-cache rejection.
func TestWorkflowIndex(t *testing.T) {
	store := workflowTestStore(t)
	index := workflowTestIndex(t, ":memory:")

	set, err := store.Create(context.Background(), "Plan", workflowTestSet())
	if err != nil {
		t.Fatal(err)
	}

	workflowTestGroupedTask(t, store, set.ID, "CHR-2", 2)

	first := workflowTestGroupedTask(t, store, set.ID, "CHR-1", 1)
	if err := index.RebuildWorkflowIndex(context.Background(), store); err != nil {
		t.Fatal(err)
	}

	q := WorkflowQuery{Kind: "ticket", OwnerID: "slice-1", Limit: 1}

	page := workflowTestQuery(t, index, store, q)
	if page.Records[0].Task.Key != "CHR-1" || page.NextOffset != 1 {
		t.Fatal("first task page lost durable order")
	}

	q.Offset, q.SnapshotHash = page.NextOffset, page.SnapshotHash

	page = workflowTestQuery(t, index, store, q)
	if page.Records[0].Task.Key != "CHR-2" || page.NextOffset != -1 {
		t.Fatal("second task page lost durable order")
	}

	slices := workflowTestQuery(t, index, store, WorkflowQuery{Kind: "slice", OwnerID: set.ID, Limit: 128})
	if len(slices.Slices) != 1 || slices.Slices[0].ID != "slice-1" {
		t.Fatal("slice index lost owning set")
	}

	first.Task.Status = ticketcore.StatusBacklog
	workflowTestWrite(t, store, first)

	if _, err := index.QueryWorkflow(context.Background(), store, q); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatalf("stale index returned work: %v", err)
	}
}

// TestWorkflowCacheDeletion proves files retain state after deleting the SQL cache.
func TestWorkflowCacheDeletion(t *testing.T) {
	store := workflowTestStore(t)
	r := workflowTestCreate(t, store, "CHR-1")
	path := filepath.Join(t.TempDir(), "cache.sqlite")

	index := workflowTestIndex(t, path)
	if err := index.RebuildWorkflowIndex(context.Background(), store); err != nil {
		t.Fatal(err)
	}

	closer, ok := index.(interface{ Close() error })
	if !ok {
		t.Fatal("index lacks close")
	}

	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	rebuilt := workflowTestIndex(t, path)
	if err := rebuilt.RebuildWorkflowIndex(context.Background(), store); err != nil {
		t.Fatal(err)
	}

	page := workflowTestQuery(t, rebuilt, store, WorkflowQuery{Kind: "ticket", LooseOnly: true, Limit: 128})
	if len(page.Records) != 1 || page.Records[0].Revision != r.Revision || page.Records[0].SourceHash != r.SourceHash {
		t.Fatal("cache deletion changed shared workflow state")
	}
}

// TestWorkflowBounds rejects oversized admission without allocating an envelope.
func TestWorkflowBounds(t *testing.T) {
	store := workflowTestStore(t)
	r := workflowTestTask("CHR-1")

	r.Body = make([]byte, workflowstate.FileLimit+1)
	if _, err := store.Create(context.Background(), "Oversized", r); err == nil {
		t.Fatal("oversized workflow admitted")
	}

	if _, err := os.Stat(filepath.Join(store.layout.Root(), "ticket")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("invalid admission allocated a partial envelope")
	}

	if err := (WorkflowSnapshot{Records: make([]WorkflowRecord, workflowstate.RecordLimit+1)}).Validate(); err == nil {
		t.Fatal("namespace record bound was ignored")
	}
}

// TestWorkflowNamespaces keeps identical native keys independent across namespaces.
func TestWorkflowNamespaces(t *testing.T) {
	store := workflowTestStore(t)
	index := workflowTestIndex(t, ":memory:")
	workflowTestCreate(t, store, "CHR-1")

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: store.layout.ProjectDir(), Namespace: "team"})
	if err != nil {
		t.Fatal(err)
	}

	other := NewWorkflowStore(layout, "project-1")
	workflowTestCreate(t, other, "CHR-1")

	for _, scoped := range []*WorkflowStore{store, other} {
		if err := index.RebuildWorkflowIndex(context.Background(), scoped); err != nil {
			t.Fatal(err)
		}
	}

	for _, scoped := range []*WorkflowStore{store, other} {
		page := workflowTestQuery(t, index, scoped, WorkflowQuery{Kind: "ticket", Limit: 128})
		if len(page.Records) != 1 || page.Records[0].Namespace != scoped.layout.Namespace() {
			t.Fatal("namespace query crossed ownership")
		}
	}
}
