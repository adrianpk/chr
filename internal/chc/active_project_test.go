package chc

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestActiveProjectStoreLoadsEmptyState(t *testing.T) {
	store := NewActiveProjectStore(newTestUserEnvironment(t))

	_, err := store.LoadActiveProject()
	if !errors.Is(err, ErrActiveProjectNotSelected) {
		t.Fatalf("LoadActiveProject error = %v, want ErrActiveProjectNotSelected", err)
	}
}

func TestActiveProjectStoreSetsAndLoadsActiveProject(t *testing.T) {
	store := NewActiveProjectStore(newTestUserEnvironment(t))

	activeProject, err := NewActiveProject("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	if err := store.SetActiveProject(activeProject); err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	loaded, err := store.LoadActiveProject()
	if err != nil {
		t.Fatalf("LoadActiveProject returned error: %v", err)
	}

	assertActiveProject(t, loaded, activeProject)
}

func TestActiveProjectStoreReplacesActiveProject(t *testing.T) {
	store := NewActiveProjectStore(newTestUserEnvironment(t))

	first, err := NewActiveProject("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	second, err := NewActiveProject("project-2")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	if err := store.SetActiveProject(first); err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	if err := store.SetActiveProject(second); err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	loaded, err := store.LoadActiveProject()
	if err != nil {
		t.Fatalf("LoadActiveProject returned error: %v", err)
	}

	assertActiveProject(t, loaded, second)
}

func TestActiveProjectStoreRejectsEmptyProjectID(t *testing.T) {
	store := NewActiveProjectStore(newTestUserEnvironment(t))

	if err := store.SetActiveProject(ActiveProject{}); err == nil {
		t.Fatal("SetActiveProject returned nil error")
	}
}

func TestActiveProjectStoreClearsActiveProject(t *testing.T) {
	store := NewActiveProjectStore(newTestUserEnvironment(t))

	activeProject, err := NewActiveProject("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	if err := store.SetActiveProject(activeProject); err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	if err := store.ClearActiveProject(); err != nil {
		t.Fatalf("ClearActiveProject returned error: %v", err)
	}

	_, err = store.LoadActiveProject()
	if !errors.Is(err, ErrActiveProjectNotSelected) {
		t.Fatalf("LoadActiveProject error = %v, want ErrActiveProjectNotSelected", err)
	}
}

func TestActiveProjectStoreUsesDevelopmentConfigPath(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	store := NewActiveProjectStore(env)

	activeProject, err := NewActiveProject("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	if err := store.SetActiveProject(activeProject); err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, "config", "chc", "active-project"))
	assertPathNotExists(t, filepath.Join(root, "chc", "active-project"))
}

func newTestUserEnvironment(t *testing.T) UserEnvironment {
	t.Helper()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		ConfigDir: filepath.Join(t.TempDir(), "config", "chc"),
		DataDir:   filepath.Join(t.TempDir(), "data", "chc"),
		CacheDir:  filepath.Join(t.TempDir(), "cache", "chc"),
		StateDir:  filepath.Join(t.TempDir(), "state", "chc"),
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	return env
}

func assertActiveProject(t *testing.T, got, want ActiveProject) {
	t.Helper()

	if got.ProjectID != want.ProjectID {
		t.Fatalf("ProjectID = %q, want %q", got.ProjectID, want.ProjectID)
	}
}

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat path %s error = %v, want not exist", path, err)
	}
}
