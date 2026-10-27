package core

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func newTestTicketIndex(t *testing.T) ticketcore.Index {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	registry, err := projectcore.NewRegistryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewProjectRegistryFromDB returned error: %v", err)
	}

	addTestProject(t, registry, projectcore.Project{
		ID:           "project-1",
		Name:         "chr",
		Path:         "/tmp/chr",
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	index, err := ticketcore.NewIndexFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewTicketIndexFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := index.Close()
		if err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	})

	return index
}

func writeRebuildTestTicket(t *testing.T, projectDir, relativePath string, ticket ticketcore.Ticket) {
	t.Helper()

	content, err := ticketcore.NewRenderer().Render(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	writeDiscoveryTestFile(t, projectDir, filepath.Join(".chr", "default", "tickets", relativePath))

	path := filepath.Join(projectDir, ".chr", "default", "tickets", filepath.FromSlash(relativePath))
	writeFile(t, path, content)
}

func rebuildTestCreatedAt() time.Time {
	return time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
}

func rebuildTestUpdatedAt() time.Time {
	return time.Date(2026, 5, 19, 10, 30, 0, 0, time.UTC)
}

func rebuildTestIndexedAt() time.Time {
	return time.Date(2026, 5, 19, 11, 0, 0, 0, time.UTC)
}

func writeDiscoveryTestFile(t *testing.T, projectDir, relativePath string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create test file dir: %v", err)
	}

	err = os.WriteFile(path, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
