package daemon

import (
	"fmt"
	"time"
)

// Status defines a package boundary type.
type Status string

const (
	// Stopped defines a package boundary constant.
	Stopped Status = "stopped"
	// Starting defines a package boundary constant.
	Starting Status = "starting"
	// Running defines a package boundary constant.
	Running Status = "running"
	// Stopping defines a package boundary constant.
	Stopping Status = "stopping"
	// Stale defines a package boundary constant.
	Stale Status = "stale"
	// Failed defines a package boundary constant.
	Failed Status = "failed"
)

// State describes a package boundary value.
type State struct {
	Status          Status
	PID             int
	StartedAt       *time.Time
	HeartbeatAt     *time.Time
	StopRequestedAt *time.Time
	FailureReason   string
	UpdatedAt       time.Time
}

// NewStoppedState constructs the corresponding package service or value.
func NewStoppedState(updatedAt time.Time) State {
	return State{
		Status:    Stopped,
		UpdatedAt: updatedAt,
	}
}

// ValidateState validates the corresponding package boundary value.
func ValidateState(state State) error {
	if !state.Status.Valid() {
		return fmt.Errorf("daemon status is invalid: %s", state.Status)
	}

	if state.UpdatedAt.IsZero() {
		return fmt.Errorf("daemon updated time is required")
	}

	err := validateStateProcess(state)
	if err != nil {
		return err
	}

	return validateStateTimes(state)
}

// Valid reports whether the value is one of the supported constants.
func (s Status) Valid() bool {
	switch s {
	case Stopped,
		Starting,
		Running,
		Stopping,
		Stale,
		Failed:
		return true
	default:
		return false
	}
}

func validateStateProcess(state State) error {
	switch state.Status {
	case Starting, Running, Stopping:
		if state.PID <= 0 {
			return fmt.Errorf("daemon pid is required for %s status", state.Status)
		}
	default:
		if state.PID < 0 {
			return fmt.Errorf("daemon pid cannot be negative")
		}
	}

	return nil
}

func validateStateTimes(state State) error {
	err := validateStateStartedAt(state)
	if err != nil {
		return err
	}

	err = validateStateHeartbeatAt(state)
	if err != nil {
		return err
	}

	return validateStateStopRequestedAt(state)
}

func validateStateStartedAt(state State) error {
	switch state.Status {
	case Starting, Running, Stopping:
		if state.StartedAt == nil {
			return fmt.Errorf("daemon started time is required for %s status", state.Status)
		}
	}

	return nil
}

func validateStateHeartbeatAt(state State) error {
	if state.StartedAt != nil && state.HeartbeatAt != nil && state.HeartbeatAt.Before(*state.StartedAt) {
		return fmt.Errorf("daemon heartbeat before start")
	}

	if state.HeartbeatAt != nil && state.StartedAt == nil {
		return fmt.Errorf("daemon heartbeat requires started time")
	}

	return nil
}

func validateStateStopRequestedAt(state State) error {
	if state.StartedAt != nil && state.StopRequestedAt != nil && state.StopRequestedAt.Before(*state.StartedAt) {
		return fmt.Errorf("daemon stop requested before start")
	}

	if state.StopRequestedAt != nil && state.Status != Stopping && state.Status != Stopped {
		return fmt.Errorf("daemon stop request is invalid for %s status", state.Status)
	}

	return nil
}
