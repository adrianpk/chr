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
- execution report path.
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

Execution records also store workspace execution metadata:

- `workspace_mode`: `project` or `worktree`.
- `workspace_project_checkout_dir`: the user-visible project checkout.
- `workspace_execution_dir`: the workspace mounted into Docker for this
  execution.

In `project` mode, the project checkout and execution workspace are the same
directory. In `worktree` mode, Chryso prepares a ticket-specific Git
worktree and uses that path as the execution workspace.

## Ticket Start Preflight

`chr ticket start <ticket-key>` prepares one ticket execution without invoking
an agent harness.

The command resolves the project through the standard project resolver, refreshes
the ticket index, selects the requested indexed ticket, loads workspace runtime
configuration, and validates the runtime boundary. Runtime validation must pass
before a durable execution record is created.

On success, the command creates a `pending` execution record with:

- the resolved project ID;
- the selected ticket key and absolute ticket path;
- runtime image, Dockerfile, container workdir, and validation command;
- execution-scoped run, result, log, and validation artifact paths;
- required lifecycle timestamps.

The execution ID is used as the runtime run ID, so artifact paths are scoped
under the configured runtime run directory.

The run directory can later contain:

- `events.jsonl`: append-only agent event stream.
- `execution-report.md`: execution summary and copied evidence strategy when
  enabled.
- `result-application.md`: result application report.
- runtime and validation logs from the configured harness.

`--dry-run` prints the same preflight output and runtime validation checks but
does not create an execution record. Ticket index reconciliation may still update
the user-level index because ticket selection always uses the indexed ticket
view.

If runtime validation fails, the command prints the failed checks and exits with
an error without creating an execution record.

During TUI work actions, Dash reads persisted event streams and displays recent
human-readable events plus paths for the ticket, execution report, runtime log,
and validation log when those references are available.

## Harness Invocation Contract

A harness invocation is a side-effect free rendering of the Docker commands
needed to build and run the configured runtime for one ticket execution.

The invocation requires:

- execution ID;
- ticket key;
- runtime plan.

The build command renders:

- `docker build`;
- the runtime Dockerfile path;
- the configured runtime image tag;
- the project directory as build context.

The run command renders:

- `docker run --rm`;
- a deterministic container name derived from the execution ID;
- runtime plan mounts;
- the configured container workdir;
- execution environment variables;
- the Codex worker shell command.

The mounted source is the runtime plan execution workspace. In worktree mode
this is the prepared ticket worktree, not the user's active checkout.

The validation command renders:

- `docker run --rm`;
- a deterministic validation container name derived from the execution ID;
- runtime plan mounts;
- the configured container workdir;
- validation artifact environment variables;
- the configured runtime validation command.

The run environment includes:

- `CHR_EXECUTION_ID`;
- `CHR_TICKET_KEY`;
- `CHR_TICKET_PATH`;
- `CHR_WORKER_CONTEXT_PATH`;
- `CHR_RESULT_PATH`;
- `CHR_LOG_PATH`.
- `CHR_VALIDATION_PATH`.
- `OPENAI_API_KEY`, inherited by name when present.

Runtime artifact paths are scoped by ticket key and execution ID when both are
available:

`<artifact_dir>/<ticket_key>/<execution_id>/`

The command renderer does not execute Docker, create directories, or write
artifacts. Later execution slices are responsible for invoking the rendered
commands and capturing outputs.

## Codex Worker Command Contract

A Codex worker command is a side-effect free rendering of the command that runs
inside the runtime container.

The worker command requires:

- execution ID;
- ticket key;
- ticket path;
- worker context artifact path;
- result artifact path;
- runtime log artifact path;
- validation artifact path.

The command invokes:

```sh
codex exec --skip-git-repo-check --dangerously-bypass-approvals-and-sandbox < "$CHR_WORKER_CONTEXT_PATH"
```

The runtime image must provide:

- a POSIX shell at `/bin/sh`;
- the `codex` executable on `PATH`;
- access to any credentials or configuration required by `codex exec`;
- the project workspace mounted at the configured container workdir.

The Docker runtime is the sandbox boundary. The Codex worker disables Codex's
inner sandbox because nested namespace creation is not available in all runtime
containers.

The rendered command exports these environment variables for the worker process:

- `CHR_EXECUTION_ID`;
- `CHR_TICKET_KEY`;
- `CHR_TICKET_PATH`;
- `CHR_WORKER_CONTEXT_PATH`;
- `CHR_RESULT_PATH`;
- `CHR_LOG_PATH`;
- `CHR_VALIDATION_PATH`.
- `OPENAI_API_KEY`, inherited by name when present.

The worker command renderer does not execute Codex, run Docker, create context
files, or write result artifacts. Later slices are responsible for context
materialization and harness integration.

Harness invocation now renders the Docker run command so the container executes
the Codex worker command. The worker context path and artifact paths are passed
as container paths through environment variables and through the shell command.

## Codex Worker Context

The Codex worker context is a deterministic Markdown artifact written before
the worker runs.

The context includes:

- execution ID;
- ticket key;
- ticket path;
- result artifact path;
- runtime log artifact path;
- validation artifact path;
- the JSON result contract the worker must satisfy.

The runtime plan reserves `context.md` in the execution run artifact directory:

`<artifact_dir>/<ticket_key>/<execution_id>/context.md`

Context materialization creates parent directories and writes the context file.
It does not execute Codex, run validation, or write the result envelope.

## Agent Harness Invocation

In non-dry-run mode, `chr ticket start <ticket-key>` creates the pending
execution record and then invokes the rendered harness commands.

The execution runner:

- creates the execution run artifact directory;
- marks the execution `running` before invoking the harness;
- runs the rendered Docker build command;
- runs the rendered Docker run command;
- runs the rendered Docker validation command;
- writes build and worker stdout and stderr to the runtime log artifact;
- writes validation stdout and stderr to the validation artifact;
- marks the execution `succeeded` when all commands complete and the result
  envelope is valid;
- marks the execution `failed` and records the failure reason when any command
  fails.

The runner updates `started_at`, `finished_at`, and `updated_at` timestamps as
the execution moves through the lifecycle.

`--dry-run` still performs preflight and prints the planned runtime output,
including execution-scoped artifact paths. It does not create an execution
record, create artifact directories, build the image, invoke Codex, capture a
result envelope, or run validation.

## Execution Result Artifact

The harness run command must write a JSON result envelope to `CHR_RESULT_PATH`.

Envelope fields:

- `execution_id`: execution attempt ID.
- `ticket_key`: selected ticket key.
- `status`: `succeeded` or `failed`.
- `summary`: human-readable result summary.
- `changed_files`: optional list of changed file objects.

Changed file objects contain:

- `path`: workspace-relative changed file path.
- `summary`: optional concise change summary.

Changed file paths must be relative to the core. Absolute paths and paths
that escape the workspace are rejected.

The worker must write this envelope before it exits successfully. If the worker
process fails before writing a valid envelope, the execution failure kind is
`worker`, and the planned result path is printed as diagnostic context.

## Changed File Discovery

The changed-file discoverer inspects the execution workspace with:

```sh
git status --porcelain=v1 --untracked-files=all
```

It returns workspace-relative changed paths as execution result changed-file
entries. The discoverer does not mutate the worktree, stage files, commit,
revert, or clean files.

Operational runtime artifacts under `.chr/user/runtime/` are ignored so worker
context, logs, result envelopes, and validation artifacts are not reported as
workspace edits.

The runner reads the result artifact after the harness commands complete. A
missing result file, malformed JSON, invalid status, mismatched execution ID, or
mismatched ticket key marks the execution `failed` and records the diagnostic in
`failure_reason`.

When the result status is `failed`, the execution is marked `failed` even if the
Docker commands exited successfully. When the result status is `succeeded`, the
execution is marked `succeeded`.

Process stdout and stderr continue to be written to `CHR_LOG_PATH`.

## Failure Handling

Execution failures use stable failure kind prefixes in `failure_reason`.

Failure kinds:

- `preflight`: project resolution, ticket reconciliation, ticket lookup, or
  workspace config failed before runtime validation.
- `runtime`: runtime boundary validation failed before creating an execution
  record.
- `harness`: Docker build invocation failed.
- `worker`: Codex worker invocation failed before post-worker validation.
- `validation`: the validation invocation failed, or the harness completed but
  the result envelope reported `failed`.
- `result_capture`: result artifact reading, JSON decoding, envelope validation,
  or identity matching failed.

`ticket start` prints failure diagnostics when an execution record exists or a
runtime preflight record can be rendered:

- `failure_kind`
- `failure_reason`
- `diagnostic_log_path`
- `diagnostic_result_path`
- `diagnostic_validation_path`

Diagnostic paths are planned artifact locations. A path may be printed even when
the artifact was not created because the failure happened before that artifact
could be written.

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

## Reporting

Execution reporting is documented in:

- `docs/reference/execution-reports.md` for the execution report format;
- `docs/reference/reports.md` for `chr report list` and `chr report show`;
- `docs/how-to/find-execution-and-loop-evidence.md` for practical evidence
  lookup.

Execution records are linked to `projects(id)` and are deleted when their
project is deleted.
