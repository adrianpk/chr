# Slice 6: Branch and PR Lifecycle

Status: delivered
Slicing: Isolated Ticket Worktrees
Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`
Branch: `feat/slice-6-branch-and-pr-lifecycle`
PR: #333

## Purpose

Connect worktree-mode ticket execution to the one-ticket, one-branch, one-PR
lifecycle without moving push or pull request ownership into the worker
container.

## Delivered Behavior

- Ticket execution now selects an implementation branch before runtime
  materialization.
- Existing ticket `branch` metadata remains authoritative.
- Tickets without branch metadata derive a deterministic `chr/<ticket-slug>`
  implementation branch.
- Worktree-mode ticket execution prepares the ticket worktree before runtime
  validation and worker context materialization.
- Missing ticket branches are created from the configured worktree base ref.
- Existing ticket branches must contain the configured base ref; unrelated
  branches block with a clear diagnostic.
- Derived branch metadata is written to the ticket copy inside the prepared
  worktree, leaving the user's checkout untouched.
- Execution records, runtime plans, worktree metadata, and agent events expose
  the prepared branch and execution workspace.
- `run-all` remains sequential and prepares only the selected ticket worktree,
  not the full ready queue.

## Implementation Notes

- Branch selection lives in `internal/core` so execution, review, and later
  cleanup paths can reuse the same ticket-to-branch rule.
- Worktree preparation happens in the ticket-start preflight path before
  `ValidateRuntime`, so the runtime plan points at a real execution workspace.
- Dry-run ticket start builds the same plan without mutating Git or the
  filesystem.
- Host-side result application continues to read from execution workspace
  metadata in worktree mode. The worker container still does not push branches
  or create pull requests.
- Project-mode execution keeps the existing project checkout behavior.

## Contracts Added or Changed

- `SelectTicketBranch` resolves ticket branch metadata or derives a stable
  branch from ticket key and title.
- `WorktreeGit` now checks whether an existing ticket branch contains a base
  ref through `merge-base --is-ancestor`.
- Worktree preparation blocks existing branches that do not contain the
  selected base ref.
- `ticket start` and service-driven ticket execution prepare worktrees when
  `work.workspace_mode` is `worktree`.
- Worktree-mode execution emits a `branch_prepared` agent event before
  `execution_created`.

## Files of Interest

- `internal/core/ticket_branch.go`
- `internal/core/worktree_git.go`
- `internal/core/worktree_service.go`
- `internal/app/ticket.go`
- `internal/app/agent_work_events.go`
- `internal/app/service_work_test.go`
- `internal/app/cli_test.go`

## Validation

- `go test ./internal/core -run 'TestSelectTicketBranch|TestWorktreeService|TestChangedFileDiscoverer|TestNewResultApplication'` (pass)
- `go test ./internal/app -run 'TestServiceStartAgentWorkTicketPreparesWorktreeWorkspace|TestServiceStartAgentWorkLoopPreparesOnlySelectedWorktree|TestServiceStartAgentWorkTicketReusesTicketExecution|TestServiceStartAllReadyAgentWorkEmitsLifecycleEvents|TestTicketApplyResultDiscoversWorktreeChangedFiles|TestTicketApplyResultAppliesExecutionResult'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- This slice does not add a new provider-specific PR creation command. It
  preserves the host-side ownership boundary and records the branch/worktree
  state needed by the existing review and later merge lifecycle.
- Review repair and post-merge worktree cleanup remain in Slice 7.
- Worktree mode remains opt-in until the final QA proving slice validates the
  full cycle end to end.
