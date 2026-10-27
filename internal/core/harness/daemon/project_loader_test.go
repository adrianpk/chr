package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

func TestLoadDaemonProjectsReturnsEnabledProjectsWithReadyStatus(t *testing.T) {
	registry := newTestProjectRegistry(t)
	projectDir := newDaemonLoaderWorkspace(t)
	registeredAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	addTestProject(t, registry, projectcore.Project{
		ID:           "project-1",
		Name:         "ready",
		Path:         projectDir,
		Enabled:      true,
		RegisteredAt: registeredAt,
	})
	addTestProject(t, registry, projectcore.Project{
		ID:           "project-2",
		Name:         "disabled",
		Path:         newDaemonLoaderWorkspace(t),
		Enabled:      false,
		RegisteredAt: registeredAt,
	})

	projects, err := LoadDaemonProjects(context.Background(), registry)
	if err != nil {
		t.Fatalf("LoadDaemonProjects returned error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}

	if projects[0].Project.Name != "ready" || projects[0].Status != DaemonProjectStatusReady {
		t.Fatalf("project = %#v, want ready enabled project", projects[0])
	}
}

func TestLoadDaemonProjectsClassifiesMissingPath(t *testing.T) {
	registry := newTestProjectRegistry(t)

	addTestProject(t, registry, projectcore.Project{
		ID:           "project-1",
		Name:         "missing",
		Path:         filepath.Join(t.TempDir(), "missing"),
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	})

	projects, err := LoadDaemonProjects(context.Background(), registry)
	if err != nil {
		t.Fatalf("LoadDaemonProjects returned error: %v", err)
	}

	assertDaemonProjectStatus(t, projects, DaemonProjectStatusMissing, "project path does not exist")
}

func TestLoadDaemonProjectsClassifiesNonDirectoryPath(t *testing.T) {
	registry := newTestProjectRegistry(t)
	path := filepath.Join(t.TempDir(), "project-file")

	err := os.WriteFile(path, []byte("not a directory\n"), 0o644)
	if err != nil {
		t.Fatalf("write project file: %v", err)
	}

	addTestProject(t, registry, projectcore.Project{
		ID:           "project-1",
		Name:         "invalid",
		Path:         path,
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	})

	projects, err := LoadDaemonProjects(context.Background(), registry)
	if err != nil {
		t.Fatalf("LoadDaemonProjects returned error: %v", err)
	}

	assertDaemonProjectStatus(t, projects, DaemonProjectStatusInvalid, "project path is not a directory")
}

func TestLoadDaemonProjectsClassifiesMissingWorkspace(t *testing.T) {
	registry := newTestProjectRegistry(t)

	addTestProject(t, registry, projectcore.Project{
		ID:           "project-1",
		Name:         "invalid",
		Path:         t.TempDir(),
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	})

	projects, err := LoadDaemonProjects(context.Background(), registry)
	if err != nil {
		t.Fatalf("LoadDaemonProjects returned error: %v", err)
	}

	assertDaemonProjectStatus(t, projects, DaemonProjectStatusInvalid, "project workspace is missing")
}

func newDaemonLoaderWorkspace(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, workspacecore.DirectoryName), 0o755)
	if err != nil {
		t.Fatalf("create workspace directory: %v", err)
	}

	return projectDir
}

func assertDaemonProjectStatus(t *testing.T, projects []DaemonProject, status DaemonProjectStatus, reason string) {
	t.Helper()

	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}

	if projects[0].Status != status || projects[0].Reason != reason {
		t.Fatalf("project = %#v, want status %q reason %q", projects[0], status, reason)
	}
}
