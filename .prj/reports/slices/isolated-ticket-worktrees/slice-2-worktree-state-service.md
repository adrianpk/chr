# Slice 2: Worktree State Service

Status: reviewing
Slicing: Isolated Ticket Worktrees
Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`
Branch: `feat/slice-2-worktree-state-service`
PR: pending

## Purpose

Add the core worktree planning and state service needed before readiness,
runtime, and ticket execution can move away from the user's active checkout.

## Delivered Behavior

- Core can build a deterministic worktree plan from ticket key, ticket branch,
  repository configuration, and workspace settings.
- The default base ref resolves through the canonical remote integration branch,
  normally `origin/dev`.
- Core can fetch the canonical remote, resolve the base SHA, create the ticket
  branch when missing, and add a ticket worktree under the configured root.
- Existing worktrees can be inspected and classified as missing, clean, dirty,
  mismatched, branch-in-use, or orphaned.
- Clean and dirty worktrees are reusable by the service; unsafe states block
  preparation with diagnostics.
- Optional JSON metadata can record advisory worktree identity and execution
  hints without replacing Git as the source of truth.

## Implementation Notes

- The service is isolated in `internal/core` and does not invoke the runtime
  worker or change ticket execution yet.
- Git operations sit behind a small `WorktreeGit` interface so service behavior
  is testable without shelling out in unit tests.
- Worktree paths are derived from configured root plus sanitized ticket key, so
  repeated planning for the same ticket is stable.
- Preparation fetches the configured canonical remote before resolving the base
  ref when the workspace base policy is `remote_integration_branch`.

## Contracts Added or Changed

- Added `WorktreePlanOptions` and `WorktreePlan`.
- Added `WorktreeGit`, `OSWorktreeGit`, and `WorktreeGitEntry`.
- Added `WorktreeService.Prepare` and `WorktreeService.Inspect`.
- Added `WorktreeState`, `WorktreeDiagnosticCode`, `WorktreeInspection`, and
  `WorktreeDiagnostic`.
- Added `WorktreeMetadata`, `WorktreeMetadataPath`, `ReadWorktreeMetadata`, and
  `WriteWorktreeMetadata`.

## Files of Interest

- `internal/core/worktree_plan.go`
- `internal/core/worktree_git.go`
- `internal/core/worktree_service.go`
- `internal/core/worktree_state.go`
- `internal/core/worktree_metadata.go`
- `internal/core/worktree_service_test.go`
- `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
- `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`

## Validation

- `go test ./internal/core -run 'Worktree'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The service is not wired into readiness, doctor, runtime mounting, result
  collection, or branch/PR lifecycle yet. Those integrations are planned for
  later slices in this slicing.
