package core

import (
	"strings"
	"testing"
	"time"
)

func TestValidateDaemonStateAcceptsStoppedState(t *testing.T) {
	state := NewStoppedDaemonState(time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC))

	err := ValidateDaemonState(state)
	if err != nil {
		t.Fatalf("ValidateDaemonState returned error: %v", err)
	}
}

func TestValidateDaemonStateAcceptsRunningState(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(time.Minute)

	state := DaemonState{
		Status:      DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	}

	err := ValidateDaemonState(state)
	if err != nil {
		t.Fatalf("ValidateDaemonState returned error: %v", err)
	}
}

func TestValidateDaemonStateAcceptsStoppingState(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	stopRequestedAt := startedAt.Add(time.Minute)

	state := DaemonState{
		Status:          DaemonStatusStopping,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	}

	err := ValidateDaemonState(state)
	if err != nil {
		t.Fatalf("ValidateDaemonState returned error: %v", err)
	}
}

func TestValidateDaemonStateRejectsInvalidStatus(t *testing.T) {
	state := NewStoppedDaemonState(time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC))
	state.Status = DaemonStatus("unknown")

	assertDaemonStateValidationError(t, state, "daemon status is invalid")
}

func TestValidateDaemonStateRejectsRunningStateWithoutPID(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	state := DaemonState{
		Status:    DaemonStatusRunning,
		StartedAt: &startedAt,
		UpdatedAt: startedAt,
	}

	assertDaemonStateValidationError(t, state, "daemon pid is required")
}

func TestValidateDaemonStateRejectsRunningStateWithoutStartedAt(t *testing.T) {
	state := DaemonState{
		Status:    DaemonStatusRunning,
		PID:       1234,
		UpdatedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	}

	assertDaemonStateValidationError(t, state, "daemon started time is required")
}

func TestValidateDaemonStateRejectsInvalidTimes(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(-time.Minute)

	state := DaemonState{
		Status:      DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   startedAt,
	}

	assertDaemonStateValidationError(t, state, "daemon heartbeat before start")
}

func TestValidateDaemonStateRejectsStopRequestForRunningState(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	stopRequestedAt := startedAt.Add(time.Minute)

	state := DaemonState{
		Status:          DaemonStatusRunning,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	}

	assertDaemonStateValidationError(t, state, "daemon stop request is invalid")
}

func assertDaemonStateValidationError(t *testing.T, state DaemonState, want string) {
	t.Helper()

	err := ValidateDaemonState(state)
	if err == nil {
		t.Fatal("ValidateDaemonState returned nil error")
	}

	if !strings.Contains(err.Error(), want) {
		t.Fatalf("ValidateDaemonState error = %v, want %q", err, want)
	}
}
