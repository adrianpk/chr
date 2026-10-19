package codex

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/harness"
)

// WorkerCommand binds host-selected environment and input to the Codex CLI.
func (Runner) WorkerCommand(request harness.WorkerRequest) (harness.WorkerCommand, error) {
	if strings.TrimSpace(request.InputPath) == "" {
		return harness.WorkerCommand{}, fmt.Errorf("worker input path is required")
	}

	return harness.WorkerCommand{
		Program:   Program,
		Args:      workerArguments(),
		Env:       append([]harness.EnvironmentVariable(nil), request.Env...),
		InputPath: request.InputPath,
	}, nil
}

func workerArguments() []string {
	return []string{"exec", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox"}
}

// WorkflowWorker consumes host-frozen JSON through the existing native lane.
func (Runner) WorkflowWorker() []string {
	return append(append([]string{Program}, workerArguments()...), "-")
}
