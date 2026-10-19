package codex

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// Execute the rendered command so shell quoting protects paths and arguments.
func TestHarnessCodexWorker(t *testing.T) {
	dir := t.TempDir()

	input := filepath.Join(dir, "worker's input")
	if err := os.WriteFile(input, []byte("frozen context"), 0o600); err != nil {
		t.Fatal(err)
	}

	command, err := (Runner{}).WorkerCommand(harness.WorkerRequest{
		InputPath: input,
		Env:       []harness.EnvironmentVariable{{Name: "CHR_EXECUTION_ID", Value: "a'b;$()"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if command.Program != Program || !slices.Equal(command.Args, workerArguments()) {
		t.Fatalf("worker command = %#v", command)
	}

	fixture := filepath.Join(dir, "worker")
	if err := os.WriteFile(fixture, []byte("#!/bin/sh\nprintf '%s\\n' \"$CHR_EXECUTION_ID\"\ncat\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	command.Program = fixture

	output, err := exec.CommandContext(t.Context(), "sh", "-c", process.WorkerShellScript(command)).CombinedOutput()
	if err != nil || string(output) != "a'b;$()\nfrozen context" {
		t.Fatalf("worker output = %q, %v", output, err)
	}

	if got := strings.Join((Runner{}).WorkflowWorker(), " "); got != "codex exec --skip-git-repo-check --dangerously-bypass-approvals-and-sandbox -" {
		t.Fatalf("native command = %q", got)
	}
}
