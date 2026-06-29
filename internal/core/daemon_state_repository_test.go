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

func TestDaemonStateRepositoryUpsertsGetsUpdatesAndClearsState(t *testing.T) {
	repository := newTestDaemonStateRepository(t)
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(time.Minute)
	state := testDaemonState(startedAt)

	err := repository.UpsertDaemonState(context.Background(), state)
	if err != nil {
		t.Fatalf("UpsertDaemonState returned error: %v", err)
	}

	got, err := repository.GetDaemonState(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState returned error: %v", err)
	}

	assertDaemonState(t, got, state)

	err = repository.UpdateDaemonHeartbeat(context.Background(), heartbeatAt)
	if err != nil {
		t.Fatalf("UpdateDaemonHeartbeat returned error: %v", err)
	}

	got, err = repository.GetDaemonState(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState after heartbeat returned error: %v", err)
	}

	state.HeartbeatAt = &heartbeatAt
	state.UpdatedAt = heartbeatAt
	assertDaemonState(t, got, state)

	stopped := NewStoppedDaemonState(heartbeatAt.Add(time.Minute))

	err = repository.UpsertDaemonState(context.Background(), stopped)
	if err != nil {
		t.Fatalf("second UpsertDaemonState returned error: %v", err)
	}

	got, err = repository.GetDaemonState(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState after upsert returned error: %v", err)
	}

	assertDaemonState(t, got, stopped)

	err = repository.ClearDaemonState(context.Background())
	if err != nil {
		t.Fatalf("ClearDaemonState returned error: %v", err)
	}

	_, err = repository.GetDaemonState(context.Background())
	if !errors.Is(err, ErrDaemonStateNotFound) {
		t.Fatalf("GetDaemonState error = %v, want ErrDaemonStateNotFound", err)
	}
}

func TestDaemonStateRepositoryRequestsStop(t *testing.T) {
	repository := newTestDaemonStateRepository(t)
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	stopRequestedAt := startedAt.Add(time.Minute)
	state := testDaemonState(startedAt)

	err := repository.UpsertDaemonState(context.Background(), state)
	if err != nil {
		t.Fatalf("UpsertDaemonState returned error: %v", err)
	}

	err = repository.RequestDaemonStop(context.Background(), stopRequestedAt)
	if err != nil {
		t.Fatalf("RequestDaemonStop returned error: %v", err)
	}

	got, err := repository.GetDaemonState(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState after stop request returned error: %v", err)
	}

	state.Status = DaemonStatusStopping
	state.StopRequestedAt = &stopRequestedAt
	state.UpdatedAt = stopRequestedAt
	assertDaemonState(t, got, state)
}

func TestDaemonStateRepositoryReturnsNotFound(t *testing.T) {
	repository := newTestDaemonStateRepository(t)

	_, err := repository.GetDaemonState(context.Background())
	if !errors.Is(err, ErrDaemonStateNotFound) {
		t.Fatalf("GetDaemonState error = %v, want ErrDaemonStateNotFound", err)
	}

	err = repository.UpdateDaemonHeartbeat(context.Background(), time.Now().UTC())
	if !errors.Is(err, ErrDaemonStateNotFound) {
		t.Fatalf("UpdateDaemonHeartbeat error = %v, want ErrDaemonStateNotFound", err)
	}

	err = repository.RequestDaemonStop(context.Background(), time.Now().UTC())
	if !errors.Is(err, ErrDaemonStateNotFound) {
		t.Fatalf("RequestDaemonStop error = %v, want ErrDaemonStateNotFound", err)
	}
}

func TestDaemonStateRepositoryRejectsInvalidState(t *testing.T) {
	repository := newTestDaemonStateRepository(t)
	state := testDaemonState(time.Now().UTC())
	state.PID = 0

	err := repository.UpsertDaemonState(context.Background(), state)
	if err == nil || err.Error() != "daemon pid is required for running status" {
		t.Fatalf("UpsertDaemonState error = %v, want pid validation", err)
	}
}

func TestNewDaemonStateRepositoryUsesUserEnvironmentDataDir(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := NewDaemonStateRepository(context.Background(), env, testMigrationFS())
	if err != nil {
		t.Fatalf("NewDaemonStateRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(root, "data", "chr", "core.db"))
}

func newTestDaemonStateRepository(t *testing.T) DaemonStateRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewDaemonStateRepositoryFromDB(context.Background(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewDaemonStateRepositoryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close repository: %v", err)
		}
	})

	return repository
}

func testDaemonState(startedAt time.Time) DaemonState {
	return DaemonState{
		Status:    DaemonStatusRunning,
		PID:       1234,
		StartedAt: &startedAt,
		UpdatedAt: startedAt,
	}
}

func assertDaemonState(t *testing.T, got, want DaemonState) {
	t.Helper()

	gotStartedAt := optionalTimeValue(got.StartedAt)
	wantStartedAt := optionalTimeValue(want.StartedAt)
	gotHeartbeatAt := optionalTimeValue(got.HeartbeatAt)
	wantHeartbeatAt := optionalTimeValue(want.HeartbeatAt)
	gotStopRequestedAt := optionalTimeValue(got.StopRequestedAt)
	wantStopRequestedAt := optionalTimeValue(want.StopRequestedAt)

	got.StartedAt = nil
	want.StartedAt = nil
	got.HeartbeatAt = nil
	want.HeartbeatAt = nil
	got.StopRequestedAt = nil
	want.StopRequestedAt = nil

	matches := got == want &&
		gotStartedAt.Equal(wantStartedAt) &&
		gotHeartbeatAt.Equal(wantHeartbeatAt) &&
		gotStopRequestedAt.Equal(wantStopRequestedAt)
	if !matches {
		t.Fatalf("DaemonState = %#v, want %#v", got, want)
	}
}
