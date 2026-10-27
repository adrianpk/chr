package app

import (
	"bytes"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramRecoveryOperator exercises CLI poll/delivery retries and read-only
// status against the same real store, retaining counters, payloads and identity.
func TestTelegramRecoveryOperator(t *testing.T) {
	receiver, config := telegramAppTestReceiver(t, telegramcore.TelegramReceiverAPI{})
	if err := telegramcore.BindTelegramReceiverIdentity(receiver.env, receiver.botID); err != nil {
		t.Fatal(err)
	}

	update := telegramcore.TelegramReceivedUpdate{UpdateID: 1, Digest: strings.Repeat("a", 64), Kind: "message", Peer: telegramcore.TelegramPeerIdentity{BotID: 12345, ChatID: 100001, UserID: 100001}, MessageID: 1, ChatType: "private", Text: "/help"}

	now := time.Now()
	if err := receiver.store.RetainBatch(t.Context(), receiver.botID, []telegramcore.TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	if err := receiver.store.QueueReadResponse(t.Context(), update, "secret retained result", config.Digest(), now); err != nil {
		t.Fatal(err)
	}

	seedTelegramRecoveryOperator(t, receiver)

	diagnostics, err := telegramcore.InspectTelegramReceiverDiagnostics(t.Context(), receiver.env, receiver.botID)
	if err != nil || len(diagnostics.Deliveries) != 1 {
		t.Fatal(diagnostics, err)
	}

	id := diagnostics.Deliveries[0].ID
	root := filepath.Dir(filepath.Dir(receiver.env.ConfigDir()))

	var output bytes.Buffer
	if err := run(t.Context(), []string{"telegram", "status", "--dev-root", root}, &output, &output); err != nil {
		t.Fatal(err)
	}

	for _, part := range []string{"installed_build:", "lifetime_poll_attempts: 11", "pending_deliveries: 1", "cycle=3 lifetime_attempts=7", id} {
		if !strings.Contains(output.String(), part) {
			t.Fatal("missing retained diagnostic", part, output.String())
		}
	}

	if strings.Contains(output.String(), "secret retained result") {
		t.Fatal("status leaked body")
	}

	checkTelegramRecoveryOperatorRetries(t, receiver, root, id)
}

func seedTelegramRecoveryOperator(t *testing.T, receiver telegramReceiver) {
	t.Helper()

	db, err := sql.Open("sqlite", database.UserDatabasePath(receiver.env))
	if err != nil {
		t.Fatal(err)
	}

	_, pollErr := db.ExecContext(t.Context(), `UPDATE telegram_receivers SET health='blocked',retry_attempt=5,total_retry_attempts=11,retry_cycle=4,last_error_code='credentials_or_access' WHERE bot_id=12345`)

	_, sendErr := db.ExecContext(t.Context(), `UPDATE telegram_outbox SET state='blocked',attempts=5,total_attempts=7,retry_cycle=3,last_error_code='request_failed' WHERE bot_id=12345`)
	if err := errors.Join(pollErr, sendErr, db.Close()); err != nil {
		t.Fatal(err)
	}
}

func checkTelegramRecoveryOperatorRetries(t *testing.T, receiver telegramReceiver, root, id string) {
	t.Helper()

	poll := []string{"telegram", "retry", "--dev-root", root, "--poll"}
	checkTelegramRecoveryLiveRetry(t, receiver, poll)

	var output bytes.Buffer

	for _, args := range [][]string{poll, {"telegram", "retry", "--dev-root", root, "--delivery", id}} {
		if err := run(t.Context(), args, &output, &output); err != nil {
			t.Fatal(err)
		}

		if err := run(t.Context(), args, &output, &output); err == nil {
			t.Fatal("second cycle opened without exhaustion")
		}
	}

	checkTelegramRecoveryOperatorHistory(t, receiver)
}

func checkTelegramRecoveryOperatorHistory(t *testing.T, receiver telegramReceiver) {
	t.Helper()

	after, err := telegramcore.InspectTelegramReceiverDiagnostics(t.Context(), receiver.env, receiver.botID)
	if err != nil {
		t.Fatal(err)
	}

	actor := "local-uid-" + strconv.Itoa(os.Getuid())
	if after.Receiver.RetryCycle != 5 || after.Receiver.RetryAttempt != 0 || after.Receiver.TotalRetryAttempts != 11 || after.Receiver.RetryActor != actor {
		t.Fatal("poll retry lost history", after)
	}

	checkTelegramRecoveryDeliveryHistory(t, after.Deliveries[0], actor)
}

func checkTelegramRecoveryDeliveryHistory(t *testing.T, item telegramcore.TelegramDeliveryDiagnostic, actor string) {
	t.Helper()

	if item.Cycle != 4 || item.Attempts != 0 || item.TotalAttempts != 7 || item.RetryActor != actor || item.RetryStartedAt.IsZero() {
		t.Fatal("delivery retry lost history", item)
	}
}

func checkTelegramRecoveryLiveRetry(t *testing.T, receiver telegramReceiver, poll []string) {
	t.Helper()

	lock, err := telegramcore.AcquireTelegramConfigurationLock(receiver.env)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	refusal := run(t.Context(), poll, &output, &output)
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}

	if refusal == nil || !strings.Contains(refusal.Error(), "local_receiver_conflict") {
		t.Fatal("live receiver retry accepted", refusal)
	}
}
