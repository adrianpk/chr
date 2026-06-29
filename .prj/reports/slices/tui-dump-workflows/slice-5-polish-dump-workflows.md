# Slice 5: Polish Dump Workflows

Status: delivered
Slicing: TUI Dump Workflows
Plan: `.prj/plans/slicing-tui-dump-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-dump-workflows-tracker.md`
Branch: `feat/slice-5-polish-dump-workflows`
PR: `#179`

## Purpose

Close the planned dump workflow slicing with command, empty-state, and naming
polish before moving into follow-up tuning for full Markdown editing.

## Delivered Behavior

- The command registry exposes create, append, and edit dump workflows.
- Empty Dumps views tell the user to use the command palette to create a dump.
- The selected-path operation is now presented as `Edit dump` instead of
  `Open dump`.
- Tests cover dump workflow command registration.
- Tests cover empty dump rendering.
- Tests cover planning the edit dump command for a selected dump.

## Implementation Notes

- This slice intentionally does not add a real external editor launcher.
- `Edit dump` currently resolves the selected dump path through the service.
- The naming change aligns the UX direction with the domain model:
  a dump is a full Markdown document that should be editable, not only appended.

## Contracts Added or Changed

- Renamed the selected dump path command from `open-dump` to `edit-dump`.
- Renamed the TUI dump action constant from open to edit.
- Improved empty state text for Dumps.

## Files of Interest

- `internal/chrtui/command_palette.go`
- `internal/chrtui/dump_actions.go`
- `internal/chrtui/dumps.go`
- `internal/chrtui/model_test.go`
- `internal/chrapp/tui_adapters.go`

## Validation

- `make verify`

## Risks and Follow-ups

- Real full-Markdown editing still needs a follow-up tuning slice with a
  service-owned editor or launcher abstraction.
- Visual refinement should happen after edit behavior exists, using the captured
  screenshots as review evidence.
