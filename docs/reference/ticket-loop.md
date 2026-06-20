# Ticket Loop

## Purpose

The ticket loop is a foreground runner that processes eligible tickets from one
managed project.

The first loop boundary is ticket selection. Execution, result application,
review policy, merge policy, status reporting, and daemon coordination are
separate concerns.

## Integration Cycle Policy

Serialized integration loop work treats one ticket integration cycle as the unit
of progress:

```text
select one ticket
develop
apply review policy
merge or block
rehydrate state
select the next ticket
```

The cycle policy is derived from `review.mode`:

- `human_review`: develop one ticket, create the review handoff, and stop before
  later normal work.
- `agentic_review`: develop one ticket, run high-level review, merge approved
  work, and then continue.
- `trusted_auto_merge`: develop one ticket and skip the separate review role,
  but still merge before continuing.

The cycle policy model is side-effect free. It does not by itself execute,
review, merge, or mutate tickets.

## Eligible Ticket Query

The eligible ticket query reads from the existing ticket index. It does not
scan ticket files, refresh the index, update ticket state, create execution
records, or mutate the core.

Required input:

- project ID.

Optional input:

- eligible ticket statuses.
- metadata filters.
- ordering.

When no statuses are provided, the default eligible status is:

```text
ready_for_dev
```

Callers may provide explicit statuses to widen or narrow selection. Status
values must be valid ticket lifecycle statuses. The legacy `ready` status
remains supported when passed explicitly, but default loop selection uses
`ready_for_dev`.

Metadata filters can narrow the eligible tickets by:

- priority;
- severity;
- project area;
- tags;
- type;
- use case.

Priority, severity, and type filters use the same validation rules as ticket
metadata. Project area, tag, and use-case filters are free-form exact matches.

Empty filter dimensions are ignored. Multiple values inside one filter
dimension are treated as alternatives, except tags: every requested tag must be
present on the ticket.

When no ordering is provided, returned tickets are sorted by `created_at`
oldest-first, with ticket key as the final tie-breaker.

Supported order values:

- `key`;
- `created_at`;
- `priority`;
- `severity`;
- `priority_severity`.

Priority and severity order use this rank from highest to lowest:

```text
critical
high
normal
low
```

All ordering modes are deterministic. Priority and severity ordering use
`created_at` as the first tie-breaker, and all modes eventually tie-break by
ticket key.

## Scope

The eligible ticket query is a core package boundary. CLI parsing, ordering
policies, loop run persistence, repeated execution, stop conditions, and
reporting are implemented by later slices.

## Loop Preview Command

Preview the foreground loop selection:

```sh
chc ticket loop --dry-run
```

The command resolves the project, refreshes the ticket index, queries eligible
tickets, and prints the planned selection.

Supported preview flags:

- `--status`;
- `--priority`;
- `--severity`;
- `--project-area`;
- `--tag`;
- `--type`;
- `--use-case`;
- `--order`.
- `--max-iterations`.

The preview command does not create loop records, create execution records,
invoke the runtime worker, apply results, or mutate ticket files.

Without `--dry-run`, the command creates a loop run, executes eligible tickets
one at a time through the existing single-ticket execution path, applies each
execution result, and prints the loop summary.

Loop execution output includes:

- loop run ID;
- loop run status;
- eligible ticket count;
- processed ticket count;
- skipped ticket count;
- blocked ticket count;
- failed ticket count;
- execution IDs for attempted tickets.

Loop report formatting is documented in `docs/reference/loop-reports.md`.

The foreground loop reuses:

- runtime validation;
- execution record creation;
- harness invocation;
- execution result capture;
- ticket result application.

Loop stop conditions:

- no eligible tickets completes the run with zero processed tickets;
- `--max-iterations` caps how many eligible tickets are attempted;
- the first blocked ticket result stops the run and marks the loop blocked;
- the first execution failure stops the run and marks the loop failed;
- duplicate ticket keys in one selected run are attempted only once.

When `--max-iterations` is zero or omitted, the loop uses every selected
eligible ticket until another stop condition is reached.

Skipped tickets are eligible tickets that were not attempted because a stop
condition ended the run first.

## Loop Run Records

A loop run record represents one foreground loop attempt for one project.

Loop run records are stored in the user-level SQLite database and are separate
from individual ticket execution records.

Required identity and lifecycle fields:

- `id`;
- `project_id`;
- `status`;
- `created_at`;
- `updated_at`.

Supported statuses:

- `pending`;
- `running`;
- `succeeded`;
- `failed`;
- `blocked`.

Counters tracked on the loop run:

- eligible tickets;
- processed tickets;
- blocked tickets;
- failed tickets.

Optional lifecycle fields:

- `started_at`;
- `finished_at`;
- failure reason.

Loop run records do not replace execution records. Later execution slices link
loop iterations to existing per-ticket execution records.

## Reporting

Loop reporting is documented in:

- `docs/reference/loop-reports.md` for the loop report format;
- `docs/reference/reports.md` for `chc report list` and `chc report show`;
- `docs/how-to/find-execution-and-loop-evidence.md` for practical evidence
  lookup.
