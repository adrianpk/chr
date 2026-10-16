package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTelegramRecoveryEnvironment preserves private model configuration during
// token replacement and rejects ambiguous assignments without leaking values.
func TestTelegramRecoveryEnvironment(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	config := telegramReceiverTestConfig()
	t.Setenv(config.BotTokenEnv, "")

	first := "12345:example_test_token_123456789"
	second := "67890:example_test_token_987654321"
	path := filepath.Join(env.ConfigDir(), "telegram.env")

	original := "OPENAI_API_KEY=private-model-key\nCHR_CONVERSATION_GENERATION_BACKEND=openai\n" + config.BotTokenEnv + "=" + first + "\n"
	if err := writeTelegramPrivateFile(path, []byte(original)); err != nil {
		t.Fatal(err)
	}

	if token, err := TelegramReceiverToken(env, config); err != nil || token != first {
		t.Fatal("multi-assignment file rejected", err)
	}

	if err := SaveTelegramReceiverToken(env, config, second, false); err == nil {
		t.Fatal("credential replaced implicitly")
	}

	if err := SaveTelegramReceiverToken(env, config, second, true); err != nil {
		t.Fatal(err)
	}

	checkTelegramRecoveryEnvironmentReplaced(t, env, config, path, first, second)

	checkTelegramRecoveryEnvironmentRejected(t, env, config, path, first, second)
}

func checkTelegramRecoveryEnvironmentRejected(t *testing.T, env UserEnvironment, config TelegramReceiverConfig, path, first, second string) {
	t.Helper()

	for _, test := range []struct{ name, body string }{
		{"duplicate", config.BotTokenEnv + "=" + first + "\n" + config.BotTokenEnv + "=" + second + "\n"},
		{"invalid", "invalid-name=secret-value\n"},
		{"oversize", strings.Repeat("x", 65537)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(test.body), 0o600); err != nil {
				t.Fatal(err)
			}

			if _, err := TelegramReceiverToken(env, config); err == nil || strings.Contains(err.Error(), "secret-value") {
				t.Fatal("unsafe private configuration accepted", err)
			}
		})
	}
}

func checkTelegramRecoveryEnvironmentReplaced(t *testing.T, env UserEnvironment, config TelegramReceiverConfig, path, first, second string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "OPENAI_API_KEY=private-model-key\n") || strings.Contains(string(data), first) {
		t.Fatal("replacement lost unrelated private environment", err)
	}

	if token, err := TelegramReceiverToken(env, config); err != nil || token != second {
		t.Fatal("replacement unavailable", err)
	}
}
