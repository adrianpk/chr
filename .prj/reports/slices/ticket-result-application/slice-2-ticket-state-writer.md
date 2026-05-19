# Slice 2: Ticket State Writer

Status: delivered
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-2-ticket-state-writer`
PR: `#42`

## Purpose

Update ticket Markdown status from one result application.

## Delivered Behavior

The core package now exposes a ticket result state writer that applies the
ticket-facing status computed by `ResultApplication`.

The writer updates:

- `status`;
- `updated_at`.

It preserves existing metadata, comments, and body sections.

## Implementation Notes

The writer delegates frontmatter mutation to the existing ticket status updater.
This keeps status validation and Markdown update behavior consistent with
`chc ticket status`.

This slice intentionally writes only state. Comments, validation evidence,
changed files, and blocked reasons are handled by later slices.

## Contracts Added or Changed

New core types:

- `TicketResultStateWriter`
- `WriteResultStateOptions`

New function:

- `NewTicketResultStateWriter`

## Files of Interest

- `internal/chc/result_state_writer.go`
- `internal/chc/result_state_writer_test.go`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

No CLI invokes the writer yet. Slice 8 is planned to expose result application
through `chc ticket apply-result`.
