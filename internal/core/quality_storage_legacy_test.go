package core

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type qualityLegacyRecords struct {
	Execution    ExecutionRecord         `json:"execution"`
	Remote       RemoteCommandRecord     `json:"remote"`
	Notification NotificationAuditRecord `json:"notification"`
	Ticket       IndexedTicket           `json:"ticket"`
	Conversation TelegramConversation    `json:"conversation"`
}

func openQualityLegacyDB(t *testing.T) *sql.DB {
	t.Helper()
	data, err := os.ReadFile("testdata/quality-storage/legacy.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "legacy.sqlite")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	db.SetMaxOpenConns(1)
	if err := configureSQLiteDatabase(t.Context(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

func readQualityLegacyRecords(t *testing.T, db *sql.DB, want qualityLegacyRecords) qualityLegacyRecords {
	t.Helper()
	var got qualityLegacyRecords
	var err error
	got.Execution, err = (&sqliteExecutionRepository{db: db}).GetExecution(t.Context(), want.Execution.ID)
	if err != nil {
		t.Fatal(err)
	}
	got.Remote, err = (&sqliteRemoteCommandRecordRepository{db: db}).GetRemoteCommandRecord(t.Context(), want.Remote.Request.ID)
	if err != nil {
		t.Fatal(err)
	}
	got.Notification, err = (&sqliteNotificationEventRepository{db: db}).GetNotificationEvent(t.Context(), want.Notification.Event.ID)
	if err != nil {
		t.Fatal(err)
	}
	got.Ticket, err = (&sqliteTicketIndex{db: db}).GetProjectTicketByKey(t.Context(), want.Ticket.ProjectID, want.Ticket.Key)
	if err != nil {
		t.Fatal(err)
	}
	got.Conversation, err = scanTelegramConversation(db.QueryRowContext(t.Context(), "SELECT "+telegramConversationColumns+" FROM telegram_conversations WHERE id=?", want.Conversation.ID))
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// TestQualityStorageLegacy reads records frozen by the pre-refactor writers,
// retaining every decoded value, NULL timestamp and empty versus nil list.
func TestQualityStorageLegacy(t *testing.T) {
	data, err := os.ReadFile("testdata/quality-storage/legacy-records.json")
	if err != nil {
		t.Fatal(err)
	}
	var want qualityLegacyRecords
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	db := openQualityLegacyDB(t)
	got := readQualityLegacyRecords(t, db, want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy records changed:\ngot %#v\nwant %#v", got, want)
	}
}
