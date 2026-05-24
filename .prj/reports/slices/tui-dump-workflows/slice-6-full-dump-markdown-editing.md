# Slice 6: Full Dump Markdown Editing

Status: reviewing
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-6-full-dump-markdown-editing`
PR: pending

## Purpose

Convert the dump update workflow from append-oriented behavior to
full-document Markdown editing while keeping dump path resolution and editor
command construction outside the TUI presentation layer.

## Delivered Behavior

- `Edit dump` resolves the selected dump through the service boundary.
- The service returns an editor command for the complete Markdown dump file.
- Bubble Tea executes the editor through `tea.ExecProcess`, then refreshes dump
  state when the editor exits.
- The selected dump is preserved by relative path after refresh when the dump
  still exists.
- The visible command palette no longer exposes `append-dump` as the primary
  dump update workflow.

## Implementation Notes

- `internal/chcapp.service` owns the dump editor launcher dependency and
  resolves `$VISUAL`, `$EDITOR`, or `vi`.
- `internal/chcapp.tuiDumpActionExecutor` converts the service editor command
  into an executable Bubble Tea command result.
- `internal/chctui.Model` handles editor completion as a message and refreshes
  through the dump action executor.

## Contracts Added or Changed

- Added `DumpEditorLauncher` and `DumpEditorCommand` to the dump service
  boundary.
- Added `DumpActionRefresh` as an internal post-editor refresh action.
- Added optional `EditorCommand` and `SelectedPath` fields to TUI dump action
  results.

## Files of Interest

- `internal/chcapp/service_dump.go`
- `internal/chcapp/tui_adapters.go`
- `internal/chctui/dump_actions.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model_test.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui`
- `go test ./internal/chc ./internal/chcapp ./internal/chctui`
- `make verify`

## Risks and Follow-ups

- `$VISUAL` and `$EDITOR` are split with `strings.Fields`; quoted editor
  command arguments are not parsed with shell semantics.
