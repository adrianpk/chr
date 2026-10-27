package core

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// Receiver and notification admissions contend for the same final 2 MiB on
// independent connections. The loser cannot leave a partial owner or snapshot.
func TestTelegramPresentationFailureSharedQuota(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	repository := openTelegramNotificationTestRepository(t, env)
	seedTelegramDocumentQuota(t, store, 31)

	update := telegramRetainedTestUpdate(t, 1, "document")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	input := telegramNotificationTestInput(t)
	input.Document.Payload = bytes.Repeat([]byte("n"), tgmessage.DocumentLimit)
	input.Document.SourceHash = contenthash.Sum(input.Document.Payload)
	response := telegramTestDocumentResponse()
	response.Document.Payload = bytes.Repeat([]byte("r"), tgmessage.DocumentLimit)
	response.Document.SourceHash = contenthash.Sum(response.Document.Payload)

	ready, results := make(chan struct{}), make(chan error, 2)
	go func() { <-ready; results <- repository.CreateTelegramNotification(t.Context(), input) }()
	go func() {
		<-ready

		results <- store.QueueDocumentResponse(t.Context(), update, response, telegramReceiverTestConfig().Digest(), time.Now())
	}()

	close(ready)

	successful := 0

	for range 2 {
		err := <-results
		if err == nil {
			successful++
		} else if !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
			t.Fatal(err)
		}
	}

	if successful != 1 {
		t.Fatal("shared quota admitted wrong number of owners", successful)
	}

	checkTelegramNotificationQuota(t, repository)
	checkTelegramNotificationQuotaRollback(t, repository, store)
}

func checkTelegramNotificationQuota(t *testing.T, repository *sqliteNotificationEventRepository) {
	t.Helper()

	var usage int64

	err := repository.db.QueryRowContext(t.Context(), `SELECT (SELECT COALESCE(SUM(length(payload)),0) FROM telegram_document_snapshots)+(SELECT COALESCE(SUM(length(payload)),0) FROM telegram_notification_documents)`).Scan(&usage)
	if err != nil || usage != tgmessage.DocumentQuota {
		t.Fatal("shared quota changed", usage, err)
	}
}

func checkTelegramNotificationQuotaRollback(t *testing.T, repository *sqliteNotificationEventRepository, store *TelegramReceiverStore) {
	t.Helper()

	var events, documents, replies int
	if err := repository.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM notification_events").Scan(&events); err != nil {
		t.Fatal(err)
	}

	if err := repository.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_notification_documents").Scan(&documents); err != nil {
		t.Fatal(err)
	}

	if err := store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_outbox").Scan(&replies); err != nil {
		t.Fatal(err)
	}

	if events != documents || events+replies != 1 {
		t.Fatal("quota left partial admission", events, documents, replies)
	}

	input := telegramNotificationTestInput(t)

	input.Record.Event.ID = "rejected-notification"
	if err := repository.CreateTelegramNotification(t.Context(), input); !errors.Is(err, tgmessage.ErrDocumentUnavailable) {
		t.Fatal("full database admitted another export", err)
	}

	if _, err := repository.GetNotificationEvent(t.Context(), input.Record.Event.ID); !errors.Is(err, ErrNotificationEventNotFound) {
		t.Fatal("failed reservation kept its event", err)
	}
}

// Cleanup removes at most twenty aged, definitely terminal images. Pending and
// uncertain bytes survive, and transport receipts and content hashes remain.
func TestTelegramPresentationNotificationRetention(t *testing.T) {
	repository := openTelegramNotificationTestRepository(t, telegramReceiverTestEnv(t))
	input := telegramNotificationTestInput(t)
	now := time.Now().UTC()
	old := now.Add(-8 * 24 * time.Hour)

	for index := range 24 {
		input.Record.Event.ID = fmt.Sprintf("retained-%02d", index)

		input.Record.Event.CreatedAt, input.Record.Event.UpdatedAt = old, old
		if err := repository.CreateTelegramNotification(t.Context(), input); err != nil {
			t.Fatal(err)
		}

		finishTelegramNotificationRetention(t, repository, input, index, old, now)
	}

	if err := repository.notificationTransaction(t.Context(), func(tx *sql.Tx) error { return purgeTelegramNotificationDocuments(t.Context(), tx, now) }); err != nil {
		t.Fatal(err)
	}

	checkTelegramNotificationRetention(t, repository, contenthash.Sum(input.Document.Payload))
}

func finishTelegramNotificationRetention(t *testing.T, repository NotificationEventRepository, input TelegramNotificationInput, index int, old, now time.Time) {
	t.Helper()

	if index == 22 {
		return
	} // Unresolved original dispatch.

	if reserved, err := repository.ReserveTelegramNotification(t.Context(), input.Record.Event.ID, input.Target, false); err != nil || !reserved {
		t.Fatal(reserved, err)
	}

	ack, code, uncertain := tgmessage.Acknowledgement{MessageID: 77, FileID: "original-file"}, "", false
	if index == 21 {
		ack, code, uncertain = tgmessage.Acknowledgement{}, "transport_failure", true
	}

	if index == 23 {
		old = now
	}

	if err := repository.FinishTelegramNotification(t.Context(), input.Record.Event.ID, input.Target, ack, code, uncertain, old); err != nil {
		t.Fatal(err)
	}
}

func checkTelegramNotificationRetention(t *testing.T, repository *sqliteNotificationEventRepository, hash string) {
	t.Helper()

	var purged, retained, receipts, hashes int
	for _, check := range []struct {
		query string
		count *int
	}{
		{"SELECT COUNT(*) FROM telegram_notification_documents WHERE payload IS NULL", &purged},
		{"SELECT COUNT(*) FROM telegram_notification_documents WHERE payload IS NOT NULL", &retained},
		{"SELECT COUNT(*) FROM notification_events WHERE telegram_message_id=77 AND telegram_file_id='original-file'", &receipts},
		{"SELECT COUNT(*) FROM telegram_notification_documents WHERE content_hash='" + hash + "'", &hashes},
	} {
		if err := repository.db.QueryRowContext(t.Context(), check.query).Scan(check.count); err != nil {
			t.Fatal(err)
		}
	}

	if purged != 20 || retained != 4 || receipts != 22 || hashes != 24 {
		t.Fatal("cleanup changed live bytes or original receipts", purged, retained, receipts, hashes)
	}

	var live int
	if err := repository.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_notification_documents WHERE owner_id IN ('retained-21','retained-22','retained-23') AND payload IS NOT NULL").Scan(&live); err != nil || live != 3 {
		t.Fatal("cleanup removed unresolved or recent bytes", live, err)
	}
}
