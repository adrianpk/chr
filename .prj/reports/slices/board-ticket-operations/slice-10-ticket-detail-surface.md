# Slice 10: Ticket Detail Surface

Status: in review
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-10-ticket-detail-surface`
PR: pending

## Purpose

Make the Board detail pane useful for selected tickets without leaving the TUI.

## Delivered Behavior

- Selected ticket detail shows synopsis when ticket metadata provides it.
- Selected ticket detail shows source kind, source reference, and slice metadata
  when available.
- Selected ticket detail preserves existing status, priority, severity, type,
  use-case, project-area, checklist, comment, attention, and path information.
- The TUI loads a bounded Markdown preview for the selected ticket.
- Body preview, checklist item preview, and recent comment preview are rendered
  only when available.
- Missing or unreadable ticket Markdown keeps rendering safe with empty preview
  data.
- Ticket filtering now matches synopsis and source metadata.

## Implementation Notes

- Markdown detail parsing is intentionally bounded for the normal right pane:
  four body lines, four checklist items, and three comment headings.
- Ticket body preview skips YAML frontmatter and comment sections.
- Comment preview uses headings under `# Comments`.
- Checklist preview uses Markdown task list lines and keeps the existing done
  count behavior.

## Contracts Added or Changed

- Board detail pane now consumes `IndexedTicket.Synopsis`.
- Board detail pane now consumes `IndexedTicket.SourceKind`.
- Board detail pane now consumes `IndexedTicket.SourceRef`.
- Board detail pane now consumes `IndexedTicket.Slice`.
- Ticket filtering includes synopsis and source metadata.

## Files of Interest

- `internal/chrtui/model.go`
- `internal/chrtui/model_test.go`

## Validation

- `go test ./internal/chrtui`
- `git diff --check`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- The Markdown preview is intentionally shallow; maximized Board rendering may
  later expose more context.
- Comment preview currently keys off `# Comments` and `##` headings, matching
  the current ticket comment format.
