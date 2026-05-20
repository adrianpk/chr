# Slice 1: Daemon Command Skeleton

Status: delivered
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-1-daemon-command-skeleton`
PR: #64

## Purpose

Add the user-facing daemon command group without introducing lifecycle
persistence or process ownership yet.

## Delivered Behavior

The CLI now includes:

- `chc daemon status`;
- `chc daemon start --dry-run`;
- `chc daemon stop --dry-run`.

`daemon status` resolves the user environment and reports the skeleton daemon as
stopped.

`daemon start` and `daemon stop` require `--dry-run` in this slice and print the
planned action plus user-level data and state directories.

## Implementation Notes

The command skeleton is intentionally side-effect free. It does not create
daemon state, acquire locks, start background work, or stop any process.

## Contracts Added or Changed

Added CLI command group:

- `daemon`

Added CLI commands:

- `daemon status`;
- `daemon start --dry-run`;
- `daemon stop --dry-run`.

## Files of Interest

- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./cmd/chc -run 'TestDaemon'`: pass
- `make verify`: pass

## Risks and Follow-ups

Lifecycle state, process ownership, and real start or stop behavior are
deliberately deferred to later daemon coordination slices.
