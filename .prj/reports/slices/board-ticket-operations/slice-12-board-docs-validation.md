# Slice 12: Board Docs Validation

Status: in review
Slicing: Board Ticket Operations
Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`
Tracker: `.prj/trackers/slicing-board-ticket-operations-tracker.md`
Branch: `feat/slice-12-board-docs-validation`
PR: pending

## Purpose

Bring user-facing and reference documentation up to date for the completed
Board ticket operations slice set.

## Delivered Behavior

- Updated the user guide Board section with current Board columns, attention
  mode, compact and maximized behavior, selected-ticket detail, and CLI
  equivalents.
- Updated the TUI reference with Board keybindings and selected-ticket detail
  behavior.
- Updated the command reference with `chc ticket board`, required ticket
  synopsis examples, and common ticket status operations.
- Updated ticket status and validation references for normal Board statuses and
  accepted legacy statuses.
- Updated ticket parser and ticket index references with synopsis, Board
  column, source metadata, and slice metadata.
- Updated SQLite migration reference with the Board metadata migration file and
  `ticket_key_sequences` role.
- Updated project workspace reference with Markdown ticket, user-level SQLite,
  and `.chc/user/` responsibilities.
- Updated CLI/TUI parity matrix with Board navigation, attention, maximized
  Board, selected-ticket detail, and Board query service parity.
- Updated tutorial and lab ticket creation examples to include required
  synopsis metadata and valid priority/severity values.

## Implementation Notes

- This slice is documentation-only.
- `gofmt` is not applicable because no Go files changed.
- Validation still ran the targeted packages and full verification gate.

## Contracts Added or Changed

- Board documentation now treats `backlog`, `ready_for_dev`, `in_progress`,
  `ready_for_review`, `ready_for_merge`, and `done` as the normal Board status
  vocabulary.
- Legacy status mapping is documented explicitly.
- Ticket creation examples now include `--synopsis`.
- `chc ticket board` is documented as the CLI equivalent for Board column
  inspection.

## Files of Interest

- `docs/how-to/user-guide/index.md`
- `docs/reference/tui.md`
- `docs/reference/commands.md`
- `docs/reference/cli-tui-action-matrix.md`
- `docs/reference/project-workspace.md`
- `docs/reference/sqlite-migrations.md`
- `docs/reference/ticket-index.md`
- `docs/reference/ticket-parser.md`
- `docs/reference/ticket-status.md`
- `docs/reference/ticket-validation.md`
- `docs/tutorials/getting-started.md`
- `docs/how-to/full-system-lab-runbook.md`

## Validation

- `git diff --check`
- `go test ./internal/chc ./internal/chcapp ./internal/chctui`
- `go test ./...`
- `make verify`

## Risks and Follow-ups

- Full multi-operator `.chc/operators/<operator-id>/...` layout remains
  deferred.
- Manual Markdown sync and LLM-assisted empty synopsis generation remain
  deferred.
- Done-ticket archival heuristics remain deferred.
