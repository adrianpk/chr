# Slice 3: Safe Readiness Repair

Status: delivered
Slicing: Configurable Work Readiness
Plan: `ops/plans/2026-06-18-configurable-work-readiness-plan.md`
Tracker: `ops/trackers/2026-06-18-configurable-work-readiness-tracker.md`
Branch: `feat/slice-3-safe-readiness-repair`
PR: `#315`

## Purpose

Add non-destructive repair execution to the work readiness layer so recoverable
Git and ChrysoCoord runtime blockers can be fixed before starting agent work.

## Delivered Behavior

- `chc work doctor --fix-safe` now applies configured safe repairs and prints
  the resulting readiness report.
- Readiness diagnostics now include planned repair metadata.
- ChrysoCoord runtime artifacts can be excluded from Git dirtiness through the
  local `.git/info/exclude` file.
- Missing remote refs can trigger a fetch of the configured remote.
- Existing integration branches can be switched to when branch mutation is safe.
- Missing integration branches can be created only when policy allows branch
  creation.
- Behind integration branches can be fast-forwarded from the configured remote
  ref when the current branch matches and branch mutation is safe.
- Branch-mutating repairs are skipped when user-visible worktree changes or an
  in-progress Git operation make the state unsafe.

## Implementation Notes

- The repair loop is bounded and rechecks readiness after each repair pass. This
  lets one run handle staged discoveries such as `missing_remote_ref` followed
  by `branch_behind_remote` after fetch.
- Runtime artifact repair writes only to `.git/info/exclude`, keeping project
  `.gitignore` and user code untouched.
- Branch repairs use safe Git primitives only: `fetch`, `switch`, `switch -c`,
  and `merge --ff-only`.
- The implementation does not reset, stash, force-push, invent remotes, or
  resolve conflicts.

## Contracts Added or Changed

- Added `RepairWorkReadiness`.
- Added `WorkReadinessRepairOptions`.
- Added `WorkReadinessRepairResult`.
- Added `WorkReadinessAppliedRepair`.
- Extended `WorkReadinessRepairGit` with safe repair primitives.
- Added the CLI flag `chc work doctor --fix-safe`.
- Doctor output now prints repair metadata for diagnostics and applied repairs
  when `--fix-safe` is used.

## Files of Interest

- `internal/core/work_readiness.go`
- `internal/core/work_readiness_service.go`
- `internal/core/work_readiness_service_test.go`
- `internal/app/work.go`
- `internal/app/cli_test.go`

## Validation

- `go test ./internal/core -run 'WorkReadiness|WorkSelectionPreflight'` - pass
- `go test ./internal/app -run 'WorkDoctor'` - pass
- `go test ./internal/core ./internal/app` - pass
- `make lint-strict` - pass
- `make verify` - pass

## Risks and Follow-ups

- Dash still needs a first-class readiness view and repair action; Slice 4
  covers that UI work.
- Safe repairs remain intentionally narrow. Divergent histories, user changes,
  missing canonical remotes, and in-progress Git operations still require
  manual intervention.
