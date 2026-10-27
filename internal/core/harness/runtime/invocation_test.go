package runtime

import (
	"path/filepath"
	"slices"
	"testing"

	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

func TestNewHarnessInvocationRendersDockerBuildCommand(t *testing.T) {
	dir := t.TempDir()
	plan := testHarnessRuntimePlan(dir)

	invocation, err := NewInvocation(InvocationOptions{
		Worker:      codex.Runner{},
		ExecutionID: "exec-123",
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(dir, ".chr", "default", "tickets", "CHR-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantArgs := []string{
		"build",
		"--file",
		filepath.Join(dir, ".chr", "default", "runtime", "agent", "Dockerfile"),
		"--tag",
		"chr-agent-runtime:local",
		dir,
	}

	if invocation.Build.Program != "docker" {
		t.Fatalf("Build.Program = %q, want docker", invocation.Build.Program)
	}

	if !slices.Equal(invocation.Build.Args, wantArgs) {
		t.Fatalf("Build.Args = %#v, want %#v", invocation.Build.Args, wantArgs)
	}
}

func TestNewHarnessInvocationRendersDockerRunCommand(t *testing.T) {
	dir := t.TempDir()
	plan := testHarnessRuntimePlan(dir)

	invocation, err := NewInvocation(InvocationOptions{
		Worker:      codex.Runner{},
		ExecutionID: "exec-123",
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(dir, ".chr", "default", "tickets", "CHR-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantArgs := []string{
		"run",
		"--rm",
		"--name",
		"chr-exec-123",
		"--volume",
		dir + ":/workspace",
		"--workdir",
		"/workspace",
		"--env",
		"CHR_EXECUTION_ID=exec-123",
		"--env",
		"CHR_TICKET_KEY=CHR-1",
		"--env",
		"CHR_TICKET_PATH=/workspace/.chr/default/tickets/CHR-1.md",
		"--env",
		"CHR_WORKER_CONTEXT_PATH=/workspace/ops/default/run/CHR-1/exec-123/context.md",
		"--env",
		"CHR_RESULT_PATH=/workspace/ops/default/run/CHR-1/exec-123/result.json",
		"--env",
		"CHR_LOG_PATH=/workspace/ops/default/run/CHR-1/exec-123/runtime.log",
		"--env",
		"CHR_VALIDATION_PATH=/workspace/ops/default/run/CHR-1/exec-123/validation.log",
		"--env",
		"OPENAI_API_KEY",
		"chr-agent-runtime:local",
		"/bin/sh",
		"-lc",
		"CHR_EXECUTION_ID='exec-123' CHR_TICKET_KEY='CHR-1' CHR_TICKET_PATH='/workspace/.chr/default/tickets/CHR-1.md' CHR_WORKER_CONTEXT_PATH='/workspace/ops/default/run/CHR-1/exec-123/context.md' CHR_RESULT_PATH='/workspace/ops/default/run/CHR-1/exec-123/result.json' CHR_LOG_PATH='/workspace/ops/default/run/CHR-1/exec-123/runtime.log' CHR_VALIDATION_PATH='/workspace/ops/default/run/CHR-1/exec-123/validation.log' 'codex' 'exec' '--skip-git-repo-check' '--dangerously-bypass-approvals-and-sandbox' < '/workspace/ops/default/run/CHR-1/exec-123/context.md'",
	}

	if invocation.Run.Program != "docker" {
		t.Fatalf("Run.Program = %q, want docker", invocation.Run.Program)
	}

	if !slices.Equal(invocation.Run.Args, wantArgs) {
		t.Fatalf("Run.Args = %#v, want %#v", invocation.Run.Args, wantArgs)
	}
}

func TestNewHarnessInvocationRendersDockerValidationCommand(t *testing.T) {
	dir := t.TempDir()
	plan := testHarnessRuntimePlan(dir)

	invocation, err := NewInvocation(InvocationOptions{
		Worker:      codex.Runner{},
		ExecutionID: "exec-123",
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(dir, ".chr", "default", "tickets", "CHR-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantArgs := []string{
		"run",
		"--rm",
		"--name",
		"chr-exec-123-validation",
		"--volume",
		dir + ":/workspace",
		"--workdir",
		"/workspace",
		"--env",
		"CHR_EXECUTION_ID=exec-123",
		"--env",
		"CHR_VALIDATION_PATH=/workspace/ops/default/run/CHR-1/exec-123/validation.log",
		"chr-agent-runtime:local",
		"/bin/sh",
		"-lc",
		"make verify",
	}

	if invocation.Validation.Program != "docker" {
		t.Fatalf("Validation.Program = %q, want docker", invocation.Validation.Program)
	}

	if !slices.Equal(invocation.Validation.Args, wantArgs) {
		t.Fatalf("Validation.Args = %#v, want %#v", invocation.Validation.Args, wantArgs)
	}
}

func TestNewHarnessInvocationRendersWorktreeDockerRunCommand(t *testing.T) {
	projectDir := t.TempDir()
	executionWorkspaceDir := filepath.Join(projectDir, ".chr", "default", "worktrees", "CHR-1")

	plan, err := ResolvePlan(PlanOptions{
		ProjectDir: projectDir,
		Workspace: workspacecore.ExecutionConfig{
			Mode:            workspacecore.ExecutionWorktree,
			WorktreeRoot:    workspacecore.DefaultWorktreeRoot,
			WorktreeBase:    workspacecore.BaseRemoteIntegrationBranch,
			WorktreeCleanup: workspacecore.CleanupAfterMerge,
		},
		ExecutionWorkspaceDir: executionWorkspaceDir,
		TicketKey:             "CHR-1",
		RunID:                 "exec-123",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	invocation, err := NewInvocation(InvocationOptions{
		Worker:      codex.Runner{},
		ExecutionID: "exec-123",
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(executionWorkspaceDir, ".chr", "default", "tickets", "CHR-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantVolume := executionWorkspaceDir + ":/workspace"
	if !slices.Contains(invocation.Run.Args, wantVolume) {
		t.Fatalf("Run.Args missing worktree volume %q: %#v", wantVolume, invocation.Run.Args)
	}

	for _, want := range []string{
		"CHR_TICKET_PATH=/workspace/.chr/default/tickets/CHR-1.md",
		"CHR_WORKER_CONTEXT_PATH=/workspace/ops/default/run/CHR-1/exec-123/context.md",
		"CHR_RESULT_PATH=/workspace/ops/default/run/CHR-1/exec-123/result.json",
		"CHR_LOG_PATH=/workspace/ops/default/run/CHR-1/exec-123/runtime.log",
		"CHR_VALIDATION_PATH=/workspace/ops/default/run/CHR-1/exec-123/validation.log",
	} {
		if !slices.Contains(invocation.Run.Args, want) {
			t.Fatalf("Run.Args missing %q: %#v", want, invocation.Run.Args)
		}
	}
}

func TestNewHarnessInvocationRendersWorktreeDockerValidationCommand(t *testing.T) {
	projectDir := t.TempDir()
	executionWorkspaceDir := filepath.Join(projectDir, ".chr", "default", "worktrees", "CHR-1")

	plan, err := ResolvePlan(PlanOptions{
		ProjectDir: projectDir,
		Config: Config{
			ValidationCommand: "go test ./...",
		},
		Workspace: workspacecore.ExecutionConfig{
			Mode:            workspacecore.ExecutionWorktree,
			WorktreeRoot:    workspacecore.DefaultWorktreeRoot,
			WorktreeBase:    workspacecore.BaseRemoteIntegrationBranch,
			WorktreeCleanup: workspacecore.CleanupAfterMerge,
		},
		ExecutionWorkspaceDir: executionWorkspaceDir,
		TicketKey:             "CHR-1",
		RunID:                 "exec-123",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	invocation, err := NewInvocation(InvocationOptions{
		Worker:      codex.Runner{},
		ExecutionID: "exec-123",
		TicketKey:   "CHR-1",
		TicketPath:  filepath.Join(executionWorkspaceDir, ".chr", "default", "tickets", "CHR-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	for _, want := range []string{
		executionWorkspaceDir + ":/workspace",
		"/workspace",
		"CHR_EXECUTION_ID=exec-123",
		"CHR_VALIDATION_PATH=/workspace/ops/default/run/CHR-1/exec-123/validation.log",
		"chr-agent-runtime:local",
		"go test ./...",
	} {
		if !slices.Contains(invocation.Validation.Args, want) {
			t.Fatalf("Validation.Args missing %q: %#v", want, invocation.Validation.Args)
		}
	}
}

func TestNewHarnessInvocationRejectsMissingIdentity(t *testing.T) {
	_, err := NewInvocation(InvocationOptions{
		Worker:     codex.Runner{},
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(t.TempDir(), ".chr", "default", "tickets", "CHR-1.md"),
		Plan:       testHarnessRuntimePlan(t.TempDir()),
	})
	if err == nil {
		t.Fatal("NewHarnessInvocation returned nil error")
	}
}

func testHarnessRuntimePlan(projectDir string) Plan {
	runDir := filepath.Join(projectDir, "ops", "default", "run", "CHR-1", "exec-123")

	return Plan{
		ProjectDir:       projectDir,
		DockerfilePath:   filepath.Join(projectDir, ".chr", "default", "runtime", "agent", "Dockerfile"),
		Image:            "chr-agent-runtime:local",
		ContainerWorkdir: "/workspace",
		Mounts: []harness.RuntimeMount{
			{
				Source:      projectDir,
				Target:      "/workspace",
				ReadOnly:    false,
				Description: "project workspace",
			},
		},
		RunArtifactDir:    runDir,
		WorkerContextPath: filepath.Join(runDir, "context.md"),
		ResultPath:        filepath.Join(runDir, "result.json"),
		LogPath:           filepath.Join(runDir, "runtime.log"),
		ValidationPath:    filepath.Join(runDir, "validation.log"),
		ValidationCommand: "make verify",
	}
}
