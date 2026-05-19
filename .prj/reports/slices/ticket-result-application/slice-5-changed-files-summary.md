# Slice 5: Changed Files Summary

Status: reviewing
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-5-changed-files-summary`
PR: pending

## Purpose

Record changed-file summaries from execution results in ticket-visible state.

## Delivered Behavior

The core package now exposes a changed files summary writer that appends one
comment to the ticket `# Comments` section from a `ResultApplication`.

The comment includes:

- execution ID;
- changed file path;
- changed file summary when present.

Empty changed-file lists are treated as no-op and do not mutate the ticket.

## Implementation Notes

The writer uses the existing ticket comment writer, so it preserves comment
formatting and updates `updated_at` only when a summary is written.

## Contracts Added or Changed

New core types:

- `TicketResultChangedFilesWriter`
- `WriteChangedFilesOptions`

New functions:

- `NewTicketResultChangedFilesWriter`
- `RenderResultApplicationChangedFilesBody`

## Files of Interest

- `internal/chc/result_changed_files_writer.go`
- `internal/chc/result_changed_files_writer_test.go`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pending

## Risks and Follow-ups

The changed-file summary uses comment text instead of a dedicated ticket
section. This keeps the ticket schema unchanged for the current slice.
