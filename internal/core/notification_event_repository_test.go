package core

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	notifycore "forge.adrianpk.com/chr/internal/core/notify"
)

func TestNotificationEventRepositoryCreatesGetsUpdatesAndListsEvents(t *testing.T) {
	repository := newTestNotificationEventRepository(t)
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	record := testNotificationAuditRecord(now)

	err := repository.Create(context.Background(), record)
	if err != nil {
		t.Fatalf("CreateNotificationEvent returned error: %v", err)
	}

	got, err := repository.Get(context.Background(), record.Event.ID)
	if err != nil {
		t.Fatalf("GetNotificationEvent returned error: %v", err)
	}

	assertNotificationAuditRecord(t, got, record)

	deliveredAt := now.Add(time.Minute)
	record.Event.Status = notifycore.StatusDelivered
	record.Event.DeliveredAt = &deliveredAt
	record.Event.UpdatedAt = deliveredAt
	record.Channel = "telegram"

	err = repository.Update(context.Background(), record)
	if err != nil {
		t.Fatalf("UpdateNotificationEvent returned error: %v", err)
	}

	got, err = repository.Get(context.Background(), record.Event.ID)
	if err != nil {
		t.Fatalf("GetNotificationEvent after update returned error: %v", err)
	}

	assertNotificationAuditRecord(t, got, record)

	records, err := repository.ListProject(context.Background(), record.Event.ProjectID)
	if err != nil {
		t.Fatalf("ListProjectNotificationEvents returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertNotificationAuditRecord(t, records[0], record)
}

func TestNotificationEventRepositoryReturnsNotFound(t *testing.T) {
	repository := newTestNotificationEventRepository(t)

	_, err := repository.Get(context.Background(), "missing")
	if !errors.Is(err, notifycore.ErrEventNotFound) {
		t.Fatalf("GetNotificationEvent error = %v, want ErrNotificationEventNotFound", err)
	}

	err = repository.Update(
		context.Background(),
		testNotificationAuditRecord(time.Now().UTC()),
	)
	if !errors.Is(err, notifycore.ErrEventNotFound) {
		t.Fatalf("UpdateNotificationEvent error = %v, want ErrNotificationEventNotFound", err)
	}
}

func TestNotificationEventRepositoryRejectsInvalidRecords(t *testing.T) {
	repository := newTestNotificationEventRepository(t)
	record := testNotificationAuditRecord(time.Now().UTC())
	record.Event.Payload.Summary = ""

	err := repository.Create(context.Background(), record)
	if err == nil || err.Error() != "notification summary is required" {
		t.Fatalf("CreateNotificationEvent error = %v, want summary validation", err)
	}
}

func TestNewNotificationEventRepositoryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := notifycore.NewRepository(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewNotificationEventRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chr", "core.db"))
}

func newTestNotificationEventRepository(t *testing.T) notifycore.Repository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := notifycore.NewRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("close sqlite after repository error: %v", closeErr)
		}

		t.Fatalf("NewNotificationEventRepositoryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	return repository
}

func testNotificationAuditRecord(now time.Time) notifycore.AuditRecord {
	return notifycore.AuditRecord{
		Event: notifycore.Event{
			ID:        "notification-1",
			ProjectID: "project-1",
			Type:      notifycore.EventTicketBlocked,
			Severity:  notifycore.SeverityWarning,
			Status:    notifycore.StatusPending,
			Source:    notifycore.SourceTicketExecution,
			References: notifycore.References{
				TicketKey:   "CHR-1",
				ExecutionID: "exec-1",
			},
			Payload: notifycore.Payload{
				Summary: "Ticket CHR-1 is blocked",
				Detail:  "Validation failed after repeated attempts.",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func assertNotificationAuditRecord(t *testing.T, got, want notifycore.AuditRecord) {
	t.Helper()

	gotDeliveredAt := optionalTimeValue(got.Event.DeliveredAt)
	wantDeliveredAt := optionalTimeValue(want.Event.DeliveredAt)

	got.Event.DeliveredAt = nil
	want.Event.DeliveredAt = nil

	matches := got == want && gotDeliveredAt.Equal(wantDeliveredAt)
	if !matches {
		t.Fatalf("NotificationAuditRecord = %#v, want %#v", got, want)
	}
}
