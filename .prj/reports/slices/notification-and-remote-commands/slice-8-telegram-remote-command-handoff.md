# Slice 8: Telegram Remote Command Handoff

Status: in review
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-8-telegram-remote-command-handoff`
PR: pending

## Purpose

Connect Telegram update input to the bounded remote command request model
without adding a new execution engine.

## Delivered Behavior

Telegram `getUpdates` fixture payloads can now be decoded into remote command
records. Supported command text becomes an accepted record with a CLI-equivalent
handoff command. Unsupported command text becomes a rejected record with a
rejection reason.

The CLI adds:

```sh
chc notify telegram poll --fixture updates.json --dry-run
chc notify telegram poll --fixture updates.json
```

Dry-run prints parsed records without persistence. Non-dry-run persists accepted
and rejected records in user-level SQLite state.

## Contracts Added or Changed

- Added Telegram update fixture decoding.
- Added Telegram update to remote command request conversion.
- Added accepted/rejected remote command records.
- Added remote command record repository and SQLite migration.
- Added CLI-equivalent handoff text for accepted commands.
- Added `notify telegram poll --fixture`.

## Files of Interest

- `internal/chc/telegram_remote_command.go`
- `internal/chc/remote_command_record.go`
- `internal/chc/remote_command_record_repository.go`
- `internal/chcapp/notify.go`
- `cmd/chc/assets/migration/sqlite/20260521000007-create_remote_command_records.sql`
- `docs/reference/remote-commands.md`

## Validation

- `go test ./internal/chc -run 'Test.*(RemoteCommand|Telegram|Migration)'`: pass
- `go test ./internal/chcapp -run 'TestNotifyTelegramPoll'`: pass
- `make verify`: pass

## Risks and Follow-ups

- This slice intentionally uses fixtures and does not perform live Telegram
  polling.
- Persisted handoff records do not execute commands automatically.
  Command execution from accepted records should remain a separate, explicit
  boundary if added later.
