package core

import (
	"fmt"
	"strings"
)

const (
	// CodexWorkerProgram defines a package boundary constant.
	CodexWorkerProgram = "codex"
)

// CodexWorkerCommandOptions describes a package boundary value.
type CodexWorkerCommandOptions struct {
	ExecutionID    string
	TicketKey      string
	TicketPath     string
	ContextPath    string
	ResultPath     string
	LogPath        string
	ValidationPath string
}

// CodexWorkerCommand describes a package boundary value.
type CodexWorkerCommand struct {
	Program string
	Args    []string
	Env     []CodexWorkerEnv
}

// CodexWorkerEnv describes a package boundary value.
type CodexWorkerEnv struct {
	Name  string
	Value string
}

// NewCodexWorkerCommand constructs the corresponding package service or value.
func NewCodexWorkerCommand(opts CodexWorkerCommandOptions) (CodexWorkerCommand, error) {
	err := validateCodexWorkerCommandOptions(opts)
	if err != nil {
		return CodexWorkerCommand{}, err
	}

	return CodexWorkerCommand{
		Program: CodexWorkerProgram,
		Args:    codexWorkerArguments(),
		Env: []CodexWorkerEnv{
			{Name: "CHR_EXECUTION_ID", Value: opts.ExecutionID},
			{Name: "CHR_TICKET_KEY", Value: opts.TicketKey},
			{Name: "CHR_TICKET_PATH", Value: opts.TicketPath},
			{Name: "CHR_WORKER_CONTEXT_PATH", Value: opts.ContextPath},
			{Name: "CHR_RESULT_PATH", Value: opts.ResultPath},
			{Name: "CHR_LOG_PATH", Value: opts.LogPath},
			{Name: "CHR_VALIDATION_PATH", Value: opts.ValidationPath},
		},
	}, nil
}

func codexWorkerArguments() []string {
	return []string{"exec", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox"}
}

// CodexWorkflowWorker consumes the host-frozen JSON input on standard input.
// The native executor owns its process, isolated lane and validation evidence.
func CodexWorkflowWorker() []string {
	return append(append([]string{CodexWorkerProgram}, codexWorkerArguments()...), "-")
}

// RenderCodexWorkerShellScript renders the corresponding package data.
func RenderCodexWorkerShellScript(command CodexWorkerCommand) string {
	parts := make([]string, 0, len(command.Env)+len(command.Args)+2)

	for _, env := range command.Env {
		parts = append(parts, env.Name+"="+shellQuote(env.Value))
	}

	parts = append(parts, shellQuote(command.Program))
	for _, arg := range command.Args {
		parts = append(parts, shellQuote(arg))
	}

	parts = append(parts, "<", "\"$CHR_WORKER_CONTEXT_PATH\"")

	return strings.Join(parts, " ")
}

func validateCodexWorkerCommandOptions(opts CodexWorkerCommandOptions) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "execution id", value: opts.ExecutionID},
		{name: "ticket key", value: opts.TicketKey},
		{name: "ticket path", value: opts.TicketPath},
		{name: "worker context path", value: opts.ContextPath},
		{name: "result path", value: opts.ResultPath},
		{name: "log path", value: opts.LogPath},
		{name: "validation path", value: opts.ValidationPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("codex worker %s is required", field.name)
		}
	}

	return nil
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}

	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
