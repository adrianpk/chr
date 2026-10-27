package daemon

import (
	"strings"
	"testing"
	"time"
)

func TestValidateDaemonStateAcceptsStoppedState(t *testing.T) {
	state := NewStoppedState(time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC))

	err := ValidateState(state)
	if err != nil {
		t.Fatalf("ValidateDaemonState returned error: %v", err)
	}
}

func TestValidateDaemonStateAcceptsRunningState(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(time.Minute)

	state := State{
		Status:      Running,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	}

	err := ValidateState(state)
	if err != nil {
		t.Fatalf("ValidateDaemonState returned error: %v", err)
	}
}

func TestValidateDaemonStateAcceptsStoppingState(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	stopRequestedAt := startedAt.Add(time.Minute)

	state := State{
		Status:          Stopping,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	}

	err := ValidateState(state)
	if err != nil {
		t.Fatalf("ValidateDaemonState returned error: %v", err)
	}
}

func TestValidateDaemonStateRejectsInvalidStatus(t *testing.T) {
	state := NewStoppedState(time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC))
	state.Status = Status("unknown")

	assertDaemonStateValidationError(t, state, "daemon status is invalid")
}

func TestValidateDaemonStateRejectsRunningStateWithoutPID(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	state := State{
		Status:    Running,
		StartedAt: &startedAt,
		UpdatedAt: startedAt,
	}

	assertDaemonStateValidationError(t, state, "daemon pid is required")
}

func TestValidateDaemonStateRejectsRunningStateWithoutStartedAt(t *testing.T) {
	state := State{
		Status:    Running,
		PID:       1234,
		UpdatedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
	}

	assertDaemonStateValidationError(t, state, "daemon started time is required")
}

func TestValidateDaemonStateRejectsInvalidTimes(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(-time.Minute)

	state := State{
		Status:      Running,
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

	state := State{
		Status:          Running,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	}

	assertDaemonStateValidationError(t, state, "daemon stop request is invalid")
}

func assertDaemonStateValidationError(t *testing.T, state State, want string) {
	t.Helper()

	err := ValidateState(state)
	if err == nil {
		t.Fatal("ValidateDaemonState returned nil error")
	}

	if !strings.Contains(err.Error(), want) {
		t.Fatalf("ValidateDaemonState error = %v, want %q", err, want)
	}
}
