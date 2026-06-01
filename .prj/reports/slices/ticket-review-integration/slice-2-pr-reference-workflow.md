# Slice 2: PR Reference Workflow

Status: delivered
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-2-pr-reference-workflow`
PR: #274

## Purpose

Make branch and pull request reference attachment an explicit workflow operation
instead of relying on direct ticket frontmatter edits.

## Delivered Behavior

- `chc ticket review-ref` can attach a ticket implementation branch and a pull
  request review reference.
- Existing pull request references are updated when the URL matches, or when
  provider, type, and provider ID match.
- Moving a ticket to `ready_for_review` requires branch metadata and at least
  one `pull_request` review reference.
- Result application preserves existing branch and pull request metadata.
- The TUI action adapter has an operation path for attaching review references
  when a future interactive input surface supplies branch and PR values.

## Implementation Notes

- Ticket review reference updates operate on frontmatter only and preserve the
  ticket body.
- The new writer can insert missing `branch` and `review_refs` fields, so older
  tickets can be upgraded without a full renderer rewrite.
- The `ready_for_review` requirement is enforced on status transition rather
  than global ticket parsing, so older indexed tickets remain readable.

## Contracts Added or Changed

- New CLI command:
  `chc ticket review-ref <ticket-key> --branch <branch> --provider <provider> --url <url> [--id <id>]`.
- `ready_for_review` status transitions now require:
  `branch` plus one `review_refs` entry with `type: "pull_request"` and a URL.
- TUI ticket actions now include an `attach_review_ref` operation request shape.

## Files of Interest

- `internal/chc/ticket_review_metadata.go`
- `internal/chc/ticket_status.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/service_artifact_ticket.go`
- `internal/chcapp/tui_adapters_spec_plan_ticket.go`
- `internal/chctui/ticket_actions.go`
- `docs/reference/ticket-review-metadata.md`
- `docs/reference/ticket-validation.md`

## Validation

- `go test ./internal/chc -run 'Ticket(Review|Status)|ReviewMetadata'`
- `go test ./internal/chcapp -run 'Ticket(ReviewRef|Status|ApplyResult)'`
- `go test ./internal/chc ./internal/chcapp ./internal/chctui`
- `make verify`

## Risks and Follow-ups

- The TUI exposes the operation adapter but does not yet include a text input
  workflow for branch and pull request URL entry.
- Provider-specific validation of URL readability remains in the later provider
  diagnostics slices.
