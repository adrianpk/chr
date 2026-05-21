package chc

import (
	"strings"
	"testing"
	"time"
)

func TestValidateNotificationEventAcceptsValidEvent(t *testing.T) {
	event := validNotificationEvent()

	err := ValidateNotificationEvent(event)
	if err != nil {
		t.Fatalf("ValidateNotificationEvent returned error: %v", err)
	}
}

func TestValidateNotificationEventRejectsInvalidEvents(t *testing.T) {
	for _, test := range invalidNotificationEventTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateNotificationEvent(test.event)
			if err == nil {
				t.Fatal("ValidateNotificationEvent returned nil error")
			}

			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ValidateNotificationEvent error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestNotificationEnumsValidateKnownValues(t *testing.T) {
	for _, eventType := range []NotificationEventType{
		NotificationEventTicketBlocked,
		NotificationEventValidationFailed,
		NotificationEventLoopFinished,
		NotificationEventLoopBlocked,
		NotificationEventLoopFailed,
		NotificationEventDaemonAttention,
		NotificationEventRemoteCommand,
		NotificationEventManualTest,
	} {
		if !eventType.Valid() {
			t.Fatalf("NotificationEventType.Valid() = false for %s", eventType)
		}
	}

	for _, severity := range []NotificationSeverity{
		NotificationSeverityInfo,
		NotificationSeverityWarning,
		NotificationSeverityError,
		NotificationSeverityCritical,
	} {
		if !severity.Valid() {
			t.Fatalf("NotificationSeverity.Valid() = false for %s", severity)
		}
	}

	for _, status := range []NotificationStatus{
		NotificationStatusPending,
		NotificationStatusSkipped,
		NotificationStatusDelivered,
		NotificationStatusFailed,
	} {
		if !status.Valid() {
			t.Fatalf("NotificationStatus.Valid() = false for %s", status)
		}
	}

	for _, source := range []NotificationSource{
		NotificationSourceTicketExecution,
		NotificationSourceTicketLoop,
		NotificationSourceDaemon,
		NotificationSourceRemoteCommand,
		NotificationSourceManual,
	} {
		if !source.Valid() {
			t.Fatalf("NotificationSource.Valid() = false for %s", source)
		}
	}
}

func invalidNotificationEventTests() []struct {
	name    string
	event   NotificationEvent
	wantErr string
} {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	earlier := now.Add(-time.Minute)

	return []struct {
		name    string
		event   NotificationEvent
		wantErr string
	}{
		{name: "id", event: notificationEventWith(func(event *NotificationEvent) { event.ID = "" }), wantErr: "notification event id is required"},
		{name: "project id", event: notificationEventWith(func(event *NotificationEvent) { event.ProjectID = "" }), wantErr: "notification event project id is required"},
		{name: "type", event: notificationEventWith(func(event *NotificationEvent) { event.Type = NotificationEventType("unknown") }), wantErr: "notification event type is invalid"},
		{name: "severity", event: notificationEventWith(func(event *NotificationEvent) { event.Severity = NotificationSeverity("unknown") }), wantErr: "notification severity is invalid"},
		{name: "status", event: notificationEventWith(func(event *NotificationEvent) { event.Status = NotificationStatus("unknown") }), wantErr: "notification status is invalid"},
		{name: "source", event: notificationEventWith(func(event *NotificationEvent) { event.Source = NotificationSource("unknown") }), wantErr: "notification source is invalid"},
		{name: "summary", event: notificationEventWith(func(event *NotificationEvent) { event.Payload.Summary = "" }), wantErr: "notification summary is required"},
		{name: "daemon pid", event: notificationEventWith(func(event *NotificationEvent) { event.References.DaemonPID = -1 }), wantErr: "notification daemon pid cannot be negative"},
		{name: "ticket reference", event: notificationEventWith(func(event *NotificationEvent) { event.References.TicketKey = "" }), wantErr: "notification ticket key is required"},
		{name: "execution reference", event: notificationEventWith(func(event *NotificationEvent) {
			event.Type = NotificationEventValidationFailed
			event.References.TicketKey = ""
			event.References.ExecutionID = ""
		}), wantErr: "notification execution id is required"},
		{name: "loop reference", event: notificationEventWith(func(event *NotificationEvent) {
			event.Type = NotificationEventLoopFailed
			event.References.TicketKey = ""
			event.References.LoopRunID = ""
		}), wantErr: "notification loop run id is required"},
		{name: "created at", event: notificationEventWith(func(event *NotificationEvent) { event.CreatedAt = time.Time{} }), wantErr: "notification created time is required"},
		{name: "updated at", event: notificationEventWith(func(event *NotificationEvent) { event.UpdatedAt = time.Time{} }), wantErr: "notification updated time is required"},
		{name: "delivered before created", event: notificationEventWith(func(event *NotificationEvent) {
			event.Status = NotificationStatusDelivered
			event.DeliveredAt = &earlier
		}), wantErr: "notification delivered before it was created"},
		{name: "delivered time status", event: notificationEventWith(func(event *NotificationEvent) {
			event.Status = NotificationStatusPending
			event.DeliveredAt = &now
		}), wantErr: "notification delivered time is invalid"},
		{name: "failure reason", event: notificationEventWith(func(event *NotificationEvent) {
			event.Status = NotificationStatusFailed
			event.FailureReason = ""
		}), wantErr: "notification failure reason is required"},
	}
}

func validNotificationEvent() NotificationEvent {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)

	return NotificationEvent{
		ID:        "notification-1",
		ProjectID: "project-1",
		Type:      NotificationEventTicketBlocked,
		Severity:  NotificationSeverityWarning,
		Status:    NotificationStatusPending,
		Source:    NotificationSourceTicketExecution,
		References: NotificationReferences{
			TicketKey:   "CHC-1",
			ExecutionID: "execution-1",
		},
		Payload: NotificationPayload{
			Summary: "Ticket CHC-1 is blocked",
			Detail:  "Validation failed after the configured retry threshold.",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func notificationEventWith(update func(*NotificationEvent)) NotificationEvent {
	event := validNotificationEvent()
	update(&event)

	return event
}
