# Slice 4: Append And Edit TUI Dumps

Status: reviewing
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-4-append-edit-tui-dumps`
PR: `#178`

## Purpose

Allow selected dumps to grow from the TUI through service-owned operations and
add a service boundary for resolving the selected dump path.

## Delivered Behavior

- The command palette exposes `Append dump`.
- The command palette exposes `Open dump`.
- Append acts on the selected dump and writes through the service.
- Appended content is added to the existing Markdown body without replacing
  frontmatter.
- Open resolves the selected dump path through the service and reports the
  project-relative path.
- The TUI refreshes dump state after append/open actions.
- The service rejects dump paths outside `.chc/dumps`.

## Implementation Notes

- Append and open are separate dump actions under the existing dump action
  executor.
- Open currently resolves and reports the path through the service. It does not
  launch an external editor because no launcher/editor abstraction exists in the
  codebase yet.
- `SaveDump` was added at the service level for direct content replacement, but
  no TUI text editing surface uses it in this slice.
- Path validation is project-relative and constrained to `.chc/dumps`.

## Contracts Added or Changed

- Added `service.AppendDump`.
- Added `service.SaveDump`.
- Added `service.OpenDump`.
- Added `DumpActionAppend` and `DumpActionOpen`.
- Added command palette entries for append and open dump operations.

## Files of Interest

- `internal/chcapp/service_dump.go`
- `internal/chcapp/tui_adapters.go`
- `internal/chcapp/service_test.go`
- `internal/chctui/dump_actions.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model_test.go`

## Validation

- `go test ./internal/chcapp ./internal/chctui`
- `make verify`

## Risks and Follow-ups

- Open does not launch `$EDITOR` or another external tool yet. A later UX or
  integration slice should introduce a service-owned launcher abstraction before
  wiring external edit behavior.
- Append currently uses deterministic default content. Rich input belongs to a
  later TUI input refinement.
