// Package daemon owns host lifecycle state, start guards and persistent stop/heartbeat operations.
package daemon

import (
	"fmt"
	"time"
)

// DefaultStaleAfter defines a package boundary constant.
const DefaultStaleAfter = 5 * time.Minute

// StartGuardDecision describes a package boundary value.
type StartGuardDecision struct {
	Allowed bool
	Stale   bool
	Reason  string
	State   State
}

// EvaluateStartGuard executes the corresponding package operation.
func EvaluateStartGuard(state State, now time.Time, staleAfter time.Duration) (StartGuardDecision, error) {
	err := ValidateState(state)
	if err != nil {
		return StartGuardDecision{}, err
	}

	if staleAfter <= 0 {
		return StartGuardDecision{}, fmt.Errorf("daemon stale duration must be positive")
	}

	decision := StartGuardDecision{State: state}

	switch state.Status {
	case Stopped, Failed, Stale:
		decision.Allowed = true

		return decision, nil
	case Starting, Running, Stopping:
		return evaluateActiveStartGuard(decision, now, staleAfter), nil
	default:
		return StartGuardDecision{}, fmt.Errorf("unsupported daemon status: %s", state.Status)
	}
}

func evaluateActiveStartGuard(decision StartGuardDecision, now time.Time, staleAfter time.Duration) StartGuardDecision {
	lastSeenAt := stateLastSeenAt(decision.State)
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

func stateLastSeenAt(state State) time.Time {
	if state.HeartbeatAt != nil {
		return *state.HeartbeatAt
	}

	if state.StartedAt != nil {
		return *state.StartedAt
	}

	return state.UpdatedAt
}
