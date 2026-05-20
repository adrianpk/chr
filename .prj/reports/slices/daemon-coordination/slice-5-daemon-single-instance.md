# Slice 5: Single-Instance Guard

Status: delivered
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-5-daemon-single-instance`
PR: #68

## Purpose

Prevent duplicate daemon ownership for one user environment.

## Delivered Behavior

The core package now includes a daemon start guard.

The guard:

- allows start when daemon state is missing, stopped, failed, or stale;
- rejects start when a fresh daemon state is starting, running, or stopping;
- classifies active daemon state as stale when the heartbeat or start time is
  older than the stale threshold.

`chc daemon start --dry-run` now prints the guard decision without writing
state or acquiring ownership.

## Implementation Notes

The default stale threshold is five minutes.

This slice does not mutate stale daemon state. It only exposes the decision that
later ownership and recovery slices can use.

## Contracts Added or Changed

Added core guard:

- `EvaluateDaemonStartGuard`;
- `DaemonStartGuardDecision`;
- `DefaultDaemonStaleAfter`.

Changed CLI dry-run output:

- `daemon start --dry-run` includes `allowed`, `stale`, `reason`, and current
  daemon status fields when state exists.

## Files of Interest

- `internal/chc/daemon_guard.go`
- `internal/chc/daemon_guard_test.go`
- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./internal/chc -run 'TestEvaluateDaemonStartGuard'`: pass
- `go test ./cmd/chc -run 'TestDaemonStart'`: pass
- `make verify`: pass

## Risks and Follow-ups

Real ownership acquisition and persisted stale recovery remain deferred to later
daemon coordination slices.
