package core

import (
	"fmt"
	"time"
)

// NotificationEventType defines a package boundary type.
type NotificationEventType string

const (
	// NotificationEventTicketBlocked defines a package boundary constant.
	NotificationEventTicketBlocked NotificationEventType = "ticket_blocked"
	// NotificationEventValidationFailed defines a package boundary constant.
	NotificationEventValidationFailed NotificationEventType = "validation_failed"
	// NotificationEventLoopFinished defines a package boundary constant.
	NotificationEventLoopFinished NotificationEventType = "loop_finished"
	// NotificationEventLoopBlocked defines a package boundary constant.
	NotificationEventLoopBlocked NotificationEventType = "loop_blocked"
	// NotificationEventLoopFailed defines a package boundary constant.
	NotificationEventLoopFailed NotificationEventType = "loop_failed"
	// NotificationEventDaemonAttention defines a package boundary constant.
	NotificationEventDaemonAttention NotificationEventType = "daemon_attention"
	// NotificationEventRemoteCommand defines a package boundary constant.
	NotificationEventRemoteCommand NotificationEventType = "remote_command"
	// NotificationEventManualTest defines a package boundary constant.
	NotificationEventManualTest NotificationEventType = "manual_test"
)

// NotificationSeverity defines a package boundary type.
type NotificationSeverity string

const (
	// NotificationSeverityInfo defines a package boundary constant.
	NotificationSeverityInfo NotificationSeverity = "info"
	// NotificationSeverityWarning defines a package boundary constant.
	NotificationSeverityWarning NotificationSeverity = "warning"
	// NotificationSeverityError defines a package boundary constant.
	NotificationSeverityError NotificationSeverity = "error"
	// NotificationSeverityCritical defines a package boundary constant.
	NotificationSeverityCritical NotificationSeverity = "critical"
)

// NotificationStatus defines a package boundary type.
type NotificationStatus string

const (
	// NotificationStatusPending defines a package boundary constant.
	NotificationStatusPending NotificationStatus = "pending"
	// NotificationStatusSkipped defines a package boundary constant.
	NotificationStatusSkipped NotificationStatus = "skipped"
	// NotificationStatusDelivered defines a package boundary constant.
	NotificationStatusDelivered NotificationStatus = "delivered"
	// NotificationStatusFailed defines a package boundary constant.
	NotificationStatusFailed NotificationStatus = "failed"
)

// NotificationSource defines a package boundary type.
type NotificationSource string

const (
	// NotificationSourceTicketExecution defines a package boundary constant.
	NotificationSourceTicketExecution NotificationSource = "ticket_execution"
	// NotificationSourceTicketLoop defines a package boundary constant.
	NotificationSourceTicketLoop NotificationSource = "ticket_loop"
	// NotificationSourceDaemon defines a package boundary constant.
	NotificationSourceDaemon NotificationSource = "daemon"
	// NotificationSourceRemoteCommand defines a package boundary constant.
	NotificationSourceRemoteCommand NotificationSource = "remote_command"
	// NotificationSourceManual defines a package boundary constant.
	NotificationSourceManual NotificationSource = "manual"
)

// NotificationEvent describes a package boundary value.
type NotificationEvent struct {
	ID            string
	ProjectID     string
	Type          NotificationEventType
	Severity      NotificationSeverity
	Status        NotificationStatus
	Source        NotificationSource
	References    NotificationReferences
	Payload       NotificationPayload
	FailureReason string
	CreatedAt     time.Time
	DeliveredAt   *time.Time
	UpdatedAt     time.Time
}

// NotificationReferences describes a package boundary value.
type NotificationReferences struct {
	TicketKey   string
	ExecutionID string
	LoopRunID   string
	DaemonPID   int
}

// NotificationPayload describes a package boundary value.
type NotificationPayload struct {
	Summary string
	Detail  string
}

// ValidateNotificationEvent validates the corresponding package boundary value.
func ValidateNotificationEvent(event NotificationEvent) error {
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

	err := validateNotificationPayload(event.Payload)
	if err != nil {
		return err
	}

	err = validateNotificationReferences(event)
	if err != nil {
		return err
	}

	return validateNotificationTimes(event)
}

// Valid reports whether the value is one of the supported constants.
func (t NotificationEventType) Valid() bool {
	switch t {
	case NotificationEventTicketBlocked,
		NotificationEventValidationFailed,
		NotificationEventLoopFinished,
		NotificationEventLoopBlocked,
		NotificationEventLoopFailed,
		NotificationEventDaemonAttention,
		NotificationEventRemoteCommand,
		NotificationEventManualTest:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s NotificationSeverity) Valid() bool {
	switch s {
	case NotificationSeverityInfo,
		NotificationSeverityWarning,
		NotificationSeverityError,
		NotificationSeverityCritical:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s NotificationStatus) Valid() bool {
	switch s {
	case NotificationStatusPending,
		NotificationStatusSkipped,
		NotificationStatusDelivered,
		NotificationStatusFailed:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (s NotificationSource) Valid() bool {
	switch s {
	case NotificationSourceTicketExecution,
		NotificationSourceTicketLoop,
		NotificationSourceDaemon,
		NotificationSourceRemoteCommand,
		NotificationSourceManual:
		return true
	default:
		return false
	}
}

func validateNotificationPayload(payload NotificationPayload) error {
	if payload.Summary == "" {
		return fmt.Errorf("notification summary is required")
	}

	return nil
}

func validateNotificationReferences(event NotificationEvent) error {
	if event.References.DaemonPID < 0 {
		return fmt.Errorf("notification daemon pid cannot be negative")
	}

	if event.Type == NotificationEventTicketBlocked && event.References.TicketKey == "" {
		return fmt.Errorf("notification ticket key is required for ticket_blocked events")
	}

	if event.Type == NotificationEventValidationFailed && event.References.ExecutionID == "" {
		return fmt.Errorf("notification execution id is required for validation_failed events")
	}

	if isLoopNotification(event.Type) && event.References.LoopRunID == "" {
		return fmt.Errorf("notification loop run id is required for %s events", event.Type)
	}

	return nil
}

func isLoopNotification(eventType NotificationEventType) bool {
	switch eventType {
	case NotificationEventLoopFinished,
		NotificationEventLoopBlocked,
		NotificationEventLoopFailed:
		return true
	default:
		return false
	}
}

func validateNotificationTimes(event NotificationEvent) error {
	if event.CreatedAt.IsZero() {
		return fmt.Errorf("notification created time is required")
	}

	if event.UpdatedAt.IsZero() {
		return fmt.Errorf("notification updated time is required")
	}

	if event.DeliveredAt != nil && event.DeliveredAt.Before(event.CreatedAt) {
		return fmt.Errorf("notification delivered before it was created")
	}

	if event.DeliveredAt != nil && event.Status != NotificationStatusDelivered {
		return fmt.Errorf("notification delivered time is invalid for %s status", event.Status)
	}

	if event.Status == NotificationStatusFailed && event.FailureReason == "" {
		return fmt.Errorf("notification failure reason is required for failed status")
	}

	return nil
}
