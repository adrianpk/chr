package core

import notifycore "forge.adrianpk.com/chr/internal/core/notify"

import (
	"bytes"
	"database/sql"
	"os"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func telegramNotificationTestInput(t *testing.T) notifycore.TelegramNotificationInput {
	t.Helper()

	record := testNotificationAuditRecord(time.Now().UTC())
	record.Channel = "telegram"

	target, err := notifycore.ResolveTelegramNotificationTarget(projectcore.Project{ID: record.Event.ProjectID, Path: t.TempDir(), Enabled: true}, DefaultOpsNamespace, "@example_channel", contenthash.Sum([]byte("configured sender")))
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("# Complete worker-reported details\n\nLiteral `*_` 😀\n")

	return notifycore.TelegramNotificationInput{Record: record, Target: target, Message: tgmessage.Blocks{Family: "47", InitialText: "Ticket blocked in Example project", ContentKind: "markdown", Content: tgmessage.DocumentExcerpt(payload), SecondaryText: "Complete original export attached. Integration requires its canonical receipt."}, Document: &tgmessage.Document{ProjectID: target.ProjectID, Namespace: target.Namespace, ProjectBindingDigest: target.ProjectBindingDigest, SourceRef: "exports/notification.md", SourceHash: contenthash.Sum(payload), Filename: "notification.md", MediaType: "text/markdown", Payload: payload}}
}

// A real database freezes the event's presentation and bytes together. Updating
// source data, replaying admission or changing destination cannot replace them.
func TestTelegramPresentationNotificationSnapshot(t *testing.T) {
	repository := openTelegramNotificationTestRepository(t, telegramReceiverTestEnv(t))
	input := telegramNotificationTestInput(t)

	original := bytes.Clone(input.Document.Payload)
	if err := repository.CreateTelegramNotification(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	input.Document.Payload = []byte("different source")

	input.Record.Event.Payload.Detail = "Changed after the original event"
	if err := repository.UpdateNotificationEvent(t.Context(), input.Record); err != nil {
		t.Fatal(err)
	}

	delivery := telegramNotificationRetained(t, repository, input.Record.Event.ID, input.Target)
	if delivery.Document == nil || !bytes.Equal(delivery.Document.Payload, original) || delivery.Presentation.DocumentHash != contenthash.Sum(original) {
		t.Fatal("notification replaced its original image", delivery)
	}

	if err := repository.CreateTelegramNotification(t.Context(), input); err == nil {
		t.Fatal("replay replaced original admission")
	}

	checkTelegramNotificationTargets(t, repository, input)
	checkTelegramNotificationIsolation(t, repository)
}

func telegramNotificationRetained(t *testing.T, repository notifycore.NotificationEventRepository, id string, target notifycore.TelegramNotificationTarget) notifycore.TelegramNotificationDelivery {
	t.Helper()

	delivery, frozen, err := repository.ReadTelegramNotification(t.Context(), id, target)
	if err != nil || !frozen {
		t.Fatal("frozen notification unavailable", delivery, err)
	}

	return delivery
}

func checkTelegramNotificationTargets(t *testing.T, repository notifycore.NotificationEventRepository, input notifycore.TelegramNotificationInput) {
	t.Helper()

	for _, field := range []string{"destination", "project", "root", "namespace", "configuration"} {
		target := input.Target

		switch field {
		case "destination":
			target.Destination = "@other_channel"
		case "project":
			target.ProjectID = "other-project"
		case "root":
			target.ProjectBindingDigest = contenthash.Sum([]byte("other root"))
		case "namespace":
			target.Namespace = "other"
		case "configuration":
			target.ConfigDigest = contenthash.Sum([]byte("other sender"))
		}

		if _, _, err := repository.ReadTelegramNotification(t.Context(), input.Record.Event.ID, target); err == nil {
			t.Fatal("cross-target notification accepted", field)
		}
	}
}

// Claims serialize under the notification owner. A second dispatch or a reopen
// cannot renew allowance; only one definite-format fallback can be selected.
func TestTelegramPresentationNotificationAllowance(t *testing.T) {
	repository := openTelegramNotificationTestRepository(t, telegramReceiverTestEnv(t))

	input := telegramNotificationTestInput(t)
	if err := repository.CreateTelegramNotification(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	for _, item := range []struct{ plain, want bool }{{false, true}, {false, false}, {true, true}, {true, false}, {false, false}} {
		got, err := repository.ReserveTelegramNotification(t.Context(), input.Record.Event.ID, input.Target, item.plain)
		if err != nil || got != item.want {
			t.Fatal("notification allowance changed", item, got, err)
		}
	}

	delivery := telegramNotificationRetained(t, repository, input.Record.Event.ID, input.Target)
	if delivery.Attempts != 2 || !delivery.PlainFallback || !delivery.Uncertain {
		t.Fatal("lost original uncertainty or attempts", delivery)
	}

	finishTelegramNotificationTest(t, repository, input)
}

func finishTelegramNotificationTest(t *testing.T, repository notifycore.NotificationEventRepository, input notifycore.TelegramNotificationInput) {
	t.Helper()

	ack := tgmessage.Acknowledgement{MessageID: 1001, FileID: "original-file"}
	if err := repository.FinishTelegramNotification(t.Context(), input.Record.Event.ID, input.Target, ack, "", false, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	retained := telegramNotificationRetained(t, repository, input.Record.Event.ID, input.Target)
	if retained.Acknowledgement != ack || retained.Status != notifycore.NotificationStatusDelivered || retained.Uncertain {
		t.Fatal("lost transport acknowledgement", retained)
	}
}

type notificationTestRepository struct {
	notifycore.NotificationEventRepository
	db *sql.DB
}

func openTelegramNotificationTestRepository(t *testing.T, env UserEnvironment) *notificationTestRepository {
	t.Helper()

	if err := os.MkdirAll(env.DataDir(), 0o755); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", database.UserDatabasePath(env))
	if err != nil {
		t.Fatal(err)
	}

	if err := database.Configure(t.Context(), db); err != nil {
		t.Fatal(database.CloseOnError(db, err))
	}

	repository, err := notifycore.NewNotificationEventRepositoryFromDB(t.Context(), db, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(database.CloseOnError(db, err))
	}

	t.Cleanup(func() {
		if err := repository.Close(); err != nil {
			t.Error(err)
		}
	})

	return &notificationTestRepository{NotificationEventRepository: repository, db: db}
}

func checkTelegramNotificationIsolation(t *testing.T, repository *notificationTestRepository) {
	t.Helper()

	for _, item := range []struct {
		query string
		want  int
	}{{"SELECT COUNT(*) FROM telegram_receivers", 0}, {"SELECT COUNT(*) FROM notification_events", 1}, {"SELECT COUNT(*) FROM telegram_notification_documents", 1}} {
		var count int
		if err := repository.db.QueryRowContext(t.Context(), item.query).Scan(&count); err != nil || count != item.want {
			t.Fatal("notification allocated another owner or duplicated admission", item.query, count, err)
		}
	}
}
