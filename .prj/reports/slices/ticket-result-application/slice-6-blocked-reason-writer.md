# Slice 6: Blocked Reason Writer

Status: delivered
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-6-blocked-reason-writer`
PR: `#46`

## Purpose

Record blocked or failure reasons from execution outcomes in ticket-visible
state.

## Delivered Behavior

The core package now exposes a blocked reason writer that prepends one entry to
the ticket `# Blocker` section from a `ResultApplication`.

The entry includes:

- execution ID;
- failure kind;
- failure reason;
- result artifact path;
- runtime log artifact path;
- validation artifact path.

Applications that do not target `blocked` are treated as no-op and do not
mutate the ticket.

## Implementation Notes

The writer updates `updated_at` when a blocker entry is written. Existing
blocker history is preserved and new entries are prepended.

## Contracts Added or Changed

New core types:

- `TicketResultBlockedReasonWriter`
- `WriteBlockedReasonOptions`

New function:

- `NewTicketResultBlockedReasonWriter`

## Files of Interest

- `internal/chc/result_blocked_reason_writer.go`
- `internal/chc/result_blocked_reason_writer_test.go`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

The blocker section format is intentionally compact Markdown. Result application
orchestration will decide when this writer runs with the other ticket writers.
