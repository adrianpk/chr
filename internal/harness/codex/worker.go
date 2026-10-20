package codex

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/harness"
)

// WorkerCommand binds host-selected environment and input to the Codex CLI.
func (r Runner) WorkerCommand(request harness.WorkerRequest) (harness.WorkerCommand, error) {
	if strings.TrimSpace(request.InputPath) == "" {
		return harness.WorkerCommand{}, fmt.Errorf("worker input path is required")
	}

	return harness.WorkerCommand{
		ForwardEnv: []string{"OPENAI_API_KEY"},
		Program:    r.workerProgram(),
		Args:       r.workerArguments(),
		Env:        append([]harness.EnvironmentVariable(nil), request.Env...),
		InputPath:  request.InputPath,
	}, nil
}

func workerArguments() []string {
	return []string{"exec", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox"}
}

// WorkflowWorker consumes host-frozen JSON through the existing native lane.
func (r Runner) WorkflowWorker() []string {
	return append(append([]string{r.workerProgram()}, r.workerArguments()...), "-")
}

func (r Runner) workerProgram() string {
	if r.WorkerProgram != "" {
		return r.WorkerProgram
	}

	return Program
}

func (r Runner) workerArguments() []string {
	args := workerArguments()
	if r.WorkerModel != "" && r.WorkerModel != "cli_default" {
		args = append(args, "--model", r.WorkerModel)
	}

	return args
}
