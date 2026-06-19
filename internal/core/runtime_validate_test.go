package core

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

func TestValidateRuntimeChecksExecutionWorkspace(t *testing.T) {
	projectDir := t.TempDir()
	createRuntimeValidationWorkspace(t, projectDir)

	executionWorkspaceDir := filepath.Join(projectDir, ".chc", "user", "worktrees", "CHC-1")
	if err := os.MkdirAll(executionWorkspaceDir, 0o755); err != nil {
		t.Fatalf("create execution workspace dir: %v", err)
	}

	createRuntimeValidationWorkspace(t, executionWorkspaceDir)

	result, err := ValidateRuntime(RuntimeValidationOptions{
		ProjectDir: projectDir,
		Workspace: WorkWorkspaceConfig{
			Mode:            WorkWorkspaceModeWorktree,
			WorktreeRoot:    DefaultWorktreeRoot,
			WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
			WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
		},
		ExecutionWorkspaceDir: executionWorkspaceDir,
		TicketKey:             "CHC-1",
		RunID:                 "exec-1",
	})
	if err != nil {
		t.Fatalf("ValidateRuntime returned error: %v", err)
	}

	if !RuntimeValidationPassed(result.Checks) {
		t.Fatalf("RuntimeValidationPassed = false, checks = %#v", result.Checks)
	}

	if result.Plan.ExecutionWorkspaceDir != executionWorkspaceDir {
		t.Fatalf("ExecutionWorkspaceDir = %q, want %q", result.Plan.ExecutionWorkspaceDir, executionWorkspaceDir)
	}

	assertRuntimeValidationCheck(t, result.Checks, "execution_workspace", executionWorkspaceDir)
	assertRuntimeValidationCheck(t, result.Checks, "workspace", filepath.Join(executionWorkspaceDir, ".chc"))
	assertRuntimeValidationCheck(t, result.Checks, "artifact_dir", filepath.Join(executionWorkspaceDir, ".chc", "user", "runtime", "runs"))
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

func assertRuntimeValidationCheck(t *testing.T, checks []RuntimeValidationCheck, name, path string) {
	t.Helper()

	for _, check := range checks {
		if check.Name != name {
			continue
		}

		if check.Path != path {
			t.Fatalf("%s path = %q, want %q", name, check.Path, path)
		}

		if !check.OK {
			t.Fatalf("%s OK = false, message = %q", name, check.Message)
		}

		return
	}

	t.Fatalf("missing runtime validation check %s", name)
}
