# Slice 7: Board Query Service

Status: delivered
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-7-board-query-service`
PR: #219

## Purpose

Introduce a shared Board read model so CLI and TUI ticket views use the same
query boundary before compact Board navigation is implemented.

## Delivered Behavior

- Added a core `BoardQuery` over the SQLite ticket index.
- Board results include normal columns, counts, ordered tickets, total count,
  normal count, and attention count.
- Normal Board columns are returned in deterministic status order.
- Legacy statuses are mapped through the existing Board status alias logic.
- Tickets outside normal Board columns are counted as attention tickets and are
  kept after normal tickets in the flattened ticket list.
- Added `ticket board` as a scriptable CLI inspection command.
- TUI ticket loading now uses the service Board read model instead of directly
  listing indexed tickets.
- Existing service ticket reload paths now reuse the Board read model.

## Implementation Notes

- This slice does not change compact Board navigation or visual column
  rendering. Later slices own those UI changes.
- Attention is counted but not exposed as a navigable Board view yet.
- The flattened `Board.Tickets` list preserves existing TUI behavior while
  making Board column data available for later UI work.

## Contracts Added or Changed

- `BoardQuery`
- `BoardQueryOptions`
- `Board`
- `BoardColumn`
- `NormalBoardColumns`
- `BoardColumnTitle`
- `service.Board`
- CLI command: `ticket board`

## Files of Interest

- `internal/chc/board_query.go`
- `internal/chc/board_query_test.go`
- `internal/chcapp/service_artifact_ticket.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/app.go`

## Validation

- `go test ./internal/chc ./internal/chcapp ./internal/chctui`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- Board attention reasons are still not modeled; Slice 9 owns that view.
- TUI Board still renders grouped text from `TicketView` until compact and
  maximized Board slices consume the column model directly.
