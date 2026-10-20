package process

import (
	"strings"

	"forge.adrianpk.com/chr/internal/harness"
)

// WorkerShellScript quotes an explicit command and its input redirection.
func WorkerShellScript(command harness.WorkerCommand) string {
	parts := make([]string, 0, len(command.Env)+len(command.Args)+3)
	for _, env := range command.Env {
		parts = append(parts, env.Name+"="+shellQuote(env.Value))
	}

	parts = append(parts, shellQuote(command.Program))
	for _, arg := range command.Args {
		parts = append(parts, shellQuote(arg))
	}

	parts = append(parts, "<", shellQuote(command.InputPath))

	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
