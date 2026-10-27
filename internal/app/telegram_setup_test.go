package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramReceiverSetup proves setup uses private environment references and never echoes tokens.
func TestTelegramReceiverSetup(t *testing.T) {
	root := t.TempDir()
	token := "12345:example_test_token_123456789"
	t.Setenv("EXAMPLE_BOT_TOKEN", token)

	args := []string{"telegram", "setup", "--dev-root", root, "--user-id", "100001", "--operator-id", "maintainer", "--project-id", "example-project", "--bot-token-env", "EXAMPLE_BOT_TOKEN"}

	var output bytes.Buffer
	if err := run(t.Context(), args, &output, &output); err != nil {
		t.Fatal(err)
	}

	if err := run(t.Context(), []string{"telegram", "status", "--dev-root", root}, &output, &output); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(root, "config", "chr", "telegram.yml"))
	if err != nil || strings.Contains(string(data)+output.String(), token) {
		t.Fatal("token exposed", err)
	}

	if _, err := os.Stat(filepath.Join(root, "data", "chr", "core.db")); !os.IsNotExist(err) {
		t.Fatal("setup or status initialized work state")
	}

	args = args[:len(args)-2]

	args = append(args, "--replace")
	if err := run(t.Context(), args, &output, &output); err == nil || !strings.Contains(err.Error(), "terminal") {
		t.Fatal("unmasked nonterminal input accepted", err)
	}
}

// TestTelegramReceiverDoctor reports unavailable project runtime without calling a model or changing the registry.
func TestTelegramReceiverDoctor(t *testing.T) {
	receiver, _ := telegramAppTestReceiver(t, telegramcore.ReceiverAPI{})
	addTelegramAppTestProjects(t, receiver.env)

	var output bytes.Buffer

	cmd := (commandFactory{deps: testAppDependencies(), io: IO{Stdin: strings.NewReader(""), Stdout: &output, Stderr: &output}}).rootCommand()
	// The development root is the ancestor above config/chr.
	root := filepath.Dir(filepath.Dir(receiver.env.ConfigDir()))
	if err := cmd.Run(t.Context(), []string{"chr", "telegram", "doctor", "--dev-root", root}); err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{"runtime_ready=false", "model_connection: unchecked", "connection: unchecked"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatal("missing readiness diagnostic", expected)
		}
	}

	if _, err := telegramcore.BoundBotID(receiver.env); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("local doctor bound a bot", err)
	}
}
