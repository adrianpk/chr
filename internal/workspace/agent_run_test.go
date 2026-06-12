package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeAgentCommandExecutor struct {
	Command           AgentCommand
	Stdin             string
	Result            AgentCommandResult
	OutputFileContent string
	Err               error
}

func (e *fakeAgentCommandExecutor) RunAgentCommand(
	_ context.Context,
	command AgentCommand,
	stdin string,
	_ time.Time,
) (AgentCommandResult, error) {
	e.Command = command
	e.Stdin = stdin

	err := e.writeOutputFile(command)
	if err != nil {
		return e.Result, err
	}

	return e.Result, e.Err
}

func (e *fakeAgentCommandExecutor) writeOutputFile(command AgentCommand) error {
	if e.OutputFileContent == "" {
		return nil
	}

	outputPath := fakeAgentOutputPath(command.Args)
	if outputPath == "" {
		return nil
	}

	return os.WriteFile(outputPath, []byte(e.OutputFileContent), 0o644)
}

func TestCodexAgentRunnerWritesRunBundleAndReadsOutputFile(t *testing.T) {
	projectDir := t.TempDir()
	executor := &fakeAgentCommandExecutor{OutputFileContent: `{"ok":true}`}
	runner := CodexAgentRunner{CommandExecutor: executor}

	result, err := runner.RunAgent(context.Background(), AgentRunRequest{
		ID:         "agent-run-1",
		Purpose:    "spec_generation",
		ProjectDir: projectDir,
		Prompt:     "Generate a spec.",
		OutputName: "output.json",
		Program:    "codex-test",
		Model:      "gpt-test",
	})
	if err != nil {
		t.Fatalf("RunAgent returned error: %v", err)
	}

	if executor.Command.Program != "codex-test" {
		t.Fatalf("program = %q, want codex-test", executor.Command.Program)
	}

	if !strings.Contains(strings.Join(executor.Command.Args, " "), "--sandbox read-only") {
		t.Fatalf("command args do not set read-only sandbox: %#v", executor.Command.Args)
	}

	if executor.Stdin != "Generate a spec." {
		t.Fatalf("stdin = %q, want prompt", executor.Stdin)
	}

	for _, path := range []string{
		result.PromptPath,
		result.StdoutPath,
		result.StderrPath,
		result.CommandPath,
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected run bundle file %s: %v", path, err)
		}
	}

	if !strings.HasPrefix(result.RunDir, filepath.Join(projectDir, ".chc", "user", "tmp", "agent-runs")) {
		t.Fatalf("run dir = %q, want userspace tmp agent-runs path", result.RunDir)
	}

	if string(result.Output) != `{"ok":true}` {
		t.Fatalf("output = %q, want output file content", result.Output)
	}
}

func TestCodexAgentRunnerFallsBackToStdoutWhenOutputFileIsMissing(t *testing.T) {
	projectDir := t.TempDir()
	executor := &fakeAgentCommandExecutor{
		Result: AgentCommandResult{Stdout: []byte("stdout result")},
	}
	runner := CodexAgentRunner{CommandExecutor: executor}

	result, err := runner.RunAgent(context.Background(), AgentRunRequest{
		ID:         "agent-run-2",
		Purpose:    "ticket_generation",
		ProjectDir: projectDir,
		Prompt:     "Generate a ticket.",
	})
	if err != nil {
		t.Fatalf("RunAgent returned error: %v", err)
	}

	if string(result.Output) != "stdout result" {
		t.Fatalf("output = %q, want stdout fallback", result.Output)
	}
}

func TestValidateAgentRunRequestRequiresCoreFields(t *testing.T) {
	err := ValidateAgentRunRequest(AgentRunRequest{})
	if err == nil {
		t.Fatal("ValidateAgentRunRequest returned nil error")
	}
}

func fakeAgentOutputPath(args []string) string {
	for index, arg := range args {
		if arg == "--output-last-message" && index+1 < len(args) {
			return args[index+1]
		}
	}

	return ""
}
