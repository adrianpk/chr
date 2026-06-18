# Slice 1: Work Policy Config

Status: delivered
Slicing: Configurable Work Readiness
Plan: `ops/plans/2026-06-18-configurable-work-readiness-plan.md`
Tracker: `ops/trackers/2026-06-18-configurable-work-readiness-tracker.md`
Branch: `feat/slice-1-work-policy-config`
PR: #313

## Purpose

Move work branch assumptions into workspace configuration so projects can keep
ChrysoCoord's default `dev` workflow or configure a different branch policy.

## Delivered Behavior

- Workspace config now includes a `work` policy section with defaults matching
  the previous hardcoded `dev` plus `origin` behavior.
- Work-selection preflight evaluates the configured branch policy instead of
  hardcoding `dev` and `origin/dev`.
- `integration_branch`, `current_branch`, and `no_git_enforcement` modes are
  recognized.
- App and TUI-facing work paths load the workspace work policy before selecting
  agent work.
- Review-repair ticket ordering remains independent from Git branch policy.

## Implementation Notes

- The first slice keeps the existing `stale_dev` block code for compatibility.
  Later slices add clearer diagnostics and doctor output.
- `RequireUpdatedDev` remains as a compatibility path for existing direct core
  tests and callers that have not yet moved to config-driven policy.
- `current_branch` mode checks cleanliness but does not enforce a specific
  branch name or remote ref.
- `no_git_enforcement` skips Git checks but still honors review-repair ordering.

## Contracts Added or Changed

- Added `WorkConfig`, `WorkBranchPolicy`, and `WorkBranchPolicyMode`.
- Added default work policy:
  - `mode: integration_branch`
  - `integration_branch: dev`
  - `canonical_remote: origin`
  - `require_clean_worktree: true`
  - `require_remote_tracking: true`
  - `allow_safe_repair: true`
  - `allow_branch_creation: false`
- Added `ValidateWorkConfig` and `ValidateWorkBranchPolicy`.
- Generalized `WorkSelectionGit` from `DevContainsOriginDev` to
  `BranchContainsRef`.

## Files of Interest

- `internal/core/config.go`
- `internal/core/work_config.go`
- `internal/core/work_selection_preflight.go`
- `internal/core/work_selection_preflight_test.go`
- `internal/app/work_policy.go`
- `internal/app/ticket.go`
- `internal/app/service_work.go`

## Validation

- `go test ./internal/core -run 'WorkspaceConfig|WorkConfig|ReviewWorkflow|NotificationConfig'` - pass
- `go test ./internal/core -run 'WorkSelectionPreflight'` - pass
- `go test ./internal/app -run 'Work|TicketLoop|TicketStart'` - pass
- `go test ./internal/core ./internal/app` - pass
- `make lint-strict` - pass
- `make verify` - pass

## Risks and Follow-ups

- Diagnostics are still coarse in this slice; Slice 2 adds the dedicated work
  doctor and classified readiness output.
- Safe repair is not implemented in this slice; Slice 3 covers repair actions.
