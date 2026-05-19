# Ticket Execution

## Execution Record

A ticket execution record represents one attempt to work on one ticket.

The record is durable execution metadata. It is separate from ticket Markdown
result application. Later result-application commands may use execution records
to update tickets, comments, status, or validation evidence.

## Required Identity Fields

- `id`: execution attempt identifier.
- `project_id`: registered project identifier.
- `ticket_key`: ticket key selected for execution.
- `status`: execution lifecycle status.
- `created_at`: time the execution record was created.
- `updated_at`: time the execution record was last changed.

## Optional References

- `ticket_path`: absolute path to the selected ticket file.
- runtime image, Dockerfile, container workdir, and validation command.
- run artifact directory.
- result artifact path.
- runtime log path.
- validation artifact path.
- failure reason.

## Status Values

Supported execution statuses:

- `pending`: execution record exists but harness work has not started.
- `running`: harness work has started.
- `succeeded`: harness work completed successfully.
- `failed`: execution attempted and failed.
- `blocked`: preflight or execution gating prevented the attempt.

## Time Semantics

`created_at` and `updated_at` are required.

`started_at` and `finished_at` are optional because preflight records may exist
before harness work starts. When both are present, `finished_at` must not be
earlier than `started_at`.

## Runtime References

Execution records copy runtime references from the computed runtime plan at the
time of execution. This keeps later review tied to the runtime boundary that was
used for that attempt, even if project config changes later.

## Persistence

Execution records are stored in the user-level SQLite database alongside the
project registry and ticket index.

The `execution_records` table is keyed by execution `id` and stores:

- project and ticket references;
- execution status;
- runtime image, Dockerfile, workdir, and validation command;
- run, result, log, and validation artifact paths;
- failure reason;
- lifecycle timestamps.

The repository supports:

- creating an execution record;
- updating an execution record;
- getting one execution record by ID;
- listing execution records for one project ticket.

Execution records are linked to `projects(id)` and are deleted when their
project is deleted.
