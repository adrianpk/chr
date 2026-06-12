package core

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestLoopRunRepositoryCreatesGetsUpdatesAndListsLoopRuns(t *testing.T) {
	repository := newTestLoopRunRepository(t)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	record := testLoopRunRecord(now)

	err := repository.CreateLoopRun(context.Background(), record)
	if err != nil {
		t.Fatalf("CreateLoopRun returned error: %v", err)
	}

	got, err := repository.GetLoopRun(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("GetLoopRun returned error: %v", err)
	}

	assertLoopRunRecord(t, got, record)

	startedAt := now.Add(time.Minute)
	finishedAt := now.Add(2 * time.Minute)
	record.Status = LoopRunStatusSucceeded
	record.StartedAt = &startedAt
	record.FinishedAt = &finishedAt
	record.EligibleCount = 3
	record.ProcessedCount = 3
	record.BlockedCount = 1
	record.UpdatedAt = finishedAt

	err = repository.UpdateLoopRun(context.Background(), record)
	if err != nil {
		t.Fatalf("UpdateLoopRun returned error: %v", err)
	}

	got, err = repository.GetLoopRun(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("GetLoopRun after update returned error: %v", err)
	}

	assertLoopRunRecord(t, got, record)

	records, err := repository.ListProjectLoopRuns(context.Background(), record.ProjectID)
	if err != nil {
		t.Fatalf("ListProjectLoopRuns returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertLoopRunRecord(t, records[0], record)
}

func TestLoopRunRepositoryReturnsNotFound(t *testing.T) {
	repository := newTestLoopRunRepository(t)

	_, err := repository.GetLoopRun(context.Background(), "missing")
	if !errors.Is(err, ErrLoopRunNotFound) {
		t.Fatalf("GetLoopRun error = %v, want ErrLoopRunNotFound", err)
	}

	err = repository.UpdateLoopRun(context.Background(), testLoopRunRecord(time.Now().UTC()))
	if !errors.Is(err, ErrLoopRunNotFound) {
		t.Fatalf("UpdateLoopRun error = %v, want ErrLoopRunNotFound", err)
	}
}

func TestLoopRunRepositoryRejectsInvalidRecords(t *testing.T) {
	repository := newTestLoopRunRepository(t)
	record := testLoopRunRecord(time.Now().UTC())
	record.EligibleCount = -1

	err := repository.CreateLoopRun(context.Background(), record)
	if err == nil || err.Error() != "loop run eligible count cannot be negative" {
		t.Fatalf("CreateLoopRun error = %v, want negative count validation", err)
	}
}

func TestNewLoopRunRepositoryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := NewLoopRunRepository(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewLoopRunRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chc", "core.db"))
}

func newTestLoopRunRepository(t *testing.T) LoopRunRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	registry, err := NewProjectRegistryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewProjectRegistryFromDB returned error: %v", err)
	}

	addTestProject(t, registry, Project{
		ID:           "project-1",
		Name:         "chrysocoord",
		Path:         "/tmp/chrysocoord",
		Enabled:      true,
		RegisteredAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	repository, err := NewLoopRunRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewLoopRunRepositoryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	return repository
}

func testLoopRunRecord(now time.Time) LoopRunRecord {
	return LoopRunRecord{
		ID:             "loop-1",
		ProjectID:      "project-1",
		Status:         LoopRunStatusPending,
		EligibleCount:  2,
		ProcessedCount: 0,
		BlockedCount:   0,
		FailedCount:    0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func assertLoopRunRecord(t *testing.T, got, want LoopRunRecord) {
	t.Helper()

	gotStartedAt := optionalTimeValue(got.StartedAt)
	wantStartedAt := optionalTimeValue(want.StartedAt)
	gotFinishedAt := optionalTimeValue(got.FinishedAt)
	wantFinishedAt := optionalTimeValue(want.FinishedAt)

	got.StartedAt = nil
	want.StartedAt = nil
	got.FinishedAt = nil
	want.FinishedAt = nil

	matches := got == want &&
		gotStartedAt.Equal(wantStartedAt) &&
		gotFinishedAt.Equal(wantFinishedAt)
	if !matches {
		t.Fatalf("LoopRunRecord = %#v, want %#v", got, want)
	}
}
