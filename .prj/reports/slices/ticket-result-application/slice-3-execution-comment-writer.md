# Slice 3: Execution Comment Writer

Status: delivered
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-3-execution-comment-writer`
PR: pending

## Purpose

Append execution summary comments to ticket Markdown.

## Delivered Behavior

The core package now exposes an execution comment writer that appends one
comment from a `ResultApplication`.

The comment body includes:

- execution ID;
- execution status;
- result summary;
- result artifact path;
- runtime log artifact path;
- validation artifact path;
- failure reason when present.

Existing comments and ticket sections are preserved.

## Implementation Notes

The writer delegates Markdown insertion to the existing ticket commenter so the
comment format and `updated_at` behavior remain consistent with `chc ticket
comment`.

The result application model now renders the canonical execution comment body.

## Contracts Added or Changed

New core types:

- `TicketResultCommentWriter`
- `WriteResultCommentOptions`

New functions:

- `NewTicketResultCommentWriter`
- `RenderResultApplicationCommentBody`

## Files of Interest

- `internal/chc/result_application.go`
- `internal/chc/result_application_test.go`
- `internal/chc/result_comment_writer.go`
- `internal/chc/result_comment_writer_test.go`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice writes a summary comment only. Dedicated validation evidence,
changed-file summary, and blocked-reason sections are planned in later slices.
