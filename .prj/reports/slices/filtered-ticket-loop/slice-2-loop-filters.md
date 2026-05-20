# Slice 2: Loop Filters

Status: in review
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-2-loop-filters`
PR: pending

## Purpose

Extend the loop ticket query with metadata filters while keeping selection
independent from CLI parsing.

## Delivered Behavior

`TicketLoopQueryOptions` now accepts `TicketLoopFilters`.

The query can filter eligible tickets by:

- priority;
- severity;
- project area;
- tags;
- type;
- use case.

Priority, severity, and type filters use existing ticket metadata validation
rules. Project area, tag, and use-case filters are free-form exact matches.

Empty filter dimensions are ignored. Multiple values within one dimension match
as alternatives, except tags: each requested tag must be present on the ticket.

## Implementation Notes

Filtering runs after project/status eligibility and before deterministic key
sorting. The implementation remains read-only and continues to use the existing
ticket index boundary.

## Contracts Added or Changed

New core type:

- `TicketLoopFilters`

Changed core type:

- `TicketLoopQueryOptions`

## Files of Interest

- `internal/chc/ticket_loop_query.go`
- `internal/chc/ticket_loop_query_test.go`
- `docs/reference/ticket-loop.md`

## Validation

- `go test ./internal/chc -run TestTicketLoopQuery`: pass
- `make verify`: pass

## Risks and Follow-ups

CLI flag parsing is intentionally not part of this slice. Slice 4 wires the
query and filter options into the foreground loop command.
