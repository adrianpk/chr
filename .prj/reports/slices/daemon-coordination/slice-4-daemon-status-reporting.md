# Slice 4: Daemon Status Reporting

Status: in review
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-4-daemon-status-reporting`
PR: #67

## Purpose

Expose persisted daemon lifecycle state through `chc daemon status`.

## Delivered Behavior

`chc daemon status` now reads the daemon state repository.

When state is missing, the command reports:

- `daemon: stopped`;
- `state: missing`.

When state exists, the command reports:

- daemon status;
- `state: stored`;
- PID;
- started time;
- heartbeat time;
- stop request time;
- failure reason;
- updated time;
- user-level data and state directories.

## Implementation Notes

The status command remains read-only. It does not classify process liveness or
mutate stale state; that belongs to later daemon ownership and recovery slices.

## Contracts Added or Changed

Changed CLI behavior:

- `daemon status` reads persisted daemon lifecycle state when present.
- Missing daemon state is reported as stopped with `state: missing`.

## Files of Interest

- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./cmd/chc -run 'TestDaemonStatus'`: pass
- `make verify`: pass

## Risks and Follow-ups

Process liveness checks and stale-state mutation are intentionally deferred to
the single-instance and stale recovery slices.
