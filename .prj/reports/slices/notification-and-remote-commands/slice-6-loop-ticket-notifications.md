# Slice 6: Loop and Ticket Notifications

Status: delivered
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-6-loop-ticket-notifications`
PR: `#113`

## Purpose

Record and optionally dispatch notification events for ticket and foreground
loop outcomes that require user visibility.

## Delivered Behavior

Ticket result application now records a `ticket_blocked` notification when an
applied execution result leaves the ticket blocked.

Foreground ticket loops now record notification events when:

- a loop finishes after processing at least one ticket;
- a loop stops on a blocked ticket result;
- a loop fails during ticket execution.

Workflow notifications include the managed project identifier and the relevant
ticket, execution, or loop references.

## Dispatch Boundary

Workflow notifications use the existing notification configuration:

- Telegram disabled: persist the audit record as `skipped`.
- Telegram enabled and valid: persist as `pending`, send through the configured
  Telegram sender, then update to `delivered` or `failed`.
- Telegram enabled but invalid: persist as `failed` with the configuration
  error.

Telegram dispatch failures are recorded on the notification audit event. They do
not replace the ticket or loop result.

## Contracts Added or Changed

- Added workflow notification emission from ticket result application.
- Added workflow notification emission from foreground loop completion.
- Added loop notification rendering for finished, blocked, and failed loops.
- Reused the notification audit repository and Telegram sender from previous
  slices.
- Documented workflow-triggered notification behavior.

## Files of Interest

- `internal/chcapp/notification_emit.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/cli_test.go`
- `docs/reference/notifications.md`
- `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`

## Validation

- `go test ./internal/chcapp -run 'TestTicketLoop(DispatchesConfiguredFinishedNotification|StopsOnExecutionFailure|RecordsFailedNotification|RecordsBlockedNotifications)|TestTicketApplyResultRecordsBlockedNotification'`: pass
- `make verify`: pass

## Risks and Follow-ups

- Loop-blocked notification emission is wired to the loop result path that stops
  on blocked ticket application. Current worker validation normally turns failed
  execution results into loop failures before result application.
- Remote command intake remains planned for Slice 7.
