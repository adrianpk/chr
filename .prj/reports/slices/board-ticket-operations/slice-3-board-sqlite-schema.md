# Slice 3: Board SQLite Schema

Status: reviewing
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-3-board-sqlite-schema`
PR: #215

## Purpose

Add the SQLite schema foundation for Board ticket metadata and future
transaction-safe ticket key allocation.

## Delivered Behavior

- Added a migration for Board metadata fields on `indexed_tickets`.
- Added query indexes for Board column, use case, project area, updated
  timestamp, and source kind.
- Added `ticket_key_sequences` as the allocation foundation for future
  transaction-safe ticket key generation.
- Extended migration tests to assert the Board fields and sequence table exist.
- Updated the SQLite migration reference with the Board schema decision.

## Implementation Notes

- The schema extends `indexed_tickets` directly. A parallel Board read model was
  not introduced because `indexed_tickets` is already the operational ticket
  metadata index used by CLI and TUI flows.
- This slice adds schema only. Repository read/write wiring for the new fields
  remains in Slice 4.
- The key sequence table is schema foundation only. Ticket creation still uses
  the existing allocation behavior until Slice 5 adopts the new boundary.

## Contracts Added or Changed

- `indexed_tickets.synopsis`
- `indexed_tickets.board_column`
- `indexed_tickets.source_kind`
- `indexed_tickets.source_ref`
- `indexed_tickets.slice`
- `ticket_key_sequences(project_id, prefix, next_number, updated_at)`

## Files of Interest

- `cmd/chc/assets/migration/sqlite/20260526000008-add_board_ticket_metadata.sql`
- `internal/chc/migration_test.go`
- `docs/reference/sqlite-migrations.md`

## Validation

- `go test ./internal/chc -run 'TestMigratorAppliesPendingMigrations|TestReadSQLiteMigrations'`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- Existing repository queries do not populate or scan the new columns yet.
  Slice 4 owns that wiring.
- `ticket_key_sequences` is not consumed yet. Slice 5 owns ticket creation
  adoption.
