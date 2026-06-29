# Slice 1: TUI App Foundation

Status: delivered
Slicing: TUI Operational Console
Plan: `.prj/plans/slicing-tui-operational-console-plan.md`
Tracker: `.prj/trackers/slicing-tui-operational-console-tracker.md`
Branch: `feat/slice-1-tui-app-foundation`
PR: #139

## Purpose

Create the first Bubble Tea application boundary and entry behavior for the
local Chryso operational console.

This slice intentionally avoids domain mutations and real board/document data
loading. It establishes the package boundary, command routing, and placeholder
navigation shell that later slices can extend.

## Delivered Behavior

- Added Bubble Tea as the TUI runtime dependency.
- Added `internal/chrtui` as the TUI package boundary.
- Added a base Bubble Tea model with update and view behavior.
- Added placeholder navigation for Dashboard, Board, Tickets, Documents,
  Authoring, Agent, Runtime, Reports, and Help.
- Added project-aware rendering with resolved project ID and project directory.
- Added `chr tui`.
- Changed `chr` with no subcommand to open the TUI.
- Preserved normal `chr --help` behavior.
- Passed stdin through the CLI assembly so the TUI can receive interactive input
  and can be smoke-tested with scripted input.

## Implementation Notes

`cmd/chr/main.go` remains assembly-only. It passes stdin, stdout, stderr, and
embedded assets into `internal/chrapp`.

`internal/chrapp` owns command routing and project context resolution. TUI
rendering and Bubble Tea model logic live in `internal/chrtui`.

The first TUI view is plain text by design. Styling, richer layout, data loading,
and command palette behavior are left for later slices.

## Files of Interest

- `cmd/chr/main.go`
- `internal/chrapp/app.go`
- `internal/chrapp/environment.go`
- `internal/chrtui/model.go`
- `internal/chrtui/model_test.go`
- `internal/chrapp/cli_test.go`

## Validation

```sh
go test ./internal/chrtui ./internal/chrapp
make verify
```

Smoke tests:

```sh
go run ./cmd/chr tui --help
printf 'q' | go run ./cmd/chr tui --dir "$tmpdir"
printf 'q' | (cd "$tmpdir" && "$tmpbin")
```

All commands pass.

## Risks and Follow-ups

- The initial view is a shell with placeholders. Slice 2 should load and render
  real ticket board data.
- The first version has no styling layer. Later slices can improve layout while
  keeping the same model boundary.
