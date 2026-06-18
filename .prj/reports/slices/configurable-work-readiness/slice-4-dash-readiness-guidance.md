# Slice 4: Dash Readiness Guidance

Status: delivered
Slicing: Configurable Work Readiness
Plan: `ops/plans/2026-06-18-configurable-work-readiness-plan.md`
Tracker: `ops/trackers/2026-06-18-configurable-work-readiness-tracker.md`
Branch: `feat/slice-4-dash-readiness-guidance`
PR: `#316`

## Purpose

Surface work readiness directly in Dash so users can see why agent work cannot
start, apply safe repairs from the same operational screen, and retry the work
action without reading raw Git failures.

## Delivered Behavior

- TUI options now load a `WorkReadinessView` from `core.CheckWorkReadiness`.
- Dash refresh keeps tickets, observability, and readiness diagnostics current.
- Dash renders readiness status, current branch, expected branch, diagnostics,
  applied repairs, and safe-repair guidance.
- Dash exposes `x` for safe readiness repair when a safe repair is available.
- Ticket execution and run-all actions are blocked before execution when
  readiness is blocked.
- Safe repair uses the same `core.RepairWorkReadiness` path as
  `chc work doctor --fix-safe`.
- QA reset creates a ready `dev` branch with a local `origin/dev` remote ref.
- QA smoke checks `chc work doctor --dir <project>` before manual TUI runs.

## Implementation Notes

- Dash does not auto-run agent work after repair. The user confirms repair, then
  retries the original work action.
- Readiness blocking happens in `startWorkAction`, so Dash keys and command
  palette work actions share the same guard.
- The TUI repair action reloads tickets, observability, and readiness after the
  repair attempt.
- QA reset keeps the local remote under ignored `qa/tmp/` state so reference
  projects stay independent from the parent repository.

## Contracts Added or Changed

- Added TUI readiness projection types in `internal/tui/work_readiness.go`.
- Extended `tui.Options`, agent work refresh results, and work action results
  with readiness state.
- Added `WorkActionRepairReadiness`.
- Added Dash key `x` for safe readiness repair.
- Extended QA smoke to require `status: ready` from `chc work doctor`.

## Files of Interest

- `internal/tui/work_readiness.go`
- `internal/tui/dash.go`
- `internal/tui/work_actions.go`
- `internal/app/tui_adapters_work.go`
- `internal/app/tui_adapters_agent_work.go`
- `scripts/qa/reset.sh`
- `scripts/qa/smoke.sh`
- `docs/reference/tui.md`

## Validation

- `go test ./internal/tui ./internal/app` - pass
- `make qa-smoke-tictactoe` - pass
- `make qa-smoke-blog` - pass
- `make verify` - pass

## Risks and Follow-ups

- Dash repair remains intentionally limited to safe repairs reported by the core
  readiness service. Divergence, user changes, missing remotes, and in-progress
  Git operations still require manual action.
- Expanded container or worker log inspection remains separate from readiness
  repair; this slice only covers work-start readiness and safe repair feedback.
