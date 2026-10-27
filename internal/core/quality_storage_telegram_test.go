package core

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func qualityStorageNow() time.Time { return time.Date(2026, 10, 26, 11, 12, 13, 0, time.UTC) }

func qualityReceiverInput(t *testing.T) telegramcore.ReceivedUpdate {
	t.Helper()

	update := telegramRetainedTestUpdate(t, 31, "distinct document request")
	update.MessageID = 32
	update.ReplyToMessageID = 33

	return update
}

func qualityNotificationInput(t *testing.T, db *sql.DB) notifycore.TelegramInput {
	t.Helper()

	input := telegramNotificationTestInput(t)
	input.Record = qualityLegacyExpected(t).Notification

	target := &input.Target
	if err := db.QueryRowContext(t.Context(), "SELECT project_id,telegram_namespace,telegram_project_binding_digest,telegram_destination,telegram_config_digest FROM notification_events WHERE id=?", input.Record.Event.ID).Scan(&target.ProjectID, &target.Namespace, &target.ProjectBindingDigest, &target.Destination, &target.ConfigDigest); err != nil {
		t.Fatal(err)
	}

	input.Document.ProjectID = target.ProjectID
	input.Document.Namespace = target.Namespace
	input.Document.ProjectBindingDigest = target.ProjectBindingDigest

	return input
}

// TestQualityStorageTelegram rewrites the frozen admissions through current
// public operations and compares all columns, including quota and owner images.
func TestQualityStorageTelegram(t *testing.T) {
	t.Run("receiver", func(t *testing.T) {
		db := openQualityLegacyDB(t)
		want := qualityOwnerImages(t, db)
		qualityStorageExec(t, db, "DELETE FROM telegram_outbox")
		qualityStorageExec(t, db, "DELETE FROM telegram_document_snapshots")
		qualityStorageExec(t, db, "DELETE FROM telegram_updates")

		store := qualityReceiverStore(t, db)

		update := qualityReceiverInput(t)
		if err := store.RetainBatch(t.Context(), 12345, []telegramcore.ReceivedUpdate{update}, qualityStorageNow()); err != nil {
			t.Fatal(err)
		}

		if err := store.QueueDocumentResponse(t.Context(), update, telegramTestDocumentResponse(), telegramReceiverTestConfig().Digest(), qualityStorageNow()); err != nil {
			t.Fatal(err)
		}

		if got := qualityOwnerImages(t, db); !reflect.DeepEqual(got, want) {
			t.Fatalf("receiver admission changed stored columns: got %#v want %#v", got, want)
		}
	})
	t.Run("notification", func(t *testing.T) {
		db := openQualityLegacyDB(t)
		input := qualityNotificationInput(t, db)
		want := qualityOwnerImages(t, db)
		qualityStorageExec(t, db, "UPDATE notification_events SET telegram_document_id=NULL")
		qualityStorageExec(t, db, "DELETE FROM telegram_notification_documents")
		qualityStorageExec(t, db, "DELETE FROM notification_events")

		if err := qualityNotificationRepository(t, db).CreateTelegram(t.Context(), input); err != nil {
			t.Fatal(err)
		}

		if got := qualityOwnerImages(t, db); !reflect.DeepEqual(got, want) {
			t.Fatalf("notification admission changed stored columns: got %#v want %#v", got, want)
		}
	})
}

// TestQualityStorageRollback injects failure after each multi-record mutation
// starts and proves the captured owner and table images survive unchanged.
func TestQualityStorageRollback(t *testing.T) {
	for _, test := range []struct {
		name    string
		prepare func(*testing.T, *sql.DB) func() error
	}{
		{"document bind", qualityDocumentFault},
		{"notification bind", qualityNotificationFault},
		{"ticket tags", qualityTicketFault},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := openQualityLegacyDB(t)
			run := test.prepare(t, db)

			want := qualityOwnerImages(t, db)
			if err := run(); err == nil || !strings.Contains(err.Error(), "injected ") {
				t.Fatalf("storage fault did not reach its injected boundary: %v", err)
			}

			if got := qualityOwnerImages(t, db); !reflect.DeepEqual(got, want) {
				t.Fatalf("rollback changed retained owner records: got %#v want %#v", got, want)
			}
		})
	}
}

func qualityDocumentFault(t *testing.T, db *sql.DB) func() error {
	t.Helper()

	store := qualityReceiverStore(t, db)

	update := telegramRetainedTestUpdate(t, 41, "rollback document")
	if err := store.RetainBatch(t.Context(), 12345, []telegramcore.ReceivedUpdate{update}, qualityStorageNow()); err != nil {
		t.Fatal(err)
	}

	qualityStorageExec(t, db, `CREATE TRIGGER quality_abort BEFORE UPDATE OF document_id ON telegram_outbox BEGIN SELECT RAISE(ABORT,'injected document bind failure'); END`)

	return func() error {
		return store.QueueDocumentResponse(t.Context(), update, telegramTestDocumentResponse(), telegramReceiverTestConfig().Digest(), qualityStorageNow())
	}
}

func qualityNotificationFault(t *testing.T, db *sql.DB) func() error {
	t.Helper()

	input := qualityNotificationInput(t, db)
	input.Record.Event.ID = "rollback-notification"

	qualityStorageExec(t, db, `CREATE TEMP TRIGGER quality_abort BEFORE UPDATE OF telegram_presentation ON notification_events BEGIN SELECT RAISE(ABORT,'injected notification bind failure'); END`)

	return func() error {
		return qualityNotificationRepository(t, db).CreateTelegram(t.Context(), input)
	}
}

func qualityTicketFault(t *testing.T, db *sql.DB) func() error {
	t.Helper()

	ticket := qualityLegacyExpected(t).Ticket
	ticket.Title = "should roll back"

	qualityStorageExec(t, db, `CREATE TEMP TRIGGER quality_abort BEFORE INSERT ON indexed_ticket_tags BEGIN SELECT RAISE(ABORT,'injected ticket tag failure'); END`)

	return func() error { return qualityTicketIndex(t, db).UpsertTicket(t.Context(), ticket) }
}
