# Slice 9: Attention View

Status: in review
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-9-attention-view`
PR: pending

## Purpose

Add a virtual Board attention view outside normal column navigation.

## Delivered Behavior

- Added Board attention tickets with plain-text reasons.
- Blocked tickets are classified as attention tickets.
- Tickets outside normal Board columns are classified as attention tickets.
- Board attention tickets remain sorted by ticket key.
- TUI Board supports `alt+a` to toggle attention mode.
- Repeating `alt+a` returns to the previous normal Board column.
- Attention rows render with a restrained `!` prefix.
- Selected ticket detail shows plain-text attention reasons.
- Board key helper text documents the attention toggle.

## Implementation Notes

- Attention mode is intentionally outside left/right normal column navigation.
- The implementation keeps the detectors conservative: blocked status and
  statuses outside normal Board columns.
- Rich reviewer, PR, validation, age, and comment-based attention signals remain
  outside this slice.

## Contracts Added or Changed

- `Board.Attention`
- `BoardAttentionTicket`
- `BoardAttentionReasons`
- TUI attention toggle: `alt+a`

## Files of Interest

- `internal/chc/board_query.go`
- `internal/chc/board_query_test.go`
- `internal/chctui/model.go`
- `internal/chctui/model_test.go`

## Validation

- `go test ./internal/chc ./internal/chctui`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- Invalid Markdown tickets that cannot be parsed still surface through existing
  reconciliation blockers, not through the Board attention list.
- Later slices may add richer attention reasons when ticket detail, reviewer,
  validation, and execution state are available in the Board model.
