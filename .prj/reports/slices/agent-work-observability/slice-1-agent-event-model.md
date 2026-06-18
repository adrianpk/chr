# Slice 1: Agent Event Model

Status: reviewing
Slicing: Agent Work Observability
Plan: `ops/plans/2026-06-18-agent-work-observability-plan.md`
Tracker: `ops/trackers/2026-06-18-agent-work-observability-tracker.md`
Branch: `feat/slice-1-agent-event-model`
PR: pending

## Purpose

Add the durable core event model and file-backed JSONL stream primitives needed
for live agent-work observability.

## Delivered Behavior

Core now defines an `AgentEvent` contract with schema version, timestamp,
project id, optional loop/execution/ticket ids, lifecycle phase, severity level,
human-readable message, and optional structured refs.

The package also exposes file-backed append and read behavior for
`events.jsonl` streams. Writers append one validated event per JSON line under
runtime run directories. Readers return valid events in order and preserve
malformed-line diagnostics instead of failing the whole stream.

## Implementation Notes

This slice intentionally does not wire event emission into work execution yet.
It only adds the shared model and persistence boundary required by later
slices.

Malformed JSON lines and structurally invalid event lines are both reported in
the read result. Missing event files read as empty streams so future TUI polling
can tolerate streams that have not been created yet.

## Contracts Added or Changed

- `core.AgentEvent`
- `core.AgentEventPhase`
- `core.AgentEventLevel`
- `core.AgentEventRefs`
- `core.AgentEventSink`
- `core.FileAgentEventSink`
- `core.AgentEventStream`
- `core.AppendAgentEvent`
- `core.ReadAgentEventStream`
- `core.AgentEventStreamPath`

## Files of Interest

- `internal/core/agent_event.go`
- `internal/core/agent_event_stream.go`
- `internal/core/agent_event_test.go`

## Validation

- `go test ./internal/core` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- No runtime code emits events yet; that is Slice 2.
- The TUI still renders the static running screen until Slice 3 wires the event
  read model into Dash.

