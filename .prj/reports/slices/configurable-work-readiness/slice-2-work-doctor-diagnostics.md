# Slice 2: Work Doctor Diagnostics

Status: delivered
Slicing: Configurable Work Readiness
Plan: `ops/plans/2026-06-18-configurable-work-readiness-plan.md`
Tracker: `ops/trackers/2026-06-18-configurable-work-readiness-tracker.md`
Branch: `feat/slice-2-work-doctor-diagnostics`
PR: `#314`

## Purpose

Expose work readiness diagnostics without starting agent work, so users can see
why a work run is blocked before launching an expensive agent loop.

## Delivered Behavior

- Added a core `WorkReadinessReport` model with status, diagnostics, severity,
  policy fields, current branch, and future repair metadata.
- Added a core readiness service that evaluates configured work policy against
  Git state.
- Added `chr work doctor` to print readiness status for the resolved project.
- Classified common blockers such as wrong branch, missing integration branch,
  dirty worktree, visible chr runtime artifacts, missing remote ref,
  branch behind remote, Git operation in progress, Git unavailable, and invalid
  work policy.
- Missing remote refs are reported as `missing_remote_ref`, not surfaced as raw
  `exit status 128`.

## Implementation Notes

- The doctor is read-only. It does not start work and does not repair state in
  this slice.
- `no_git_enforcement` mode returns ready without calling Git.
- Runtime artifacts and user-visible worktree changes are split into separate
  diagnostics so Slice 3 can repair chr-owned runtime dirtiness without
  hiding user edits.
- CLI output uses stable line-oriented fields consistent with existing work
  commands.

## Contracts Added or Changed

- Added `WorkReadinessStatus`.
- Added `WorkReadinessDiagnosticCode`.
- Added `WorkReadinessSeverity`.
- Added `WorkReadinessRepairCode`.
- Added `CheckWorkReadiness`.
- Added `chr work doctor`.

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

- Safe repair is intentionally deferred to Slice 3.
- Dash still surfaces work readiness through the existing error path; Slice 4
  adds TUI guidance.
