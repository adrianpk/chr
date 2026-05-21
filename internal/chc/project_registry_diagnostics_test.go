package chc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectRegistryDiagnosticsReportsProjectState(t *testing.T) {
	env := newTestUserEnvironment(t)
	registry, err := NewProjectRegistry(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewProjectRegistry returned error: %v", err)
	}

	defer func() {
		err = registry.Close()
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	}()

	okDir := newDiagnosticProjectWorkspace(t)
	disabledDir := newDiagnosticProjectWorkspace(t)
	missingWorkspaceDir := t.TempDir()
	missingDir := filepath.Join(t.TempDir(), "missing")

	addDiagnosticProject(t, registry, Project{
		ID:           "project-ok",
		Name:         "ok",
		Path:         okDir,
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
	})
	addDiagnosticProject(t, registry, Project{
		ID:           "project-disabled",
		Name:         "disabled",
		Path:         disabledDir,
		Enabled:      false,
		RegisteredAt: time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
	})
	addDiagnosticProject(t, registry, Project{
		ID:           "project-no-workspace",
		Name:         "no-workspace",
		Path:         missingWorkspaceDir,
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
	})
	addDiagnosticProject(t, registry, Project{
		ID:           "project-missing",
		Name:         "missing",
		Path:         missingDir,
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
	})

	activeProject, err := NewActiveProject("project-ok")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	err = NewActiveProjectStore(env).SetActiveProject(activeProject)
	if err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	diagnostics, err := NewProjectRegistryDiagnosticsService(
		registry,
		NewActiveProjectStore(env),
	).ListDiagnostics(context.Background())
	if err != nil {
		t.Fatalf("ListDiagnostics returned error: %v", err)
	}

	assertDiagnosticStatus(t, diagnostics, "ok", ProjectRegistryDiagnosticOK, true, true, true)
	assertDiagnosticStatus(t, diagnostics, "disabled", ProjectRegistryDiagnosticDisabled, false, true, true)
	assertDiagnosticStatus(t, diagnostics, "no-workspace", ProjectRegistryDiagnosticNoWorkspace, false, true, false)
	assertDiagnosticStatus(t, diagnostics, "missing", ProjectRegistryDiagnosticMissingPath, false, false, false)
}

func newDiagnosticProjectWorkspace(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, DirectoryName), 0o755)
	if err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	return projectDir
}

func addDiagnosticProject(t *testing.T, registry ProjectRegistry, project Project) {
	t.Helper()

	err := registry.AddProject(context.Background(), project)
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}
}

func assertDiagnosticStatus(
	t *testing.T,
	diagnostics []ProjectRegistryDiagnostic,
	name string,
	status ProjectRegistryDiagnosticStatus,
	active bool,
	pathExists bool,
	workspaceExists bool,
) {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if diagnostic.Project.Name != name {
			continue
		}

		if diagnostic.Status != status {
			t.Fatalf("%s status = %q, want %q", name, diagnostic.Status, status)
		}

		if diagnostic.Active != active {
			t.Fatalf("%s active = %t, want %t", name, diagnostic.Active, active)
		}

		if diagnostic.PathExists != pathExists {
			t.Fatalf("%s path exists = %t, want %t", name, diagnostic.PathExists, pathExists)
		}

		if diagnostic.WorkspaceExists != workspaceExists {
			t.Fatalf("%s workspace exists = %t, want %t", name, diagnostic.WorkspaceExists, workspaceExists)
		}

		return
	}

	t.Fatalf("diagnostic for %s not found: %#v", name, diagnostics)
}
