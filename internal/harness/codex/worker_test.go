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

// Native and container commands carry the same host-selected executable/model;
// the empty/default selection preserves the original legacy command exactly.
func TestHarnessCodexWorkerSelection(t *testing.T) {
	for _, tc := range []struct{ name, model string }{{"legacy", ""}, {"CLI default", "cli_default"}, {"exact", "exact model"}} {
		t.Run(tc.name, func(t *testing.T) {
			model := tc.model
			runner := Runner{WorkerProgram: "/host/selected-codex", WorkerModel: model}
			native := runner.WorkflowWorker()

			container, err := runner.WorkerCommand(harness.WorkerRequest{InputPath: "/input"})
			if err != nil || container.Program != native[0] || !slices.Equal(container.Args, native[1:len(native)-1]) {
				t.Fatal("worker transports disagree on selection", err)
			}

			hasModel := slices.Contains(native, "--model")
			if hasModel != (model == "exact model") {
				t.Fatal("worker lost explicit/default model semantics", native)
			}
		})
	}
}
