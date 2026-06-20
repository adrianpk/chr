# Slice 5: Finite Repair Cycle

Status: delivered
Slicing: Serialized Integration Loop
Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`
Tracker: `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`
Branch: `feat/slice-5-finite-repair-cycle`
PR: `#340`

## Purpose

Repair the current ticket when agentic review requests changes, without
allowing unbounded development-review loops or letting later tickets bypass the
blocked repair.

## Delivered Behavior

- `needs_changes` now triggers one immediate repair execution for the same
  ticket before later normal work can be selected.
- Repair reuses the same ticket branch, pull request metadata, and worktree
  path behavior.
- The default repair limit is one repair execution after the initial review
  request.
- If the repair review still returns `needs_changes`, the loop blocks and later
  tickets do not start.
- A successful repair is reviewed again, merged, and only then may the next
  ticket cycle begin.
- Repair worker context includes local blocking findings from the review result
  in addition to the remote review comments URL.

## Implementation Notes

- Durable repair count is derived from existing `review_round` metadata:
  review round `1` means no repair has been consumed, round `2` means the
  default single repair has been consumed.
- The app loop treats repair as part of the same integration cycle, so
  `ProcessedCount` still counts tickets, not raw executions.
- The final merge uses the repair execution record when repair succeeds.
- Worktree path remapping now accepts a ticket path that is already inside the
  target worktree, which lets repair reuse the active worktree safely.
- Missing repair handoff metadata remains a preflight blocker through the
  worker context materializer.

## Contracts Added or Changed

- `core.DefaultTicketReviewRepairLimit` defines the default repair limit.
- `core.IsTicketReviewRepair`, `core.TicketReviewRepairCount`,
  `core.TicketReviewRepairAllowed`, and
  `core.TicketReviewRepairLimitExceeded` define the review-repair policy
  boundary.
- `CodexWorkerRepairHandoff` now carries `BlockingFindings`.
- `StartAgentWorkLoop` now performs at most one repair execution for an
  agentic-review `needs_changes` ticket before either merging or blocking.

## Files of Interest

- `internal/core/ticket_review_repair.go`
- `internal/core/codex_worker_context.go`
- `internal/app/service_work.go`
- `internal/app/ticket.go`
- `internal/app/service_work_test.go`
- `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`

## Validation

- `go test ./internal/core -run 'TicketReviewRepair|RepairHandoff|CodexWorkerContext'` (pass)
- `go test ./internal/app -run 'Repair|AgenticReview|StartAllReadyAgentWork|WorkRunAll'` (pass)
- `go test ./internal/app ./internal/core` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The repair limit is currently fixed at the default contract value. Making it
  configurable can be handled as a later policy slice if needed.
- Dash/event/report visualization for repair phases remains deferred to Slice 6.
