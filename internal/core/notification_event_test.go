package core

import notifycore "forge.adrianpk.com/chr/internal/core/notify"

import (
	"strings"
	"testing"
	"time"
)

func TestValidateNotificationEventAcceptsValidEvent(t *testing.T) {
	event := validNotificationEvent()

	err := notifycore.ValidateNotificationEvent(event)
	if err != nil {
		t.Fatalf("ValidateNotificationEvent returned error: %v", err)
	}
}

func TestValidateNotificationEventRejectsInvalidEvents(t *testing.T) {
	for _, test := range invalidNotificationEventTests() {
		t.Run(test.name, func(t *testing.T) {
			err := notifycore.ValidateNotificationEvent(test.event)
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
	for _, eventType := range []notifycore.NotificationEventType{
		notifycore.NotificationEventTicketBlocked,
		notifycore.NotificationEventValidationFailed,
		notifycore.NotificationEventLoopFinished,
		notifycore.NotificationEventLoopBlocked,
		notifycore.NotificationEventLoopFailed,
		notifycore.NotificationEventDaemonAttention,
		notifycore.NotificationEventRemoteCommand,
		notifycore.NotificationEventManualTest,
	} {
		if !eventType.Valid() {
			t.Fatalf("NotificationEventType.Valid() = false for %s", eventType)
		}
	}

	for _, severity := range []notifycore.NotificationSeverity{
		notifycore.NotificationSeverityInfo,
		notifycore.NotificationSeverityWarning,
		notifycore.NotificationSeverityError,
		notifycore.NotificationSeverityCritical,
	} {
		if !severity.Valid() {
			t.Fatalf("NotificationSeverity.Valid() = false for %s", severity)
		}
	}

	for _, status := range []notifycore.NotificationStatus{
		notifycore.NotificationStatusPending,
		notifycore.NotificationStatusSkipped,
		notifycore.NotificationStatusDelivered,
		notifycore.NotificationStatusFailed,
	} {
		if !status.Valid() {
			t.Fatalf("NotificationStatus.Valid() = false for %s", status)
		}
	}

	for _, source := range []notifycore.NotificationSource{
		notifycore.NotificationSourceTicketExecution,
		notifycore.NotificationSourceTicketLoop,
		notifycore.NotificationSourceDaemon,
		notifycore.NotificationSourceRemoteCommand,
		notifycore.NotificationSourceManual,
	} {
		if !source.Valid() {
			t.Fatalf("NotificationSource.Valid() = false for %s", source)
		}
	}
}

func invalidNotificationEventTests() []struct {
	name    string
	event   notifycore.NotificationEvent
	wantErr string
} {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	earlier := now.Add(-time.Minute)

	return []struct {
		name    string
		event   notifycore.NotificationEvent
		wantErr string
	}{
		{name: "id", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.ID = "" }), wantErr: "notification event id is required"},
		{name: "project id", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.ProjectID = "" }), wantErr: "notification event project id is required"},
		{name: "type", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.Type = notifycore.NotificationEventType("unknown") }), wantErr: "notification event type is invalid"},
		{name: "severity", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.Severity = notifycore.NotificationSeverity("unknown") }), wantErr: "notification severity is invalid"},
		{name: "status", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.Status = notifycore.NotificationStatus("unknown") }), wantErr: "notification status is invalid"},
		{name: "source", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.Source = notifycore.NotificationSource("unknown") }), wantErr: "notification source is invalid"},
		{name: "summary", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.Payload.Summary = "" }), wantErr: "notification summary is required"},
		{name: "daemon pid", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.References.DaemonPID = -1 }), wantErr: "notification daemon pid cannot be negative"},
		{name: "ticket reference", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.References.TicketKey = "" }), wantErr: "notification ticket key is required"},
		{name: "execution reference", event: notificationEventWith(func(event *notifycore.NotificationEvent) {
			event.Type = notifycore.NotificationEventValidationFailed
			event.References.TicketKey = ""
			event.References.ExecutionID = ""
		}), wantErr: "notification execution id is required"},
		{name: "loop reference", event: notificationEventWith(func(event *notifycore.NotificationEvent) {
			event.Type = notifycore.NotificationEventLoopFailed
			event.References.TicketKey = ""
			event.References.LoopRunID = ""
		}), wantErr: "notification loop run id is required"},
		{name: "created at", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.CreatedAt = time.Time{} }), wantErr: "notification created time is required"},
		{name: "updated at", event: notificationEventWith(func(event *notifycore.NotificationEvent) { event.UpdatedAt = time.Time{} }), wantErr: "notification updated time is required"},
		{name: "delivered before created", event: notificationEventWith(func(event *notifycore.NotificationEvent) {
			event.Status = notifycore.NotificationStatusDelivered
			event.DeliveredAt = &earlier
		}), wantErr: "notification delivered before it was created"},
		{name: "delivered time status", event: notificationEventWith(func(event *notifycore.NotificationEvent) {
			event.Status = notifycore.NotificationStatusPending
			event.DeliveredAt = &now
		}), wantErr: "notification delivered time is invalid"},
		{name: "failure reason", event: notificationEventWith(func(event *notifycore.NotificationEvent) {
			event.Status = notifycore.NotificationStatusFailed
			event.FailureReason = ""
		}), wantErr: "notification failure reason is required"},
	}
}

func validNotificationEvent() notifycore.NotificationEvent {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)

	return notifycore.NotificationEvent{
		ID:        "notification-1",
		ProjectID: "project-1",
		Type:      notifycore.NotificationEventTicketBlocked,
		Severity:  notifycore.NotificationSeverityWarning,
		Status:    notifycore.NotificationStatusPending,
		Source:    notifycore.NotificationSourceTicketExecution,
		References: notifycore.NotificationReferences{
			TicketKey:   "CHR-1",
			ExecutionID: "execution-1",
		},
		Payload: notifycore.NotificationPayload{
			Summary: "Ticket CHR-1 is blocked",
			Detail:  "Validation failed after the configured retry threshold.",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func notificationEventWith(update func(*notifycore.NotificationEvent)) notifycore.NotificationEvent {
	event := validNotificationEvent()
	update(&event)

	return event
}
