package runtime

import (
	"os"
	"path/filepath"
	"testing"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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

	err := os.MkdirAll(filepath.Join(dir, "ops", "default", "run"), 0o755)
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

	artifactDir := filepath.Join(dir, "var", "chr", "runs")

	err = os.MkdirAll(artifactDir, 0o755)
	if err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}

	result, err := ValidateRuntime(RuntimeValidationOptions{
		ProjectDir: dir,
		Config: RuntimeConfig{
			DockerfilePath: "ops/runtime.Dockerfile",
			ArtifactDir:    "var/chr/runs",
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

	executionWorkspaceDir := filepath.Join(projectDir, ".chr", "default", "worktrees", "CHR-1")
	if err := os.MkdirAll(executionWorkspaceDir, 0o755); err != nil {
		t.Fatalf("create execution workspace dir: %v", err)
	}

	createRuntimeValidationWorkspace(t, executionWorkspaceDir)

	result, err := ValidateRuntime(RuntimeValidationOptions{
		ProjectDir: projectDir,
		Workspace: workspacecore.WorkWorkspaceConfig{
			Mode:            workspacecore.WorkWorkspaceModeWorktree,
			WorktreeRoot:    workspacecore.DefaultWorktreeRoot,
			WorktreeBase:    workspacecore.WorktreeBasePolicyRemoteIntegrationBranch,
			WorktreeCleanup: workspacecore.WorktreeCleanupPolicyAfterMerge,
		},
		ExecutionWorkspaceDir: executionWorkspaceDir,
		TicketKey:             "CHR-1",
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
	assertRuntimeValidationCheck(t, result.Checks, "workspace", filepath.Join(executionWorkspaceDir, ".chr"))
	assertRuntimeValidationCheck(t, result.Checks, "artifact_dir", filepath.Join(executionWorkspaceDir, "ops", "default", "run"))
}

func createRuntimeValidationWorkspace(t *testing.T, dir string) {
	t.Helper()

	createWorkspaceRoot(t, dir)

	dockerfilePath := filepath.Join(dir, ".chr", "default", "runtime", "agent", "Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create Dockerfile dir: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM ubuntu:24.04\n"), 0o644)
	if err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}

	err = os.MkdirAll(filepath.Join(dir, "ops", "default", "run"), 0o755)
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

func createWorkspaceRoot(t *testing.T, dir string) {
	t.Helper()
	if err := os.Mkdir(filepath.Join(dir, workspacecore.DirectoryName), 0o755); err != nil {
		t.Fatalf("create workspace root: %v", err)
	}
}
