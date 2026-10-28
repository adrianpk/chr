package workflow

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func addTestProject(t *testing.T, registry projectcore.Registry, project projectcore.Project) {
	t.Helper()

	err := registry.AddProject(context.Background(), project)
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}
}
