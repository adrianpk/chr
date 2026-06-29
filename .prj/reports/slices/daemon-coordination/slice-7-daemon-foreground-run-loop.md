# Slice 7: Daemon Foreground Run Loop

Status: delivered
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-7-daemon-foreground-run-loop`
PR: #70

## Purpose

Run daemon coordination as a supervised foreground process before OS service
integration exists.

## Delivered Behavior

`chr daemon start` now runs a bounded foreground loop.

The command:

- evaluates the daemon start guard;
- rejects fresh active daemon state;
- writes running daemon state with the current PID;
- updates daemon heartbeat state once per iteration;
- loads enabled registered projects;
- refreshes ticket indexes for ready projects;
- skips unavailable projects;
- stops after `--max-iterations`;
- writes stopped daemon state on normal completion.

## Implementation Notes

The default iteration limit is `1`. `--max-iterations` must be greater than
zero.

This slice uses index refresh as the foreground-safe loop boundary. It does not
execute ticket loops or observe stop requests yet.

## Contracts Added or Changed

Changed CLI behavior:

- `daemon start` no longer requires `--dry-run`.
- `daemon start --max-iterations N` bounds foreground daemon iterations.

## Files of Interest

- `cmd/chr/main.go`
- `cmd/chr/main_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./cmd/chr -run 'TestDaemonStart'`: pass
- `make verify`: pass

## Risks and Follow-ups

Graceful stop requests and persisted stale recovery remain Slice 8.
