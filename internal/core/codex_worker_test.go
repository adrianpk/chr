package core

import (
	"slices"
	"strings"
	"testing"
)

func TestNewCodexWorkerCommandDefinesCodexInvocation(t *testing.T) {
	command, err := NewCodexWorkerCommand(testCodexWorkerCommandOptions())
	if err != nil {
		t.Fatalf("NewCodexWorkerCommand returned error: %v", err)
	}

	if command.Program != "codex" {
		t.Fatalf("Program = %q, want codex", command.Program)
	}

	wantArgs := []string{
		"exec",
		"--skip-git-repo-check",
		"--dangerously-bypass-approvals-and-sandbox",
	}
	if !slices.Equal(command.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", command.Args, wantArgs)
	}

	assertCodexWorkerEnv(t, command, "CHC_EXECUTION_ID", "exec-1")
	assertCodexWorkerEnv(t, command, "CHC_TICKET_KEY", "CHC-1")
	assertCodexWorkerEnv(t, command, "CHC_TICKET_PATH", "/workspace/.chc/user/tickets/CHC-1.md")
	assertCodexWorkerEnv(t, command, "CHC_WORKER_CONTEXT_PATH", "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/context.md")
	assertCodexWorkerEnv(t, command, "CHC_RESULT_PATH", "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/result.json")
	assertCodexWorkerEnv(t, command, "CHC_LOG_PATH", "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/runtime.log")
	assertCodexWorkerEnv(t, command, "CHC_VALIDATION_PATH", "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/validation.log")
}

func TestRenderCodexWorkerShellScriptQuotesInputs(t *testing.T) {
	opts := testCodexWorkerCommandOptions()
	opts.TicketPath = "/workspace/.chc/user/tickets/CHC-1 Bob's fix.md"

	command, err := NewCodexWorkerCommand(opts)
	if err != nil {
		t.Fatalf("NewCodexWorkerCommand returned error: %v", err)
	}

	script := RenderCodexWorkerShellScript(command)
	for _, want := range []string{
		"CHC_EXECUTION_ID='exec-1'",
		"CHC_TICKET_PATH='/workspace/.chc/user/tickets/CHC-1 Bob'\"'\"'s fix.md'",
		"'codex' 'exec' '--skip-git-repo-check' '--dangerously-bypass-approvals-and-sandbox'",
		`< "$CHC_WORKER_CONTEXT_PATH"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("script does not contain %q:\n%s", want, script)
		}
	}
}

func TestNewCodexWorkerCommandRejectsMissingRequiredFields(t *testing.T) {
	opts := testCodexWorkerCommandOptions()
	opts.ResultPath = ""

	_, err := NewCodexWorkerCommand(opts)
	if err == nil {
		t.Fatal("NewCodexWorkerCommand returned nil error")
	}
}

func assertCodexWorkerEnv(t *testing.T, command CodexWorkerCommand, name, want string) {
	t.Helper()

	for _, env := range command.Env {
		if env.Name == name {
			if env.Value != want {
				t.Fatalf("%s = %q, want %q", name, env.Value, want)
			}

			return
		}
	}

	t.Fatalf("missing env %s", name)
}

func testCodexWorkerCommandOptions() CodexWorkerCommandOptions {
	return CodexWorkerCommandOptions{
		ExecutionID:    "exec-1",
		TicketKey:      "CHC-1",
		TicketPath:     "/workspace/.chc/user/tickets/CHC-1.md",
		ContextPath:    "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/context.md",
		ResultPath:     "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/result.json",
		LogPath:        "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/runtime.log",
		ValidationPath: "/workspace/.chc/user/runtime/runs/CHC-1/exec-1/validation.log",
	}
}
