# Slice 5: Telegram Dispatch Boundary

Status: delivered
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-5-telegram-dispatch-boundary`
PR: `#112`

## Purpose

Add the first concrete notification channel boundary for sending notification
events through Telegram.

## Delivered Behavior

The core package now includes:

- a Telegram message renderer for notification events;
- a Telegram sender with injectable HTTP client;
- Telegram input validation for token, chat ID, and message text;
- HTTP error reporting with response status capture.

The CLI now supports non-dry-run:

```sh
chr notify telegram test
```

The command:

- validates notification config;
- resolves the bot token from the configured environment variable;
- creates a notification audit record;
- sends the Telegram test message;
- updates the audit record to `delivered` or `failed`;
- prints the notification ID and HTTP status when delivery succeeds.

## Implementation Notes

Tests use `httptest` and an injectable `TelegramSender`, so no real Telegram
credentials or network calls are required.

The notification event remains channel-neutral. Telegram-specific behavior lives
in the sender and in the audit record channel field.

## Contracts Added or Changed

- Added `TelegramMessage`.
- Added `TelegramSender`.
- Added `TelegramSendResult`.
- Added `TelegramHTTPClient`.
- Added `RenderTelegramNotificationMessage`.
- Added non-dry-run `chr notify telegram test`.
- Updated notification command documentation.

## Files of Interest

- `internal/chr/telegram_notification.go`
- `internal/chr/telegram_notification_test.go`
- `internal/chrapp/notify.go`
- `internal/chrapp/cli_test.go`
- `docs/reference/notifications.md`
- `docs/reference/commands.md`

## Validation

- `go test ./internal/chr -run TestTelegram`: pass
- `go test ./internal/chrapp -run TestNotify`: pass
- `make verify`: pass

## Risks and Follow-ups

- This slice only dispatches the manual Telegram setup-check message.
- Workflow-triggered loop and ticket notifications are planned for Slice 6.
- Remote command intake is planned for later slices.
