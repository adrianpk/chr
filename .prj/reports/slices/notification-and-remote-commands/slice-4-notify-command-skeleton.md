# Slice 4: Notify Command Skeleton

Status: delivered
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-4-notify-command-skeleton`
PR: `#111`

## Purpose

Expose setup-check notification commands without treating manual notification as
the primary workflow.

## Delivered Behavior

The CLI now includes a `notify` command group with:

- `chr notify status`;
- `chr notify telegram test --dry-run`.

`notify status` resolves the project, loads `.chr/config.yml`, and reports
Telegram notification readiness.

`notify telegram test --dry-run` renders the planned Telegram test payload from
project configuration without contacting Telegram.

Non-dry-run `notify telegram test` is intentionally blocked until the Telegram
dispatch boundary is implemented.

## Implementation Notes

The command skeleton reuses the existing project context resolver, so commands
can operate from `--dir`, active project selection, or the current managed
directory.

The dry-run Telegram command validates notification config before rendering the
payload. Invalid enabled Telegram config is surfaced before any transport layer
exists.

## Contracts Added or Changed

- Added `chr notify`.
- Added `chr notify status`.
- Added `chr notify telegram test --dry-run`.
- Added CLI tests for readiness reporting, invalid config reporting, dry-run
  payload rendering, and non-dry-run blocking.

## Files of Interest

- `internal/chrapp/app.go`
- `internal/chrapp/notify.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/commands.md`
- `docs/reference/notifications.md`

## Validation

- `go test ./internal/chrapp -run TestNotify`: pass
- `make verify`: pass

## Risks and Follow-ups

- This slice does not send Telegram messages.
- This slice does not create notification audit records for setup checks.
- The Telegram dispatch boundary is planned for Slice 5.
