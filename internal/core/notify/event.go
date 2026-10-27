package notify

import (
	"fmt"
	"time"
)

// EventType defines a package boundary type.
type EventType string

const (
	// EventTicketBlocked defines a package boundary constant.
	EventTicketBlocked EventType = "ticket_blocked"
	// EventValidationFailed defines a package boundary constant.
	EventValidationFailed EventType = "validation_failed"
	// EventLoopFinished defines a package boundary constant.
	EventLoopFinished EventType = "loop_finished"
	// EventLoopBlocked defines a package boundary constant.
	EventLoopBlocked EventType = "loop_blocked"
	// EventLoopFailed defines a package boundary constant.
	EventLoopFailed EventType = "loop_failed"
	// EventDaemonAttention defines a package boundary constant.
	EventDaemonAttention EventType = "daemon_attention"
	// EventRemoteCommand defines a package boundary constant.
	EventRemoteCommand EventType = "remote_command"
	// EventManualTest defines a package boundary constant.
	EventManualTest EventType = "manual_test"
)

// Severity defines a package boundary type.
type Severity string

const (
	// SeverityInfo defines a package boundary constant.
	SeverityInfo Severity = "info"
	// SeverityWarning defines a package boundary constant.
	SeverityWarning Severity = "warning"
	// SeverityError defines a package boundary constant.
	SeverityError Severity = "error"
	// SeverityCritical defines a package boundary constant.
	SeverityCritical Severity = "critical"
)

// Status defines a package boundary type.
type Status string

const (
	// StatusPending defines a package boundary constant.
	StatusPending Status = "pending"
	// StatusSkipped defines a package boundary constant.
	StatusSkipped Status = "skipped"
	// StatusDelivered defines a package boundary constant.
	StatusDelivered Status = "delivered"
	// StatusFailed defines a package boundary constant.
	StatusFailed Status = "failed"
)

// Source defines a package boundary type.
type Source string

const (
	// SourceTicketExecution defines a package boundary constant.
	SourceTicketExecution Source = "ticket_execution"
	// SourceTicketLoop defines a package boundary constant.
	SourceTicketLoop Source = "ticket_loop"
	// SourceDaemon defines a package boundary constant.
	SourceDaemon Source = "daemon"
	// SourceRemoteCommand defines a package boundary constant.
	SourceRemoteCommand Source = "remote_command"
	// SourceManual defines a package boundary constant.
	SourceManual Source = "manual"
)

// Event describes a package boundary value.
type Event struct {
	ID            string
	ProjectID     string
	Type          EventType
	Severity      Severity
	Status        Status
	Source        Source
	References    References
	Payload       Payload
	FailureReason string
	CreatedAt     time.Time
	DeliveredAt   *time.Time
	UpdatedAt     time.Time
}

// References describes a package boundary value.
type References struct {
	TicketKey   string
	ExecutionID string
	LoopRunID   string
	DaemonPID   int
}

// Payload describes a package boundary value.
type Payload struct {
	Summary string
	Detail  string
}

// ValidateEvent validates the corresponding package boundary value.
func ValidateEvent(event Event) error {
	if event.ID == "" {
		return fmt.Errorf("notification event id is required")
	}

	if event.ProjectID == "" {
		return fmt.Errorf("notification event project id is required")
	}

	if !event.Type.Valid() {
		return fmt.Errorf("notification event type is invalid: %s", event.Type)
	}

	if !event.Severity.Valid() {
		return fmt.Errorf("notification severity is invalid: %s", event.Severity)
	}

	if !event.Status.Valid() {
		return fmt.Errorf("notification status is invalid: %s", event.Status)
	}

	if !event.Source.Valid() {
		return fmt.Errorf("notification source is invalid: %s", event.Source)
	}

	err := validatePayload(event.Payload)
	if err != nil {
		return err
	}

	err = validateReferences(event)
	if err != nil {
		return err
	}

	return validateTimes(event)
}

// Valid reports whether the value is one of the supported constants.
func (t EventType) Valid() bool {
	switch t {
	case EventTicketBlocked,
		EventValidationFailed,
		EventLoopFinished,
		EventLoopBlocked,
		EventLoopFailed,
		EventDaemonAttention,
		EventRemoteCommand,
		EventManualTest:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s Severity) Valid() bool {
	switch s {
	case SeverityInfo,
		SeverityWarning,
		SeverityError,
		SeverityCritical:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s Status) Valid() bool {
	switch s {
	case StatusPending,
		StatusSkipped,
		StatusDelivered,
		StatusFailed:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s Source) Valid() bool {
	switch s {
	case SourceTicketExecution,
		SourceTicketLoop,
		SourceDaemon,
		SourceRemoteCommand,
		SourceManual:
		return true
	default:
		return false
	}
}

func validatePayload(payload Payload) error {
	if payload.Summary == "" {
		return fmt.Errorf("notification summary is required")
	}

	return nil
}

func validateReferences(event Event) error {
	if event.References.DaemonPID < 0 {
		return fmt.Errorf("notification daemon pid cannot be negative")
	}

	if event.Type == EventTicketBlocked && event.References.TicketKey == "" {
		return fmt.Errorf("notification ticket key is required for ticket_blocked events")
	}

	if event.Type == EventValidationFailed && event.References.ExecutionID == "" {
		return fmt.Errorf("notification execution id is required for validation_failed events")
	}

	if isLoop(event.Type) && event.References.LoopRunID == "" {
		return fmt.Errorf("notification loop run id is required for %s events", event.Type)
	}

	return nil
}

func isLoop(eventType EventType) bool {
	switch eventType {
	case EventLoopFinished,
		EventLoopBlocked,
		EventLoopFailed:
		return true
	default:
		return false
	}
}

func validateTimes(event Event) error {
	if event.CreatedAt.IsZero() {
		return fmt.Errorf("notification created time is required")
	}

	if event.UpdatedAt.IsZero() {
		return fmt.Errorf("notification updated time is required")
	}

	if event.DeliveredAt != nil && event.DeliveredAt.Before(event.CreatedAt) {
		return fmt.Errorf("notification delivered before it was created")
	}

	if event.DeliveredAt != nil && event.Status != StatusDelivered {
		return fmt.Errorf("notification delivered time is invalid for %s status", event.Status)
	}

	if event.Status == StatusFailed && event.FailureReason == "" {
		return fmt.Errorf("notification failure reason is required for failed status")
	}

	return nil
}
