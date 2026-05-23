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

func TestServiceSaveProjectMetadata(t *testing.T) {
	projectDir := t.TempDir()
	operations := newTestService(t, projectDir)

	result, err := operations.SaveProjectMetadata(chc.ProjectMetadata{
		SchemaVersion: chc.ProjectMetadataSchemaVersion,
		Name:          "Saved Project",
		Description:   "Saved through service.",
	})
	if err != nil {
		t.Fatalf("SaveProjectMetadata returned error: %v", err)
	}

	if result.Message != "Project metadata saved." {
		t.Fatalf("Message = %q, want Project metadata saved.", result.Message)
	}

	if result.Metadata.Name != "Saved Project" {
		t.Fatalf("Metadata.Name = %q, want Saved Project", result.Metadata.Name)
	}

	loaded, err := chc.NewProjectMetadataService(projectDir).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Name != "Saved Project" {
		t.Fatalf("loaded.Name = %q, want Saved Project", loaded.Name)
	}
}

func TestServiceSaveCaseMetadataReloadsCases(t *testing.T) {
	projectDir := t.TempDir()
	operations := newTestService(t, projectDir)

	result, err := operations.SaveCaseMetadata(chc.CaseMetadata{
		SchemaVersion: chc.CaseMetadataSchemaVersion,
		Name:          "Saved Case",
		Description:   "Saved through service.",
	})
	if err != nil {
		t.Fatalf("SaveCaseMetadata returned error: %v", err)
	}

	if result.Message != "Case metadata saved." {
		t.Fatalf("Message = %q, want Case metadata saved.", result.Message)
	}

	if result.Case.Slug != "saved-case" {
		t.Fatalf("Case.Slug = %q, want saved-case", result.Case.Slug)
	}

	if len(result.Cases) != 1 || result.Cases[0].Slug != "saved-case" {
		t.Fatalf("Cases = %#v, want saved-case", result.Cases)
	}
}

func newTestService(t *testing.T, projectDir string) service {
	t.Helper()

	env, err := chcTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chcTestUserEnvironment returned error: %v", err)
	}

	return newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  "service-project",
	})
}
