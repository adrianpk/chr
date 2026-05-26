# Slice 8: Compact Board Columns

Status: delivered
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-8-compact-board-columns`
PR: #220

## Purpose

Turn the compact Board view into a one-column-at-a-time operational surface.

## Delivered Behavior

- Added active normal Board column state to the TUI model.
- Board defaults to the first normal column with tickets, or `backlog` when no
  normal tickets exist.
- Left/right and h/l navigate through normal Board columns.
- Compact Board ticket rows are filtered to the active column.
- Up/down selection operates inside the active column.
- Changing columns selects the first visible ticket in that column.
- Board header renders a human-readable column label and position.
- Ticket detail continues to render in the right pane for the selected ticket.
- Board key helper text now advertises column navigation.

## Implementation Notes

- This slice intentionally leaves attention outside the normal column cycle.
- Maximized multi-column Board rendering remains unchanged and belongs to a
  later slice.
- Ticket list view still shows all tickets; only Board applies active-column
  filtering.

## Contracts Added or Changed

- `Model.boardColumn`
- `TicketView.BoardColumn`
- Board navigation keys: `left`, `right`, `h`, `l`

## Files of Interest

- `internal/chctui/model.go`
- `internal/chctui/model_test.go`

## Validation

- `go test ./internal/chctui`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- Empty normal columns render an empty-column message until attention handling
  exists.
- Attention tickets are intentionally hidden from compact normal navigation
  until Slice 9 adds the virtual attention view.
