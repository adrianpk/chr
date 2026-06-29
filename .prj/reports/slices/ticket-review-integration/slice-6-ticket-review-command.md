# Slice 6: Ticket Review Command

Status: delivered
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-6-ticket-review-command`
PR: #278

## Purpose

Run an inspection-only ticket review over `ready_for_review` work without
starting Docker, launching Codex, or reimplementing the ticket.

## Delivered Behavior

- `chr ticket review <key>` applies an explicit review result to a
  `ready_for_review` ticket.
- The command assembles ticket and remote pull request evidence through the
  remote review adapter.
- Clean passes update local review metadata and move the ticket to
  `ready_for_merge` without writing remote comments.
- `needs_changes` requires at least one finding, writes remote finding comments,
  updates local review metadata, and moves the ticket back to `ready_for_dev`.
- `--dry-run` reads evidence and prints the plan without mutating the ticket or
  writing remote comments.

## Implementation Notes

- The reviewer policy is represented as an inspection-only command contract.
- The command accepts explicit `--result` and `--finding` inputs; it does not
  invoke an LLM reviewer in this slice.
- Remote comments are written only for supplied findings.
- The command reuses the provider adapter from Slice 5.

## Contracts Added or Changed

- Added `TicketReviewCommandResult` and `RunTicketReviewCommand`.
- Added CLI command:
  `chr ticket review <key> --result <result> [--finding <text>]`.
- Supported results are `pass`, `needs_changes`, `insufficient_evidence`, and
  `needs_human_decision`.
- Review command output is line-oriented and machine-readable.

## Files of Interest

- `internal/chr/ticket_review_command.go`
- `internal/chrapp/ticket.go`
- `internal/chrapp/cli_test.go`

## Validation

- `git diff --check`
- `go test ./internal/chr -run 'TicketReview|ReviewRemote|ReviewProvider|Config'`
- `go test ./internal/chrapp -run 'TicketReview|TicketReviewMode|TicketReviewRef'`
- `go test ./internal/chr ./internal/chrapp`
- `golangci-lint run ./internal/chr ./internal/chrapp`

## Risks and Follow-ups

- A later slice must add serialized work preflight so later development cannot
  skip a ticket moved back to `ready_for_dev`.
- Agent-authored review decisions still need prompt integration once the
  Codex-driven reviewer flow is wired into the command.
