package chc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	AgentProviderCodex = "codex"

	defaultCodexAgentCommand = "codex"
	defaultCodexAgentTimeout = 20 * time.Minute
)

type AgentRunRequest struct {
	ID                 string
	Purpose            string
	ProjectDir         string
	RunDir             string
	Prompt             string
	OutputSchema       []byte
	OutputName         string
	Program            string
	Model              string
	Timeout            time.Duration
	DeferResponseTrace bool
}

type AgentRunResult struct {
	ID          string
	Provider    string
	Model       string
	RunDir      string
	PromptPath  string
	OutputPath  string
	StdoutPath  string
	StderrPath  string
	CommandPath string
	Command     []string
	Output      []byte
	Stdout      []byte
	Stderr      []byte
	StartedAt   time.Time
	Latency     time.Duration
}

type AgentRunner interface {
	RunAgent(ctx context.Context, request AgentRunRequest) (AgentRunResult, error)
}

type AgentCommand struct {
	Program string
	Args    []string
	Dir     string
}

type AgentCommandResult struct {
	Stdout  []byte
	Stderr  []byte
	Latency time.Duration
}

type AgentCommandExecutor interface {
	RunAgentCommand(ctx context.Context, command AgentCommand, stdin string, startedAt time.Time) (AgentCommandResult, error)
}

type OSAgentCommandExecutor struct{}

type CodexAgentRunner struct {
	CommandExecutor AgentCommandExecutor
}

type preparedAgentRun struct {
	Request    AgentRunRequest
	Result     AgentRunResult
	SchemaPath string
	StartedAt  time.Time
}

func (r CodexAgentRunner) RunAgent(ctx context.Context, request AgentRunRequest) (AgentRunResult, error) {
	err := ValidateAgentRunRequest(request)
	if err != nil {
		return AgentRunResult{}, err
	}

	run, err := prepareAgentRun(request)
	if err != nil {
		return AgentRunResult{}, err
	}

	recordAgentRunRequest(ctx, run)

	command := codexAgentCommand(run)
	result := run.Result

	result.Command = append([]string{command.Program}, command.Args...)

	err = writeAgentRunCommand(result.CommandPath, command)
	if err != nil {
		return result, err
	}

	execResult, execErr := r.commandExecutor().RunAgentCommand(ctx, command, request.Prompt, run.StartedAt)
	result.Stdout = execResult.Stdout
	result.Stderr = execResult.Stderr
	result.Latency = execResult.Latency

	logErr := writeAgentRunLogs(result)
	if logErr != nil {
		recordAgentRunResponseIfEnabled(ctx, run, result, "error", logErr)

		return result, logErr
	}

	if execErr != nil {
		err = fmt.Errorf("invoke Codex agent run: %w: %s", execErr, strings.TrimSpace(string(result.Stderr)))
		recordAgentRunResponseIfEnabled(ctx, run, result, "error", err)

		return result, err
	}

	output, err := readAgentRunOutput(result.OutputPath, result.Stdout)
	if err != nil {
		recordAgentRunResponseIfEnabled(ctx, run, result, "error", err)

		return result, err
	}

	result.Output = output
	recordAgentRunResponseIfEnabled(ctx, run, result, "ready", nil)

	return result, nil
}

func (OSAgentCommandExecutor) RunAgentCommand(
	ctx context.Context,
	command AgentCommand,
	stdin string,
	startedAt time.Time,
) (AgentCommandResult, error) {
	cmd := exec.CommandContext(ctx, command.Program, command.Args...)
	cmd.Dir = command.Dir
	cmd.Stdin = strings.NewReader(stdin)

	var stdout bytes.Buffer

	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return AgentCommandResult{
		Stdout:  stdout.Bytes(),
		Stderr:  stderr.Bytes(),
		Latency: time.Since(startedAt),
	}, err
}

func ValidateAgentRunRequest(request AgentRunRequest) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "id", value: request.ID},
		{name: "purpose", value: request.Purpose},
		{name: "project dir", value: request.ProjectDir},
		{name: "prompt", value: request.Prompt},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("agent run %s is required", field.name)
		}
	}

	return nil
}

func prepareAgentRun(request AgentRunRequest) (preparedAgentRun, error) {
	startedAt := time.Now().UTC()

	runDir := strings.TrimSpace(request.RunDir)
	if runDir == "" {
		runDir = UserPath(request.ProjectDir, "tmp", "agent-runs", request.ID)
	}

	err := os.MkdirAll(runDir, 0o755)
	if err != nil {
		return preparedAgentRun{}, fmt.Errorf("create agent run directory: %w", err)
	}

	promptPath := filepath.Join(runDir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte(request.Prompt), 0o644); err != nil {
		return preparedAgentRun{}, fmt.Errorf("write agent run prompt: %w", err)
	}

	schemaPath := ""
	if len(request.OutputSchema) > 0 {
		schemaPath = filepath.Join(runDir, "output-schema.json")
		if err := os.WriteFile(schemaPath, request.OutputSchema, 0o644); err != nil {
			return preparedAgentRun{}, fmt.Errorf("write agent run output schema: %w", err)
		}
	}

	outputName := strings.TrimSpace(request.OutputName)
	if outputName == "" {
		outputName = "output.md"
	}

	result := AgentRunResult{
		ID:          request.ID,
		Provider:    AgentProviderCodex,
		Model:       codexAgentModel(request),
		RunDir:      runDir,
		PromptPath:  promptPath,
		OutputPath:  filepath.Join(runDir, outputName),
		StdoutPath:  filepath.Join(runDir, "stdout.log"),
		StderrPath:  filepath.Join(runDir, "stderr.log"),
		CommandPath: filepath.Join(runDir, "command.json"),
		StartedAt:   startedAt,
	}

	return preparedAgentRun{
		Request:    request,
		Result:     result,
		SchemaPath: schemaPath,
		StartedAt:  startedAt,
	}, nil
}

func codexAgentCommand(run preparedAgentRun) AgentCommand {
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

	return AgentCommand{
		Program: codexAgentProgram(request),
		Args:    args,
		Dir:     request.ProjectDir,
	}
}

func writeAgentRunCommand(path string, command AgentCommand) error {
	payload, err := json.MarshalIndent(command, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agent run command: %w", err)
	}

	err = os.WriteFile(path, payload, 0o644)
	if err != nil {
		return fmt.Errorf("write agent run command: %w", err)
	}

	return nil
}

func writeAgentRunLogs(result AgentRunResult) error {
	if err := os.WriteFile(result.StdoutPath, result.Stdout, 0o644); err != nil {
		return fmt.Errorf("write agent run stdout: %w", err)
	}

	if err := os.WriteFile(result.StderrPath, result.Stderr, 0o644); err != nil {
		return fmt.Errorf("write agent run stderr: %w", err)
	}

	return nil
}

func readAgentRunOutput(path string, stdout []byte) ([]byte, error) {
	output, err := os.ReadFile(path)
	if err == nil {
		return output, nil
	}

	if len(stdout) > 0 {
		return stdout, nil
	}

	return nil, fmt.Errorf("read agent run output: %w", err)
}

func (r CodexAgentRunner) commandExecutor() AgentCommandExecutor {
	if r.CommandExecutor != nil {
		return r.CommandExecutor
	}

	return OSAgentCommandExecutor{}
}

func codexAgentProgram(request AgentRunRequest) string {
	if program := strings.TrimSpace(request.Program); program != "" {
		return program
	}

	if program := strings.TrimSpace(os.Getenv("CHC_CODEX_COMMAND")); program != "" {
		return program
	}

	return defaultCodexAgentCommand
}

func codexAgentModel(request AgentRunRequest) string {
	if model := strings.TrimSpace(request.Model); model != "" {
		return model
	}

	return strings.TrimSpace(os.Getenv("CHC_CODEX_MODEL"))
}

func codexAgentTimeout(request AgentRunRequest) time.Duration {
	if request.Timeout > 0 {
		return request.Timeout
	}

	if value := strings.TrimSpace(os.Getenv("CHC_CODEX_TIMEOUT_SECONDS")); value != "" {
		seconds, err := strconv.Atoi(value)
		if err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}

	return defaultCodexAgentTimeout
}

func recordAgentRunRequest(ctx context.Context, run preparedAgentRun) {
	request := run.Request

	RecordLLMTraceEvent(ctx, LoadLLMTraceConfigOptions{ProjectDir: request.ProjectDir}, map[string]any{
		"event":                 "llm_request",
		"llm_call_id":           run.Result.ID,
		"request_id":            request.ID,
		"action":                request.Purpose,
		"provider":              AgentProviderCodex,
		"model":                 run.Result.Model,
		"prompt_path":           run.Result.PromptPath,
		"output_path":           run.Result.OutputPath,
		"request_payload_bytes": len(request.Prompt),
		"request_payload":       request.Prompt,
	})
}

func recordAgentRunResponse(ctx context.Context, run preparedAgentRun, result AgentRunResult, status string, runErr error) {
	recordAgentRunResponseEvent(ctx, run.Request, result, status, runErr)
}

func recordAgentRunResponseIfEnabled(
	ctx context.Context,
	run preparedAgentRun,
	result AgentRunResult,
	status string,
	runErr error,
) {
	if run.Request.DeferResponseTrace {
		return
	}

	recordAgentRunResponse(ctx, run, result, status, runErr)
}

func recordAgentRunResponseEvent(
	ctx context.Context,
	request AgentRunRequest,
	result AgentRunResult,
	status string,
	runErr error,
) {
	event := map[string]any{
		"event":                  "llm_response",
		"llm_call_id":            result.ID,
		"request_id":             request.ID,
		"action":                 request.Purpose,
		"provider":               AgentProviderCodex,
		"model":                  result.Model,
		"latency_ms":             result.Latency.Milliseconds(),
		"status":                 status,
		"output_payload_bytes":   len(result.Output),
		"response_payload_bytes": len(result.Stdout) + len(result.Stderr),
		"output_payload":         string(result.Output),
		"response_payload":       string(result.Output),
		"stdout":                 string(result.Stdout),
		"stderr":                 string(result.Stderr),
	}
	if runErr != nil {
		event["error"] = runErr.Error()
	}

	RecordLLMTraceEvent(ctx, LoadLLMTraceConfigOptions{ProjectDir: request.ProjectDir}, event)
}

func runCodexAgent(ctx context.Context, runner AgentRunner, request AgentRunRequest) (AgentRunResult, error) {
	timeout := codexAgentTimeout(request)
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := runner.RunAgent(runCtx, request)
	if runCtx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("invoke Codex agent run: deadline exceeded after %s", timeout)
	}

	return result, err
}
