package core

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func telegramReceiverTestConfig() TelegramReceiverConfig {
	return TelegramReceiverConfig{
		Enabled: true, BotTokenEnv: "EXAMPLE_BOT_TOKEN", PollTimeoutSeconds: 30,
		AllowedUsers: []TelegramAllowedUser{{TelegramUserID: 100001, OperatorID: "maintainer", ProjectIDs: []string{"example-project"}}},
	}
}

func telegramReceiverTestEnv(t *testing.T) UserEnvironment {
	t.Helper()

	env, err := NewUserEnvironment(UserEnvironmentOptions{DevMode: true, DevRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	return env
}

// TestTelegramReceiverConfig protects setup preservation, private modes and strict authority parsing.
func TestTelegramReceiverConfig(t *testing.T) {
	env := telegramReceiverTestEnv(t)

	config := telegramReceiverTestConfig()
	if err := SaveTelegramReceiverConfig(env, config, false); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(env.ConfigDir(), "telegram.yml")

	prior, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	config.AllowedUsers[0].OperatorID = "replacement"
	if err := SaveTelegramReceiverConfig(env, config, false); err == nil {
		t.Fatal("implicit replacement accepted")
	}

	current, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(prior, current) {
		t.Fatal("configuration changed after denied replacement", err)
	}

	if err := SaveTelegramReceiverConfig(env, config, true); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadTelegramReceiverConfig(env)
	if err != nil || loaded.Digest() != config.Digest() {
		t.Fatal("round trip changed configuration", err)
	}

	assertTelegramPrivateMode(t, path, 0o600)
	assertTelegramPrivateMode(t, env.ConfigDir(), 0o700)
}

// TestTelegramReceiverConfigErrors rejects unsafe YAML and invalid numeric authority.
func TestTelegramReceiverConfigErrors(t *testing.T) {
	body := string(renderTelegramReceiverConfig(telegramReceiverTestConfig()))

	cases := map[string]string{
		"unknown":     body + "  token: forbidden\n",
		"duplicate":   body + "  enabled: true\n",
		"bad indent":  strings.Replace(body, "      operator_id:", "     operator_id:", 1),
		"alias":       strings.Replace(body, "maintainer", "*authority", 1),
		"no projects": strings.Replace(body, "        - example-project\n", "", 1),
		"zero sender": strings.Replace(body, "100001", "0", 1),
		"long poll":   strings.Replace(body, "seconds: 30", "seconds: 51", 1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			config, err := parseTelegramReceiverConfig([]byte(data))
			if err == nil && config.Validate() == nil {
				t.Fatal("unsafe configuration accepted")
			}
		})
	}
}

// TestTelegramReceiverToken keeps credentials private and rejects implicit replacement.
func TestTelegramReceiverToken(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	config := telegramReceiverTestConfig()

	token := "12345:example_test_token_123456789"
	if err := SaveTelegramReceiverToken(env, config, token, false); err != nil {
		t.Fatal(err)
	}

	got, err := TelegramReceiverToken(env, config)
	if err != nil || got != token {
		t.Fatal("private token unavailable", err)
	}

	if err := SaveTelegramReceiverToken(env, config, token+"x", false); err == nil {
		t.Fatal("credential replaced implicitly")
	}

	assertTelegramPrivateMode(t, filepath.Join(env.ConfigDir(), "telegram.env"), 0o600)
	t.Setenv(config.BotTokenEnv, "malformed-secret")

	if _, err := TelegramReceiverToken(env, config); err == nil || strings.Contains(err.Error(), "malformed-secret") {
		t.Fatal("malformed token accepted or exposed")
	}
}

func assertTelegramPrivateMode(t *testing.T, path string, mode os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != mode {
		t.Fatalf("private mode mismatch for %s: %v", path, err)
	}
}
