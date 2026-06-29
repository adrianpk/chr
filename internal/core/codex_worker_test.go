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

	assertCodexWorkerEnv(t, command, "CHR_EXECUTION_ID", "exec-1")
	assertCodexWorkerEnv(t, command, "CHR_TICKET_KEY", "CHR-1")
	assertCodexWorkerEnv(t, command, "CHR_TICKET_PATH", "/workspace/.chr/user/tickets/CHR-1.md")
	assertCodexWorkerEnv(t, command, "CHR_WORKER_CONTEXT_PATH", "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/context.md")
	assertCodexWorkerEnv(t, command, "CHR_RESULT_PATH", "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/result.json")
	assertCodexWorkerEnv(t, command, "CHR_LOG_PATH", "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/runtime.log")
	assertCodexWorkerEnv(t, command, "CHR_VALIDATION_PATH", "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/validation.log")
}

func TestRenderCodexWorkerShellScriptQuotesInputs(t *testing.T) {
	opts := testCodexWorkerCommandOptions()
	opts.TicketPath = "/workspace/.chr/user/tickets/CHR-1 Bob's fix.md"

	command, err := NewCodexWorkerCommand(opts)
	if err != nil {
		t.Fatalf("NewCodexWorkerCommand returned error: %v", err)
	}

	script := RenderCodexWorkerShellScript(command)
	for _, want := range []string{
		"CHR_EXECUTION_ID='exec-1'",
		"CHR_TICKET_PATH='/workspace/.chr/user/tickets/CHR-1 Bob'\"'\"'s fix.md'",
		"'codex' 'exec' '--skip-git-repo-check' '--dangerously-bypass-approvals-and-sandbox'",
		`< "$CHR_WORKER_CONTEXT_PATH"`,
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
		TicketKey:      "CHR-1",
		TicketPath:     "/workspace/.chr/user/tickets/CHR-1.md",
		ContextPath:    "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/context.md",
		ResultPath:     "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/result.json",
		LogPath:        "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/runtime.log",
		ValidationPath: "/workspace/.chr/user/runtime/runs/CHR-1/exec-1/validation.log",
	}
}
