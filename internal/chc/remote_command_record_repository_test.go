package chc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRemoteCommandRecordRepositoryCreatesGetsAndListsRecords(t *testing.T) {
	repository := newTestRemoteCommandRecordRepository(t)
	record := testRemoteCommandRecord(t)

	err := repository.CreateRemoteCommandRecord(context.Background(), record)
	if err != nil {
		t.Fatalf("CreateRemoteCommandRecord returned error: %v", err)
	}

	got, err := repository.GetRemoteCommandRecord(context.Background(), record.Request.ID)
	if err != nil {
		t.Fatalf("GetRemoteCommandRecord returned error: %v", err)
	}

	assertRemoteCommandRecord(t, got, record)

	records, err := repository.ListRemoteCommandRecords(context.Background())
	if err != nil {
		t.Fatalf("ListRemoteCommandRecords returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRemoteCommandRecord(t, records[0], record)
}

func TestRemoteCommandRecordRepositoryReturnsNotFound(t *testing.T) {
	repository := newTestRemoteCommandRecordRepository(t)

	_, err := repository.GetRemoteCommandRecord(context.Background(), "missing")
	if !errors.Is(err, ErrRemoteCommandRecordNotFound) {
		t.Fatalf("GetRemoteCommandRecord error = %v, want ErrRemoteCommandRecordNotFound", err)
	}
}

func newTestRemoteCommandRecordRepository(t *testing.T) RemoteCommandRecordRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewRemoteCommandRecordRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("close sqlite after repository error: %v", closeErr)
		}

		t.Fatalf("NewRemoteCommandRecordRepositoryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	return repository
}

func testRemoteCommandRecord(t *testing.T) RemoteCommandRecord {
	t.Helper()

	request, err := ParseRemoteCommandRequest(remoteParseOptions("loop start --status ready --max-iterations 1"))
	if err != nil {
		t.Fatalf("ParseRemoteCommandRequest returned error: %v", err)
	}

	record, err := NewAcceptedRemoteCommandRecord(request)
	if err != nil {
		t.Fatalf("NewAcceptedRemoteCommandRecord returned error: %v", err)
	}

	return record
}

func assertRemoteCommandRecord(t *testing.T, got, want RemoteCommandRecord) {
	t.Helper()

	if got.Status != want.Status || got.HandoffCommand != want.HandoffCommand {
		t.Fatalf("record = %#v, want %#v", got, want)
	}

	if got.Request.ID != want.Request.ID || got.Request.Intent != want.Request.Intent {
		t.Fatalf("request = %#v, want %#v", got.Request, want.Request)
	}
}
