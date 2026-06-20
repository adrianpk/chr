# Slice 2: One-Ticket Cycle Runner

Status: delivered
Slicing: Serialized Integration Loop
Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`
Tracker: `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`
Branch: `feat/slice-2-one-ticket-cycle-runner`
PR: `#337`

## Purpose

Replace development-only multi-ticket processing in `work run-all` with a
serialized integration cycle runner that selects one ticket, applies
development, rehydrates state, and stops before later work if the selected
ticket still requires review or merge.

## Delivered Behavior

- `work run-all` now selects one ticket per integration cycle.
- The runner re-reads ticket index state and workspace config before any
  continuation.
- Successful development in the default `human_review` mode stops the loop as a
  blocked integration handoff instead of starting the next ready ticket.
- Dry-run output identifies the first cycle ticket and the number of remaining
  eligible tickets.
- `--max-iterations` is available on `work run-all` and applies to integration
  cycles.

## Implementation Notes

- The historical `ticket loop` command still uses its existing loop execution
  path. This slice scopes the new barrier to `work run-all`.
- The runner reuses existing loop records, execution records, event emission,
  result application, and notification behavior.
- A blocked loop in this slice can mean an expected integration barrier, not
  necessarily a failed development execution.

## Contracts Added or Changed

- `StartAgentWorkLoop` delegates to a serialized integration cycle runner.
- `PreviewAgentWorkLoop` previews the first selected cycle rather than a stale
  multi-ticket queue.
- `work run-all --dry-run` prints `cycle_ticket` and `remaining_eligible`.
- `work run-all --max-iterations` rejects negative values and counts
  integration cycles.

## Files of Interest

- `internal/app/service_work.go`
- `internal/app/work.go`
- `internal/app/service_work_test.go`
- `internal/app/cli_test.go`
- `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`

## Validation

- `go test ./internal/app -run 'AgentWorkLoop|WorkRunAll'` (pass)
- `go test ./internal/app -run 'AgentWorkLoop|WorkRunAll|StartAllReadyAgentWork'` (pass)
- `go test ./internal/app ./internal/core` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Agentic review is not implemented in this slice. Slice 3 adds the reviewer
  phase.
- Merge-before-next-ticket behavior is not implemented in this slice. Slice 4
  adds merge orchestration before later normal work can continue.
