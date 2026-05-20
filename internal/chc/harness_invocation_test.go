package chc

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
		TicketPath:  filepath.Join(dir, ".chc", "tickets", "CHC-1.md"),
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("NewHarnessInvocation returned error: %v", err)
	}

	wantArgs := []string{
		"build",
		"--file",
		filepath.Join(dir, ".chc", "runtime", "agent", "Dockerfile"),
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
		TicketPath:  filepath.Join(dir, ".chc", "tickets", "CHC-1.md"),
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
		"CHC_TICKET_PATH=/workspace/.chc/tickets/CHC-1.md",
		"--env",
		"CHC_WORKER_CONTEXT_PATH=/workspace/.chc/runtime/runs/CHC-1/exec-123/context.md",
		"--env",
		"CHC_RESULT_PATH=/workspace/.chc/runtime/runs/CHC-1/exec-123/result.json",
		"--env",
		"CHC_LOG_PATH=/workspace/.chc/runtime/runs/CHC-1/exec-123/runtime.log",
		"--env",
		"CHC_VALIDATION_PATH=/workspace/.chc/runtime/runs/CHC-1/exec-123/validation.log",
		"chc-agent-runtime:local",
		"/bin/sh",
		"-lc",
		"CHC_EXECUTION_ID='exec-123' CHC_TICKET_KEY='CHC-1' CHC_TICKET_PATH='/workspace/.chc/tickets/CHC-1.md' CHC_WORKER_CONTEXT_PATH='/workspace/.chc/runtime/runs/CHC-1/exec-123/context.md' CHC_RESULT_PATH='/workspace/.chc/runtime/runs/CHC-1/exec-123/result.json' CHC_LOG_PATH='/workspace/.chc/runtime/runs/CHC-1/exec-123/runtime.log' CHC_VALIDATION_PATH='/workspace/.chc/runtime/runs/CHC-1/exec-123/validation.log' 'codex' 'exec' '--skip-git-repo-check' '--full-auto' < \"$CHC_WORKER_CONTEXT_PATH\"",
	}

	if invocation.Run.Program != "docker" {
		t.Fatalf("Run.Program = %q, want docker", invocation.Run.Program)
	}

	if !slices.Equal(invocation.Run.Args, wantArgs) {
		t.Fatalf("Run.Args = %#v, want %#v", invocation.Run.Args, wantArgs)
	}
}

func TestNewHarnessInvocationRejectsMissingIdentity(t *testing.T) {
	_, err := NewHarnessInvocation(HarnessInvocationOptions{
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(t.TempDir(), ".chc", "tickets", "CHC-1.md"),
		Plan:       testHarnessRuntimePlan(t.TempDir()),
	})
	if err == nil {
		t.Fatal("NewHarnessInvocation returned nil error")
	}
}

func testHarnessRuntimePlan(projectDir string) RuntimePlan {
	runDir := filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1", "exec-123")

	return RuntimePlan{
		ProjectDir:        projectDir,
		DockerfilePath:    filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"),
		Image:             "chc-agent-runtime:local",
		ContainerWorkdir:  "/workspace",
		Mounts:            runtimeMounts(projectDir, "/workspace"),
		RunArtifactDir:    runDir,
		WorkerContextPath: filepath.Join(runDir, "context.md"),
		ResultPath:        filepath.Join(runDir, "result.json"),
		LogPath:           filepath.Join(runDir, "runtime.log"),
		ValidationPath:    filepath.Join(runDir, "validation.log"),
		ValidationCommand: "make verify",
	}
}
