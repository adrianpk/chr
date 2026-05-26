# Slice 4: Parity Tests

Status: delivered
Slicing: CLI TUI Surface Parity
Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`
Tracker: `.prj/trackers/slicing-cli-tui-surface-parity-tracker.md`
Branch: `feat/slice-4-parity-tests`
PR: #207

## Purpose

Add focused test coverage for the CLI/TUI parity surfaces introduced or audited
by the previous slices.

## Delivered Behavior

- Added root command registration coverage for metadata, dump, quiz, spec, plan,
  and plan tracking CLI command paths.
- Added CLI behavior coverage for use-case metadata creation and spec/plan list
  service views.
- Added TUI adapter coverage for dump create/edit/save dispatch through dump
  services.
- Added TUI adapter coverage for runtime quiz start/answer/mode dispatch through
  quiz runtime services.
- Added a test evidence section to the CLI/TUI action matrix.

## Implementation Notes

The tests exercise command and adapter seams without adding new production
behavior. Model-backed paths use existing fake quiz model fixtures where needed.

No new service abstractions were introduced.

## Contracts Added or Changed

- CLI parity command paths must remain registered.
- TUI dump and runtime quiz adapters must continue dispatching to the same
  service-backed operations used by CLI commands.
- The parity matrix now carries concrete test evidence for audited matched
  surfaces.

## Files of Interest

- `internal/chcapp/cli_test.go`
- `internal/chcapp/tui_adapter_parity_test.go`
- `docs/reference/cli-tui-action-matrix.md`

## Validation

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass after splitting one long test into helper functions.

## Risks and Follow-ups

- Slice 5 should update user-facing documentation for the current CLI and TUI
  surface.
- Workspace graph refresh remains the main known `missing_cli` parity gap.
