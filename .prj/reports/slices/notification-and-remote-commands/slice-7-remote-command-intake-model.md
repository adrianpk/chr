# Slice 7: Remote Command Intake Model

Status: in review
Slicing: Notification and Remote Commands
Plan: `.prj/plans/slicing-notification-and-remote-commands-plan.md`
Tracker: `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`
Branch: `feat/slice-7-remote-command-intake-model`
PR: pending

## Purpose

Define a bounded, channel-neutral remote command request contract before adding
Telegram update polling or command handoff.

## Delivered Behavior

The core package now includes a remote command request model with:

- sender channel and external sender identity;
- optional project selector;
- supported command intent;
- ticket key and comment fields for ticket commands;
- loop filters and max iteration limit for loop commands;
- raw text and timestamps.

The parser accepts a small command grammar:

```text
status
@project status
loop start --status ready --priority high --max-iterations 1
ticket start CHC-12
ticket comment CHC-12 blocked on CI logs
```

## Safety Boundary

The parser rejects unknown roots, unknown ticket actions, positional loop
arguments, unknown loop flags, missing sender identity, missing timestamps, and
missing required ticket/comment arguments.

The model does not execute commands and does not expose shell access. Later
channel-specific slices can persist requests and map accepted intents to
existing local command behavior.

## Contracts Added or Changed

- Added `RemoteCommandRequest`.
- Added `RemoteCommandSender`.
- Added `RemoteCommandLoopOptions`.
- Added `RemoteCommandIntent`.
- Added `ParseRemoteCommandRequest`.
- Added `ValidateRemoteCommandRequest`.
- Added remote command reference documentation.

## Files of Interest

- `internal/chc/remote_command.go`
- `internal/chc/remote_command_test.go`
- `docs/reference/remote-commands.md`
- `docs/reference/notifications.md`
- `docs/index.md`
- `.prj/trackers/slicing-notification-and-remote-commands-tracker.md`

## Validation

- `go test ./internal/chc -run 'Test.*RemoteCommand'`: pass
- `make verify`: pass

## Risks and Follow-ups

- This slice does not persist remote command requests.
- This slice does not parse Telegram update payloads.
- Telegram command handoff remains planned for Slice 8.
