# Slice 2: Notification Configuration

Status: reviewing
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-2-notification-config`
PR: `#109`

## Purpose

Add project-local notification configuration before introducing audit
persistence, Telegram dispatch, or remote command intake.

## Delivered Behavior

Workspace configuration now includes a notification section with Telegram
settings:

- Telegram is disabled by default.
- Telegram can be enabled through `.chc/config.yml`.
- Bot tokens are referenced by environment variable name, not stored in the
  config file.
- Enabled Telegram config requires both token environment variable name and chat
  ID.
- Validation checks local config shape only and does not contact external
  services.

## Implementation Notes

The existing config loader is a small section-based parser, so this slice keeps
the notification shape flat under `notifications:`. The supported Telegram keys
are:

- `telegram_enabled`
- `telegram_bot_token_env`
- `telegram_chat_id`

This avoids introducing a new YAML parser or changing existing runtime and
ticket configuration behavior.

## Contracts Added or Changed

- Added `NotificationConfig`.
- Added `TelegramNotificationConfig`.
- Added `DefaultNotificationConfig`.
- Added `ValidateNotificationConfig`.
- Extended `WorkspaceConfig` with `Notifications`.
- Added `docs/reference/notification-config.md`.

## Files of Interest

- `internal/chc/config.go`
- `internal/chc/config_test.go`
- `internal/chc/notification_config.go`
- `docs/reference/notification-config.md`
- `docs/reference/project-workspace.md`
- `docs/index.md`

## Validation

- `go test ./internal/chc -run 'TestLoadWorkspaceConfig|TestValidateNotificationConfig'`: pass
- `make verify`: pass

## Risks and Follow-ups

- This slice does not dispatch notifications.
- This slice does not persist notification audit records.
- Telegram token environment lookup is deferred to the dispatch boundary.
