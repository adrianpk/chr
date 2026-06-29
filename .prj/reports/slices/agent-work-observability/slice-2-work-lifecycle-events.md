# Slice 2: Work Lifecycle Events

Status: delivered
Slicing: Agent Work Observability
Plan: `ops/plans/2026-06-18-agent-work-observability-plan.md`
Tracker: `ops/trackers/2026-06-18-agent-work-observability-tracker.md`
Branch: `feat/slice-2-work-lifecycle-events`
PR: #309

## Purpose

Emit durable agent-work lifecycle events from the existing work loop and ticket
execution paths without changing execution, validation, or result application
semantics.

## Delivered Behavior

Agent work now writes append-only JSONL events while `work run-all`, TUI
ready-loop actions, and single-ticket agent work execute.

Loop runs write a loop-scoped stream under the runtime artifact tree. Ticket
executions write execution-scoped streams under their existing run directories.
When a ticket execution belongs to a loop, execution events are also fanned out
to the loop stream so the loop stream can render aggregate progress.

The emitted lifecycle covers loop start and finish, ticket selection, execution
creation, runtime preparation, worker start and finish, validation start and
finish, result application, ticket status updates, blocked outcomes, failed
outcomes, and changed-file detection when the worker reports changed files.

Event writes are best-effort. A persistence failure in the event stream does not
replace or hide the underlying work result.

## Implementation Notes

`core.ExecutionRunner` now accepts an optional event sink and loop id. It emits
execution lifecycle events around the existing harness build, worker, validation,
and result-capture boundaries.

The app layer constructs file sinks from existing runtime artifact paths and
uses a fanout sink for loop-owned executions. Result application events are
emitted after the existing ticket result writers and execution report writers
complete.

## Contracts Added or Changed

- `core.AgentEventFanoutSink`
- `core.NewAgentEventFanoutSink`
- `core.ExecutionRunner.EventSink`
- `core.RunExecutionOptions.LoopID`
- agent-work events are written to:
  - `.chr/user/runtime/runs/<loop-id>/events.jsonl`
  - `.chr/user/runtime/runs/<ticket-key>/<execution-id>/events.jsonl`

## Files of Interest

- `internal/app/agent_work_events.go`
- `internal/app/service_work.go`
- `internal/app/ticket.go`
- `internal/app/service_work_test.go`
- `internal/core/agent_event_sink.go`
- `internal/core/execution_runner.go`

## Validation

- `go test ./internal/app` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Events are persisted but not rendered in Dash until Slice 3.
- Detailed log lookup remains a later slice.
