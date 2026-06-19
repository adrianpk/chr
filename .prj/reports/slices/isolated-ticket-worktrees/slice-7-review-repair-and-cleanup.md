# Slice 7: Review Repair and Cleanup

Status: reviewing
Slicing: Isolated Ticket Worktrees
Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`
Branch: `feat/slice-7-review-repair-and-cleanup`
PR: #334

## Purpose

Close the lifecycle gap after ticket branches and PRs exist: review repair
work should reuse the ticket branch workspace, and merged tickets should clean
up only when the ticket worktree state is safe.

## Delivered Behavior

- Review-repair tickets reuse existing ticket branch metadata when worktree
  execution starts.
- Missing review-repair worktrees are prepared from the recorded ticket branch
  instead of deriving a new branch.
- Successful ticket merge can remove a clean ticket worktree when
  `work.worktree_cleanup` is `after_merge`.
- Cleanup is idempotent for already-missing worktrees and prunes stale metadata
  when present.
- Cleanup blocks dirty, mismatched, branch-in-use, and orphaned worktree states
  instead of deleting them.
- Readiness now reports orphaned ticket worktrees with a dedicated diagnostic.
- Safe repair can remove an empty orphaned ticket worktree path, but refuses to
  remove a non-empty orphaned path.

## Implementation Notes

- Cleanup lives in `internal/core` as a lifecycle operation that shares the
  existing worktree planning and inspection boundary.
- The ticket merge command invokes cleanup only after a provider merge succeeds
  and the ticket is written as merged.
- The app command prints cleanup status when cleanup ran, giving the operator a
  concise post-merge signal without making cleanup mandatory in project mode.
- Orphaned worktree repair is intentionally narrow: only empty directories are
  removed automatically.
- Review-repair branch reuse is tested at service level because branch metadata
  is already attached when review refs are added.

## Contracts Added or Changed

- `CleanupMergedTicketWorktree` adds a core lifecycle contract for post-merge
  worktree cleanup.
- `WorktreeGit` now exposes `RemoveWorktree` for `git worktree remove`.
- `TicketMergeCommandResult` includes `WorktreeCleanup`.
- `ticket merge` can print `worktree_cleanup` and `worktree_cleanup_detail`.
- Work readiness adds `orphaned_ticket_worktree` and
  `remove_orphaned_worktree`.

## Files of Interest

- `internal/core/worktree_cleanup.go`
- `internal/core/ticket_merge_command.go`
- `internal/core/work_readiness.go`
- `internal/core/work_readiness_service.go`
- `internal/app/ticket.go`
- `internal/app/service_work_test.go`
- `internal/core/worktree_service_test.go`
- `internal/core/ticket_merge_command_test.go`
- `internal/core/work_readiness_service_test.go`

## Validation

- `go test ./internal/core -run 'TestRepairWorkReadinessRemovesEmptyOrphanedTicketWorktree|TestRepairWorkReadinessDoesNotRemoveNonEmptyOrphanedTicketWorktree|TestRepairWorkReadinessPrunesStaleWorktreeMetadata|TestCheckWorkReadinessBlocksDirtyTicketWorktree'` (pass)
- `go test ./internal/core -run 'Worktree|WorkReadiness|TicketMergeCommand'` (pass)
- `go test ./internal/app -run 'TestTicketMerge|TestServiceStartAgentWorkTicketReusesReviewRepairBranch|TestServiceStartAgentWorkTicketPreparesWorktreeWorkspace|TestServiceStartAgentWorkLoopPreparesOnlySelectedWorktree|TestWorkDoctor'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- This slice does not add provider API logic to recover a branch from a PR
  whose ticket metadata lacks `branch`; the existing review-ref attach flow
  records branch metadata separately.
- Worktree mode remains opt-in until the final QA proving slice validates the
  complete cycle end to end.
