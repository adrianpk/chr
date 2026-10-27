package core

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/core/database"
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

type qualityLegacyRecords struct {
	Execution    ExecutionRecord           `json:"execution"`
	Remote       RemoteCommandRecord       `json:"remote"`
	Notification notifycore.AuditRecord    `json:"notification"`
	Ticket       ticketcore.Indexed        `json:"ticket"`
	Conversation telegramcore.Conversation `json:"conversation"`
}

func openQualityLegacyDB(t *testing.T) *sql.DB {
	t.Helper()

	data, err := os.ReadFile("testdata/quality-storage/legacy.sqlite")
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "core.db")
	if err := os.WriteFile(path, data, 0o600); err != nil {
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

	if err := database.Configure(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	qualityStorageExec(t, db, "PRAGMA foreign_keys=ON")

	return db
}

func readQualityLegacyRecords(t *testing.T, db *sql.DB, want qualityLegacyRecords) qualityLegacyRecords {
	t.Helper()

	var (
		got qualityLegacyRecords
		err error
	)

	got.Execution, err = (&sqliteExecutionRepository{db: db}).GetExecution(t.Context(), want.Execution.ID)
	if err != nil {
		t.Fatal(err)
	}

	got.Remote, err = (&sqliteRemoteCommandRecordRepository{db: db}).GetRemoteCommandRecord(t.Context(), want.Remote.Request.ID)
	if err != nil {
		t.Fatal(err)
	}

	got.Notification, err = qualityNotificationRepository(t, db).Get(t.Context(), want.Notification.Event.ID)
	if err != nil {
		t.Fatal(err)
	}

	got.Ticket, err = qualityTicketIndex(t, db).GetProjectByKey(t.Context(), want.Ticket.ProjectID, want.Ticket.Key)
	if err != nil {
		t.Fatal(err)
	}

	got.Conversation = qualityReadConversation(t, db, want.Conversation)

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

	before := qualityOwnerImages(t, db)
	if err := database.NewMigrator(os.DirFS("../../cmd/chr")).Migrate(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	if after := qualityOwnerImages(t, db); !reflect.DeepEqual(after, before) {
		t.Fatal("opening the old database changed retained rows")
	}

	var violations int
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil || violations != 0 {
		t.Fatalf("legacy fixture foreign keys: %d, %v", violations, err)
	}

	got := readQualityLegacyRecords(t, db, want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy records changed:\ngot %#v\nwant %#v", got, want)
	}
}

func qualityTicketIndex(t *testing.T, db *sql.DB) ticketcore.Index {
	t.Helper()

	index, err := ticketcore.NewIndexFromDB(t.Context(), db, testMigrationFS())
	if err != nil {
		t.Fatal(err)
	}

	return index
}
