package daemon

import (
	"fmt"
	"time"
)

// DaemonStatus defines a package boundary type.
type DaemonStatus string

const (
	// DaemonStatusStopped defines a package boundary constant.
	DaemonStatusStopped DaemonStatus = "stopped"
	// DaemonStatusStarting defines a package boundary constant.
	DaemonStatusStarting DaemonStatus = "starting"
	// DaemonStatusRunning defines a package boundary constant.
	DaemonStatusRunning DaemonStatus = "running"
	// DaemonStatusStopping defines a package boundary constant.
	DaemonStatusStopping DaemonStatus = "stopping"
	// DaemonStatusStale defines a package boundary constant.
	DaemonStatusStale DaemonStatus = "stale"
	// DaemonStatusFailed defines a package boundary constant.
	DaemonStatusFailed DaemonStatus = "failed"
)

// DaemonState describes a package boundary value.
type DaemonState struct {
	Status          DaemonStatus
	PID             int
	StartedAt       *time.Time
	HeartbeatAt     *time.Time
	StopRequestedAt *time.Time
	FailureReason   string
	UpdatedAt       time.Time
}

// NewStoppedDaemonState constructs the corresponding package service or value.
func NewStoppedDaemonState(updatedAt time.Time) DaemonState {
	return DaemonState{
		Status:    DaemonStatusStopped,
		UpdatedAt: updatedAt,
	}
}

// ValidateDaemonState validates the corresponding package boundary value.
func ValidateDaemonState(state DaemonState) error {
	if !state.Status.Valid() {
		return fmt.Errorf("daemon status is invalid: %s", state.Status)
	}

	if state.UpdatedAt.IsZero() {
		return fmt.Errorf("daemon updated time is required")
	}

	err := validateDaemonStateProcess(state)
	if err != nil {
		return err
	}

	return validateDaemonStateTimes(state)
}

// Valid reports whether the value is one of the supported constants.
func (s DaemonStatus) Valid() bool {
	switch s {
	case DaemonStatusStopped,
		DaemonStatusStarting,
		DaemonStatusRunning,
		DaemonStatusStopping,
		DaemonStatusStale,
		DaemonStatusFailed:
		return true
	default:
		return false
	}
}

func validateDaemonStateProcess(state DaemonState) error {
	switch state.Status {
	case DaemonStatusStarting, DaemonStatusRunning, DaemonStatusStopping:
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

func validateDaemonStateTimes(state DaemonState) error {
	err := validateDaemonStateStartedAt(state)
	if err != nil {
		return err
	}

	err = validateDaemonStateHeartbeatAt(state)
	if err != nil {
		return err
	}

	return validateDaemonStateStopRequestedAt(state)
}

func validateDaemonStateStartedAt(state DaemonState) error {
	switch state.Status {
	case DaemonStatusStarting, DaemonStatusRunning, DaemonStatusStopping:
		if state.StartedAt == nil {
			return fmt.Errorf("daemon started time is required for %s status", state.Status)
		}
	}

	return nil
}

func validateDaemonStateHeartbeatAt(state DaemonState) error {
	if state.StartedAt != nil && state.HeartbeatAt != nil && state.HeartbeatAt.Before(*state.StartedAt) {
		return fmt.Errorf("daemon heartbeat before start")
	}

	if state.HeartbeatAt != nil && state.StartedAt == nil {
		return fmt.Errorf("daemon heartbeat requires started time")
	}

	return nil
}

func validateDaemonStateStopRequestedAt(state DaemonState) error {
	if state.StartedAt != nil && state.StopRequestedAt != nil && state.StopRequestedAt.Before(*state.StartedAt) {
		return fmt.Errorf("daemon stop requested before start")
	}

	if state.StopRequestedAt != nil && state.Status != DaemonStatusStopping && state.Status != DaemonStatusStopped {
		return fmt.Errorf("daemon stop request is invalid for %s status", state.Status)
	}

	return nil
}
