# Slice 3: Notification Audit Repository

Status: delivered
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-3-notification-audit-repository`
PR: `#110`

## Purpose

Persist notification events and dispatch outcomes in user-level state before
adding manual notification commands or Telegram transport behavior.

## Delivered Behavior

The core package now has a notification audit repository backed by SQLite.

The repository can:

- create notification audit records;
- update notification audit records;
- get one notification audit record by ID;
- list notification audit records for a project ordered by creation time;
- return a typed not-found error;
- validate notification records before writing.

The audit record keeps the core notification event channel-neutral while adding
an optional `channel` field for later dispatch slices.

## Implementation Notes

The migration creates `notification_events` in the user-level SQLite database.
The table stores event identity, project, type, severity, status, source,
channel, references, summary, detail, failure reason, and lifecycle timestamps.

The repository follows the same shape as the execution and loop repositories:

- construction through `UserEnvironment.DataDir()`;
- migration through the shared migrator;
- single SQLite connection;
- focused create/update/get/list methods.

## Contracts Added or Changed

- Added `NotificationAuditRecord`.
- Added `NotificationEventRepository`.
- Added `NewNotificationEventRepository`.
- Added `NewNotificationEventRepositoryFromDB`.
- Added `ErrNotificationEventNotFound`.
- Added SQLite migration `20260521000006-create_notification_events.sql`.
- Updated notification reference documentation with audit record behavior.

## Files of Interest

- `cmd/chc/assets/migration/sqlite/20260521000006-create_notification_events.sql`
- `internal/chc/notification_event_repository.go`
- `internal/chc/notification_event_repository_test.go`
- `internal/chc/migration_test.go`
- `docs/reference/notifications.md`

## Validation

- `go test ./internal/chc -run 'TestNotificationEventRepository|TestNewNotificationEventRepository|TestMigrator'`: pass
- `make verify`: pass

## Risks and Follow-ups

- This slice does not dispatch notifications.
- This slice does not add a CLI command surface.
- Channel-specific delivery metadata remains intentionally minimal until the
  Telegram dispatch boundary is implemented.
