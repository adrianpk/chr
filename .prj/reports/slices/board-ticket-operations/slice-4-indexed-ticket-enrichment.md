# Slice 4: Indexed Ticket Enrichment

Status: reviewing
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-4-indexed-ticket-enrichment`
PR: pending

## Purpose

Wire enriched Board ticket metadata through the parser, reconciliation path, and
SQLite ticket index repository.

## Delivered Behavior

- `IndexedTicket` includes `BoardColumn`.
- Ticket parsing normalizes deterministic Board columns from ticket status.
- SQLite upsert/list/get reads and writes synopsis, Board column, source
  metadata, and slice metadata.
- Legacy tickets without synopsis remain indexable.
- Blocked and unknown statuses stay outside deterministic Board column mapping,
  preserving attention-ready behavior for later slices.
- Index tests assert enriched metadata round-trips through SQLite.

## Implementation Notes

- Unknown statuses are still rejected by existing ticket validation. Attention
  for invalid external Markdown continues to flow through existing parse-error
  blockers until the attention view adds explicit presentation.
- `blocked` remains a valid status but does not normalize to a normal Board
  column.
- This slice does not introduce Board query APIs. Slice 7 owns the shared Board
  read model.

## Contracts Added or Changed

- `IndexedTicket.BoardColumn`
- `TicketIndex.UpsertTicket` persists Board metadata columns.
- `TicketIndex.ListProjectTickets` and `GetProjectTicketByKey` scan Board
  metadata columns.

## Files of Interest

- `internal/chc/indexed_ticket.go`
- `internal/chc/ticket_parser.go`
- `internal/chc/ticket_index.go`
- `internal/chc/ticket_index_test.go`

## Validation

- `go test ./internal/chc -run 'TestTicketIndex|TestTicketParser|TestTicketIndexRebuild|TestTicketReconciliation'`
- `go test ./internal/chc -run 'TestTicketIndex'`

## Risks and Follow-ups

- Existing CLI output does not display the new metadata yet.
- Existing TUI Board still groups by raw status until later Board query and UI
  slices.
