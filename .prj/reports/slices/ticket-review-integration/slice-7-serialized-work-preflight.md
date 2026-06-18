# Slice 7: Serialized Work Preflight

Status: delivered
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-7-serialized-work-preflight`
PR: #279

## Purpose

Prevent agent work selection from skipping unresolved review repairs or running
concurrent development and review operations in the same workspace.

## Delivered Behavior

- Development and review execution paths acquire a workspace-level agent
  operation lock before mutating ticket or review state.
- Work selection detects `ready_for_dev` tickets with `changes_requested` or
  `needs_changes` review metadata.
- Loop and run-all selection prioritize review repair tickets before later
  development work.
- Starting a non-repair ticket is blocked while a repair ticket is pending.
- Work selection blocks git workspaces that are on a branch other than `dev` or
  have uncommitted changes.

## Implementation Notes

- The reusable preflight contract lives in `internal/chc` so CLI, TUI, and
  future adapters share one rule set.
- Non-git workspaces are allowed through the git freshness check to preserve
  existing test and scratch workspace behavior.
- The lock is file-based under `.chc/user/runtime/agent-operation.lock`.
- Dry-run review paths do not take the operation lock.

## Contracts Added or Changed

- Added `AcquireAgentOperationLock` and `AgentOperationLock.Release`.
- Added `RunWorkSelectionPreflight` with machine-readable block reason codes.
- Added block reason codes:
  - `review_repair_pending`
  - `branch_policy_blocked`
- Added a local style check for newly added multiline Go function signatures.

## Files of Interest

- `internal/chc/work_selection_preflight.go`
- `internal/chc/work_selection_preflight_test.go`
- `internal/chcapp/ticket.go`
- `internal/chcapp/service_work.go`
- `scripts/check-go-func-signatures.sh`
- `Makefile`

## Validation

- `scripts/check-go-func-signatures.sh`
- `go test ./internal/chc -run 'WorkSelection|AgentOperationLock'`
- `go test ./internal/chcapp -run 'AgentWork|TicketLoop|TicketReview'`
- `go test ./internal/chc ./internal/chcapp`
- `golangci-lint run ./internal/chc ./internal/chcapp`

## Risks and Follow-ups

- The git freshness check verifies local `dev` and clean worktree state, but it
  does not yet verify remote `origin/dev` freshness.
- Slice 8 still needs repair handoff context that points repair workers back to
  the existing branch and pull request comments.
