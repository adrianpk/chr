# Slice 1: Ticket Metadata Contract

Status: reviewing
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-1-ticket-metadata-contract`
PR: pending

## Purpose

Establish the durable ticket metadata contract needed by later Board slices
before changing SQLite schema, ticket creation services, or TUI rendering.

## Delivered Behavior

- Board status constants now exist for the normal Board flow.
- Legacy ticket statuses remain valid.
- Legacy statuses can be mapped to normal Board columns when the mapping is
  deterministic.
- `Ticket` and `IndexedTicket` now expose synopsis, source metadata, and slice
  metadata.
- Ticket rendering writes the new frontmatter fields.
- Ticket parsing reads the new frontmatter fields into `IndexedTicket`.
- Validation tests cover Board-style tickets, legacy tickets without synopsis,
  and unknown status rejection.

## Implementation Notes

- `DefaultTicketStatus` still resolves to the legacy `open` status. Changing
  creation defaults to `backlog` belongs to Slice 5.
- `blocked` remains a valid legacy ticket status, but it does not map to a
  normal Board column. A later attention slice should surface it as
  attention-worthy.
- SQLite read/write code is intentionally unchanged in this slice. Schema and
  repository persistence changes are reserved for Slice 3 and Slice 4.
- New metadata fields render even when empty so generated Markdown has a stable
  frontmatter shape.

## Contracts Added or Changed

- Added normal Board statuses:
  - `backlog`
  - `ready_for_dev`
  - `in_progress`
  - `ready_for_review`
  - `ready_for_merge`
  - `done`
- Preserved legacy statuses:
  - `open`
  - `ready`
  - `reviewing`
  - `blocked`
  - `solved`
  - `merged`
  - `closed`
- Added `BoardTicketStatusFor` for deterministic normal-column mapping.
- Added ticket fields:
  - `Synopsis`
  - `SourceKind`
  - `SourceRef`
  - `Slice`

## Files of Interest

- `internal/chc/ticket.go`
- `internal/chc/ticket_status.go`
- `internal/chc/indexed_ticket.go`
- `internal/chc/ticket_parser.go`
- `internal/chc/ticket_test.go`
- `internal/chc/ticket_parser_test.go`
- `internal/chc/ticket_status_test.go`
- `internal/chc/ticket_validation_test.go`

## Validation

- `go test ./internal/chc -run 'TestTicketStatus|TestTicketRenderer|TestTicketParser'`
- `go test ./internal/chc -run 'TestTicketRenderer'`
- `go test ./internal/chc -run 'TestTicketParser'`
- `go test ./internal/chc -run 'TestValidateIndexedTicket|TestTicketStatus'`

## Risks and Follow-ups

- SQLite does not persist the new fields yet. That is planned in Slice 3 and
  Slice 4.
- Ticket creation still defaults to `open`. That is planned for Slice 5.
