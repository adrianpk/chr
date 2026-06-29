# Slice 5: Dash Realtime Progress

Status: reviewing

## Purpose

Keep Dash agent-work progress current while work runs by reloading persisted
ticket and observability state on a small polling cadence.

## Delivered

- Added a Dash agent-work refresh executor contract.
- Added a two-second Dash refresh tick.
- Added a TUI app adapter that reloads tickets and observability from persisted
  ticket, execution, loop, daemon, and runtime state.
- Wired refresh completion into the model so Dash updates queue, current work,
  and recent outcome state.
- Added tests for Dash refresh, non-Dash no-op refresh behavior, and running,
  failed, and blocked render states.
- Updated TUI documentation and CLI/TUI action matrix notes.

## Validation

- `go test ./internal/chrtui ./internal/chrapp`
- `make verify`
- `git diff --check`
