# Slice 1: Notification Event Model

Status: reviewing
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-1-notification-event-model`
PR: `#108`

## Purpose

Define the channel-neutral notification event contract before adding Telegram,
email, audit persistence, remote commands, or CLI dispatch behavior.

## Delivered Behavior

The core package now exposes a notification event model with:

- event type, severity, status, and source enums;
- event identity and project ownership;
- optional ticket, execution, loop, and daemon references;
- channel-neutral summary and detail payload fields;
- validation for required fields, references, lifecycle timestamps, and failed
  dispatch reasons.

## Implementation Notes

The event model deliberately avoids channel-specific data. Telegram chat IDs,
email recipients, transport response IDs, and delivery attempts belong to later
configuration, dispatch, or audit records.

Reference validation is tied to event type where the relationship is required:

- blocked-ticket events require a ticket key;
- validation-failed events require an execution ID;
- loop outcome events require a loop run ID.

## Contracts Added or Changed

- Added `NotificationEvent`.
- Added `NotificationReferences`.
- Added `NotificationPayload`.
- Added `NotificationEventType`.
- Added `NotificationSeverity`.
- Added `NotificationStatus`.
- Added `NotificationSource`.
- Added `ValidateNotificationEvent`.
- Added `docs/reference/notifications.md`.

## Files of Interest

- `internal/chc/notification_event.go`
- `internal/chc/notification_event_test.go`
- `docs/reference/notifications.md`
- `docs/index.md`

## Validation

- `go test ./internal/chc -run TestNotification`: pass
- `make verify`: pass

## Risks and Follow-ups

- Notification persistence is not part of this slice.
- Channel configuration and Telegram dispatch are planned for later slices.
- Remote command intake is planned for later slices.
