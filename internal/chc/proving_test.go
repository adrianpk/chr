package chc

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseProvingObjective(t *testing.T) {
	objective, err := ParseProvingObjective(`---
objective_id: proving-TODO-1
ticket_key: TODO-1
kind: bugfix
artifact_root: .chc/runtime/runs/proving/TODO-1
context_path: .chc/runtime/runs/proving/TODO-1/context.md
result_path: .chc/runtime/runs/proving/TODO-1/result.json
log_path: .chc/runtime/runs/proving/TODO-1/agent.log
validation_path: .chc/runtime/runs/proving/TODO-1/validation.log
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
	objectivePath := filepath.Join(DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")

	result, err := PlanProvingSandboxRun(projectDir, objectivePath)
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
		"CHC_EXECUTION_ID=proving-TODO-1",
		"CHC_TICKET_KEY=TODO-1",
		"CHC_WORKER_CONTEXT_PATH=/workspace/.chc/runtime/runs/proving/TODO-1/context.md",
	} {
		if !harnessArgsContain(result.Run.Args, want) {
			t.Fatalf("run args do not contain %q:\n%#v", want, result.Run.Args)
		}
	}
}

func TestRunProvingSandboxDryRunDoesNotExecuteCommands(t *testing.T) {
	projectDir := newTestProvingProject(t)
	objectivePath := filepath.Join(DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")
	executor := &recordingHarnessExecutor{}

	result, err := RunProvingSandbox(context.Background(), ProvingSandboxRunOptions{
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
	objectivePath := filepath.Join(DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md")

	result, err := PlanProvingSandboxRun(projectDir, objectivePath)
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

func newTestProvingProject(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, DirectoryName, "config.yml"), `version: 1
workspace_root: .chc
canonical_assets: markdown
tickets:
  key_prefix: TODO
runtime:
  image: chc-todolist-agent:local
  dockerfile: .chc/runtime/agent/Dockerfile
  workdir: /workspace
  artifact_dir: .chc/runtime/runs
  validation_command: make verify
`)
	writeTestFile(t, filepath.Join(dir, DirectoryName, "runtime", "agent", "Dockerfile"), "FROM golang:1.22-bookworm\n")
	writeTestFile(t, filepath.Join(dir, DirectoryName, "tickets", "TODO-1-fix-task-id-reuse.md"), "# Ticket\n")
	writeTestFile(t, filepath.Join(dir, DirectoryName, "proving", "objectives", "TODO-1-fix-task-id-reuse.md"), `---
objective_id: proving-TODO-1
ticket_key: TODO-1
kind: bugfix
artifact_root: .chc/runtime/runs/proving/TODO-1
context_path: .chc/runtime/runs/proving/TODO-1/context.md
result_path: .chc/runtime/runs/proving/TODO-1/result.json
log_path: .chc/runtime/runs/proving/TODO-1/agent.log
validation_path: .chc/runtime/runs/proving/TODO-1/validation.log
max_attempts: 3
---

# Fix Task ID Reuse
`)

	return dir
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create test file directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
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
