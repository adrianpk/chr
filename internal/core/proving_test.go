package core

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

func TestParseProvingObjective(t *testing.T) {
	objective, err := ParseProvingObjective(`---
objective_id: proving-TODO-1
ticket_key: TODO-1
kind: bugfix
artifact_root: ops/default/run/00001-proving-todo-1
context_path: ops/default/run/00001-proving-todo-1/files/context.md
result_path: ops/default/run/00001-proving-todo-1/evidence/result.json
log_path: ops/default/run/00001-proving-todo-1/evidence/agent.log
validation_path: ops/default/run/00001-proving-todo-1/evidence/validation.log
max_attempts: 3
---

# Fix Task ID Reuse
`)
	if err != nil {
		t.Fatalf("ParseProvingObjective returned error: %v", err)
	}

	if objective.ID != "proving-TODO-1" {
		t.Fatalf("ID = %q, want proving-TODO-1", objective.ID)
	}

	if objective.MaxAttempts != 3 {
		t.Fatalf("MaxAttempts = %d, want 3", objective.MaxAttempts)
	}
}

func TestPlanProvingSandboxRunRendersDockerCommands(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")

	result, err := PlanProvingSandboxRun(projectDir, objectivePath, codex.Runner{})
	if err != nil {
		t.Fatalf("PlanProvingSandboxRun returned error: %v", err)
	}

	if result.Build.Program != "docker" {
		t.Fatalf("Build.Program = %q, want docker", result.Build.Program)
	}

	if result.Run.Program != "docker" {
		t.Fatalf("Run.Program = %q, want docker", result.Run.Program)
	}

	for _, want := range []string{
		"CHR_EXECUTION_ID=proving-TODO-1",
		"CHR_TICKET_KEY=TODO-1",
		"CHR_WORKER_CONTEXT_PATH=/workspace/ops/default/run/00001-proving-todo-1/files/context.md",
	} {
		if !harnessArgsContain(result.Run.Args, want) {
			t.Fatalf("run args do not contain %q:\n%#v", want, result.Run.Args)
		}
	}
}

func TestRunProvingSandboxDryRunDoesNotExecuteCommands(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")
	executor := &recordingHarnessExecutor{}

	result, err := RunProvingSandbox(context.Background(), ProvingSandboxRunOptions{
		Worker:        codex.Runner{},
		ProjectDir:    projectDir,
		ObjectivePath: objectivePath,
		DryRun:        true,
		Executor:      executor,
	})
	if err != nil {
		t.Fatalf("RunProvingSandbox returned error: %v", err)
	}

	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}

	if len(executor.commands) != 0 {
		t.Fatalf("commands = %#v, want none", executor.commands)
	}
}

func TestRenderProvingWorkerContextIncludesObjective(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")

	result, err := PlanProvingSandboxRun(projectDir, objectivePath, codex.Runner{})
	if err != nil {
		t.Fatalf("PlanProvingSandboxRun returned error: %v", err)
	}

	context := RenderProvingWorkerContext(result)
	for _, want := range []string{
		"objective_id: proving-TODO-1",
		"ticket_key: TODO-1",
		"Fix Task ID Reuse",
		"Write the result JSON",
	} {
		if !strings.Contains(context, want) {
			t.Fatalf("context does not contain %q:\n%s", want, context)
		}
	}
}

func TestRunProvingValidationCapturesFailure(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")
	writeTestFile(t, filepath.Join(projectDir, "Makefile"), `.PHONY: compile lint test verify
compile:
	@true
lint:
	@true
test:
	@echo failing test
	@exit 2
verify:
	@$(MAKE) compile lint test
`)

	result, err := RunProvingValidation(context.Background(), projectDir, objectivePath)
	if err != nil {
		t.Fatalf("RunProvingValidation returned error: %v", err)
	}

	if result.Passed {
		t.Fatal("Passed = true, want false")
	}

	if len(result.Commands) != 4 {
		t.Fatalf("len(Commands) = %d, want 4", len(result.Commands))
	}

	if result.Commands[2].ExitCode == 0 {
		t.Fatalf("test command exit code = %d, want non-zero", result.Commands[2].ExitCode)
	}

	content, err := os.ReadFile(result.LogPath)
	if err != nil {
		t.Fatalf("read validation log: %v", err)
	}

	if !strings.Contains(string(content), "failing test") {
		t.Fatalf("validation log does not contain failure output:\n%s", string(content))
	}
}

func TestRunProvingRalphLoopStopsOnValidationPass(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")
	executor := &recordingHarnessExecutor{}
	validations := 0

	result, err := RunProvingRalphLoop(context.Background(), ProvingRalphLoopOptions{
		Worker:        codex.Runner{},
		ProjectDir:    projectDir,
		ObjectivePath: objectivePath,
		Executor:      executor,
		ValidationRunner: func(_ context.Context, _ string, objective ProvingObjective) (ProvingValidationResult, error) {
			validations++

			return ProvingValidationResult{
				Objective: objective,
				Passed:    true,
				LogPath:   projectcore.ResolvePath(projectDir, objective.ValidationPath),
				Commands: []ProvingValidationCommandResult{
					{Command: "make verify", ExitCode: 0},
				},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("RunProvingRalphLoop returned error: %v", err)
	}

	if result.Status != ProvingRalphLoopStatusPassed {
		t.Fatalf("Status = %q, want %q", result.Status, ProvingRalphLoopStatusPassed)
	}

	if result.StopReason != ProvingRalphLoopStopValidationPassed {
		t.Fatalf("StopReason = %q, want %q", result.StopReason, ProvingRalphLoopStopValidationPassed)
	}

	if len(result.Attempts) != 1 {
		t.Fatalf("len(Attempts) = %d, want 1", len(result.Attempts))
	}

	if validations != 1 {
		t.Fatalf("validations = %d, want 1", validations)
	}

	if len(executor.commands) != 2 {
		t.Fatalf("commands = %d, want docker build and run", len(executor.commands))
	}

	_, err = os.Stat(result.SummaryPath)
	if err != nil {
		t.Fatalf("ralph loop summary was not written: %v", err)
	}
}

func TestRunProvingRalphLoopFeedsFailureIntoNextAttempt(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")
	executor := &recordingHarnessExecutor{}
	validations := 0

	result, err := RunProvingRalphLoop(context.Background(), ProvingRalphLoopOptions{
		Worker:        codex.Runner{},
		ProjectDir:    projectDir,
		ObjectivePath: objectivePath,
		Executor:      executor,
		ValidationRunner: func(_ context.Context, _ string, objective ProvingObjective) (ProvingValidationResult, error) {
			validations++
			passed := validations == 2
			exitCode := 0
			stdout := "ok\n"

			if !passed {
				exitCode = 2
				stdout = "expected ids not to be reused\n"
			}

			return ProvingValidationResult{
				Objective: objective,
				Passed:    passed,
				LogPath:   projectcore.ResolvePath(projectDir, objective.ValidationPath),
				Commands: []ProvingValidationCommandResult{
					{Command: "make test", ExitCode: exitCode, Stdout: stdout},
				},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("RunProvingRalphLoop returned error: %v", err)
	}

	if result.Status != ProvingRalphLoopStatusPassed {
		t.Fatalf("Status = %q, want %q", result.Status, ProvingRalphLoopStatusPassed)
	}

	if len(result.Attempts) != 2 {
		t.Fatalf("len(Attempts) = %d, want 2", len(result.Attempts))
	}

	contextContent, err := os.ReadFile(result.Attempts[1].ContextPath)
	if err != nil {
		t.Fatalf("read second attempt context: %v", err)
	}

	for _, want := range []string{
		"Previous Validation Failure",
		"expected ids not to be reused",
		"Do not weaken validation",
	} {
		if !strings.Contains(string(contextContent), want) {
			t.Fatalf("second attempt context does not contain %q:\n%s", want, string(contextContent))
		}
	}
}

func TestRunProvingRalphLoopStopsOnRepeatedFailure(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")

	result, err := RunProvingRalphLoop(context.Background(), ProvingRalphLoopOptions{
		Worker:        codex.Runner{},
		ProjectDir:    projectDir,
		ObjectivePath: objectivePath,
		Executor:      &recordingHarnessExecutor{},
		ValidationRunner: func(_ context.Context, _ string, objective ProvingObjective) (ProvingValidationResult, error) {
			return ProvingValidationResult{
				Objective: objective,
				Passed:    false,
				LogPath:   projectcore.ResolvePath(projectDir, objective.ValidationPath),
				Commands: []ProvingValidationCommandResult{
					{Command: "make test", ExitCode: 2, Stdout: "same failure\n"},
				},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("RunProvingRalphLoop returned error: %v", err)
	}

	if result.Status != ProvingRalphLoopStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, ProvingRalphLoopStatusFailed)
	}

	if result.StopReason != ProvingRalphLoopStopRepeatedFailure {
		t.Fatalf("StopReason = %q, want %q", result.StopReason, ProvingRalphLoopStopRepeatedFailure)
	}

	if len(result.Attempts) != 2 {
		t.Fatalf("len(Attempts) = %d, want 2", len(result.Attempts))
	}
}

func newTestProvingProject(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, workspacecore.DirectoryName, "config.yml"), `version: 1
workspace_root: .chr
canonical_assets: markdown
tickets:
  key_prefix: TODO
runtime:
  image: chr-todolist-agent:local
  dockerfile: .chr/default/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: ops/default/run
  validation_command: make verify
`)
	writeTestFile(t, filepath.Join(dir, workspacecore.DirectoryName, UserDirectoryName, "runtime", "agent", "Dockerfile"), "FROM golang:1.22-bookworm\n")
	writeTestFile(t, filepath.Join(dir, workspacecore.DirectoryName, UserDirectoryName, "tickets", "TODO-1-fix-task-id-reuse.md"), "# Ticket\n")
	writeTestFile(t, filepath.Join(dir, workspacecore.DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md"), `---
objective_id: proving-TODO-1
ticket_key: TODO-1
kind: bugfix
artifact_root: ops/default/run/00001-proving-todo-1
context_path: ops/default/run/00001-proving-todo-1/files/context.md
result_path: ops/default/run/00001-proving-todo-1/evidence/result.json
log_path: ops/default/run/00001-proving-todo-1/evidence/agent.log
validation_path: ops/default/run/00001-proving-todo-1/evidence/validation.log
max_attempts: 3
---

# Fix Task ID Reuse
`)

	return dir
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create test file directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

type recordingHarnessExecutor struct {
	commands []HarnessCommand
}

func (e *recordingHarnessExecutor) Run(
	_ context.Context,
	command HarnessCommand,
	_, _ io.Writer,
) error {
	e.commands = append(e.commands, command)

	return nil
}

func harnessArgsContain(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}

	return false
}
