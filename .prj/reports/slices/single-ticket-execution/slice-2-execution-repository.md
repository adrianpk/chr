# Slice 2: Execution Repository

Status: delivered
Slicing: Single Ticket Execution
Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
Branch: `feat/slice-2-execution-repository`
PR: `#35`

## Purpose

Persist execution records so later `ticket start` slices can durably record one
ticket execution attempt.

## Delivered Behavior

The user-level SQLite database now has an `execution_records` table. The core
package exposes an execution repository that can create, update, fetch, and list
execution records for one project ticket.

## Implementation Notes

The repository follows the existing project registry and ticket index patterns:
it uses the shared migrator, opens the user-level `chr.db`, and provides a
`FromDB` constructor for focused tests.

## Contracts Added or Changed

New migration:

- `20260519000003-create_execution_records.sql`

New repository API:

- `NewExecutionRepository`
- `NewExecutionRepositoryFromDB`
- `CreateExecution`
- `UpdateExecution`
- `GetExecution`
- `ListTicketExecutions`

Lookup errors use `ErrExecutionNotFound`.

## Files of Interest

- `cmd/chr/assets/migration/sqlite/20260519000003-create_execution_records.sql`
- `internal/chr/execution_repository.go`
- `internal/chr/execution_repository_test.go`
- `internal/chr/migration_test.go`
- `docs/reference/ticket-execution.md`

## Validation

- `go test ./internal/chr`: pass
- `make verify`: pass

## Risks and Follow-ups

No CLI uses the repository yet. Slice 3 is planned to create pending execution
records from `chr ticket start` preflight.
