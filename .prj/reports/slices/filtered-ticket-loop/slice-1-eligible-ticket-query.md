# Slice 1: Eligible Ticket Query

Status: in review
Slicing: Filtered Ticket Loop
Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
Branch: `feat/slice-1-eligible-ticket-query`
PR: pending

## Purpose

Add the core read-only ticket selection boundary that later loop slices can use
to choose eligible tickets.

## Delivered Behavior

The core package now exposes `TicketLoopQuery`.

The query:

- requires a project ID;
- reads tickets from the existing `TicketIndex`;
- defaults eligible statuses to `ready`;
- supports explicit eligible statuses;
- validates explicit statuses with the existing ticket lifecycle status rules;
- returns tickets sorted by ticket key;
- does not mutate ticket files, execution records, or workspace state.

## Implementation Notes

This slice intentionally does not add CLI flags, metadata filters, ordering
policies, loop persistence, or execution. Those concerns are assigned to later
loop slices.

## Contracts Added or Changed

New core types:

- `TicketLoopQuery`
- `TicketLoopQueryOptions`

New function:

- `NewTicketLoopQuery`

## Files of Interest

- `internal/chc/ticket_loop_query.go`
- `internal/chc/ticket_loop_query_test.go`
- `docs/reference/ticket-loop.md`

## Validation

- `go test ./internal/chc -run TestTicketLoopQuery`: pass
- `make verify`: pass

## Risks and Follow-ups

The default eligible status is `ready`. Slice 2 adds richer filters, including
explicit status filters for callers that want to include `open` or other
lifecycle states.
