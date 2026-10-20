// Package codex implements Codex generation and worker transport.
package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

const (
	// Program is the legacy Codex executable name.
	Program = "codex"
	// Identity is the harness identity, separate from its model provider.
	Identity = "codex"
	// WorkerProtocol identifies the bounded native worker argv contract.
	WorkerProtocol = "codex-exec/v1"
	defaultTimeout = 20 * time.Minute
)

// Runner implements generation and worker command capabilities without IO at construction.
type Runner struct {
	WorkerProgram   string
	WorkerModel     string
	Mounts          []harness.RuntimeMount
	CommandExecutor harness.AgentCommandExecutor
}

// Capabilities reports local adapter support; it does not certify token enforcement.
func (Runner) Capabilities() harness.Capabilities {
	return harness.Capabilities{Generation: true, Worker: true, ReadOnlyFiles: true}
}

type preparedAgentRun struct {
	Request    harness.AgentRunRequest
	Result     harness.AgentRunResult
	SchemaPath string
	StartedAt  time.Time
}

// RunAgent executes the configured operation.
func (r Runner) RunAgent(ctx context.Context, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	err := harness.ValidateAgentRunRequest(request)
	if err != nil {
		return harness.AgentRunResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, codexAgentTimeout(request))
	defer cancel()

	run, err := prepareAgentRun(request)
	if err != nil {
		return harness.AgentRunResult{}, err
	}

	if request.OnStart != nil {
		request.OnStart(run.Result)
	}

	command := codexAgentCommand(run)
	result := run.Result

	result.Command = append([]string{command.Program}, command.Args...)

	err = writeAgentRunCommand(result.CommandPath, command)
	if err != nil {
		return result, err
	}

	execResult, execErr := r.commandExecutor().RunAgentCommand(ctx, command, request.Prompt, run.StartedAt)
	result.Stdout = process.Diagnostic(execResult.Stdout)
	result.Stderr = process.Diagnostic(execResult.Stderr)
	result.Latency = execResult.Latency

	logErr := writeAgentRunLogs(result)
	if logErr != nil {
		return result, logErr
	}

	if execErr != nil {
		err = fmt.Errorf("invoke Codex agent run: %w: %s", execErr, strings.TrimSpace(string(result.Stderr)))

		return result, err
	}

	output, err := readAgentRunOutput(result.OutputPath, execResult.Stdout, outputLimit(request))
	if err != nil {
		return result, err
	}

	result.Output = output

	return result, nil
}

func prepareAgentRun(request harness.AgentRunRequest) (preparedAgentRun, error) {
	startedAt := time.Now().UTC()

	runDir := strings.TrimSpace(request.RunDir)
	if runDir == "" {
		return preparedAgentRun{}, fmt.Errorf("agent run directory is required")
	}

	err := os.MkdirAll(runDir, 0o700)
	if err != nil {
		return preparedAgentRun{}, fmt.Errorf("create agent run directory: %w", err)
	}

	promptPath := filepath.Join(runDir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte(request.Prompt), 0o600); err != nil {
		return preparedAgentRun{}, fmt.Errorf("write agent run prompt: %w", err)
	}

	schemaPath := ""
	if len(request.OutputSchema) > 0 {
		schemaPath = filepath.Join(runDir, "output-schema.json")
		if err := os.WriteFile(schemaPath, request.OutputSchema, 0o600); err != nil {
			return preparedAgentRun{}, fmt.Errorf("write agent run output schema: %w", err)
		}
	}

	outputName := strings.TrimSpace(request.OutputName)
	if outputName == "" {
		outputName = "output.md"
	}

	result := harness.AgentRunResult{
		ID:          request.ID,
		Provider:    Identity,
		Model:       codexAgentModel(request),
		RunDir:      runDir,
		PromptPath:  promptPath,
		OutputPath:  filepath.Join(runDir, outputName),
		StdoutPath:  filepath.Join(runDir, "stdout.log"),
		StderrPath:  filepath.Join(runDir, "stderr.log"),
		CommandPath: filepath.Join(runDir, "command.json"),
		StartedAt:   startedAt,
	}

	if err := os.Remove(result.OutputPath); err != nil && !os.IsNotExist(err) {
		return preparedAgentRun{}, fmt.Errorf("clear previous agent output: %w", err)
	}

	return preparedAgentRun{
		Request:    request,
		Result:     result,
		SchemaPath: schemaPath,
		StartedAt:  startedAt,
	}, nil
}

func codexAgentCommand(run preparedAgentRun) harness.AgentCommand {
	request := run.Request
	args := []string{
		"exec",
		"--cd", request.ProjectDir,
		"--sandbox", "read-only",
		"--skip-git-repo-check",
		"--ephemeral",
		"--output-last-message", run.Result.OutputPath,
	}

	if run.SchemaPath != "" {
		args = append(args, "--output-schema", run.SchemaPath)
	}

	if model := codexAgentModel(request); model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, "-")

	return harness.AgentCommand{
		Program: codexAgentProgram(request),
		Args:    args,
		Dir:     request.ProjectDir,
	}
}

func writeAgentRunCommand(path string, command harness.AgentCommand) error {
	payload, err := json.MarshalIndent(command, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agent run command: %w", err)
	}

	err = os.WriteFile(path, payload, 0o600)
	if err != nil {
		return fmt.Errorf("write agent run command: %w", err)
	}

	return nil
}

func writeAgentRunLogs(result harness.AgentRunResult) error {
	if err := os.WriteFile(result.StdoutPath, result.Stdout, 0o600); err != nil {
		return fmt.Errorf("write agent run stdout: %w", err)
	}

	if err := os.WriteFile(result.StderrPath, result.Stderr, 0o600); err != nil {
		return fmt.Errorf("write agent run stderr: %w", err)
	}

	return nil
}

func readAgentRunOutput(path string, stdout []byte, limit int) ([]byte, error) {
	output, err := process.ReadOutput(path, limit)
	if err == nil {
		return output, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read agent run output: %w", err)
	}

	if len(stdout) > limit {
		return nil, process.ErrLimit
	}

	if len(stdout) > 0 {
		return stdout, nil
	}

	return nil, fmt.Errorf("read agent run output: %w", err)
}

func outputLimit(request harness.AgentRunRequest) int {
	if request.OutputLimit > 0 && request.OutputLimit < harness.MaxStructuredOutput {
		return request.OutputLimit
	}

	return harness.MaxStructuredOutput
}

func (r Runner) commandExecutor() harness.AgentCommandExecutor {
	if r.CommandExecutor != nil {
		return r.CommandExecutor
	}

	return process.Executor{}
}

func codexAgentProgram(request harness.AgentRunRequest) string {
	if program := strings.TrimSpace(request.Program); program != "" {
		return program
	}

	if program := strings.TrimSpace(os.Getenv("CHR_CODEX_COMMAND")); program != "" {
		return program
	}

	return Program
}

func codexAgentModel(request harness.AgentRunRequest) string {
	if request.FrozenSelection {
		return request.Model
	}

	if model := strings.TrimSpace(request.Model); model != "" {
		return model
	}

	return strings.TrimSpace(os.Getenv("CHR_CODEX_MODEL"))
}

func codexAgentTimeout(request harness.AgentRunRequest) time.Duration {
	if request.Timeout > 0 {
		return request.Timeout
	}

	if value := strings.TrimSpace(os.Getenv("CHR_CODEX_TIMEOUT_SECONDS")); value != "" {
		seconds, err := strconv.Atoi(value)
		if err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}

	return defaultTimeout
}
