package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
)

// A physical pre-28 database upgrades without rewriting event status, payload,
// references or timestamps. NULL and explicit version-zero rows stay legacy.
func TestTelegramPresentationNotificationMigration(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	assets := telegramRecoveryMigrationAssets(t)
	delete(assets, "assets/migration/sqlite/20261021000028-add_telegram_notification_presentation.sql")

	old, err := NewNotificationEventRepository(t.Context(), env, assets)
	if err != nil {
		t.Fatal(err)
	}

	input := telegramNotificationTestInput(t)
	records := []NotificationAuditRecord{input.Record, input.Record}
	records[1].Event.ID = "legacy-delivered"
	records[1].Event.Status = NotificationStatusDelivered

	records[1].Event.DeliveredAt = &records[1].Event.UpdatedAt
	for _, record := range records {
		if err := old.CreateNotificationEvent(t.Context(), record); err != nil {
			t.Fatal(err)
		}
	}

	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	repository := openTelegramNotificationTestRepository(t, env)
	for _, record := range records {
		checkTelegramLegacyNotification(t, repository, record, input.Target)

		if _, err := repository.db.ExecContext(t.Context(), `UPDATE notification_events SET telegram_presentation='{"Version":0}' WHERE id=?`, record.Event.ID); err != nil {
			t.Fatal(err)
		}

		checkTelegramLegacyNotification(t, repository, record, input.Target)
	}
}

func checkTelegramLegacyNotification(t *testing.T, repository NotificationEventRepository, record NotificationAuditRecord, target TelegramNotificationTarget) {
	t.Helper()

	saved, err := repository.GetNotificationEvent(t.Context(), record.Event.ID)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Event.Payload != record.Event.Payload || saved.Event.References != record.Event.References || saved.Event.Status != record.Event.Status || database.FormatTime(saved.Event.UpdatedAt) != database.FormatTime(record.Event.UpdatedAt) {
		t.Fatal("migration changed legacy event", saved)
	}

	delivery, frozen, err := repository.ReadTelegramNotification(t.Context(), record.Event.ID, target)
	if err != nil || frozen || delivery.Document != nil || delivery.Attempts != 0 {
		t.Fatal("legacy event acquired new view or allowance", delivery, frozen, err)
	}
}

// Corrupt, future and missing images fail before a sender can obtain bytes;
// invalid new input rolls back without acknowledging or partially saving it.
func TestTelegramPresentationFailureNotificationMapping(t *testing.T) {
	for _, mutation := range []struct{ name, query string }{
		{"future", `UPDATE notification_events SET telegram_presentation=json_set(telegram_presentation,'$.version',999)`},
		{"missing manifest", `UPDATE notification_events SET telegram_presentation=NULL`},
		{"partial version zero", `UPDATE notification_events SET telegram_presentation='{"version":0}',telegram_presentation_hash=NULL,telegram_document_id=NULL`},
		{"corrupt manifest", `UPDATE notification_events SET telegram_presentation='{'`},
		{"corrupt hash", `UPDATE notification_events SET telegram_presentation_hash=printf('%064d',0)`},
		{"missing image", `UPDATE telegram_notification_documents SET payload=NULL,purged_at='2030-01-01T00:00:00Z'`},
		{"changed bytes", `UPDATE telegram_notification_documents SET payload=zeroblob(byte_count)`},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			repository := openTelegramNotificationTestRepository(t, telegramReceiverTestEnv(t))

			input := telegramNotificationTestInput(t)
			if err := repository.CreateTelegramNotification(t.Context(), input); err != nil {
				t.Fatal(err)
			}

			if _, err := repository.db.ExecContext(t.Context(), mutation.query); err != nil {
				t.Fatal(err)
			}

			if _, _, err := repository.ReadTelegramNotification(t.Context(), input.Record.Event.ID, input.Target); !errors.Is(err, ErrTelegramDocumentUnavailable) {
				t.Fatal("invalid retained image accepted", err)
			}
		})
	}
}

// Shared byte and target validation rejects malformed UTF-8, oversized exports
// and foreign roots atomically rather than silently attaching a shortened file.
func TestTelegramPresentationFailureNotificationInput(t *testing.T) {
	for _, item := range []struct {
		name   string
		change func(*TelegramNotificationInput)
	}{
		{"encoding", func(i *TelegramNotificationInput) { i.Document.Payload = []byte{0xff} }},
		{"oversized", func(i *TelegramNotificationInput) {
			i.Document.Payload = []byte(strings.Repeat("x", telegramDocumentLimit+1))
		}},
		{"wrong root", func(i *TelegramNotificationInput) {
			i.Document.ProjectBindingDigest = contenthash.Sum([]byte("other root"))
		}},
	} {
		t.Run(item.name, func(t *testing.T) {
			repository := openTelegramNotificationTestRepository(t, telegramReceiverTestEnv(t))
			input := telegramNotificationTestInput(t)
			item.change(&input)

			if err := repository.CreateTelegramNotification(t.Context(), input); !errors.Is(err, ErrTelegramDocumentUnavailable) {
				t.Fatal("invalid admission accepted", err)
			}

			if _, err := repository.GetNotificationEvent(t.Context(), input.Record.Event.ID); !errors.Is(err, ErrNotificationEventNotFound) {
				t.Fatal("invalid admission kept its owner", err)
			}
		})
	}
}

// Reopen retains a spent uncertain dispatch. It cannot start a second original
// send, and cleanup never classifies its missing acknowledgement as terminal.
func TestTelegramPresentationFailureNotificationReopen(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	repository := openTelegramNotificationTestRepository(t, env)

	input := telegramNotificationTestInput(t)
	if err := repository.CreateTelegramNotification(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	if reserved, err := repository.ReserveTelegramNotification(t.Context(), input.Record.Event.ID, input.Target, false); err != nil || !reserved {
		t.Fatal(reserved, err)
	}

	if err := repository.FinishTelegramNotification(t.Context(), input.Record.Event.ID, input.Target, TelegramDocumentAcknowledgement{}, "transport_failure", true, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := repository.Close(); err != nil {
		t.Fatal(err)
	}

	repository = openTelegramNotificationTestRepository(t, env)

	checkTelegramNotificationReopened(t, repository, input)
}

func checkTelegramNotificationReopened(t *testing.T, repository NotificationEventRepository, input TelegramNotificationInput) {
	t.Helper()

	retained := telegramNotificationRetained(t, repository, input.Record.Event.ID, input.Target)
	if retained.Status != NotificationStatusFailed || !retained.Uncertain || retained.Attempts != 1 || retained.Acknowledgement.MessageID != 0 {
		t.Fatal("reopen invented delivery or allowance", retained)
	}

	if reserved, err := repository.ReserveTelegramNotification(t.Context(), input.Record.Event.ID, input.Target, false); err != nil || reserved {
		t.Fatal("reopen renewed dispatch allowance", reserved, err)
	}
}
