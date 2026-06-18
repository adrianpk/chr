# Agent Work Observability Model

Status: draft

Date: 2026-06-18

Spec: `ops/specs/2026-06-18-agent-work-observability.md`

## Purpose

This model defines the durable event contract for agent-work observability.

The model is intentionally file-first. Events are append-only JSONL records
stored under the existing runtime artifact tree so they can be rendered by the
TUI, inspected by CLI commands, and preserved as execution evidence.

## Agent Event

`AgentEvent` describes one lifecycle observation from agent work.

Fields:

| Field | Required | Semantics |
| --- | --- | --- |
| `schema_version` | yes | Event schema version. Initial value: `1`. |
| `timestamp` | yes | UTC RFC3339 timestamp for when the event was emitted. |
| `project_id` | yes | ChrysoCoord project id. |
| `loop_id` | no | Loop/run id when the event belongs to a multi-ticket loop. |
| `execution_id` | no | Execution id when the event belongs to one ticket execution. |
| `ticket_key` | no | Ticket key such as `CHC-1`. |
| `phase` | yes | Machine-readable lifecycle phase. |
| `level` | yes | `info`, `warn`, or `error`. |
| `message` | yes | Human-readable concise event message. |
| `refs` | no | Optional structured references to files, branch, container, or commands. |

## Phases

Initial phases:

- `loop_started`
- `loop_finished`
- `ticket_selected`
- `execution_created`
- `branch_prepared`
- `runtime_prepared`
- `worker_started`
- `worker_finished`
- `changed_files_detected`
- `validation_started`
- `validation_finished`
- `result_applied`
- `ticket_status_updated`
- `work_blocked`
- `work_failed`
- `work_stopped`

Phases are stable enough for filtering and rendering. The event `message`
contains display text and may evolve without changing the phase contract.

## Levels

- `info`: normal progress.
- `warn`: degraded or recoverable state.
- `error`: failed or blocked state.

## References

`refs` is an optional object. Initial supported keys:

| Key | Semantics |
| --- | --- |
| `ticket_path` | Repository-relative ticket path. |
| `branch` | Git branch or worktree branch prepared for the ticket. |
| `container_id` | Runtime container id when known. |
| `validation_command` | Command used for validation. |
| `context_path` | Runtime context artifact path. |
| `result_path` | Runtime result JSON path. |
| `execution_report_path` | Execution report path. |
| `result_application_path` | Result application report path. |
| `runtime_log_path` | Runtime log path. |
| `validation_log_path` | Validation log path. |

Implementations may add more keys, but readers must tolerate unknown keys.

## Storage

Event files should be append-only JSONL.

Preferred paths:

```text
.chc/user/runtime/runs/<loop-id>/events.jsonl
.chc/user/runtime/runs/<ticket-key>/<execution-id>/events.jsonl
```

When there is no loop id, execution-scoped events are still written under the
execution run directory.

If a loop id exists, loop-level events should be written to the loop stream.
Ticket execution events may be written to both the loop stream and the
execution stream if the implementation needs a complete aggregate stream.

## Readback Summary

Event stream readback should summarize the latest valid event even when the
stream contains malformed lines.

The summary should include:

- event count;
- malformed line count;
- degraded flag;
- latest loop id;
- latest ticket key;
- latest execution id;
- latest phase;
- latest level;
- latest message;
- inferred status.

Known statuses are `unknown`, `running`, `succeeded`, `blocked`, `failed`, and
`stopped`.

## Rendering Model

The TUI should render recent events as lines:

```text
<HH:MM:SS> <ticket-key?> <message>
```

Rendering rules:

- show newest relevant events without raw JSON;
- keep timestamps in the user's local display context when practical;
- prefer event `message` over reconstructing prose from `phase`;
- show `ticket_key` when available;
- show error/warn levels visibly but without blocking the event stream.

## Invariants

- Event writes are append-only.
- A malformed event line must not make the whole stream unreadable.
- Missing optional refs must not fail rendering.
- Event persistence failure must not hide the underlying work result.
- Event rendering failure must not stop active work.
- Readers must tolerate unknown phases and refs.

## Retention

Event streams are runtime evidence. They should follow the same retention and
cleanup posture as existing execution reports and runtime logs unless a later
maintenance spec defines different behavior.
