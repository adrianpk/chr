package daemon

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestDaemonStateRepositoryUpsertsGetsUpdatesAndClearsState(t *testing.T) {
	repository := newTestDaemonStateRepository(t)
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(time.Minute)
	state := testDaemonState(startedAt)

	err := repository.Upsert(context.Background(), state)
	if err != nil {
		t.Fatalf("UpsertDaemonState returned error: %v", err)
	}

	got, err := repository.Get(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState returned error: %v", err)
	}

	assertDaemonState(t, got, state)

	err = repository.UpdateHeartbeat(context.Background(), heartbeatAt)
	if err != nil {
		t.Fatalf("UpdateDaemonHeartbeat returned error: %v", err)
	}

	got, err = repository.Get(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState after heartbeat returned error: %v", err)
	}

	state.HeartbeatAt = &heartbeatAt
	state.UpdatedAt = heartbeatAt
	assertDaemonState(t, got, state)

	stopped := NewStoppedState(heartbeatAt.Add(time.Minute))

	err = repository.Upsert(context.Background(), stopped)
	if err != nil {
		t.Fatalf("second UpsertDaemonState returned error: %v", err)
	}

	got, err = repository.Get(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState after upsert returned error: %v", err)
	}

	assertDaemonState(t, got, stopped)

	err = repository.Clear(context.Background())
	if err != nil {
		t.Fatalf("ClearDaemonState returned error: %v", err)
	}

	_, err = repository.Get(context.Background())
	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("GetDaemonState error = %v, want ErrDaemonStateNotFound", err)
	}
}

func TestDaemonStateRepositoryRequestsStop(t *testing.T) {
	repository := newTestDaemonStateRepository(t)
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	stopRequestedAt := startedAt.Add(time.Minute)
	state := testDaemonState(startedAt)

	err := repository.Upsert(context.Background(), state)
	if err != nil {
		t.Fatalf("UpsertDaemonState returned error: %v", err)
	}

	err = repository.RequestStop(context.Background(), stopRequestedAt)
	if err != nil {
		t.Fatalf("RequestDaemonStop returned error: %v", err)
	}

	got, err := repository.Get(context.Background())
	if err != nil {
		t.Fatalf("GetDaemonState after stop request returned error: %v", err)
	}

	state.Status = Stopping
	state.StopRequestedAt = &stopRequestedAt
	state.UpdatedAt = stopRequestedAt
	assertDaemonState(t, got, state)
}

func TestDaemonStateRepositoryReturnsNotFound(t *testing.T) {
	repository := newTestDaemonStateRepository(t)

	_, err := repository.Get(context.Background())
	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("GetDaemonState error = %v, want ErrDaemonStateNotFound", err)
	}

	err = repository.UpdateHeartbeat(context.Background(), time.Now().UTC())
	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("UpdateDaemonHeartbeat error = %v, want ErrDaemonStateNotFound", err)
	}

	err = repository.RequestStop(context.Background(), time.Now().UTC())
	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("RequestDaemonStop error = %v, want ErrDaemonStateNotFound", err)
	}
}

func TestDaemonStateRepositoryRejectsInvalidState(t *testing.T) {
	repository := newTestDaemonStateRepository(t)
	state := testDaemonState(time.Now().UTC())
	state.PID = 0

	err := repository.Upsert(context.Background(), state)
	if err == nil || err.Error() != "daemon pid is required for running status" {
		t.Fatalf("UpsertDaemonState error = %v, want pid validation", err)
	}
}

func newTestDaemonStateRepository(t *testing.T) Repository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewRepositoryFromDB(context.Background(), db, testMigrationFS())
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

func testDaemonState(startedAt time.Time) State {
	return State{
		Status:    Running,
		PID:       1234,
		StartedAt: &startedAt,
		UpdatedAt: startedAt,
	}
}

func assertDaemonState(t *testing.T, got, want State) {
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
