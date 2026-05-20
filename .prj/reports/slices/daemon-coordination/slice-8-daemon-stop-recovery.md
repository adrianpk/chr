# Slice 8: Graceful Stop and Stale Recovery

Status: in review
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-8-daemon-stop-recovery`
PR: #71

## Purpose

Let users request daemon stop and recover stale daemon lifecycle state.

## Delivered Behavior

`chc daemon stop` now persists stop requests for active daemon state.

The foreground run loop observes persisted stop requests between iterations and
exits cleanly. Normal foreground exit writes stopped daemon state.

`chc daemon status` recovers stale active daemon state by rewriting it as
`stale` with a stale failure reason.

## Implementation Notes

Stop requests apply to `starting`, `running`, and `stopping` state.

Stale recovery uses the existing daemon start guard stale threshold. The status
command remains the recovery trigger for persisted stale classification.

## Contracts Added or Changed

Changed repository behavior:

- `DaemonStateRepository.RequestDaemonStop` persists stop requests.

Changed CLI behavior:

- `daemon stop` no longer requires `--dry-run`;
- `daemon status` rewrites stale active state as `stale`;
- foreground daemon iterations stop when stop requests are present.

## Files of Interest

- `internal/chc/daemon_state_repository.go`
- `internal/chc/daemon_state_repository_test.go`
- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./cmd/chc -run 'TestDaemonStatus|TestDaemonStop|TestDaemonStart'`: pass
- `go test ./internal/chc -run 'TestDaemonStateRepository'`: pass
- `make verify`: pass

## Risks and Follow-ups

This completes the planned daemon coordination slicing. Notification, remote
command, and OS service integration remain outside this slicing.
