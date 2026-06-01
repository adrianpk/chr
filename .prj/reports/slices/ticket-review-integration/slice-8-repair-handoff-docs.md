# Slice 8: Repair Handoff and Docs

Status: delivered
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-8-repair-handoff-docs`
PR: #280

## Purpose

Connect review rejection to repair execution and document the complete ticket
review workflow.

## Delivered Behavior

- Repair worker context includes branch, pull request URL, review comments URL,
  review state, review result, and review round.
- Repair context points workers at remote pull request comments as the source of
  detailed findings instead of copying those comments into ticket comments.
- Repair context materialization blocks before worker launch when required
  review handoff metadata is missing.
- Ticket review workflow documentation now covers Forgejo setup, review modes,
  provider diagnostics, serialized work selection, review command behavior, and
  repair handoff.

## Implementation Notes

- Handoff generation is part of Codex worker context materialization.
- Non-repair tickets do not receive a repair handoff section.
- Repair detection reuses existing ticket review metadata:
  `review_state: "changes_requested"` or `review_result: "needs_changes"`.
- Remote comment contents remain on the forge. The ticket stores only the URL
  needed to locate them.

## Contracts Added or Changed

- Added `CodexWorkerRepairHandoff` to worker context rendering.
- Added pre-worker validation for repair tickets missing branch, pull request
  reference, or `review_comments_url`.
- Added `docs/reference/ticket-review-workflow.md`.
- Added `docs/how-to/use-forgejo-ticket-review.md`.

## Files of Interest

- `internal/chc/codex_worker_context.go`
- `internal/chc/codex_worker_context_test.go`
- `docs/reference/ticket-review-workflow.md`
- `docs/how-to/use-forgejo-ticket-review.md`
- `docs/reference/review-workflow-config.md`
- `docs/reference/commands.md`

## Validation

- `git diff --check`
- `scripts/check-go-func-signatures.sh`
- `go test ./internal/chc -run 'CodexWorkerContext|WorkSelection|TicketReview'`
- `make verify`
- Ticket review QA scenario:
  - initialize temporary workspace
  - create ticket
  - attach Forgejo pull request reference
  - move ticket to `ready_for_review`
  - run Forgejo provider diagnostics dry-run

## Risks and Follow-ups

- The QA scenario validates local workflow wiring and provider detection, not
  live remote comment writing.
- Full live Forgejo comment QA still depends on a disposable pull request or a
  dedicated fixture repository.
