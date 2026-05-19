# Slice 7: Markdown Index Consistency

Status: reviewing
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-7-markdown-index-consistency`
PR: pending

## Purpose

Refresh derived SQLite ticket state after result application mutates canonical
Markdown.

## Delivered Behavior

The core reconciler now supports affected-ticket reconciliation through
`ReconcileProjectTicket`.

The single-ticket refresh:

- parses one known ticket path;
- upserts the derived SQLite index row;
- returns parse failures as reconciliation blockers;
- preserves hard errors for storage failures.

## Implementation Notes

The existing project-wide reconciliation remains unchanged. Result application
callers can use the single-ticket reconciliation path after applying ticket
status, comment, validation, changed-file, or blocker writes.

## Contracts Added or Changed

New core type:

- `ReconcileProjectTicketOptions`

Updated core interface:

- `TicketReconciler`

## Files of Interest

- `internal/chc/ticket_reconciliation.go`
- `internal/chc/ticket_reconciliation_test.go`
- `docs/reference/ticket-reconciliation.md`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pending

## Risks and Follow-ups

The result apply CLI still needs to call this reconciliation path after it runs
the result application writers.
