# Slice 2: Read-Only Board And Ticket Detail

Status: delivered
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-2-tui-board-ticket-detail`
PR: #140

## Purpose

Make tickets the first useful operational surface in the TUI.

This slice keeps behavior read-only while loading current project tickets and
rendering them as navigable board and list views with selected ticket detail.

## Delivered Behavior

- Loads tickets through the existing ticket index and reconciliation path before
  starting the TUI.
- Renders the Board view as tickets grouped by status.
- Renders the Tickets view as a flat ticket list.
- Supports selected-ticket movement in ticket views.
- Shows selected ticket metadata, source path, checklist summary, and comment
  summary.
- Adds text filtering over ticket metadata.
- Keeps all ticket behavior read-only.

## Implementation Notes

`internal/chcapp` remains responsible for project resolution and ticket index
loading. It passes indexed tickets into `internal/chctui` through TUI options.

`internal/chctui` converts indexed tickets into package-local ticket view models
so rendering and key handling stay isolated from index persistence details.

Checklist and comment summaries are derived from the ticket Markdown file when
available. Missing or unreadable files produce zero summary counts instead of
blocking TUI startup.

## Contracts Added or Changed

- `chctui.Options` now accepts `Tickets []chc.IndexedTicket`.
- `chc` and `chc tui` refresh and read the project ticket index before TUI
  startup.
- TUI navigation changes to left/right or tab for view navigation in order to
  reserve up/down for ticket selection inside ticket views.

## Files of Interest

- `internal/chcapp/app.go`
- `internal/chctui/model.go`
- `internal/chctui/model_test.go`

## Validation

```sh
go test ./internal/chctui ./internal/chcapp
go test ./...
make verify
```

All commands pass.

## Risks and Follow-ups

- Board grouping is textual and status-sorted. Richer layout and styling remain
  future TUI work.
- Branch and PR fields are not shown because they are not represented in the
  current indexed ticket metadata.
