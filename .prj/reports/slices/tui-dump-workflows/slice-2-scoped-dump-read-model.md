# Slice 2: Scoped Dump Read Model

Status: delivered
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-2-scoped-dump-read-model`
PR: `#176`

## Purpose

Make Dumps a dedicated TUI read model instead of a generic document-family
projection.

This lets the TUI present project-wide and use-case-scoped dumps without
parsing Markdown or knowing dump filesystem layout rules.

## Delivered Behavior

- The TUI accepts dump-specific state through `chctui.Options`.
- The app adapter loads dumps from the service and maps them into TUI state.
- The Dumps screen renders scope, status, title, path, updated time, and
  preview.
- Project-wide dumps are shown with `project` scope.
- Use-case-scoped dumps are shown with their case slug.
- Dumps have independent selection state from generic documents.
- The selected source path command resolves the selected dump path.

## Implementation Notes

- `internal/chctui.DumpView` is intentionally separate from
  `internal/chcapp.DumpView` to keep the TUI package independent from the app
  service implementation.
- `loadTUIDumps` is the adapter boundary that converts service results into TUI
  read state.
- The TUI still receives generic `DocumentView` values for quizzes, specs,
  plans, and legacy document views; only Dumps has been split out in this
  slice.
- `loadTUIOptions` was split into smaller loading helpers to keep the new dump
  loading path within lint limits.

## Contracts Added or Changed

- Added `chctui.DumpScope` and `chctui.DumpView`.
- Added `Options.Dumps`.
- Added independent dump selection and navigation state.
- Changed `F3 Dumps` to render from dump-specific state.
- Changed selected-source-path resolution for Dumps to use selected dump state.

## Files of Interest

- `internal/chctui/dumps.go`
- `internal/chctui/model.go`
- `internal/chctui/shell.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model_test.go`
- `internal/chcapp/app.go`
- `internal/chcapp/tui_adapters.go`

## Validation

- `go test ./internal/chctui ./internal/chcapp`
- `make verify`

## Risks and Follow-ups

- Dump preview currently depends on service-provided state or falls back to a
  placeholder. Rich body preview can be improved after service read behavior is
  expanded.
- Create, append, save, and open workflows remain for later slices.
