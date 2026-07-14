# Slice 1: Workspace Execution Config

Status: delivered
Slicing: Isolated Ticket Worktrees
Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`
Branch: `feat/slice-1-workspace-execution-config`
PR: #328

## Purpose

Add the project-local configuration and runtime planning fields needed for
isolated ticket worktrees without changing the current execution behavior.

This slice makes the future workspace split explicit while keeping existing
managed projects on project-mounted execution by default.

## Delivered Behavior

- Added `work.workspace_mode`, `work.worktree_root`, `work.worktree_base`, and
  `work.worktree_cleanup` config fields.
- Added controlled values for project execution, worktree execution, worktree
  base selection, and cleanup policy.
- Defaulted workspace execution to `project`, preserving the current writable
  project checkout mount.
- Added parser and validation coverage for explicit worktree config, omitted
  defaults, and invalid workspace policy values.
- Extended `RuntimePlan` with project checkout, execution workspace, and
  workspace mode fields.
- Kept Docker mount planning unchanged in this slice; the writable mount still
  points at the project checkout.
- Updated reference docs to describe the new config fields and migration
  posture.

## Implementation Notes

- Worktree mode is parsed and validated but does not yet create Git worktrees.
- Runtime planning records `ExecutionWorkspaceDir` so later slices can switch
  mount behavior without another config contract change.
- `RuntimeValidationOptions` mirrors `RuntimePlanOptions` so validation callers
  can carry workspace metadata when later slices start using it.
- Parser and validation helpers were split to satisfy strict lint complexity
  limits.

## Contracts Added or Changed

- Added `core.WorkWorkspaceConfig`.
- Added `core.WorkWorkspaceMode`, `core.WorktreeBasePolicy`, and
  `core.WorktreeCleanupPolicy`.
- Added default workspace execution config under `DefaultWorkConfig`.
- Extended `core.RuntimePlanOptions`, `core.RuntimePlan`, and
  `core.RuntimeValidationOptions` with workspace execution metadata.
- Extended `.chr/default/config.yml` `work:` reference fields.

## Files of Interest

- `internal/core/config.go`
- `internal/core/work_config.go`
- `internal/core/config_test.go`
- `internal/core/runtime_plan.go`
- `internal/core/runtime_plan_test.go`
- `internal/core/runtime_validate.go`
- `docs/reference/work-readiness.md`
- `docs/reference/project-workspace.md`
- `docs/reference/runtime-config.md`

## Validation

- `go test ./internal/core -run 'TestLoadWorkspaceConfig.*Work|TestValidateWorkConfigRejectsInvalid.*Config|TestPlanRuntime|TestCodexCredentialMountsUseReadOnlyRootCodexPaths|TestValidateRuntime'` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Worktree mode is configuration-only in this slice. Later slices must ensure
  readiness, runtime mounts, result discovery, PR lifecycle, and cleanup all
  read from the same workspace metadata.
- Default remains `project` until isolated worktree execution is proven by the
  later QA slice.
