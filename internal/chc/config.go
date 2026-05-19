package chc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WorkspaceConfig struct {
	TicketKeyPrefix string
	Runtime         RuntimeConfig
}

type RuntimeConfig struct {
	Image             string
	DockerfilePath    string
	Workdir           string
	ArtifactDir       string
	ValidationCommand string
}

const (
	DefaultRuntimeImage             = "chc-agent-runtime:local"
	DefaultRuntimeDockerfilePath    = ".chc/runtime/agent/Dockerfile"
	DefaultRuntimeWorkdir           = "/workspace"
	DefaultRuntimeArtifactDir       = ".chc/runtime/runs"
	DefaultRuntimeValidationCommand = "make verify"
)

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Image:             DefaultRuntimeImage,
		DockerfilePath:    DefaultRuntimeDockerfilePath,
		Workdir:           DefaultRuntimeWorkdir,
		ArtifactDir:       DefaultRuntimeArtifactDir,
		ValidationCommand: DefaultRuntimeValidationCommand,
	}
}

func LoadWorkspaceConfig(projectDir string) (WorkspaceConfig, error) {
	config := WorkspaceConfig{
		TicketKeyPrefix: DefaultTicketKeyPrefix,
		Runtime:         DefaultRuntimeConfig(),
	}
	path := filepath.Join(projectDir, ".chc", "config.yml")

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}

	if err != nil {
		return WorkspaceConfig{}, fmt.Errorf("open workspace config: %w", err)
	}

	defer func() {
		//nolint:errcheck // The config file is read-only here; scanner errors are checked below.
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	section := ""

	for scanner.Scan() {
		rawLine := scanner.Text()
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
		if indent == 0 && strings.HasSuffix(line, ":") {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		if indent == 0 {
			section = ""
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = configStringValue(value)

		if section == "" && key == "key_prefix" {
			setString(&config.TicketKeyPrefix, value)
			continue
		}

		if section == "tickets" && key == "key_prefix" {
			setString(&config.TicketKeyPrefix, value)
			continue
		}

		if section == "runtime" {
			setRuntimeConfigValue(&config.Runtime, key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return WorkspaceConfig{}, fmt.Errorf("read workspace config: %w", err)
	}

	return config, nil
}

func setRuntimeConfigValue(config *RuntimeConfig, key, value string) {
	switch key {
	case "image":
		setString(&config.Image, value)
	case "dockerfile":
		setString(&config.DockerfilePath, value)
	case "workdir":
		setString(&config.Workdir, value)
	case "artifact_dir":
		setString(&config.ArtifactDir, value)
	case "validation_command":
		setString(&config.ValidationCommand, value)
	}
}

func setString(target *string, value string) {
	if value != "" {
		*target = value
	}
}

func configStringValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}
