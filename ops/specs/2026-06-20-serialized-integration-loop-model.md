# Serialized Integration Loop Model

Status: planned

Date: 2026-06-20

Spec: `ops/specs/2026-06-20-serialized-integration-loop.md`

## Purpose

This model defines the durable state needed for serialized ticket integration
cycles.

The model extends the existing loop, execution, review, merge, and ticket
metadata surfaces. It avoids relying on in-memory queue state after a ticket
cycle finishes.

## Existing Models

The implementation must reuse existing models where possible:

- `LoopRunRecord`
- `ExecutionRecord`
- `IndexedTicket`
- ticket frontmatter review fields
- ticket frontmatter merge fields
- worktree metadata
- execution reports
- loop reports

New state should be additive and compatible with existing records.

## Integration Cycle

An integration cycle represents one attempt to take one ticket from selected
work through development, review policy, merge, and finalization.

Suggested attributes:

| Attribute | Meaning |
| --- | --- |
| `id` | Stable cycle id. |
| `loop_id` | Parent loop run id. |
| `project_id` | Managed project id. |
| `ticket_key` | Selected ticket key. |
| `mode` | Review workflow mode used for this cycle. |
| `phase` | Current or final phase. |
| `status` | Final cycle status. |
| `execution_id` | Development execution record id when present. |
| `review_artifact_path` | Local review artifact path when present. |
| `review_result` | Review outcome when present. |
| `repair_count` | Number of repair cycles consumed. |
| `merge_result` | Merge result when present. |
| `blocker` | Compact blocker code when blocked. |
| `started_at` | Cycle start timestamp. |
| `finished_at` | Cycle finish timestamp when done. |
| `updated_at` | Last mutation timestamp. |

## Cycle Phases

Allowed phases:

- `selected`
- `developing`
- `development_applied`
- `review_handoff`
- `reviewing`
- `repairing`
- `ready_for_merge`
- `merging`
- `done`
- `blocked`

The current phase must always be reconstructable from durable state.

## Cycle Status

Allowed final statuses:

- `pending`
- `running`
- `done`
- `blocked`
- `failed`

`done` means the ticket reached `status: "done"` with merge metadata.

`blocked` means the active ticket could not complete review or merge safely.

`failed` means the loop infrastructure failed before it could produce a normal
blocked outcome.

## Review Artifact

Agentic review should persist a local review artifact.

Suggested path:

```text
.chr/user/reviews/<ticket-key>/<cycle-id>.md
```

The artifact should include:

- ticket key;
- cycle id;
- source branch;
- pull request reference;
- commit range or diff source;
- validation evidence inspected;
- criteria version;
- decision;
- blocking findings;
- non-blocking notes;
- created timestamp.

Review artifacts may be versioned according to project policy, but operational
state must remain under `.chr/`.

## Review Decision

Allowed review decisions:

- `pass`
- `needs_changes`
- `insufficient_evidence`
- `human_required`

Blocking findings must be structured enough for a repair worker to consume.

Suggested finding fields:

| Field | Meaning |
| --- | --- |
| `code` | Stable category such as `scope_drift`, `missing_validation`, or `contract_risk`. |
| `severity` | `blocking` or `non_blocking`. |
| `summary` | Short human-readable finding. |
| `evidence` | File, command, report, or PR evidence supporting the finding. |
| `repair_hint` | Concise instruction for the repair worker when blocking. |

## Repair Count

Repair count belongs to the ticket or cycle state, not to model memory.

The first implementation may store repair count in ticket review metadata if
that is simplest. A dedicated cycle record may duplicate it for reporting.

Default maximum:

```text
1
```

The limit must be configurable later, but the first implementation can use a
constant if configuration would add unnecessary surface area.

## Loop Report Additions

Loop reports should add per-cycle entries.

Suggested fields per entry:

- ticket key;
- mode;
- final status;
- execution id;
- review result;
- repair count;
- merge result;
- blocker;
- started and finished timestamps.

Existing processed, blocked, failed, and eligible counts remain valid, but their
meaning changes:

- processed: cycles that reached a final state;
- blocked: cycles that intentionally stopped the loop;
- failed: infrastructure failures.

## SQL Shape

If a table is introduced, use an additive migration.

Suggested table:

```sql
CREATE TABLE integration_cycles (
  id text PRIMARY KEY,
  loop_id text NOT NULL,
  project_id text NOT NULL,
  ticket_key text NOT NULL,
  mode text NOT NULL,
  phase text NOT NULL,
  status text NOT NULL,
  execution_id text NOT NULL DEFAULT '',
  review_artifact_path text NOT NULL DEFAULT '',
  review_result text NOT NULL DEFAULT '',
  repair_count integer NOT NULL DEFAULT 0,
  merge_result text NOT NULL DEFAULT '',
  blocker text NOT NULL DEFAULT '',
  started_at text NOT NULL,
  finished_at text NOT NULL DEFAULT '',
  updated_at text NOT NULL
);
```

Suggested indexes:

```sql
CREATE INDEX integration_cycles_loop_id_idx ON integration_cycles(loop_id);
CREATE INDEX integration_cycles_project_ticket_idx ON integration_cycles(project_id, ticket_key);
```

The first slice does not introduce a dedicated `integration_cycles` table.

Reason:

- existing ticket frontmatter already stores ticket, review, and merge handoff
  state;
- execution records already store ticket execution attempts and workspace
  metadata;
- loop records already store the parent run;
- Slice 1 only introduces side-effect-free planning contracts;
- persistence can remain additive in a later slice when the runner needs
  durable per-cycle records for reports or recovery.

A later slice may still add this table if report reconstruction from existing
state becomes ambiguous or too expensive.

## Invariants

- One cycle owns one ticket.
- One loop may contain many cycles.
- A cycle cannot select a second ticket.
- A successful cycle requires the ticket to be `done`.
- A blocked cycle prevents later cycle selection in the same run.
- Repair cycles reuse the same ticket, branch, pull request, and worktree.
- Review artifacts must not be the only source of ticket state.
- Ticket frontmatter remains the primary human-editable handoff state.

## Open Questions

- Is a dedicated `integration_cycles` table worth the migration cost in the
  first implementation?
- Should review artifacts be committed when project policy asks for versioned
  operational evidence?
- Should repair count live only in ticket metadata, only in cycle metadata, or
  both?
