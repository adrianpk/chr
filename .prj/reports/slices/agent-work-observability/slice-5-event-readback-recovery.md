# Slice 5: Event Readback Recovery

Status: delivered
Slicing: Agent Work Observability
Plan: `ops/plans/2026-06-18-agent-work-observability-plan.md`
Tracker: `ops/trackers/2026-06-18-agent-work-observability-tracker.md`
Branch: `feat/slice-5-event-readback-recovery`
PR: #312

## Purpose

Make persisted event streams useful after interrupted runs by summarizing the
latest valid work state even when the stream is partially malformed.

## Delivered Behavior

Core now exposes an event stream readback summary. The summary reports valid
event count, malformed line count, degraded status, latest known loop, ticket,
execution, phase, level, message, timestamp, and inferred work status.

The CLI now supports:

```text
chc work events <event-stream-path>
```

The command accepts project-relative or absolute event stream paths. It can
summarize both loop streams and execution streams. Missing streams produce an
`unknown` summary instead of an error. Malformed lines mark the summary as
degraded while preserving the latest valid event state.

The recovery path is documented alongside existing execution and loop evidence
inspection.

## Implementation Notes

Readback status is inferred from the latest valid event. Terminal events map to
`succeeded`, `blocked`, `failed`, or `stopped`; non-terminal progress maps to
`running`; empty streams map to `unknown`.

Malformed lines are counted through the existing `ReadAgentEventStream`
diagnostics and do not prevent summary output.

## Contracts Added or Changed

- `core.AgentEventStreamSummary`
- `core.AgentEventReadbackStatus`
- `core.SummarizeAgentEventStream`
- `core.AgentEventReadbackStatusForEvent`
- `chc work events <event-stream-path>`

## Files of Interest

- `internal/core/agent_event_readback.go`
- `internal/core/agent_event_readback_test.go`
- `internal/app/work.go`
- `internal/app/cli_test.go`
- `docs/how-to/find-execution-and-loop-evidence.md`
- `docs/reference/commands.md`
- `ops/specs/2026-06-18-agent-work-observability.md`
- `ops/specs/2026-06-18-agent-work-observability-model.md`

## Validation

- `go test ./internal/core` (pass)
- `go test ./internal/app` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- Readback summarizes event streams but does not resume work automatically.
- Event stream retention remains tied to runtime artifact retention.
