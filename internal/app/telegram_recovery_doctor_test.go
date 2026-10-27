package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramRecoveryDoctor uses actual HTTP identity/webhook responses and
// OS exclusion, never polling, sending, rebinding identity or starting a model.
func TestTelegramRecoveryDoctor(t *testing.T) {
	for _, test := range []telegramRecoveryDoctorCase{
		{"ready", 12345, "", false, "remote_polling: unchecked"},
		{"replacement", 67890, "", false, "bot_identity_changed"},
		{"webhook", 12345, "https://private.invalid/hook", false, "webhook_conflict"},
		{"local owner", 12345, "", true, "local_receiver_conflict"},
	} {
		t.Run(test.name, func(t *testing.T) { checkTelegramRecoveryDoctor(t, test) })
	}
}

type telegramRecoveryDoctorCase struct {
	name    string
	botID   int64
	webhook string
	locked  bool
	want    string
}

func checkTelegramRecoveryDoctor(t *testing.T, test telegramRecoveryDoctorCase) {
	t.Helper()

	receiver, _ := telegramAppTestReceiver(t, telegramcore.ReceiverAPI{})
	if err := telegramcore.BindReceiverIdentity(receiver.env, 12345); err != nil {
		t.Fatal(err)
	}

	if test.locked {
		holdTelegramRecoveryDoctorLock(t, receiver.env)
	}

	server := telegramRecoveryDoctorServer(t, test.botID, test.webhook)
	defer server.Close()

	var output bytes.Buffer

	root := filepath.Dir(filepath.Dir(receiver.env.ConfigDir()))

	command := (commandFactory{deps: testAppDependencies(), io: IO{Stdin: strings.NewReader(""), Stdout: &output, Stderr: &output}}).rootCommand()
	command.Command("telegram").Command("doctor").Action = func(ctx context.Context, cmd *cli.Command) error {
		return (commandFactory{deps: testAppDependencies(), io: IO{Stdout: &output}}).runTelegramDoctor(ctx, cmd, telegramcore.ReceiverAPI{Client: server.Client(), BaseURL: server.URL})
	}

	err := command.Run(t.Context(), []string{"chr", "telegram", "doctor", "--dev-root", root, "--check-connection"})

	checkTelegramRecoveryDoctorResult(t, test, receiver, output.String(), err)
}

func telegramRecoveryDoctorServer(t *testing.T, botID int64, webhook string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch filepath.Base(r.URL.Path) {
		case "getMe":
			if _, err := fmt.Fprintf(w, `{"ok":true,"result":{"id":%d,"is_bot":true}}`, botID); err != nil {
				t.Error(err)
			}
		case "getWebhookInfo":
			if _, err := fmt.Fprintf(w, `{"ok":true,"result":{"url":%q}}`, webhook); err != nil {
				t.Error(err)
			}
		default:
			t.Error("doctor attempted an active operation", filepath.Base(r.URL.Path))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

// TestTelegramRecoveryFailures separates correction-required exits from
// recoverable transport/process failure and never echoes an underlying secret.
func TestTelegramRecoveryFailures(t *testing.T) {
	for _, test := range []struct {
		name    string
		failure error
		code    int
		want    string
	}{
		{"configuration", &telegramServiceFailure{code: "configuration_invalid", permanent: true}, 78, "configuration_invalid"},
		{"credentials", &telegramcore.APIError{Code: "credentials_or_access", Permanent: true}, 78, "credentials_or_access"},
		{"ownership", &telegramcore.APIError{Code: "receiver_conflict", Permanent: true}, 78, "receiver_conflict"},
		{"network", &telegramcore.APIError{Code: "transport_failure"}, 1, "transport_failure"},
		{"process", errors.New("secret process body"), 1, "local_process_failure"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := telegramCommandFailure(test.failure)
			if ExitCode(err) != test.code || !strings.Contains(err.Error(), test.want) || strings.Contains(err.Error(), "secret") {
				t.Fatal("unsafe restart classification", err)
			}
		})
	}
}

// TestTelegramRecoveryProviders reports the selected provider's local
// prerequisite, including explicitly allowed projects absent from the registry.
func TestTelegramRecoveryProviders(t *testing.T) {
	receiver, config := telegramAppTestReceiver(t, telegramcore.ReceiverAPI{})
	addTelegramAppTestProjects(t, receiver.env)

	config.AllowedUsers[0].ProjectIDs = append(config.AllowedUsers[0].ProjectIDs, "missing-project")
	if err := telegramcore.SaveReceiverConfig(receiver.env, config, true); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CHR_CONVERSATION_GENERATION_BACKEND", "openai")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("CHR_CODEX_COMMAND", filepath.Join(t.TempDir(), "missing-command"))

	var output bytes.Buffer
	if err := run(t.Context(), []string{"telegram", "doctor", "--dev-root", filepath.Dir(filepath.Dir(receiver.env.ConfigDir()))}, &output, &output); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "project: missing-project registered=false") {
		t.Fatal("missing registry entry was hidden", output.String())
	}

	if !strings.Contains(output.String(), "provider=openai local_ready=false") || !strings.Contains(output.String(), "provider=codex local_ready=false") {
		t.Fatal("configured provider readiness missing", output.String())
	}

	if _, err := os.Stat(filepath.Join(receiver.env.StateDir(), "telegram", "bot-id")); !os.IsNotExist(err) {
		t.Fatal("doctor started receiver", err)
	}
}

func checkTelegramRecoveryDoctorResult(t *testing.T, test telegramRecoveryDoctorCase, receiver telegramReceiver, result string, err error) {
	t.Helper()

	if err != nil {
		result += err.Error()
	}

	if (err == nil) != (test.name == "ready") || !strings.Contains(result, test.want) {
		t.Fatal("incorrect connection diagnosis", err, result)
	}

	if strings.Contains(result, "private.invalid") || strings.Contains(result, receiver.token) {
		t.Fatal("doctor exposed remote secret")
	}

	id, err := telegramcore.BoundBotID(receiver.env)
	if err != nil || id != 12345 {
		t.Fatal("doctor rebound receiver", id, err)
	}
}

func holdTelegramRecoveryDoctorLock(t *testing.T, env core.UserEnvironment) {
	t.Helper()

	lock, err := telegramcore.AcquireConfigurationLock(env)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	})
}
