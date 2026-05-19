# Slice 4: Validation Evidence Writer

Status: delivered
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-4-validation-evidence-writer`
PR: `#44`

## Purpose

Record validation evidence from execution artifacts in ticket-visible state.

## Delivered Behavior

The core package now exposes a validation evidence writer that appends one entry
to the ticket `# Validation` section from a `ResultApplication`.

The entry includes:

- execution ID;
- validation command;
- result artifact path;
- runtime log artifact path;
- validation artifact path.

Existing validation history is preserved and new evidence is prepended.

## Implementation Notes

The writer updates `updated_at` when evidence is written. It does not write
comments, changed files, or blocked reason content.

## Contracts Added or Changed

New core types:

- `TicketResultValidationWriter`
- `WriteResultValidationOptions`

New function:

- `NewTicketResultValidationWriter`

## Files of Interest

- `internal/chc/result_validation_writer.go`
- `internal/chc/result_validation_writer_test.go`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

The validation section format is intentionally compact Markdown. Later slices
may enrich it if execution artifacts gain structured validation details.
