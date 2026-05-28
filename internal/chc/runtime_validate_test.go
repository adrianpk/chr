package chc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRuntimePassesWhenRequiredPathsExist(t *testing.T) {
	dir := t.TempDir()
	createRuntimeValidationWorkspace(t, dir)

	result, err := ValidateRuntime(RuntimeValidationOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("ValidateRuntime returned error: %v", err)
	}

	if !RuntimeValidationPassed(result.Checks) {
		t.Fatalf("RuntimeValidationPassed = false, checks = %#v", result.Checks)
	}
}

func TestValidateRuntimeFailsWhenDockerfileIsMissing(t *testing.T) {
	dir := t.TempDir()
	createWorkspaceRoot(t, dir)

	err := os.MkdirAll(filepath.Join(dir, ".chc", "user", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}

	result, err := ValidateRuntime(RuntimeValidationOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("ValidateRuntime returned error: %v", err)
	}

	if RuntimeValidationPassed(result.Checks) {
		t.Fatalf("RuntimeValidationPassed = true, checks = %#v", result.Checks)
	}
}

func TestValidateRuntimeUsesCustomConfig(t *testing.T) {
	dir := t.TempDir()
	createWorkspaceRoot(t, dir)

	dockerfilePath := filepath.Join(dir, "ops", "runtime.Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create Dockerfile dir: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM ubuntu:24.04\n"), 0o644)
	if err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}

	artifactDir := filepath.Join(dir, "var", "chc", "runs")

	err = os.MkdirAll(artifactDir, 0o755)
	if err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}

	result, err := ValidateRuntime(RuntimeValidationOptions{
		ProjectDir: dir,
		Config: RuntimeConfig{
			DockerfilePath: "ops/runtime.Dockerfile",
			ArtifactDir:    "var/chc/runs",
		},
	})
	if err != nil {
		t.Fatalf("ValidateRuntime returned error: %v", err)
	}

	if !RuntimeValidationPassed(result.Checks) {
		t.Fatalf("RuntimeValidationPassed = false, checks = %#v", result.Checks)
	}

	if result.Plan.DockerfilePath != dockerfilePath {
		t.Fatalf("DockerfilePath = %q, want %q", result.Plan.DockerfilePath, dockerfilePath)
	}
}

func createRuntimeValidationWorkspace(t *testing.T, dir string) {
	t.Helper()

	createWorkspaceRoot(t, dir)

	dockerfilePath := filepath.Join(dir, ".chc", "user", "runtime", "agent", "Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create Dockerfile dir: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM ubuntu:24.04\n"), 0o644)
	if err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}

	err = os.MkdirAll(filepath.Join(dir, ".chc", "user", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}
}
