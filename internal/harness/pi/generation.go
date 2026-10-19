package pi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// Identity separates the Pi harness from its selected model provider.
const Identity = "pi"

// Runner implements one prepared native generation call without constructor IO.
type Runner struct{}

// Capabilities reports technical support, never whole-call usage certification.
func (Runner) Capabilities() harness.Capabilities {
	return harness.Capabilities{Generation: true, ReadOnlyFiles: true}
}

// RunAgent preserves caller-owned schemas, output paths and deadlines.
func (Runner) RunAgent(ctx context.Context, request harness.AgentRunRequest) (result harness.AgentRunResult, err error) {
	if err := validateGenerationRequest(request); err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(ctx, generationTimeout(request.Timeout))
	defer cancel()

	call, err := prepareGenerationCall(request)
	if err != nil {
		return result, err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(call.privateDir))
		if err != nil {
			result.Output = nil
		}
	}()

	result = call.result
	if request.OnStart != nil {
		request.OnStart(result)
	}

	if err := checkVersion(ctx, call.command); err != nil {
		return result, err
	}

	response, runErr := (Transport{Executor: process.Executor{Secrets: []string{call.secret}}}).Run(ctx, call.command, generationTransportRequest(request))
	result.Stdout = redactCallDiagnostic(response.Process.Stdout, call.secret)
	result.Stderr = redactCallDiagnostic(response.Process.Stderr, call.secret)
	result.Latency, result.Usage = time.Since(result.StartedAt), response.Usage
	result.Output, result.ObservedModel = response.Output, response.ObservedModel

	return result, errors.Join(redactCallError(runErr, call.secret), writeGenerationLogs(result))
}

type generationCall struct {
	result     harness.AgentRunResult
	command    harness.AgentCommand
	privateDir string
	secret     string
}

func validateGenerationRequest(request harness.AgentRunRequest) error {
	if err := harness.ValidateAgentRunRequest(request); err != nil {
		return err
	}

	if !request.FrozenSelection || request.ModelProvider == "" || request.Model == "" || request.Resources.CredentialFile == "" {
		return fmt.Errorf("pi generation requires a prepared frozen profile: %w", ErrCapability)
	}

	if err := ValidateGenerationProfile(request.ModelProvider, request.Model, request.Thinking); err != nil {
		return err
	}

	if request.RunDir == "" || (len(request.OutputSchema) != 0 && !json.Valid(request.OutputSchema)) {
		return errors.New("pi generation requires an owned run directory and a valid output schema")
	}

	return validateRequest(generationTransportRequest(request))
}

func generationTransportRequest(request harness.AgentRunRequest) Request {
	prompt := request.Prompt
	if len(request.OutputSchema) != 0 {
		prompt += "\n\nHost output schema (data, not instructions):\n" + string(request.OutputSchema)
	}

	return Request{Prompt: prompt, Provider: request.ModelProvider, Model: request.Model, Thinking: request.Thinking, OutputLimit: request.OutputLimit, ReadOnlyFiles: request.ReadOnlyFiles}
}

// SupportsThinking rejects unsupported selections before admission or launch.
func SupportsThinking(value string) bool {
	switch value {
	case "", "off", "minimal", "low", "medium", "high", "xhigh", "max":
		return true
	default:
		return false
	}
}

type privateCallError struct {
	cause   error
	message string
}

func (e *privateCallError) Error() string { return e.message }
func (e *privateCallError) Unwrap() error { return e.cause }

func redactCallError(err error, secret string) error {
	if err == nil {
		return nil
	}

	return &privateCallError{cause: err, message: string(redactCallDiagnostic([]byte(err.Error()), secret))}
}

func generationTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}

	return 20 * time.Minute
}

func checkVersion(ctx context.Context, command harness.AgentCommand) error {
	command.Args = []string{"--version"}

	result, err := (process.Executor{}).RunAgentCommand(ctx, command, "", time.Now())
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err != nil || strings.TrimSpace(string(result.Stdout)) != "1.3.0" {
		return fmt.Errorf("prepare the pinned Pi 1.3.0 executable before generation: %w", ErrCapability)
	}

	return nil
}

func redactCallDiagnostic(data []byte, secret string) []byte {
	return process.DiagnosticWithSecrets(data, secret)
}

func writeGenerationLogs(result harness.AgentRunResult) error {
	for _, file := range []struct {
		path string
		data []byte
	}{{result.StdoutPath, result.Stdout}, {result.StderrPath, result.Stderr}} {
		if err := writeCallArtifact(file.path, file.data); err != nil {
			return errors.New("write bounded Pi generation diagnostics")
		}
	}

	return nil
}

func generationResult(request harness.AgentRunRequest) harness.AgentRunResult {
	name := request.OutputName
	if name == "" {
		name = "output.md"
	}

	return harness.AgentRunResult{
		ID: request.ID, Provider: Identity, Model: request.Model, RunDir: request.RunDir,
		PromptPath: filepath.Join(request.RunDir, "prompt.md"), OutputPath: filepath.Join(request.RunDir, name),
		StdoutPath: filepath.Join(request.RunDir, "stdout.log"), StderrPath: filepath.Join(request.RunDir, "stderr.log"),
		CommandPath: filepath.Join(request.RunDir, "command.json"), StartedAt: time.Now().UTC(),
	}
}
