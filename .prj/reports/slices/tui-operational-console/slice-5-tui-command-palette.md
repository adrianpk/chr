# Slice 5: Command Palette Foundation

Status: delivered
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-5-tui-command-palette`
PR: #143

## Purpose

Add a discoverable command palette foundation for safe TUI operations.

This slice exposes navigation and read-only helper commands without wiring
durable mutations.

## Delivered Behavior

- Adds `:` command palette state and rendering.
- Adds a command registry model with navigation, refresh, help, and selected
  source path commands.
- Adds text filtering for command discovery.
- Adds command selection and execution through keyboard input.
- Navigates to registered views from palette commands.
- Shows selected source paths for tickets, documents, and authoring sessions.
- Keeps refresh as an explicit read-only status message rather than a data
  mutation or process restart.

## Implementation Notes

The command palette is model-local UI state. It does not write project files or
invoke CLI commands.

The initial command registry is static and intentionally conservative. Later
slices can add mutating commands behind confirmation flows.

Selected source path handling reuses existing selected ticket, document, and
authoring session view models.

## Contracts Added or Changed

- Pressing `:` opens the command palette.
- While the palette is open, up/down selects commands, enter runs the selected
  command, and escape closes the palette.
- The palette currently supports safe navigation, help, refresh status, and
  selected source path reporting.

## Files of Interest

- `internal/chctui/command_palette.go`
- `internal/chctui/model.go`
- `internal/chctui/model_test.go`

## Validation

```sh
go test ./internal/chctui
make verify
```

All commands pass.

## Risks and Follow-ups

- Refresh currently reports intent only. A later data reload slice can connect
  it to explicit reloading without leaving read-only boundaries.
- Opening selected source paths is represented as a visible path message. A
  later shell/editor integration can perform the external open action safely.
