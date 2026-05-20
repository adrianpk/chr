# Slice 3: Loop Ordering

Status: in review
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-3-loop-ordering`
PR: pending

## Purpose

Make eligible ticket ordering explicit and deterministic for later loop
execution.

## Delivered Behavior

`TicketLoopQueryOptions` now accepts a `TicketLoopOrder`.

Supported ordering values:

- `key`;
- `created_at`;
- `priority`;
- `severity`;
- `priority_severity`.

The default order is `key`.

Priority and severity ordering rank ticket metadata from highest to lowest:

```text
critical
high
normal
low
```

Ordering remains deterministic. Priority and severity order use `created_at` as
their first tie-breaker, and all modes eventually tie-break by ticket key.

## Implementation Notes

Ordering is applied after status eligibility and metadata filters. CLI parsing
is still out of scope for this slice.

## Contracts Added or Changed

New core type:

- `TicketLoopOrder`

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

Slice 4 wires ordering into the foreground loop command. This slice only
defines the core ordering contract.
