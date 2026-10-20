# Slice 11: Maximized Board

Status: delivered
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-11-maximized-board`
PR: #223

Verified merge: `34c1e76dabefc1ee80874b38a4e049bd4931c31e`, an ancestor of dev
as checked on 2026-10-20.

## Purpose

Make the maximized Board pane behave like a broad Board view instead of a
larger copy of the compact single-column list.

## Delivered Behavior

- Maximizing the Board ticket pane renders all normal Board columns.
- Normal Board columns render as responsive rows based on terminal width.
- Each normal column shows its title, ticket count, and bounded ticket rows.
- The active normal column remains highlighted while maximized.
- Selected ticket state is preserved across maximize and restore.
- Active column state is preserved across maximize and restore.
- Attention mode remains separate from normal Board columns.
- Maximized attention mode renders a full-width attention ticket list.

## Implementation Notes

- The maximized Board uses six, three, two, or one columns per row depending on
  available terminal width.
- Attention is intentionally not included in the normal maximized column grid.
- Compact Board rendering is unchanged when the Board pane is not maximized.
- Ticket detail pane maximization is unchanged.

## Contracts Added or Changed

- Board pane maximize (`alt+m`) now has Board-specific rendering when focused on
  the ticket pane.
- Attention mode maximize (`alt+a`, then `alt+m`) renders only attention
  tickets.

## Files of Interest

- `internal/chrtui/model.go`
- `internal/chrtui/shell.go`
- `internal/chrtui/model_test.go`

## Validation

- `go test ./internal/chrtui`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- The maximized Board currently shows compact ticket rows; richer per-ticket
  metadata can be added after the Board layout is validated in daily use.
- Column scrolling remains outside this slice.
