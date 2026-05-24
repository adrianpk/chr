# Slice 6: Full Dump Markdown Editing

Status: delivered
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-6-full-dump-markdown-editing`
PR: #180

## Purpose

Convert the dump update workflow from append-oriented behavior to
full-document Markdown editing while keeping dump reads, writes, path
resolution, and external editor command construction outside the TUI
presentation layer.

## Delivered Behavior

- `Edit dump` opens the selected dump in an internal Markdown editing state.
- `ctrl+s` saves the full dump body through the service and refreshes state.
- `alt+e` opens the selected dump in the configured external editor.
- Bubble Tea executes the external editor through `tea.ExecProcess`, then
  refreshes dump state when the editor exits.
- The selected dump is preserved by relative path after refresh when the dump
  still exists.
- The visible command palette no longer exposes `append-dump` as the primary
  dump update workflow.

## Implementation Notes

- `internal/chcapp.service` owns dump read/save/open operations and the editor
  launcher dependency.
- `internal/chcapp.tuiDumpActionExecutor` maps internal edit, save, external
  edit, and refresh intents to service operations.
- `internal/chctui.Model` keeps a dump edit state for internal Markdown edits
  and handles external editor completion as a message.

## Contracts Added or Changed

- Added `DumpEditorLauncher`, `DumpEditorCommand`, and `ReadDump` to the dump
  service boundary.
- Added `DumpActionSave` for internal full-document saves.
- Added `DumpActionExternalEdit` for the `alt+e` external editor flow.
- Added `DumpActionRefresh` as an internal post-editor refresh action.
- Added optional `Content`, `EditorCommand`, and `SelectedPath` fields to TUI
  dump action results.

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

- The internal editor is deliberately rudimentary and follows the existing TUI
  metadata editing model rather than a full IDE-style Markdown editor.
- `$VISUAL` and `$EDITOR` are split with `strings.Fields`; quoted editor
  command arguments are not parsed with shell semantics.
