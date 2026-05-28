package chc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkspaceConfigUsesDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != DefaultTicketKeyPrefix {
		t.Fatalf("TicketKeyPrefix = %q, want %q", config.TicketKeyPrefix, DefaultTicketKeyPrefix)
	}

	assertDefaultRuntimeConfig(t, config.Runtime)

	if config.Notifications != DefaultNotificationConfig() {
		t.Fatalf("Notifications = %#v, want %#v", config.Notifications, DefaultNotificationConfig())
	}
}

func TestLoadWorkspaceConfigParsesTicketAndRuntimeConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `version: 1
workspace_root: .chc
tickets:
  key_prefix: PRJ
runtime:
  image: custom-agent:latest
  dockerfile: .chc/user/runtime/custom.Dockerfile
  workdir: /repo
  artifact_dir: .chc/user/runtime/custom-runs
  validation_command: go test ./...
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != "PRJ" {
		t.Fatalf("TicketKeyPrefix = %q, want PRJ", config.TicketKeyPrefix)
	}

	want := RuntimeConfig{
		Image:             "custom-agent:latest",
		DockerfilePath:    ".chc/user/runtime/custom.Dockerfile",
		Workdir:           "/repo",
		ArtifactDir:       ".chc/user/runtime/custom-runs",
		ValidationCommand: "go test ./...",
	}
	if config.Runtime != want {
		t.Fatalf("Runtime = %#v, want %#v", config.Runtime, want)
	}
}

func TestLoadWorkspaceConfigParsesNotificationConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	want := NotificationConfig{
		Telegram: TelegramNotificationConfig{
			Enabled:     true,
			BotTokenEnv: "CHC_TELEGRAM_BOT_TOKEN",
			ChatID:      "123456",
		},
	}
	if config.Notifications != want {
		t.Fatalf("Notifications = %#v, want %#v", config.Notifications, want)
	}
}

func TestLoadWorkspaceConfigKeepsNotificationsDisabledByDefault(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `notifications:
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.Notifications.Telegram.Enabled {
		t.Fatal("Telegram notifications are enabled by default")
	}
}

func TestLoadWorkspaceConfigKeepsRuntimeDefaultsForMissingKeys(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `runtime:
  image: custom-agent:latest
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.Runtime.Image != "custom-agent:latest" {
		t.Fatalf("Runtime.Image = %q, want custom-agent:latest", config.Runtime.Image)
	}

	if config.Runtime.DockerfilePath != DefaultRuntimeDockerfilePath {
		t.Fatalf("Runtime.DockerfilePath = %q, want %q", config.Runtime.DockerfilePath, DefaultRuntimeDockerfilePath)
	}
}

func TestLoadWorkspaceConfigSupportsLegacyTopLevelTicketKeyPrefix(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, "key_prefix: LEGACY\n")

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != "LEGACY" {
		t.Fatalf("TicketKeyPrefix = %q, want LEGACY", config.TicketKeyPrefix)
	}
}

func TestValidateNotificationConfigAcceptsDisabledConfig(t *testing.T) {
	err := ValidateNotificationConfig(NotificationConfig{})
	if err != nil {
		t.Fatalf("ValidateNotificationConfig returned error: %v", err)
	}
}

func TestValidateNotificationConfigAcceptsEnabledTelegramConfig(t *testing.T) {
	config := NotificationConfig{
		Telegram: TelegramNotificationConfig{
			Enabled:     true,
			BotTokenEnv: "CHC_TELEGRAM_BOT_TOKEN",
			ChatID:      "123456",
		},
	}

	err := ValidateNotificationConfig(config)
	if err != nil {
		t.Fatalf("ValidateNotificationConfig returned error: %v", err)
	}
}

func TestValidateNotificationConfigRejectsIncompleteTelegramConfig(t *testing.T) {
	for _, test := range []struct {
		name    string
		config  NotificationConfig
		wantErr string
	}{
		{
			name: "token env",
			config: NotificationConfig{
				Telegram: TelegramNotificationConfig{
					Enabled: true,
					ChatID:  "123456",
				},
			},
			wantErr: "telegram bot token environment variable is required",
		},
		{
			name: "chat id",
			config: NotificationConfig{
				Telegram: TelegramNotificationConfig{
					Enabled:     true,
					BotTokenEnv: "CHC_TELEGRAM_BOT_TOKEN",
				},
			},
			wantErr: "telegram chat id is required",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateNotificationConfig(test.config)
			if err == nil {
				t.Fatal("ValidateNotificationConfig returned nil error")
			}

			if err.Error() != test.wantErr+" when telegram notifications are enabled" {
				t.Fatalf("ValidateNotificationConfig error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func assertDefaultRuntimeConfig(t *testing.T, config RuntimeConfig) {
	t.Helper()

	if config != DefaultRuntimeConfig() {
		t.Fatalf("Runtime = %#v, want %#v", config, DefaultRuntimeConfig())
	}
}

func writeWorkspaceConfig(t *testing.T, projectDir, content string) {
	t.Helper()

	path := UserPath(projectDir, "config.yml")

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
}
