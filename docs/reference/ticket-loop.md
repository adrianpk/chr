# Ticket Loop

## Purpose

The ticket loop is a foreground runner that processes eligible tickets from one
managed project.

The first loop boundary is ticket selection. Execution, result application,
status reporting, and daemon coordination are separate concerns.

## Eligible Ticket Query

The eligible ticket query reads from the existing ticket index. It does not
scan ticket files, refresh the index, update ticket state, create execution
records, or mutate the workspace.

Required input:

- project ID.

Optional input:

- eligible ticket statuses.
- metadata filters.
- ordering.

When no statuses are provided, the default eligible status is:

```text
ready
```

Callers may provide explicit statuses to widen or narrow selection. Status
values must be valid ticket lifecycle statuses.

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

When no ordering is provided, returned tickets are sorted by ticket key.

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
