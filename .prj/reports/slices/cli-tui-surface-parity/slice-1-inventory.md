# Slice 1: Inventory

Status: reviewing
Slicing: CLI TUI Surface Parity
Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`
Tracker: `.prj/trackers/slicing-cli-tui-surface-parity-tracker.md`
Branch: `feat/slice-1-cli-tui-inventory`
PR: pending

## Purpose

Create the first complete CLI/TUI parity inventory before behavior changes.

This slice documents the current operation surfaces and identifies where TUI
actions, CLI commands, and application service paths are matched, missing, or
divergent.

## Delivered Behavior

- Added a CLI/TUI action matrix under `docs/reference/`.
- Captured direct TUI keybindings by view.
- Captured command palette actions and availability gates.
- Captured current CLI commands by domain.
- Mapped TUI adapters, CLI commands, and shared service paths.
- Classified gaps using explicit parity statuses.

## Implementation Notes

The matrix is documentation-only. It does not change runtime behavior.

The inventory shows that many TUI operations already use service methods that
can become CLI entrypoints. It also shows several existing CLI commands that are
operationally equivalent to TUI actions but bypass the same service facade.

## Contracts Added or Changed

- Added `docs/reference/cli-tui-action-matrix.md` as the canonical parity
  inventory for the current slicing.
- Linked the matrix from the docs index, command reference, and TUI reference.

## Files of Interest

- `docs/reference/cli-tui-action-matrix.md`
- `docs/reference/commands.md`
- `docs/reference/tui.md`
- `docs/index.md`
- `.prj/trackers/slicing-cli-tui-surface-parity-tracker.md`

## Validation

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Risks and Follow-ups

- Some service-path classifications are based on code inspection and should be
  confirmed during Slice 2 before refactoring.
- `missing_tui` for CLI-only administrative commands may remain intentional
  after product review.
