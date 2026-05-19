package chc

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestExecutionRepositoryCreatesGetsUpdatesAndListsExecutions(t *testing.T) {
	repository := newTestExecutionRepository(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	record := testExecutionRecord(now)

	if err := repository.CreateExecution(context.Background(), record); err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	got, err := repository.GetExecution(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("GetExecution returned error: %v", err)
	}

	assertExecutionRecord(t, got, record)

	startedAt := now.Add(time.Minute)
	finishedAt := now.Add(2 * time.Minute)
	record.Status = ExecutionStatusSucceeded
	record.StartedAt = &startedAt
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	record.Artifacts.ValidationPath = "/tmp/exec-1/validation.log"

	if err := repository.UpdateExecution(context.Background(), record); err != nil {
		t.Fatalf("UpdateExecution returned error: %v", err)
	}

	got, err = repository.GetExecution(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("GetExecution after update returned error: %v", err)
	}

	assertExecutionRecord(t, got, record)

	records, err := repository.ListTicketExecutions(context.Background(), record.ProjectID, record.TicketKey)
	if err != nil {
		t.Fatalf("ListTicketExecutions returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertExecutionRecord(t, records[0], record)
}

func TestExecutionRepositoryReturnsNotFound(t *testing.T) {
	repository := newTestExecutionRepository(t)

	_, err := repository.GetExecution(context.Background(), "missing")
	if !errors.Is(err, ErrExecutionNotFound) {
		t.Fatalf("GetExecution error = %v, want ErrExecutionNotFound", err)
	}

	err = repository.UpdateExecution(context.Background(), testExecutionRecord(time.Now().UTC()))
	if !errors.Is(err, ErrExecutionNotFound) {
		t.Fatalf("UpdateExecution error = %v, want ErrExecutionNotFound", err)
	}
}

func TestNewExecutionRepositoryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := NewExecutionRepository(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		if err := repository.Close(); err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chc", "chc.db"))
}

func newTestExecutionRepository(t *testing.T) ExecutionRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewExecutionRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			t.Fatalf("close sqlite after repository error: %v", closeErr)
		}

		t.Fatalf("NewExecutionRepositoryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		if err := repository.Close(); err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	return repository
}

func testExecutionRecord(now time.Time) ExecutionRecord {
	return ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHC-1",
		TicketPath: "/tmp/project/.chc/tickets/CHC-1.md",
		Status:     ExecutionStatusPending,
		Runtime: ExecutionRuntimeRef{
			Image:             "chc-agent-runtime:local",
			DockerfilePath:    "/tmp/project/.chc/runtime/agent/Dockerfile",
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: ExecutionArtifacts{
			RunDir:     "/tmp/project/.chc/runtime/runs/exec-1",
			ResultPath: "/tmp/project/.chc/runtime/runs/exec-1/result.json",
			LogPath:    "/tmp/project/.chc/runtime/runs/exec-1/runtime.log",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func assertExecutionRecord(t *testing.T, got, want ExecutionRecord) {
	t.Helper()

	gotStartedAt := optionalTimeValue(got.StartedAt)
	wantStartedAt := optionalTimeValue(want.StartedAt)
	gotFinishedAt := optionalTimeValue(got.FinishedAt)
	wantFinishedAt := optionalTimeValue(want.FinishedAt)

	got.StartedAt = nil
	want.StartedAt = nil
	got.FinishedAt = nil
	want.FinishedAt = nil

	if got != want ||
		!gotStartedAt.Equal(wantStartedAt) ||
		!gotFinishedAt.Equal(wantFinishedAt) {
		t.Fatalf("ExecutionRecord = %#v, want %#v", got, want)
	}
}

func optionalTimeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}

	return *value
}
