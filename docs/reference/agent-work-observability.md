# Agent Work Observability

## Purpose

Agent work observability is the contract for seeing what ChrysoCoord is doing
while ticket execution or a ready-ticket queue is running.

Observability is file-first. Runtime work writes durable reports, logs, result
artifacts, and append-only event streams under the project runtime artifact
tree. The TUI renders a human-readable live view from those records, and the CLI
can inspect the same evidence after the fact.

## Event Streams

Agent events are JSONL records written under runtime run directories.

Preferred paths:

```text
.chc/user/runtime/runs/<loop-id>/events.jsonl
.chc/user/runtime/runs/<ticket-key>/<execution-id>/events.jsonl
```

Loop-level streams describe queue progress. Execution-level streams describe
one ticket execution. Implementations may mirror execution events into a loop
stream so Dash can show a complete run-all timeline.

## Event Fields

Each event records:

| Field | Meaning |
| --- | --- |
| `schema_version` | Event schema version. |
| `timestamp` | UTC RFC3339 event time. |
| `project_id` | Project id known to ChrysoCoord. |
| `loop_id` | Loop id when the event belongs to a multi-ticket run. |
| `execution_id` | Execution id when the event belongs to one ticket run. |
| `ticket_key` | Ticket key such as `CHC-1`. |
| `phase` | Machine-readable lifecycle phase. |
| `level` | `info`, `warn`, or `error`. |
| `message` | Human-readable event message. |
| `refs` | Optional structured references to files, commands, or runtime state. |

Readers must tolerate unknown fields, unknown refs, unknown phases, and
malformed lines.

## Phases

Known phases include:

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

The `phase` is stable enough for filtering and summaries. The `message` is the
preferred display text and may evolve.

## References

Event refs may include:

| Ref | Meaning |
| --- | --- |
| `ticket_path` | Project-relative ticket path. |
| `branch` | Git branch prepared for the work. |
| `container_id` | Runtime container id when known. |
| `validation_command` | Validation command used for the run. |
| `context_path` | Runtime context artifact path. |
| `result_path` | Runtime result JSON path. |
| `execution_report_path` | Execution report path. |
| `result_application_path` | Result application report path. |
| `runtime_log_path` | Runtime log path. |
| `validation_log_path` | Validation log path. |

## TUI Rendering

Dash renders recent agent events as human-readable lines while work is active
or after recent work has finished.

Dash should show:

- current work readiness;
- ready queue;
- current loop or execution state;
- recent event messages;
- recent execution and loop outcomes;
- safe repair hints when readiness blocks work.

Dash refreshes from persisted local state. Refreshing Dash does not start work
or mutate ticket files.

## CLI Readback

Summarize an event stream:

```sh
chc work events .chc/user/runtime/runs/CHC-1/exec-1/events.jsonl
chc work events .chc/user/runtime/runs/loop-1/events.jsonl
```

The readback summary includes event counts, malformed line counts, latest known
ticket, latest execution, latest phase, latest level, latest message, and an
inferred status.

Known inferred statuses are:

- `unknown`
- `running`
- `succeeded`
- `blocked`
- `failed`
- `stopped`

Malformed lines degrade the summary but do not make the whole stream
unreadable.

## Runtime Logs

Locate detailed runtime artifacts for one execution:

```sh
chc work logs <execution-id>
```

The command prints artifact paths and existence flags for:

- run directory;
- runtime log;
- validation log;
- result file;
- execution report;
- event stream.

Use the runtime log for the closest expanded view of the Codex/Docker worker
activity. Use the event stream for the human-readable operational timeline.

## Reports

Reports summarize completed evidence:

```sh
chc report list --kind execution
chc report list --kind loop
chc report show --execution <execution-id>
chc report show --loop <loop-id>
```

Execution and loop reports are the normal entry point for reviewing what
happened after a run. Event streams and logs are the lower-level evidence.

## Invariants

- Event writes are append-only.
- Event persistence failure must not hide the underlying work result.
- Event rendering failure must not stop active work.
- Missing optional refs must not fail rendering.
- Runtime logs remain the detailed worker evidence.
- Reports remain the summarized review evidence.
