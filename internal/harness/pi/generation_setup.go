package pi

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/harness"
)

func prepareGenerationCall(request harness.AgentRunRequest) (call generationCall, err error) {
	auth, models, secret, err := loadCallResources(request.Resources, request.ModelProvider)
	if err != nil {
		return call, err
	}

	call.privateDir, err = createGenerationDirectory(request.RunDir)
	if err != nil {
		return call, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(call.privateDir))
		}
	}()

	call.result, call.secret = generationResult(request), secret

	if err := writeGenerationInput(call, request); err != nil {
		return call, err
	}

	if err := writePrivateCallFiles(call.privateDir, auth, models); err != nil {
		return call, err
	}

	if request.ReadOnlyFiles {
		if err := writeReadOnlyProfile(call.privateDir, request); err != nil {
			return call, err
		}
	}

	call.command = generationCommand(request, call.privateDir)
	call.result.Command = append([]string{call.command.Program}, call.command.Args...)

	data, err := json.Marshal(call.command)
	if err != nil {
		return call, err
	}

	return call, writeCallArtifact(call.result.CommandPath, data)
}

func createGenerationDirectory(runDir string) (string, error) {
	if err := checkCallDirectoryPath(runDir); err != nil {
		return "", err
	}

	if err := os.MkdirAll(runDir, 0o700); err != nil {
		return "", err
	}

	info, err := os.Lstat(runDir)
	if err != nil || !info.IsDir() {
		return "", errors.New("pi run directory must be host-owned")
	}

	real, err := filepath.EvalSymlinks(runDir)
	if err != nil || real != filepath.Clean(runDir) {
		return "", errors.New("pi run directory cannot traverse symlinks")
	}

	return os.MkdirTemp("", "chr-pi-call-")
}

func writeGenerationInput(call generationCall, request harness.AgentRunRequest) error {
	if err := writeCallArtifact(call.result.PromptPath, []byte(request.Prompt)); err != nil {
		return err
	}

	if len(request.OutputSchema) != 0 {
		return writeCallArtifact(filepath.Join(request.RunDir, "output-schema.json"), request.OutputSchema)
	}

	return nil
}

func writePrivateCallFiles(dir string, auth, models []byte) error {
	for _, name := range []string{"agent", "home", "tmp", "config", "cache", "data"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o700); err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "agent", "auth.json"), auth, 0o600); err != nil {
		return err
	}

	if len(models) != 0 {
		if err := os.WriteFile(filepath.Join(dir, "agent", "models.json"), models, 0o600); err != nil {
			return err
		}
	}

	return os.WriteFile(filepath.Join(dir, "agent", "settings.json"), []byte(`{"cacheWarming":"off","retry":{"enabled":false,"maxRetries":0,"provider":{"maxRetries":0}},"compaction":{"enabled":false},"packages":[],"extensions":[],"skills":[],"prompts":[],"themes":[]}`), 0o600)
}

func generationCommand(request harness.AgentRunRequest, dir string) harness.AgentCommand {
	args := []string{
		"--mode", "rpc", "--no-session", "--offline", "--no-extensions", "--no-mcp", "--no-skills", "--no-prompt-templates", "--no-themes", "--no-context-files", "--no-approve",
		"--provider", request.ModelProvider, "--model", request.Model,
	}
	if request.ReadOnlyFiles {
		args = append(args, "--no-builtin-tools", "--tools", "chr_inspect", "--extension", filepath.Join(dir, "readonly.mjs"))
	} else {
		args = append(args, "--no-tools")
	}

	if request.Thinking != "" {
		args = append(args, "--thinking", request.Thinking)
	}

	return harness.AgentCommand{Program: request.Program, Args: args, Dir: dir, Env: []harness.EnvironmentVariable{
		{Name: "PATH", Value: os.Getenv("PATH")},
		{Name: "HOME", Value: filepath.Join(dir, "home")},
		{Name: "PI_CODING_AGENT_DIR", Value: filepath.Join(dir, "agent")},
		{Name: "TMPDIR", Value: filepath.Join(dir, "tmp")},
		{Name: "XDG_CONFIG_HOME", Value: filepath.Join(dir, "config")},
		{Name: "XDG_CACHE_HOME", Value: filepath.Join(dir, "cache")},
		{Name: "XDG_DATA_HOME", Value: filepath.Join(dir, "data")},
		{Name: "PI_OFFLINE", Value: "1"},
	}}
}
