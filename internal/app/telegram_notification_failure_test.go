package app

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core/database"
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func installTelegramNotificationSender(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	previous := notifyTelegramSender
	notifyTelegramSender = telegramcore.Sender{Client: server.Client(), BaseURL: server.URL}

	t.Cleanup(func() { notifyTelegramSender = previous })
}

// Real HTTP distinguishes a definite upload failure from a lost acknowledgement
// or malformed success. Reopening cannot re-send or claim delivery for any case.
func TestTelegramPresentationFailureNotificationHTTP(t *testing.T) {
	for _, failure := range []telegramNotificationHTTPFailure{
		{"lost acknowledgement", 0, "", true},
		{"server failure", 503, `{"ok":false,"error_code":503}`, true},
		{"missing file acknowledgement", 200, `{"ok":true,"result":{"message_id":71}}`, true},
		{"access rejected", 403, `{"ok":false,"error_code":403}`, false},
	} {
		t.Run(failure.name, func(t *testing.T) { checkTelegramNotificationHTTPFailure(t, failure) })
	}
}

type telegramNotificationHTTPFailure struct {
	name      string
	status    int
	body      string
	uncertain bool
}

func checkTelegramNotificationHTTPFailure(t *testing.T, failure telegramNotificationHTTPFailure) {
	t.Helper()

	repository, input, project, env := newTelegramNotificationAppInput(t)

	var calls atomic.Int32

	installTelegramNotificationSender(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)

		if failure.status == 0 {
			closeTelegramDocumentConnection(t, w)

			return
		}

		w.WriteHeader(failure.status)

		if _, err := w.Write([]byte(failure.body)); err != nil {
			t.Error(err)
		}
	})

	if err := repository.CreateTelegram(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	if _, err := dispatchTelegramNotification(t.Context(), repository, input.Record, project, input.Target); err == nil {
		t.Fatal("failed upload claimed success")
	}

	if err := repository.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := notifycore.NewRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := reopened.Close(); err != nil {
			t.Error(err)
		}
	})
	checkTelegramNotificationFailed(t, reopened, input, failure.uncertain)

	if _, err := dispatchTelegramNotification(t.Context(), reopened, input.Record, project, input.Target); err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 {
		t.Fatal("reopen silently retried a notification", calls.Load())
	}
}

func checkTelegramNotificationFailed(t *testing.T, repository notifycore.Repository, input notifycore.TelegramInput, uncertain bool) {
	t.Helper()

	lookup, err := repository.ReadTelegram(t.Context(), input.Record.Event.ID, input.Target)

	saved, frozen := lookup.Delivery, lookup.Frozen
	if err != nil || !frozen || saved.Status != notifycore.StatusFailed || saved.Attempts != 1 || saved.Uncertain != uncertain || saved.Acknowledgement.MessageID != 0 {
		t.Fatal("failure invented delivery or allowance", saved, err)
	}

	if saved.Document == nil || !bytes.Equal(saved.Document.Payload, input.Document.Payload) {
		t.Fatal("failure replaced complete export")
	}
}

// Configuration, token and root changes block sends, including a change after
// definite format rejection. No stale destination receives the fallback upload.
func TestTelegramPresentationFailureNotificationRevocation(t *testing.T) {
	for _, change := range []string{"destination", "disabled", "token", "root", "before fallback"} {
		t.Run(change, func(t *testing.T) { checkTelegramNotificationRevocation(t, change) })
	}
}

func checkTelegramNotificationRevocation(t *testing.T, change string) {
	t.Helper()

	repository, input, project, _ := newTelegramNotificationAppInput(t)

	var calls atomic.Int32

	installTelegramNotificationSender(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		t.Setenv("CHR_TELEGRAM_BOT_TOKEN", "replacement-token")
		w.WriteHeader(http.StatusBadRequest)

		if _, err := w.Write([]byte(`{"ok":false,"error_code":400,"description":"can't parse entities"}`)); err != nil {
			t.Error(err)
		}
	})

	if err := repository.CreateTelegram(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	if change != "before fallback" {
		revokeTelegramNotification(t, project, change)
	}

	if _, err := dispatchTelegramNotification(t.Context(), repository, input.Record, project, input.Target); err == nil {
		t.Fatal("revoked target accepted")
	}

	want := int32(0)
	if change == "before fallback" {
		want = 1
	}

	if calls.Load() != want {
		t.Fatal("revoked target reached HTTP", change, calls.Load())
	}

	saved, err := repository.Get(t.Context(), input.Record.Event.ID)
	if err != nil || saved.Event.Status != notifycore.StatusFailed || saved.Event.FailureReason != "notification_target_unavailable" {
		t.Fatal("revocation claimed another outcome", saved, err)
	}
}

func revokeTelegramNotification(t *testing.T, project commandProjectContext, change string) {
	t.Helper()

	if change == "token" {
		t.Setenv("CHR_TELEGRAM_BOT_TOKEN", "replacement-token")

		return
	}

	if change == "root" {
		moved := filepath.Join(t.TempDir(), "moved")
		if err := os.Rename(project.ProjectDir, moved); err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(moved, project.ProjectDir); err != nil {
			t.Fatal(err)
		}

		return
	}

	enabled, destination := true, "@replacement"
	if change == "disabled" {
		enabled, destination = false, "123456"
	}

	writeCLIWorkspaceConfig(t, project.ProjectDir, fmt.Sprintf("notifications:\n  telegram_enabled: %t\n  telegram_bot_token_env: CHR_TELEGRAM_BOT_TOKEN\n  telegram_chat_id: %q\n", enabled, destination))
}

// Future versions and missing images cannot escape the retained reader into
// HTTP. Definite local failure is saved without a false transport receipt.
func TestTelegramPresentationFailureNotificationBeforeIO(t *testing.T) {
	for _, item := range []struct{ name, query string }{
		{"future manifest", `UPDATE notification_events SET telegram_presentation=json_set(telegram_presentation,'$.version',999)`},
		{"missing image", `UPDATE telegram_notification_documents SET payload=NULL,purged_at='2030-01-01T00:00:00Z'`},
		{"missing manifest", `UPDATE notification_events SET telegram_presentation=NULL`},
		{"partial version zero", `UPDATE notification_events SET telegram_presentation='{"version":0}',telegram_presentation_hash=NULL,telegram_document_id=NULL`},
	} {
		t.Run(item.name, func(t *testing.T) { checkTelegramNotificationBeforeIO(t, item.query) })
	}
}

func checkTelegramNotificationBeforeIO(t *testing.T, query string) {
	t.Helper()

	repository, input, project, env := newTelegramNotificationAppInput(t)

	var calls atomic.Int32

	installTelegramNotificationSender(t, func(http.ResponseWriter, *http.Request) { calls.Add(1) })

	if err := repository.CreateTelegram(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", database.UserDatabasePath(env))
	if err != nil {
		t.Fatal(err)
	}

	_, writeErr := db.ExecContext(t.Context(), query)

	closeErr := db.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal(writeErr, closeErr)
	}

	if _, err := dispatchTelegramNotification(t.Context(), repository, input.Record, project, input.Target); err == nil {
		t.Fatal("invalid retained image dispatched")
	}

	if calls.Load() != 0 {
		t.Fatal("invalid image reached HTTP")
	}

	saved, err := repository.Get(t.Context(), input.Record.Event.ID)
	if err != nil || saved.Event.Status != notifycore.StatusFailed || saved.Event.FailureReason != "notification_document_unavailable" {
		t.Fatal("invalid image claimed delivery", saved, err)
	}
}

// Concurrent callers may observe pending state together, but only the original
// event's persistent claim can reach HTTP and retain its acknowledgement.
func TestTelegramPresentationNotificationConcurrent(t *testing.T) {
	repository, input, project, _ := newTelegramNotificationAppInput(t)

	var calls atomic.Int32

	installTelegramNotificationSender(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)

		if _, err := w.Write([]byte(`{"ok":true,"result":{"message_id":71,"document":{"file_id":"original-file"}}}`)); err != nil {
			t.Error(err)
		}
	})

	if err := repository.CreateTelegram(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	ready, results := make(chan struct{}), make(chan error, 2)

	for range 2 {
		go func() {
			<-ready

			_, err := dispatchTelegramNotification(t.Context(), repository, input.Record, project, input.Target)
			results <- err
		}()
	}

	close(ready)

	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}

	checkTelegramConcurrentNotification(t, repository, input, calls.Load())
}

func checkTelegramConcurrentNotification(t *testing.T, repository notifycore.Repository, input notifycore.TelegramInput, calls int32) {
	t.Helper()

	lookup2, err := repository.ReadTelegram(t.Context(), input.Record.Event.ID, input.Target)

	saved, frozen := lookup2.Delivery, lookup2.Frozen
	if err != nil || !frozen || saved.Attempts != 1 || saved.Status != notifycore.StatusDelivered || calls != 1 {
		t.Fatal("concurrent dispatch duplicated delivery", saved, calls, err)
	}
}

// Oversized complete input keeps a failed audit record without sending a
// truncated export or surfacing a transport failure as failed domain work.
func TestTelegramPresentationFailureNotificationPreparation(t *testing.T) {
	repository, input, project, _ := newTelegramNotificationAppInput(t)
	input.Record.Event.Payload.Detail = strings.Repeat("界", 1<<20)

	var calls atomic.Int32

	installTelegramNotificationSender(t, func(http.ResponseWriter, *http.Request) { calls.Add(1) })

	config, err := workspaceconfig.Load(project.ProjectDir)
	if err != nil {
		t.Fatal(err)
	}

	_, prepareErr := createTelegramNotification(t.Context(), repository, input.Record, project, config.Notifications.Telegram)
	if prepareErr == nil {
		t.Fatal("oversized complete event admitted")
	}

	if err := workflowNotificationFailure(t.Context(), repository, input.Record.Event.ID, prepareErr); err != nil {
		t.Fatal("failed presentation changed domain outcome", err)
	}

	lookup3, err := repository.ReadTelegram(t.Context(), input.Record.Event.ID, input.Target)

	saved, frozen := lookup3.Delivery, lookup3.Frozen
	if err != nil || frozen || saved.Status != notifycore.StatusFailed || calls.Load() != 0 {
		t.Fatal("failed presentation left a partial image or claimed delivery", saved, frozen, err)
	}
}
