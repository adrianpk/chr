# Slice 3: Create Dumps From TUI

Status: reviewing
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-3-create-dumps-from-tui`
PR: `#177`

## Purpose

Allow users to create project-wide and use-case-scoped dumps from the TUI while
keeping write behavior inside the application service.

## Delivered Behavior

- The command palette exposes `Create dump`.
- The TUI plans a confirmable dump action.
- When invoked from `Cases`, the selected case slug becomes the default dump
  scope.
- When invoked outside `Cases`, the dump is project-wide.
- Confirming the action calls the dump service through a TUI adapter.
- The service writes Markdown under `.chc/dumps`.
- Created dumps use `status: "raw"`.
- Case-scoped dumps include `case: "<slug>"`; project-wide dumps omit `case`.
- The TUI refreshes its dump read model from the service result.

## Implementation Notes

- Added a dedicated dump action path instead of extending artifact actions with
  dump-specific behavior.
- The TUI action stores intent and scope only; filename, path, frontmatter, and
  writes are service-owned.
- The service creates unique filenames from the requested title and refuses to
  overwrite existing files.
- The initial TUI create flow uses deterministic default titles and content.
  Rich title/body entry belongs to later UX or edit/append slices.

## Contracts Added or Changed

- Added `service.CreateDump`.
- Added `DumpActionExecutor`, `DumpActionRequest`, `DumpActionResult`, and
  `DumpActionState` in `internal/chctui`.
- Added `CommandActionPlanDump` and the `create-dump` command.
- Added TUI adapter mapping from dump action requests to service
  `DumpCreateRequest`.

## Files of Interest

- `internal/chcapp/service_dump.go`
- `internal/chcapp/tui_adapters.go`
- `internal/chcapp/service_test.go`
- `internal/chctui/dump_actions.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model.go`
- `internal/chctui/model_test.go`
- `internal/chctui/shell.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui`
- `make verify`

## Risks and Follow-ups

- The create flow currently uses default titles and body text. Later slices
  should add richer user input, append, and edit behavior.
- The current default scope is contextual rather than an explicit scope picker.
  A later UX pass can make scope selection more visible.
