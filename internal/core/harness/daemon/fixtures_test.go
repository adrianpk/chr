package daemon

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func testMigrationFS() fs.FS { return os.DirFS(filepath.Join("..", "..", "..", "..", "cmd", "chr")) }

func newTestProjectRegistry(t *testing.T) projectcore.Registry {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	registry, err := projectcore.NewRegistryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("close sqlite after registry error: %v", closeErr)
		}

		t.Fatalf("NewProjectRegistryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := registry.Close()
		if err != nil {
			t.Fatalf("close registry: %v", err)
		}
	})

	return registry
}

func addTestProject(t *testing.T, registry projectcore.Registry, project projectcore.Project) {
	t.Helper()

	err := registry.AddProject(context.Background(), project)
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}
}

func optionalTimeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}

	return *value
}
