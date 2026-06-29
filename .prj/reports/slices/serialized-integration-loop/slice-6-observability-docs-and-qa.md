# Slice 6: Observability Docs and QA

Status: delivered
Slicing: Serialized Integration Loop
Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`
Tracker: `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`
Branch: `feat/slice-6-observability-docs-and-qa`
PR: #341

## Purpose

Make the serialized integration loop inspectable and document the final
operator-facing behavior for ticket development, review, repair, merge, and
loop-stop evidence.

## Delivered Behavior

- Agent event streams now distinguish cycle rehydration, review, repair, and
  merge phases.
- Dash shows the latest persisted integration phase and latest event message in
  the current agent-work panel.
- Loop reports include a derived stop reason.
- Agent-work observability, command reference, user guide, ticket loop,
  review workflow, loop reports, and QA runbook docs describe the serialized
  cycle.
- The changelog includes the serialized-loop observability and finite repair
  behavior.
- The tic-tac-toe QA reference includes a resettable two-ticket fixture under
  `.chr/user/fixtures/serialized-loop-tickets/`.

## Implementation Notes

- Event persistence remains best-effort and append-only. New phases are added
  to the existing schema rather than introducing a new event version.
- Dash derives phase and stop context from persisted agent events. It does not
  depend on in-memory worker state.
- Loop report stop reasons are derived from existing loop status, counts, and
  failure reason fields.
- The tic-tac-toe serialized-loop fixture is not placed in `.chr/user/tickets`
  by default, so the normal reset remains a pre-ticket exercise.

## Contracts Added or Changed

- New `AgentEventPhase` values:
  `cycle_rehydrated`, `review_started`, `review_finished`,
  `repair_started`, `repair_finished`, `merge_started`, and
  `merge_finished`.
- `ObservabilityView` now exposes `LatestEvent`.
- Dash current work lines now include `Phase` and `Last event` when available.
- `LoopReport` now includes `StopReason` and renders a `## Stop Reason`
  section.

## Files of Interest

- `internal/core/agent_event.go`
- `internal/core/loop_report.go`
- `internal/app/agent_work_events.go`
- `internal/app/service_work.go`
- `internal/tui/observability.go`
- `internal/tui/dash.go`
- `docs/reference/agent-work-observability.md`
- `docs/how-to/observe-agent-work.md`
- `docs/how-to/qa-tictactoe-agent-work-scenario.md`
- `qa/ref/003/tictactoe/.chr/user/fixtures/serialized-loop-tickets/`

## Validation

- `go test ./internal/core -run 'AgentEvent|LoopReport'` (pass)
- `go test ./internal/app -run 'StartAllReadyAgentWork|WorkRunAll'` (pass)
- `go test ./internal/tui -run 'Dash|Observability|WorkActionEventStream'` (pass)
- `go test ./internal/app ./internal/core ./internal/tui` (pass)
- `make qa-test-tictactoe` (pass)
- `make lint-strict` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- `qa-smoke-tictactoe` was not run in this slice because it resets the active
  QA project. The non-resetting `qa-test-tictactoe` gate passed.
