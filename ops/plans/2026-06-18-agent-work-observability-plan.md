# Agent Work Observability Plan

Status: planned

Active Slice: Slice 1 - agent event model

Tracker: `ops/trackers/2026-06-18-agent-work-observability-tracker.md`

Spec: `ops/specs/2026-06-18-agent-work-observability.md`

Model: `ops/specs/2026-06-18-agent-work-observability-model.md`

## Purpose

Add durable, human-readable feedback for long-running agent work without
loosening the current modal safety behavior in the TUI.

The first iteration should make foreground work observable: users should see
what the agent loop is doing, which ticket is active, and where detailed logs
live, while navigation remains locked until the action completes.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice updates its implementation report under
  `.prj/reports/slices/agent-work-observability/`.
- Each task gets one conventional commit.
- After each merged slice, return to `dev` before follow-up work.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | agent event model | `feat/slice-1-agent-event-model` | `feat(slice-1): agent event model` | `.prj/reports/slices/agent-work-observability/slice-1-agent-event-model.md` |
| Slice 2 | work lifecycle events | `feat/slice-2-work-lifecycle-events` | `feat(slice-2): work lifecycle events` | `.prj/reports/slices/agent-work-observability/slice-2-work-lifecycle-events.md` |
| Slice 3 | dash live work stream | `feat/slice-3-dash-live-work-stream` | `feat(slice-3): dash live work stream` | `.prj/reports/slices/agent-work-observability/slice-3-dash-live-work-stream.md` |
| Slice 4 | execution log lookup | `feat/slice-4-execution-log-lookup` | `feat(slice-4): execution log lookup` | `.prj/reports/slices/agent-work-observability/slice-4-execution-log-lookup.md` |
| Slice 5 | event readback recovery | `feat/slice-5-event-readback-recovery` | `feat(slice-5): event readback recovery` | `.prj/reports/slices/agent-work-observability/slice-5-event-readback-recovery.md` |

## Slice 1: Agent Event Model

Goal: add the durable event model and JSONL persistence/readback primitives.

Behavior:

- Core defines an `AgentEvent` contract aligned with the spec model.
- Events can be appended to JSONL files under runtime run directories.
- Event readers tolerate malformed lines and unknown fields.
- Unit tests cover append/read behavior and malformed-line tolerance.

Tasks:

- T1.1 Add the `AgentEvent` model and validation helpers.
- T1.2 Add JSONL append/read primitives for runtime event streams.
- T1.3 Add unit tests for persistence, ordering, and malformed-line handling.
- T1.4 Add the slice implementation report and validation evidence.

## Slice 2: Work Lifecycle Events

Goal: emit events from the existing work loop and execution lifecycle.

Behavior:

- `chc work run-all` emits loop-level events.
- Ticket execution emits ticket/execution-level events for start, worker
  finish, validation, result application, and blocked/failure outcomes.
- Existing result application behavior is unchanged.
- Runtime reports continue to be written as before.

Tasks:

- T2.1 Wire an event sink into work loop and ticket execution paths.
- T2.2 Emit loop and ticket lifecycle events at high-signal boundaries.
- T2.3 Cover lifecycle emission with service or CLI tests.
- T2.4 Add the slice implementation report and validation evidence.

## Slice 3: Dash Live Work Stream

Goal: replace the static work-running overlay with live events in Dash.

Behavior:

- While foreground work is running, Dash shows recent events and a lock notice.
- Navigation remains blocked while `workBusy` is active.
- The event stream refreshes during the running action.
- Existing Dash ready-queue behavior is preserved after the action completes.

Tasks:

- T3.1 Add a TUI read model for recent agent events.
- T3.2 Render the work-running screen as a live event stream with lock notice.
- T3.3 Refresh event state while foreground work is active.
- T3.4 Add TUI tests for modal lock behavior and event rendering.
- T3.5 Add the slice implementation report and validation evidence.

## Slice 4: Execution Log Lookup

Goal: expose a stable way to locate deeper runtime logs for an execution.

Behavior:

- Users can locate runtime and validation logs from an execution id or recent
  event ref.
- Dash event rendering can show a concise hint for deeper logs when available.
- The first implementation may be CLI-first if that keeps the TUI simple.

Tasks:

- T4.1 Add a log lookup command or extend an existing report/runtime command.
- T4.2 Connect event refs to the lookup output or hint text.
- T4.3 Add command tests for successful and missing-log lookup.
- T4.4 Add the slice implementation report and validation evidence.

## Slice 5: Event Readback Recovery

Goal: make persisted events useful after interrupted runs.

Behavior:

- A readback path summarizes latest loop/ticket/execution status from event
  streams.
- Malformed or partial event streams produce degraded summaries rather than
  hard failures.
- Documentation explains how to inspect the last known state after an
  interrupted work run.

Tasks:

- T5.1 Add event stream summary/readback behavior.
- T5.2 Surface latest known work state in CLI or Dash diagnostics.
- T5.3 Document recovery inspection for interrupted runs.
- T5.4 Add the slice implementation report and validation evidence.

