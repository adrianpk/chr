# Slice 5: Parity Docs

Status: reviewing
Slicing: CLI TUI Surface Parity
Plan: `.prj/plans/slicing-cli-tui-surface-parity-plan.md`
Tracker: `.prj/trackers/slicing-cli-tui-surface-parity-tracker.md`
Branch: `feat/slice-5-parity-docs`
PR: #208

## Purpose

Publish the current CLI/TUI parity contract in user-facing and reference
documentation.

## Delivered Behavior

- Added CLI equivalents for dumps, runtime quizzes, specs, plans, and plan
  tracking to the user guide.
- Added command reference coverage for metadata, dumps, runtime quizzes, specs,
  plans, and plan tracking.
- Added CLI equivalent notes to the TUI reference.
- Documented the CLI/TUI operation-surface architecture.
- Finalized the parity matrix gap notes for the completed slicing.

## Implementation Notes

The documentation describes existing command and TUI behavior only. No runtime
code or command behavior changed in this slice.

Validation still ran the full repository gate to catch formatting, lint, and
test regressions from documentation changes.

## Contracts Added or Changed

- CLI and TUI are documented as separate operation surfaces over shared service
  boundaries.
- The user guide and references now expose the current command equivalents for
  service-backed TUI actions.
- Remaining gaps stay explicit in the CLI/TUI action matrix.

## Files of Interest

- `docs/how-to/user-guide/index.md`
- `docs/reference/commands.md`
- `docs/reference/tui.md`
- `docs/explanation/architecture.md`
- `docs/reference/cli-tui-action-matrix.md`

## Validation

- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Risks and Follow-ups

- Workspace graph refresh remains the only high-priority `missing_cli` gap.
- Checklist item operations remain unavailable until a safe checklist writer is
  implemented.
