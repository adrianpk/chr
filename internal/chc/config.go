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
	Notifications   NotificationConfig
}

type RuntimeConfig struct {
	Image             string
	DockerfilePath    string
	Workdir           string
	ArtifactDir       string
	ValidationCommand string
}

type NotificationConfig struct {
	Telegram TelegramNotificationConfig
}

type TelegramNotificationConfig struct {
	Enabled     bool
	BotTokenEnv string
	ChatID      string
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

func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{}
}

func LoadWorkspaceConfig(projectDir string) (WorkspaceConfig, error) {
	config := WorkspaceConfig{
		TicketKeyPrefix: DefaultTicketKeyPrefix,
		Runtime:         DefaultRuntimeConfig(),
		Notifications:   DefaultNotificationConfig(),
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

	err = readWorkspaceConfig(file, &config)
	if err != nil {
		return WorkspaceConfig{}, err
	}

	return config, nil
}

func readWorkspaceConfig(file *os.File, config *WorkspaceConfig) error {
	scanner := bufio.NewScanner(file)
	section := ""

	for scanner.Scan() {
		key, value, updatedSection, ok := parseWorkspaceConfigLine(scanner.Text(), section)
		section = updatedSection

		if !ok {
			continue
		}

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

		if section == "notifications" {
			setNotificationConfigValue(&config.Notifications, key, value)
		}
	}

	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("read workspace config: %w", err)
	}

	return nil
}

func parseWorkspaceConfigLine(rawLine, section string) (string, string, string, bool) {
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", section, false
	}

	indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
	if indent == 0 && strings.HasSuffix(line, ":") {
		return "", "", strings.TrimSuffix(line, ":"), false
	}

	if indent == 0 {
		section = ""
	}

	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", section, false
	}

	return strings.TrimSpace(key), configStringValue(value), section, true
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

func setNotificationConfigValue(config *NotificationConfig, key, value string) {
	switch key {
	case "telegram_enabled":
		config.Telegram.Enabled = configBoolValue(value)
	case "telegram_bot_token_env":
		setString(&config.Telegram.BotTokenEnv, value)
	case "telegram_chat_id":
		setString(&config.Telegram.ChatID, value)
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

func configBoolValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "1", "on":
		return true
	default:
		return false
	}
}
