package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestRuntimeInitCreatesRuntimeTemplate(t *testing.T) {
	dir := t.TempDir()
	createWorkspaceRoot(t, dir)

	runtimeAssets := newTestRuntimeAssets()

	result, err := runtimeAssets.Init(context.Background(), RuntimeInitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if result.Root != filepath.Join(dir, DirectoryName, UserDirectoryName, "runtime") {
		t.Fatalf("Root = %q, want runtime root", result.Root)
	}

	assertFile(t, dir, ".chr/user/runtime/agent/Dockerfile")
	assertDir(t, dir, ".chr/user/runtime/runs")

	if !contains(result.Created, ".chr/user/runtime/agent/Dockerfile") {
		t.Fatalf("Created = %#v, want runtime Dockerfile", result.Created)
	}

	if !contains(result.Created, ".chr/user/runtime/runs") {
		t.Fatalf("Created = %#v, want runtime runs directory", result.Created)
	}
}

func TestRuntimeInitDryRunDoesNotCreateRuntimeTemplate(t *testing.T) {
	dir := t.TempDir()
	createWorkspaceRoot(t, dir)

	runtimeAssets := newTestRuntimeAssets()

	result, err := runtimeAssets.Init(context.Background(), RuntimeInitOptions{
		ProjectDir: dir,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	assertNotExists(t, dir, ".chr/user/runtime/agent/Dockerfile")
	assertNotExists(t, dir, ".chr/user/runtime/runs")

	if !contains(result.Created, ".chr/user/runtime/agent/Dockerfile") {
		t.Fatalf("Created = %#v, want runtime Dockerfile", result.Created)
	}

	if !contains(result.Created, ".chr/user/runtime/runs") {
		t.Fatalf("Created = %#v, want runtime runs directory", result.Created)
	}
}

func TestRuntimeInitSkipsExistingRuntimeTemplate(t *testing.T) {
	dir := t.TempDir()
	createWorkspaceRoot(t, dir)

	dockerfilePath := filepath.Join(dir, ".chr", "user", "runtime", "agent", "Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM custom\n"), 0o644)
	if err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	result, err := newTestRuntimeAssets().Init(context.Background(), RuntimeInitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("read runtime template: %v", err)
	}

	if string(content) != "FROM custom\n" {
		t.Fatalf("runtime template content = %q, want custom content unchanged", string(content))
	}

	if !contains(result.Skipped, ".chr/user/runtime/agent/Dockerfile") {
		t.Fatalf("Skipped = %#v, want runtime Dockerfile", result.Skipped)
	}
}

func newTestRuntimeAssets() RuntimeAssets {
	return NewRuntimeAssets(fstest.MapFS{
		"assets/workspace/profiles/standard/config.yml": {
			Data: []byte("version: 1\n"),
		},
		"assets/workspace/profiles/standard/runtime/agent/Dockerfile": {
			Data: []byte("FROM ubuntu:24.04\n"),
		},
	})
}

func createWorkspaceRoot(t *testing.T, dir string) {
	t.Helper()

	err := os.Mkdir(filepath.Join(dir, DirectoryName), 0o755)
	if err != nil {
		t.Fatalf("create workspace root: %v", err)
	}
}
