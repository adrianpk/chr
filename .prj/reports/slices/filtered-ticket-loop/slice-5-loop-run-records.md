# Slice 5: Loop Run Records

Status: in review
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-5-loop-run-records`
PR: `#60`

## Purpose

Persist foreground loop run metadata separately from per-ticket execution
records.

## Delivered Behavior

The core package now has a loop run model and SQLite repository.

Loop run records track:

- loop run ID;
- project ID;
- lifecycle status;
- eligible, processed, blocked, and failed counts;
- failure reason;
- created, started, finished, and updated timestamps.

Supported statuses:

- `pending`;
- `running`;
- `succeeded`;
- `failed`;
- `blocked`.

## Implementation Notes

The repository follows the existing execution repository pattern:

- `CreateLoopRun`;
- `UpdateLoopRun`;
- `GetLoopRun`;
- `ListProjectLoopRuns`.

This slice does not create loop runs from the CLI yet. Slice 6 wires run records
into repeated ticket execution.

## Contracts Added or Changed

New core types:

- `LoopRunStatus`
- `LoopRunRecord`
- `LoopRunRepository`

New repository constructors:

- `NewLoopRunRepository`
- `NewLoopRunRepositoryFromDB`

New migration:

- `20260520000004-create_loop_runs.sql`

## Files of Interest

- `internal/chc/loop_run.go`
- `internal/chc/loop_run_repository.go`
- `internal/chc/loop_run_repository_test.go`
- `cmd/chc/assets/migration/sqlite/20260520000004-create_loop_runs.sql`
- `docs/reference/ticket-loop.md`

## Validation

- `go test ./internal/chc -run 'TestLoopRun|TestNewLoopRun|TestMigrator'`:
  pass
- `make verify`: pass

## Risks and Follow-ups

Loop run records are not linked to execution records yet. Slice 6 adds repeated
ticket execution and iteration summaries.
