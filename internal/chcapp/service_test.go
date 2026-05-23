package chcapp

import (
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestServiceRefreshWorkspaceBuildsGraph(t *testing.T) {
	projectDir := t.TempDir()

	err := os.MkdirAll(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create .chc: %v", err)
	}

	err = os.WriteFile(
		chc.ProjectMetadataPath(projectDir),
		[]byte(chc.RenderProjectMetadata(chc.ProjectMetadata{
			Name:        "Service Project",
			Description: "Loaded by service.",
		})),
		0o644,
	)
	if err != nil {
		t.Fatalf("write project metadata: %v", err)
	}

	env, err := chcTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chcTestUserEnvironment returned error: %v", err)
	}

	operations := newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
	})

	result, err := operations.RefreshWorkspace()
	if err != nil {
		t.Fatalf("RefreshWorkspace returned error: %v", err)
	}

	if result.Message != "Workspace refreshed." {
		t.Fatalf("Message = %q, want Workspace refreshed.", result.Message)
	}

	if result.Graph.Project == nil || result.Graph.Project.Name != "Service Project" {
		t.Fatalf("Project = %#v, want Service Project", result.Graph.Project)
	}
}
