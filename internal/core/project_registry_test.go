package core

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestProjectRegistryAddsListsAndGetsProjects(t *testing.T) {
	registry := newTestProjectRegistry(t)
	registeredAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	lastIndexedAt := time.Date(2026, 5, 18, 13, 30, 0, 0, time.UTC)

	err := registry.AddProject(context.Background(), Project{
		ID:            "project-1",
		Name:          "chr",
		Path:          "/workspaces/chr",
		Enabled:       true,
		RegisteredAt:  registeredAt,
		LastIndexedAt: &lastIndexedAt,
	})
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}

	projects, err := registry.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}

	assertProject(t, projects[0], Project{
		ID:            "project-1",
		Name:          "chr",
		Path:          "/workspaces/chr",
		Enabled:       true,
		RegisteredAt:  registeredAt,
		LastIndexedAt: &lastIndexedAt,
	})

	project, err := registry.GetProjectByName(context.Background(), "chr")
	if err != nil {
		t.Fatalf("GetProjectByName returned error: %v", err)
	}

	assertProject(t, project, projects[0])

	project, err = registry.GetProjectByID(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("GetProjectByID returned error: %v", err)
	}

	assertProject(t, project, projects[0])

	project, err = registry.GetProjectByPath(context.Background(), "/workspaces/chr")
	if err != nil {
		t.Fatalf("GetProjectByPath returned error: %v", err)
	}

	assertProject(t, project, projects[0])
}

func TestProjectRegistryListsProjectsByName(t *testing.T) {
	registry := newTestProjectRegistry(t)
	registeredAt := time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)

	addTestProject(t, registry, Project{
		ID:           "project-2",
		Name:         "zeta",
		Path:         "/tmp/zeta",
		Enabled:      true,
		RegisteredAt: registeredAt,
	})
	addTestProject(t, registry, Project{
		ID:           "project-1",
		Name:         "alpha",
		Path:         "/tmp/alpha",
		Enabled:      true,
		RegisteredAt: registeredAt,
	})

	projects, err := registry.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}

	if projects[0].Name != "alpha" || projects[1].Name != "zeta" {
		t.Fatalf("project order = %q, %q; want alpha, zeta", projects[0].Name, projects[1].Name)
	}
}

func TestProjectRegistryRejectsDuplicateName(t *testing.T) {
	registry := newTestProjectRegistry(t)
	project := Project{
		ID:           "project-1",
		Name:         "chr",
		Path:         "/tmp/chr",
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC),
	}

	addTestProject(t, registry, project)

	project.ID = "project-2"
	project.Path = "/tmp/other"

	err := registry.AddProject(context.Background(), project)
	if err == nil {
		t.Fatal("AddProject returned nil error for duplicate project name")
	}
}

func TestProjectRegistryReturnsNotFoundError(t *testing.T) {
	registry := newTestProjectRegistry(t)

	_, err := registry.GetProjectByName(context.Background(), "missing")
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("GetProjectByName error = %v, want ErrProjectNotFound", err)
	}
}

func TestProjectRegistryDisablesProject(t *testing.T) {
	registry := newTestProjectRegistry(t)
	project := testRegistryProject("project-1", "example", "/tmp/example")

	addTestProject(t, registry, project)

	updated, err := registry.DisableProject(context.Background(), "example")
	if err != nil {
		t.Fatalf("DisableProject returned error: %v", err)
	}

	if updated.Enabled {
		t.Fatal("Enabled = true, want false")
	}

	stored, err := registry.GetProjectByName(context.Background(), "example")
	if err != nil {
		t.Fatalf("GetProjectByName returned error: %v", err)
	}

	if stored.Enabled {
		t.Fatal("stored Enabled = true, want false")
	}
}

func TestProjectRegistryUpdatesProjectPath(t *testing.T) {
	registry := newTestProjectRegistry(t)
	project := testRegistryProject("project-1", "example", "/tmp/example")

	addTestProject(t, registry, project)

	updated, err := registry.UpdateProjectPath(context.Background(), "example", "/tmp/moved")
	if err != nil {
		t.Fatalf("UpdateProjectPath returned error: %v", err)
	}

	if updated.Path != "/tmp/moved" {
		t.Fatalf("Path = %q, want /tmp/moved", updated.Path)
	}

	stored, err := registry.GetProjectByName(context.Background(), "example")
	if err != nil {
		t.Fatalf("GetProjectByName returned error: %v", err)
	}

	if stored.Path != "/tmp/moved" {
		t.Fatalf("stored Path = %q, want /tmp/moved", stored.Path)
	}
}

func TestNewProjectRegistryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	registry, err := NewProjectRegistry(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewProjectRegistry returned error: %v", err)
	}

	t.Cleanup(func() {
		err := registry.Close()
		if err != nil {
			t.Fatalf("close registry: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chr", "core.db"))
}

func newTestProjectRegistry(t *testing.T) ProjectRegistry {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	registry, err := NewProjectRegistryFromDB(context.Background(), db, testMigrationFS())
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

func addTestProject(t *testing.T, registry ProjectRegistry, project Project) {
	t.Helper()

	err := registry.AddProject(context.Background(), project)
	if err != nil {
		t.Fatalf("AddProject returned error: %v", err)
	}
}

func testRegistryProject(id, name, path string) Project {
	return Project{
		ID:           id,
		Name:         name,
		Path:         path,
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC),
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat path %s: %v", path, err)
	}
}

func assertProject(t *testing.T, got, want Project) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf("ID = %q, want %q", got.ID, want.ID)
	}

	if got.Name != want.Name {
		t.Fatalf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.Path != want.Path {
		t.Fatalf("Path = %q, want %q", got.Path, want.Path)
	}

	if got.Enabled != want.Enabled {
		t.Fatalf("Enabled = %t, want %t", got.Enabled, want.Enabled)
	}

	if !got.RegisteredAt.Equal(want.RegisteredAt) {
		t.Fatalf("RegisteredAt = %s, want %s", got.RegisteredAt, want.RegisteredAt)
	}

	if !sameOptionalTime(got.LastIndexedAt, want.LastIndexedAt) {
		t.Fatalf("LastIndexedAt = %v, want %v", got.LastIndexedAt, want.LastIndexedAt)
	}
}

func sameOptionalTime(got, want *time.Time) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}

	return got.Equal(*want)
}
