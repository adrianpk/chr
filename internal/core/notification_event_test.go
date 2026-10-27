package core

import (
	"strings"
	"testing"
	"time"

	notifycore "forge.adrianpk.com/chr/internal/core/notify"
)

func TestValidateNotificationEventAcceptsValidEvent(t *testing.T) {
	event := validNotificationEvent()

	err := notifycore.ValidateEvent(event)
	if err != nil {
		t.Fatalf("ValidateNotificationEvent returned error: %v", err)
	}
}

func TestValidateNotificationEventRejectsInvalidEvents(t *testing.T) {
	for _, test := range invalidNotificationEventTests() {
		t.Run(test.name, func(t *testing.T) {
			err := notifycore.ValidateEvent(test.event)
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
	for _, eventType := range []notifycore.EventType{
		notifycore.EventTicketBlocked,
		notifycore.EventValidationFailed,
		notifycore.EventLoopFinished,
		notifycore.EventLoopBlocked,
		notifycore.EventLoopFailed,
		notifycore.EventDaemonAttention,
		notifycore.EventRemoteCommand,
		notifycore.EventManualTest,
	} {
		if !eventType.Valid() {
			t.Fatalf("NotificationEventType.Valid() = false for %s", eventType)
		}
	}

	for _, severity := range []notifycore.Severity{
		notifycore.SeverityInfo,
		notifycore.SeverityWarning,
		notifycore.SeverityError,
		notifycore.SeverityCritical,
	} {
		if !severity.Valid() {
			t.Fatalf("NotificationSeverity.Valid() = false for %s", severity)
		}
	}

	for _, status := range []notifycore.Status{
		notifycore.StatusPending,
		notifycore.StatusSkipped,
		notifycore.StatusDelivered,
		notifycore.StatusFailed,
	} {
		if !status.Valid() {
			t.Fatalf("NotificationStatus.Valid() = false for %s", status)
		}
	}

	for _, source := range []notifycore.Source{
		notifycore.SourceTicketExecution,
		notifycore.SourceTicketLoop,
		notifycore.SourceDaemon,
		notifycore.SourceRemoteCommand,
		notifycore.SourceManual,
	} {
		if !source.Valid() {
			t.Fatalf("NotificationSource.Valid() = false for %s", source)
		}
	}
}

func invalidNotificationEventTests() []struct {
	name    string
	event   notifycore.Event
	wantErr string
} {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	earlier := now.Add(-time.Minute)

	return []struct {
		name    string
		event   notifycore.Event
		wantErr string
	}{
		{name: "id", event: notificationEventWith(func(event *notifycore.Event) { event.ID = "" }), wantErr: "notification event id is required"},
		{name: "project id", event: notificationEventWith(func(event *notifycore.Event) { event.ProjectID = "" }), wantErr: "notification event project id is required"},
		{name: "type", event: notificationEventWith(func(event *notifycore.Event) { event.Type = notifycore.EventType("unknown") }), wantErr: "notification event type is invalid"},
		{name: "severity", event: notificationEventWith(func(event *notifycore.Event) { event.Severity = notifycore.Severity("unknown") }), wantErr: "notification severity is invalid"},
		{name: "status", event: notificationEventWith(func(event *notifycore.Event) { event.Status = notifycore.Status("unknown") }), wantErr: "notification status is invalid"},
		{name: "source", event: notificationEventWith(func(event *notifycore.Event) { event.Source = notifycore.Source("unknown") }), wantErr: "notification source is invalid"},
		{name: "summary", event: notificationEventWith(func(event *notifycore.Event) { event.Payload.Summary = "" }), wantErr: "notification summary is required"},
		{name: "daemon pid", event: notificationEventWith(func(event *notifycore.Event) { event.References.DaemonPID = -1 }), wantErr: "notification daemon pid cannot be negative"},
		{name: "ticket reference", event: notificationEventWith(func(event *notifycore.Event) { event.References.TicketKey = "" }), wantErr: "notification ticket key is required"},
		{name: "execution reference", event: notificationEventWith(func(event *notifycore.Event) {
			event.Type = notifycore.EventValidationFailed
			event.References.TicketKey = ""
			event.References.ExecutionID = ""
		}), wantErr: "notification execution id is required"},
		{name: "loop reference", event: notificationEventWith(func(event *notifycore.Event) {
			event.Type = notifycore.EventLoopFailed
			event.References.TicketKey = ""
			event.References.LoopRunID = ""
		}), wantErr: "notification loop run id is required"},
		{name: "created at", event: notificationEventWith(func(event *notifycore.Event) { event.CreatedAt = time.Time{} }), wantErr: "notification created time is required"},
		{name: "updated at", event: notificationEventWith(func(event *notifycore.Event) { event.UpdatedAt = time.Time{} }), wantErr: "notification updated time is required"},
		{name: "delivered before created", event: notificationEventWith(func(event *notifycore.Event) {
			event.Status = notifycore.StatusDelivered
			event.DeliveredAt = &earlier
		}), wantErr: "notification delivered before it was created"},
		{name: "delivered time status", event: notificationEventWith(func(event *notifycore.Event) {
			event.Status = notifycore.StatusPending
			event.DeliveredAt = &now
		}), wantErr: "notification delivered time is invalid"},
		{name: "failure reason", event: notificationEventWith(func(event *notifycore.Event) {
			event.Status = notifycore.StatusFailed
			event.FailureReason = ""
		}), wantErr: "notification failure reason is required"},
	}
}

func validNotificationEvent() notifycore.Event {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)

	return notifycore.Event{
		ID:        "notification-1",
		ProjectID: "project-1",
		Type:      notifycore.EventTicketBlocked,
		Severity:  notifycore.SeverityWarning,
		Status:    notifycore.StatusPending,
		Source:    notifycore.SourceTicketExecution,
		References: notifycore.References{
			TicketKey:   "CHR-1",
			ExecutionID: "execution-1",
		},
		Payload: notifycore.Payload{
			Summary: "Ticket CHR-1 is blocked",
			Detail:  "Validation failed after the configured retry threshold.",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func notificationEventWith(update func(*notifycore.Event)) notifycore.Event {
	event := validNotificationEvent()
	update(&event)

	return event
}
