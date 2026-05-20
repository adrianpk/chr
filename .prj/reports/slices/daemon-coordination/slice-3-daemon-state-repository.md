# Slice 3: Daemon State Repository

Status: delivered
Slicing: Daemon Coordination
Plan: `.prj/plans/slicing-daemon-coordination-plan.md`
Tracker: `.prj/trackers/slicing-daemon-coordination-tracker.md`
Branch: `feat/slice-3-daemon-state-repository`
PR: #66

## Purpose

Persist daemon lifecycle state in the user-level SQLite database.

## Delivered Behavior

The core package now includes a daemon state repository backed by the user
database.

The repository supports:

- reading current daemon state;
- upserting daemon state;
- updating daemon heartbeat time;
- clearing daemon state.

## Implementation Notes

Daemon state is stored as a singleton row for the local user environment.

This slice does not start, stop, signal, or inspect OS processes. It only
persists lifecycle state for later daemon ownership and status slices.

## Contracts Added or Changed

Added SQLite migration:

- `20260520000005-create_daemon_state.sql`

Added core repository:

- `DaemonStateRepository`;
- `NewDaemonStateRepository`;
- `NewDaemonStateRepositoryFromDB`;
- `ErrDaemonStateNotFound`.

## Files of Interest

- `cmd/chc/assets/migration/sqlite/20260520000005-create_daemon_state.sql`
- `internal/chc/daemon_state_repository.go`
- `internal/chc/daemon_state_repository_test.go`
- `internal/chc/migration_test.go`
- `docs/reference/daemon-coordination.md`

## Validation

- `go test ./internal/chc -run 'TestDaemonStateRepository|TestMigrator'`: pass
- `make verify`: pass

## Risks and Follow-ups

The CLI skeleton still reports static daemon state. Slice 4 connects
`daemon status` to this repository.
