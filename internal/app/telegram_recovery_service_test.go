package app

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTelegramRecoveryService installs a real private unit without opening the
// work database, activating a service or leaking environment credentials.
func TestTelegramRecoveryService(t *testing.T) {
	root := filepath.Join(t.TempDir(), "space % $ fixture")
	t.Setenv("RECOVERY_SERVICE_TOKEN", "12345:example_test_token_123456789")

	args := []string{"telegram", "setup", "--dev-root", root, "--user-id", "100001", "--operator-id", "maintainer", "--project-id", "example-project", "--bot-token-env", "RECOVERY_SERVICE_TOKEN", "--install-service"}

	var output bytes.Buffer
	if err := run(t.Context(), args, &output, &output); err != nil {
		t.Fatal(err)
	}

	unitPath := filepath.Join(root, "config", "systemd", "user", "chr-telegram.service")

	data, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(unitPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatal("unit is not private", err)
	}

	for _, part := range []string{" telegram serve --dev-root ", "RestartPreventExitStatus=78", "TimeoutStopSec=35s", "EnvironmentFile=", "XDG_DATA_HOME=", "space %% $$ fixture"} {
		if !strings.Contains(string(data), part) {
			t.Fatal("unit lost explicit installation contract", part, string(data))
		}
	}

	if strings.Contains(string(data)+output.String(), os.Getenv("RECOVERY_SERVICE_TOKEN")) {
		t.Fatal("unit exposed token")
	}

	if _, err := os.Stat(filepath.Join(root, "data", "chr", "core.db")); !os.IsNotExist(err) {
		t.Fatal("installation initialized work state", err)
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(unitPath), "default.target.wants")); !os.IsNotExist(err) {
		t.Fatal("installation enabled service", err)
	}

	checkTelegramRecoveryServiceReplacement(t, args, unitPath)
	checkTelegramRecoveryUnitSyntax(t, unitPath)
}

func checkTelegramRecoveryServiceReplacement(t *testing.T, args []string, path string) {
	t.Helper()

	var output bytes.Buffer
	if err := run(t.Context(), args, &output, &output); err != nil {
		t.Fatal("idempotent install", err)
	}

	if err := os.WriteFile(path, []byte("operator unit\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := run(t.Context(), args, &output, &output); err == nil {
		t.Fatal("differing unit was replaced implicitly")
	}

	if err := run(t.Context(), append(args, "--replace"), &output, &output); err != nil {
		t.Fatal(err)
	}
}

func checkTelegramRecoveryUnitSyntax(t *testing.T, path string) {
	t.Helper()

	verifier, err := exec.LookPath("systemd-analyze")
	if err != nil {
		t.Skip("systemd unit verifier is unavailable")
	}

	command := exec.CommandContext(t.Context(), verifier, "verify", "--man=no", path)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal("systemd rejected installed unit", err, string(output))
	}
}

// TestTelegramRecoveryHelp exposes the recovery surface in installed command
// help and keeps malformed retry selection outside database initialization.
func TestTelegramRecoveryHelp(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"transport", []string{"telegram", "--help"}, "five attempts"},
		{"service", []string{"telegram", "setup", "--help"}, "--install-service"},
		{"retry", []string{"telegram", "retry", "--help"}, "--delivery"},
		{"status", []string{"telegram", "status", "--help"}, "without migration"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := run(t.Context(), test.args, &output, &output); err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(output.String(), test.want) {
				t.Fatal("missing recovery help", output.String())
			}
		})
	}

	root := t.TempDir()

	for _, flags := range [][]string{{}, {"--poll", "--delivery", "invalid"}, {"extra"}} {
		var output bytes.Buffer

		args := append([]string{"telegram", "retry", "--dev-root", root}, flags...)
		if err := run(t.Context(), args, &output, &output); err == nil {
			t.Fatal("invalid retry selection accepted")
		}
	}

	if _, err := os.Stat(filepath.Join(root, "data")); !os.IsNotExist(err) {
		t.Fatal("invalid retry created state", err)
	}
}

// TestTelegramRecoveryConfigExit exercises the serve entrypoint with invalid
// local configuration: it requires correction and never opens a work store.
func TestTelegramRecoveryConfigExit(t *testing.T) {
	root := t.TempDir()

	var output bytes.Buffer

	err := run(t.Context(), []string{"telegram", "serve", "--dev-root", root}, &output, &output)
	if err == nil || ExitCode(err) != 78 || !strings.Contains(err.Error(), "configuration_invalid") {
		t.Fatal("invalid configuration would restart", err)
	}

	if _, err := os.Stat(filepath.Join(root, "data", "chr", "core.db")); !os.IsNotExist(err) {
		t.Fatal("invalid configuration opened work store", err)
	}
}
