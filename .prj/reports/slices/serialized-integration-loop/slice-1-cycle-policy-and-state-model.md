# Slice 1: Cycle Policy and State Model

Status: reviewing
Slicing: Serialized Integration Loop
Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`
Tracker: `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`
Branch: `feat/slice-1-cycle-policy-and-state-model`
PR: pending

## Purpose

Define the core integration-cycle vocabulary and policy mapping that later
slices can use to enforce one selected ticket through development, review,
repair, and merge before selecting the next normal ticket.

## Delivered Behavior

- Added core value types for integration cycle mode, phase, status, blocker,
  and review decision.
- Added policy mapping from existing review workflow modes:
  `human_review`, `agentic_review`, and `trusted_auto_merge`.
- Added side-effect-free selected-ticket cycle planning.
- Documented that `trusted_auto_merge` skips a separate review role but still
  requires merge before later ticket work continues.

## Implementation Notes

- Slice 1 intentionally does not mutate tickets, run workers, merge branches,
  or change `run-all` behavior.
- The planner validates project identity, ticket identity, and review workflow
  config before returning a pending selected-ticket plan.
- Dedicated `integration_cycles` persistence is deferred until a runner slice
  needs durable state that existing ticket, execution, review, merge, and loop
  records cannot represent cleanly.

## Contracts Added or Changed

- `IntegrationCyclePolicyFromReviewConfig` maps configured review mode to cycle
  policy.
- `PlanIntegrationCycle` returns selected-ticket cycle metadata without IO.
- Integration cycle constants now define the supported vocabulary for later
  runner, review, repair, merge, and observability slices.
- Documentation now describes serialized cycle policy semantics in the ticket
  loop and review workflow references.

## Files of Interest

- `internal/core/integration_cycle.go`
- `internal/core/integration_cycle_test.go`
- `docs/reference/ticket-loop.md`
- `docs/reference/review-workflow-config.md`
- `ops/specs/2026-06-20-serialized-integration-loop-model.md`
- `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`

## Validation

- `go test ./internal/core` (pass)
- `go test ./internal/core -run 'ReviewWorkflow|IntegrationCycle'` (pass)
- `go test ./internal/core -run IntegrationCycle` (pass)
- `go test ./internal/core -run 'IntegrationCycle'` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The cycle contracts are not wired into `run-all` yet. Slice 2 introduces the
  one-ticket cycle runner and the first executable barrier.
- Merge behavior remains unchanged until the merge orchestration slice.
