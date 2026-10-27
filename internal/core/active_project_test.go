package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

func TestActiveProjectStoreLoadsEmptyState(t *testing.T) {
	store := projectcore.NewActiveStore(newTestUserEnvironment(t))

	_, err := store.LoadActiveProject()
	if !errors.Is(err, projectcore.ErrActiveNotSelected) {
		t.Fatalf("LoadActiveProject error = %v, want ErrActiveProjectNotSelected", err)
	}
}

func TestActiveProjectStoreSetsAndLoadsActiveProject(t *testing.T) {
	store := projectcore.NewActiveStore(newTestUserEnvironment(t))

	activeProject, err := projectcore.NewActive("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	err = store.SetActiveProject(activeProject)
	if err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	loaded, err := store.LoadActiveProject()
	if err != nil {
		t.Fatalf("LoadActiveProject returned error: %v", err)
	}

	assertActiveProject(t, loaded, activeProject)
}

func TestActiveProjectStoreReplacesActiveProject(t *testing.T) {
	store := projectcore.NewActiveStore(newTestUserEnvironment(t))

	first, err := projectcore.NewActive("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	second, err := projectcore.NewActive("project-2")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	err = store.SetActiveProject(first)
	if err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	err = store.SetActiveProject(second)
	if err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	loaded, err := store.LoadActiveProject()
	if err != nil {
		t.Fatalf("LoadActiveProject returned error: %v", err)
	}

	assertActiveProject(t, loaded, second)
}

func TestActiveProjectStoreRejectsEmptyProjectID(t *testing.T) {
	store := projectcore.NewActiveStore(newTestUserEnvironment(t))

	err := store.SetActiveProject(projectcore.Active{})
	if err == nil {
		t.Fatal("SetActiveProject returned nil error")
	}
}

func TestActiveProjectStoreClearsActiveProject(t *testing.T) {
	store := projectcore.NewActiveStore(newTestUserEnvironment(t))

	activeProject, err := projectcore.NewActive("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	err = store.SetActiveProject(activeProject)
	if err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	err = store.ClearActiveProject()
	if err != nil {
		t.Fatalf("ClearActiveProject returned error: %v", err)
	}

	_, err = store.LoadActiveProject()
	if !errors.Is(err, projectcore.ErrActiveNotSelected) {
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

	store := projectcore.NewActiveStore(env)

	activeProject, err := projectcore.NewActive("project-1")
	if err != nil {
		t.Fatalf("NewActiveProject returned error: %v", err)
	}

	err = store.SetActiveProject(activeProject)
	if err != nil {
		t.Fatalf("SetActiveProject returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, "config", "chr", "active-project"))
	assertPathNotExists(t, filepath.Join(root, "chr", "active-project"))
}

func newTestUserEnvironment(t *testing.T) UserEnvironment {
	t.Helper()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		ConfigDir: filepath.Join(t.TempDir(), "config", "chr"),
		DataDir:   filepath.Join(t.TempDir(), "data", "chr"),
		CacheDir:  filepath.Join(t.TempDir(), "cache", "chr"),
		StateDir:  filepath.Join(t.TempDir(), "state", "chr"),
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	return env
}

func assertActiveProject(t *testing.T, got, want projectcore.Active) {
	t.Helper()

	if got.ProjectID != want.ProjectID {
		t.Fatalf("ProjectID = %q, want %q", got.ProjectID, want.ProjectID)
	}
}

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat path %s error = %v, want not exist", path, err)
	}
}
