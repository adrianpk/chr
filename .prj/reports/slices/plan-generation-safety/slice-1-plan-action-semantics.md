# Slice 1: Plan Action Semantics

Status: reviewing
Slicing: Plan Generation Safety
Plan: `.prj/plans/slicing-plan-generation-safety-plan.md`
Tracker: `.prj/trackers/slicing-plan-generation-safety-tracker.md`
Branch: `feat/slice-1-plan-action-semantics`
PR: pending

## Purpose

Separate cheap plan refresh, missing-plan creation, and explicit plan
regeneration intent in the TUI before implementing archive-before-replace
behavior.

## Delivered Behavior

- `g` now means generate a missing plan only.
- Pressing `g` on a selected scope with an existing canonical plan does not
  invoke the model.
- Existing-plan generation refusal reports that regeneration is explicit.
- `G` opens a separate confirmable regenerate action.
- The command palette exposes a separate `regenerate-plan` command.
- Plan busy state blocks repeated plan keypress handling with a clear status.
- Plan action labels and busy messages distinguish generate, regenerate, and
  refresh.
- The service adapter receives a distinct regenerate intent and reports a
  regenerate-specific success message.

## Implementation Notes

- `PlanActionGenerate` now carries the internal action value
  `generate_missing_plan`.
- Regenerate currently reuses the existing plan generation operation. Archive
  creation and replacement safety belong to Slice 2.
- The TUI owns intent gating for this slice. Service-level same-scope guards
  belong to Slice 4.

## Contracts Added or Changed

- Added `PlanActionRegenerate`.
- Added the `regenerate-plan` command palette command.
- Updated Plans key help and action hints to show generate-missing versus
  regenerate behavior.
- Added canonical plan existence detection for plan action gating.

## Files of Interest

- `internal/chctui/plan_actions.go`
- `internal/chctui/plans.go`
- `internal/chctui/model.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model_test.go`
- `internal/chcapp/tui_adapters.go`

## Validation

- `go test ./...`: pass
- `make verify`: fail

`make verify` fails in `lint-strict` on existing lint debt across the plan,
spec, and TUI areas, including cyclomatic complexity, function length,
formatting policy, unused functions, and whitespace rules. The slice behavior
tests pass with `go test ./...`.

## Commits

- `9d9db1f feat(tui): separate plan generate and regenerate actions`
- `09f21f7 test(tui): cover plan action semantics`

## Risks and Follow-ups

- Regenerate is intentionally unsafe until Slice 2 adds archive-before-replace.
- Service-level duplicate generation protection remains planned for Slice 4.
