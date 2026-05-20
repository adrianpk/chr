# Slice 2: Daemon Lifecycle State Model

Status: delivered
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-2-daemon-lifecycle-state`
PR: #65

## Purpose

Define daemon lifecycle state before adding persistence or process ownership.

## Delivered Behavior

The core package now includes a daemon lifecycle model with statuses:

- `stopped`;
- `starting`;
- `running`;
- `stopping`;
- `stale`;
- `failed`.

The model tracks:

- process ID;
- start time;
- heartbeat time;
- stop request time;
- failure reason;
- update time.

## Implementation Notes

Validation keeps active process states explicit. `starting`, `running`, and
`stopping` states require a positive PID and start time.

Heartbeat and stop request times are validated against the daemon start time.
Stop requests are valid only for `stopping` or `stopped` state.

## Contracts Added or Changed

Added core model:

- `DaemonState`;
- `DaemonStatus`;
- `ValidateDaemonState`;
- `NewStoppedDaemonState`.

## Files of Interest

- `internal/chc/daemon_state.go`
- `internal/chc/daemon_state_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./internal/chc -run 'TestValidateDaemonState'`: pass
- `make verify`: pass

## Risks and Follow-ups

Persistence is intentionally deferred to Slice 3. The CLI skeleton does not use
this model until daemon state repository behavior exists.
