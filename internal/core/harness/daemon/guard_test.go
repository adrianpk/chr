package daemon

import (
	"strings"
	"testing"
	"time"
)

func TestEvaluateDaemonStartGuardAllowsStoppedFailedAndStaleStates(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	for _, state := range []State{
		NewStoppedState(now),
		{Status: Failed, FailureReason: "crashed", UpdatedAt: now},
		{Status: Stale, PID: 1234, FailureReason: "missing process", UpdatedAt: now},
	} {
		decision, err := EvaluateStartGuard(state, now, DefaultStaleAfter)
		if err != nil {
			t.Fatalf("EvaluateDaemonStartGuard returned error: %v", err)
		}

		if !decision.Allowed {
			t.Fatalf("decision = %#v, want allowed for %s", decision, state.Status)
		}
	}
}

func TestEvaluateDaemonStartGuardRejectsActiveFreshState(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(time.Minute)
	state := State{
		Status:      Running,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	}

	decision, err := EvaluateStartGuard(state, heartbeatAt.Add(time.Minute), DefaultStaleAfter)
	if err != nil {
		t.Fatalf("EvaluateDaemonStartGuard returned error: %v", err)
	}

	if decision.Allowed {
		t.Fatalf("decision = %#v, want rejected", decision)
	}

	if !strings.Contains(decision.Reason, "daemon already running with pid 1234") {
		t.Fatalf("reason = %q, want active daemon reason", decision.Reason)
	}
}

func TestEvaluateDaemonStartGuardAllowsActiveStaleState(t *testing.T) {
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	heartbeatAt := startedAt.Add(time.Minute)
	state := State{
		Status:      Running,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	}

	decision, err := EvaluateStartGuard(state, heartbeatAt.Add(10*time.Minute), DefaultStaleAfter)
	if err != nil {
		t.Fatalf("EvaluateDaemonStartGuard returned error: %v", err)
	}

	if !decision.Allowed || !decision.Stale {
		t.Fatalf("decision = %#v, want allowed stale state", decision)
	}
}

func TestEvaluateDaemonStartGuardRejectsInvalidStaleDuration(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	_, err := EvaluateStartGuard(NewStoppedState(now), now, 0)
	if err == nil {
		t.Fatal("EvaluateDaemonStartGuard returned nil error")
	}

	if !strings.Contains(err.Error(), "daemon stale duration must be positive") {
		t.Fatalf("error = %v, want stale duration validation", err)
	}
}
