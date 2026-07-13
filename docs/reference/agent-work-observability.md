# Agent Work Observability

## Purpose

Agent work observability is the contract for seeing what Chryso is doing
while ticket execution or a ready-ticket queue is running.

Observability is file-first. Runtime work writes durable reports, logs, result
artifacts, and append-only event streams under the active shared operational
namespace. The TUI renders a human-readable live view from those records, and
the CLI can inspect the same evidence after the fact.

## Event Streams

Agent events are JSONL records written under runtime run directories.

Preferred paths:

```text
ops/<namespace>/run/<loop-artifact-id>/evidence/events.jsonl
ops/<namespace>/run/<execution-artifact-id>/evidence/events.jsonl
```

`<namespace>` is `default` unless `OPS_USER` is configured.

Loop-level streams describe queue progress. Execution-level streams describe
one ticket execution. Implementations may mirror execution events into a loop
stream so Dash can show a complete run-all timeline.

## Event Fields

Each event records:

| Field | Meaning |
| --- | --- |
| `schema_version` | Event schema version. |
| `timestamp` | UTC RFC3339 event time. |
| `project_id` | Project id known to Chryso. |
| `loop_id` | Loop id when the event belongs to a multi-ticket run. |
| `execution_id` | Execution id when the event belongs to one ticket run. |
| `ticket_key` | Ticket key such as `CHR-1`. |
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
- `cycle_rehydrated`
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
- `review_started`
- `review_finished`
- `repair_started`
- `repair_finished`
- `merge_started`
- `merge_finished`
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
| `ticket_branch` | Git branch prepared for the work. |
| `review_ref` | Pull request or review surface URL. |
| `review_comments_url` | Review comments URL used for repair handoff. |
| `merge_commit` | Merge commit SHA when the provider returns one. |
| `merge_blocker` | Compact merge blocker when merge fails or blocks. |
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
- current or latest integration phase;
- latest event message or stop reason;
- recent event messages;
- ticket, execution report, runtime log, and validation log refs when present;
- recent execution and loop outcomes;
- safe repair hints when readiness blocks work.

Dash refreshes from persisted local state. Refreshing Dash does not start work
or mutate ticket files.

## CLI Readback

Summarize an event stream:

```sh
chr work events ops/default/run/<execution-artifact-id>/evidence/events.jsonl
chr work events ops/default/run/<loop-artifact-id>/evidence/events.jsonl
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
chr work logs <execution-id>
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
chr report list --kind execution
chr report list --kind loop
chr report show --execution <execution-id>
chr report show --loop <loop-id>
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
