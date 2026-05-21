package chc

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

	if result.Root != filepath.Join(dir, DirectoryName, "runtime") {
		t.Fatalf("Root = %q, want runtime root", result.Root)
	}

	assertFile(t, dir, ".chc/runtime/agent/Dockerfile")

	if !contains(result.Created, ".chc/runtime/agent/Dockerfile") {
		t.Fatalf("Created = %#v, want runtime Dockerfile", result.Created)
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

	assertNotExists(t, dir, ".chc/runtime/agent/Dockerfile")

	if !contains(result.Created, ".chc/runtime/agent/Dockerfile") {
		t.Fatalf("Created = %#v, want runtime Dockerfile", result.Created)
	}
}

func TestRuntimeInitSkipsExistingRuntimeTemplate(t *testing.T) {
	dir := t.TempDir()
	createWorkspaceRoot(t, dir)

	dockerfilePath := filepath.Join(dir, ".chc", "runtime", "agent", "Dockerfile")

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

	if !contains(result.Skipped, ".chc/runtime/agent/Dockerfile") {
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
