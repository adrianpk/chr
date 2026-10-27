package execution_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func testMigrationFS() fs.FS { return os.DirFS(filepath.Join("..", "..", "..", "cmd", "chr")) }
func addTestProject(t *testing.T, registry projectcore.Registry, project projectcore.Project) {
	t.Helper()

	err := registry.AddProject(context.Background(), project)
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}
}
