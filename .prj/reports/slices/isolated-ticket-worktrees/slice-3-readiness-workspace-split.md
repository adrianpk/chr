# Slice 3: Readiness Workspace Split

Status: delivered
Slicing: Isolated Ticket Worktrees
Plan: `ops/plans/2026-06-19-isolated-ticket-worktrees-plan.md`
Tracker: `ops/trackers/2026-06-19-isolated-ticket-worktrees-tracker.md`
Branch: `feat/slice-3-readiness-workspace-split`
PR: #330

## Purpose

Make work readiness distinguish the user's checkout from the ticket execution
workspace so future worktree execution can stop blocking on unrelated local
user changes.

## Delivered Behavior

- Readiness input now includes workspace mode plus optional ticket key and
  ticket branch context.
- In `project` mode, existing branch, remote, and clean-worktree behavior is
  preserved.
- In `worktree` mode, dirty or stale user checkout state no longer blocks work
  selection.
- In `worktree` mode, readiness fetches and resolves the configured remote base
  ref, normally `origin/dev`.
- Dirty, mismatched, branch-in-use, and orphaned ticket worktrees block
  readiness when a ticket context is available.
- `chc work doctor` and Dash expose workspace mode, user checkout state, ticket
  worktree state, base ref, and base SHA.
- Safe repair can fetch the configured remote, create a missing worktree root,
  and remove stale metadata for missing ticket worktrees without deleting
  worktrees or user changes.

## Implementation Notes

- Worktree readiness reuses the Slice 2 `WorktreeService` inspection model.
- Work selection skips local checkout branch and cleanliness gates in
  `worktree` mode because the fetched remote base becomes the operational
  source for agent work.
- Ticket worktree inspection is optional. Doctor can inspect one ticket with
  `--ticket` and can override branch metadata with `--ticket-branch`.
- Root creation and metadata pruning are safe repairs only. Ambiguous or dirty
  worktree states remain manual blockers.

## Contracts Added or Changed

- `WorkReadinessReport` now carries `Workspace`, `TicketKey`, `TicketBranch`,
  `UserCheckout`, and `TicketWorktree`.
- Added readiness diagnostics for `dirty_user_checkout`,
  `dirty_ticket_worktree`, `unsafe_ticket_worktree`,
  `remote_base_fetch_failed`, `worktree_root_missing`, and
  `stale_worktree_metadata`.
- Added repair codes `create_worktree_root` and `prune_worktree_metadata`.
- `WorkSelectionPreflightOptions` now accepts `Workspace`.
- `chc work doctor` accepts `--ticket` and `--ticket-branch`.

## Files of Interest

- `internal/core/work_readiness.go`
- `internal/core/work_readiness_service.go`
- `internal/core/work_selection_preflight.go`
- `internal/app/work.go`
- `internal/app/work_policy.go`
- `internal/app/ticket.go`
- `internal/tui/work_readiness.go`
- `internal/tui/dash.go`
- `internal/core/work_readiness_service_test.go`
- `internal/core/work_selection_preflight_test.go`
- `internal/app/cli_test.go`
- `internal/app/service_work_test.go`
- `internal/tui/work_readiness_test.go`

## Validation

- `go test ./internal/core -run 'WorkReadiness|WorkSelectionPreflight|Worktree' && go test ./internal/app -run 'WorkDoctor|PreviewAgentWork|StartAgentWork' && go test ./internal/tui -run 'WorkReadiness|Dash|Model'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Runtime mounting still uses the project checkout. Slice 4 wires the runtime
  workspace mount to the prepared execution workspace.
- Ticket branch selection is still based on available ticket branch metadata or
  explicit doctor input. Slice 6 owns the full branch and PR lifecycle.
