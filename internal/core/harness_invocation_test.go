package core

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestNewHarnessInvocationRendersDockerBuildCommand(t *testing.T) {
	dir := t.TempDir()
	plan := testHarnessRuntimePlan(dir)

	invocation, err := NewHarnessInvocation(HarnessInvocationOptions{
		ExecutionID: "exec-123",
		TicketKey:   "CHC-1",
		TicketPath:  filepath.Join(dir, ".chc", "user", "tickets", "CHC-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantArgs := []string{
		"build",
		"--file",
		filepath.Join(dir, ".chc", "user", "runtime", "agent", "Dockerfile"),
		"--tag",
		"chc-agent-runtime:local",
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

	invocation, err := NewHarnessInvocation(HarnessInvocationOptions{
		ExecutionID: "exec-123",
		TicketKey:   "CHC-1",
		TicketPath:  filepath.Join(dir, ".chc", "user", "tickets", "CHC-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantArgs := []string{
		"run",
		"--rm",
		"--name",
		"chc-exec-123",
		"--volume",
		dir + ":/workspace",
		"--workdir",
		"/workspace",
		"--env",
		"CHC_EXECUTION_ID=exec-123",
		"--env",
		"CHC_TICKET_KEY=CHC-1",
		"--env",
		"CHC_TICKET_PATH=/workspace/.chc/user/tickets/CHC-1.md",
		"--env",
		"CHC_WORKER_CONTEXT_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/context.md",
		"--env",
		"CHC_RESULT_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/result.json",
		"--env",
		"CHC_LOG_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/runtime.log",
		"--env",
		"CHC_VALIDATION_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/validation.log",
		"--env",
		"OPENAI_API_KEY",
		"chc-agent-runtime:local",
		"/bin/sh",
		"-lc",
		"CHC_EXECUTION_ID='exec-123' CHC_TICKET_KEY='CHC-1' CHC_TICKET_PATH='/workspace/.chc/user/tickets/CHC-1.md' CHC_WORKER_CONTEXT_PATH='/workspace/.chc/user/runtime/runs/CHC-1/exec-123/context.md' CHC_RESULT_PATH='/workspace/.chc/user/runtime/runs/CHC-1/exec-123/result.json' CHC_LOG_PATH='/workspace/.chc/user/runtime/runs/CHC-1/exec-123/runtime.log' CHC_VALIDATION_PATH='/workspace/.chc/user/runtime/runs/CHC-1/exec-123/validation.log' 'codex' 'exec' '--skip-git-repo-check' '--dangerously-bypass-approvals-and-sandbox' < \"$CHC_WORKER_CONTEXT_PATH\"",
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

	invocation, err := NewHarnessInvocation(HarnessInvocationOptions{
		ExecutionID: "exec-123",
		TicketKey:   "CHC-1",
		TicketPath:  filepath.Join(dir, ".chc", "user", "tickets", "CHC-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantArgs := []string{
		"run",
		"--rm",
		"--name",
		"chc-exec-123-validation",
		"--volume",
		dir + ":/workspace",
		"--workdir",
		"/workspace",
		"--env",
		"CHC_EXECUTION_ID=exec-123",
		"--env",
		"CHC_VALIDATION_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/validation.log",
		"chc-agent-runtime:local",
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
	executionWorkspaceDir := filepath.Join(projectDir, ".chc", "user", "worktrees", "CHC-1")

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: projectDir,
		Workspace: WorkWorkspaceConfig{
			Mode:            WorkWorkspaceModeWorktree,
			WorktreeRoot:    DefaultWorktreeRoot,
			WorktreeBase:    WorktreeBasePolicyRemoteIntegrationBranch,
			WorktreeCleanup: WorktreeCleanupPolicyAfterMerge,
		},
		ExecutionWorkspaceDir: executionWorkspaceDir,
		TicketKey:             "CHC-1",
		RunID:                 "exec-123",
	})
	if err != nil {
		t.Fatalf("PlanRuntime returned error: %v", err)
	}

	invocation, err := NewHarnessInvocation(HarnessInvocationOptions{
		ExecutionID: "exec-123",
		TicketKey:   "CHC-1",
		TicketPath:  filepath.Join(executionWorkspaceDir, ".chc", "user", "tickets", "CHC-1.md"),
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
		"CHC_TICKET_PATH=/workspace/.chc/user/tickets/CHC-1.md",
		"CHC_WORKER_CONTEXT_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/context.md",
		"CHC_RESULT_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/result.json",
		"CHC_LOG_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/runtime.log",
		"CHC_VALIDATION_PATH=/workspace/.chc/user/runtime/runs/CHC-1/exec-123/validation.log",
	} {
		if !slices.Contains(invocation.Run.Args, want) {
			t.Fatalf("Run.Args missing %q: %#v", want, invocation.Run.Args)
		}
	}
}

func TestNewHarnessInvocationRejectsMissingIdentity(t *testing.T) {
	_, err := NewHarnessInvocation(HarnessInvocationOptions{
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(t.TempDir(), ".chc", "user", "tickets", "CHC-1.md"),
		Plan:       testHarnessRuntimePlan(t.TempDir()),
	})
	if err == nil {
		t.Fatal("NewHarnessInvocation returned nil error")
	}
}

func testHarnessRuntimePlan(projectDir string) RuntimePlan {
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-123")

	return RuntimePlan{
		ProjectDir:       projectDir,
		DockerfilePath:   filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"),
		Image:            "chc-agent-runtime:local",
		ContainerWorkdir: "/workspace",
		Mounts: []RuntimeMount{
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
