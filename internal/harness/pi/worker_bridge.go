package pi

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

//go:embed worker.mjs
var workerExtension []byte

// WorkerBridgeRequest contains only host-selected technical settings. The
// caller supplies an isolated container, a private writable state directory,
// explicit credential/model mounts and a context tied to the host connection.
type WorkerBridgeRequest struct {
	Provider       string
	Model          string
	Thinking       string
	CredentialFile string
	ModelFile      string
	StateDir       string
	Prompt         string
}

// RunWorkerBridge performs one Pi RPC call inside its owned container. It does
// not validate a task, claim completion, retry a model or acquire capacity.
func RunWorkerBridge(ctx context.Context, request WorkerBridgeRequest) (result Result, err error) {
	if err := prepareWorkerBridge(request); err != nil {
		return result, err
	}

	privateAuth := filepath.Join(request.StateDir, "agent", "auth.json")

	var secrets []string

	err = WithWorkerCredential(ctx, request.CredentialFile, privateAuth, request.Provider, func() error {
		secrets = workerCredentialSecrets(privateAuth)

		command := workerBridgeCommand(request)

		if err := checkVersion(ctx, command); err != nil {
			return err
		}

		var runErr error

		result, runErr = (Transport{}).Run(ctx, command, Request{Prompt: request.Prompt, Provider: request.Provider, Model: request.Model, Thinking: request.Thinking, WorkerTools: true, OutputLimit: harness.MaxWorkerOutput})

		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		return errors.Join(runErr, process.ReapChildren(cleanup))
	})
	secrets = append(secrets, workerCredentialSecrets(privateAuth)...)
	result.Output = redactWorkerOutput(result.Output, secrets)
	// Protocol streams can contain refreshed secrets unknown before execution.
	// The bridge exposes normalized output only; raw records stay ephemeral.
	result.Process.Stdout, result.Process.Stderr = nil, nil

	if err != nil {
		result.Output = nil

		var provider *providerFailure
		if errors.As(err, &provider) && provider.message != "" {
			return result, &privateCallError{cause: err, message: "pi provider failed: " + string(process.DiagnosticWithSecrets([]byte(provider.message), secrets...))}
		}

		return result, workerBridgeError(err)
	}

	return result, nil
}

func workerBridgeCommand(request WorkerBridgeRequest) harness.AgentCommand {
	command := generationCommand(harness.AgentRunRequest{Program: "/usr/local/bin/pi", ModelProvider: request.Provider, Model: request.Model, Thinking: request.Thinking}, request.StateDir)
	for i, arg := range command.Args {
		if arg == "--no-tools" {
			command.Args[i] = "--no-builtin-tools"

			break
		}
	}

	command.Args = append(command.Args, "--tools", "read,bash,edit,write,grep,find,ls", "--extension", filepath.Join(request.StateDir, "worker.mjs"))

	command.Env[0].Value = "/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin"

	return command
}

func workerCredentialSecrets(path string) []string {
	data, err := readPrivateResource(path)
	if err != nil {
		return nil
	}

	var entries map[string]map[string]any
	if json.Unmarshal(data, &entries) != nil {
		return nil
	}

	var secrets []string

	for _, credential := range entries {
		for _, name := range []string{"key", "access", "refresh"} {
			if value, ok := credential[name].(string); ok && value != "" {
				secrets = append(secrets, value)
			}
		}
	}

	return secrets
}

func redactWorkerOutput(data []byte, secrets []string) []byte {
	text := string(data)
	for _, secret := range secrets {
		text = strings.ReplaceAll(text, secret, "[REDACTED]")

		escaped, err := json.Marshal(secret)
		if err == nil {
			text = strings.ReplaceAll(text, string(escaped[1:len(escaped)-1]), "[REDACTED]")
		}
	}

	return process.Sanitize([]byte(text))
}

func workerBridgeError(err error) error {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, process.ErrLimit, ErrCapability, ErrAborted, ErrProtocol, ErrProvider, ErrEarlyExit} {
		if errors.Is(err, cause) {
			return cause
		}
	}

	return errors.New("pi worker failed; check the selected runtime, credential owner and model endpoint")
}

type providerFailure struct{ message string }

func (*providerFailure) Error() string { return ErrProvider.Error() }
func (*providerFailure) Unwrap() error { return ErrProvider }

func writeWorkerToolProfile(stateDir string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	config, err := json.Marshal(map[string]string{"root": root})
	if err != nil {
		return err
	}

	if err := writeCallArtifact(filepath.Join(stateDir, "worker.json"), config); err != nil {
		return err
	}

	return writeCallArtifact(filepath.Join(stateDir, "worker.mjs"), workerExtension)
}

func prepareWorkerBridge(request WorkerBridgeRequest) error {
	if err := ValidateGenerationProfile(request.Provider, request.Model, request.Thinking); err != nil {
		return err
	}

	if err := validateRequest(Request{Prompt: request.Prompt, Provider: request.Provider, Model: request.Model, Thinking: request.Thinking, WorkerTools: true}); err != nil {
		return err
	}

	if err := process.EnableChildReaping(); err != nil {
		return err
	}

	if err := writePrivateCallFiles(request.StateDir, []byte("{}"), nil); err != nil {
		return errors.New("prepare private Pi worker state")
	}

	if err := writeWorkerToolProfile(request.StateDir); err != nil {
		return err
	}

	return copyWorkerModels(request)
}

func copyWorkerModels(request WorkerBridgeRequest) error {
	if request.ModelFile != "" {
		data, err := readPrivateResource(request.ModelFile)
		if err != nil {
			return err
		}

		if err := writeCallArtifact(filepath.Join(request.StateDir, "agent", "models.json"), data); err != nil {
			return err
		}
	}

	return nil
}
