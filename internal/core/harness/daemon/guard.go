// Package daemon owns host lifecycle state, start guards and persistent stop/heartbeat operations.
package daemon

import (
	"fmt"
	"time"
)

// DefaultDaemonStaleAfter defines a package boundary constant.
const DefaultDaemonStaleAfter = 5 * time.Minute

// DaemonStartGuardDecision describes a package boundary value.
type DaemonStartGuardDecision struct {
	Allowed bool
	Stale   bool
	Reason  string
	State   DaemonState
}

// EvaluateDaemonStartGuard executes the corresponding package operation.
func EvaluateDaemonStartGuard(state DaemonState, now time.Time, staleAfter time.Duration) (DaemonStartGuardDecision, error) {
	err := ValidateDaemonState(state)
	if err != nil {
		return DaemonStartGuardDecision{}, err
	}

	if staleAfter <= 0 {
		return DaemonStartGuardDecision{}, fmt.Errorf("daemon stale duration must be positive")
	}

	decision := DaemonStartGuardDecision{State: state}

	switch state.Status {
	case DaemonStatusStopped, DaemonStatusFailed, DaemonStatusStale:
		decision.Allowed = true

		return decision, nil
	case DaemonStatusStarting, DaemonStatusRunning, DaemonStatusStopping:
		return evaluateActiveDaemonStartGuard(decision, now, staleAfter), nil
	default:
		return DaemonStartGuardDecision{}, fmt.Errorf("unsupported daemon status: %s", state.Status)
	}
}

func evaluateActiveDaemonStartGuard(
	decision DaemonStartGuardDecision,
	now time.Time,
	staleAfter time.Duration,
) DaemonStartGuardDecision {
	lastSeenAt := daemonStateLastSeenAt(decision.State)
	if now.Sub(lastSeenAt) > staleAfter {
		decision.Allowed = true
		decision.Stale = true
		decision.Reason = "daemon state is stale"

		return decision
	}

	decision.Allowed = false
	decision.Reason = fmt.Sprintf("daemon already %s with pid %d", decision.State.Status, decision.State.PID)

	return decision
}

func daemonStateLastSeenAt(state DaemonState) time.Time {
	if state.HeartbeatAt != nil {
		return *state.HeartbeatAt
	}

	if state.StartedAt != nil {
		return *state.StartedAt
	}

	return state.UpdatedAt
}
