package ticket_test

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func testMigrationFS() fs.FS { return os.DirFS(filepath.Join("..", "..", "..", "cmd", "chr")) }

func addTestProject(t *testing.T, registry projectcore.Registry, project projectcore.Project) {
	t.Helper()

	err := registry.AddProject(context.Background(), project)
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}
}

func openTestSQLite(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	})

	return db
}

func fixedTicketTime() time.Time {
	return time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}

	return string(content)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create parent dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func newTestTicketRepository(projectDir string, opts ...ticketcore.RepositoryOption) ticketcore.Repository {
	return ticketcore.NewRepository(projectDir, func(context.Context, string) (int, error) { return 1, nil }, opts...)
}
