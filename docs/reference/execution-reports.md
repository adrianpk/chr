# Execution Reports

Execution reports are human-readable summaries of one ticket execution attempt.

They summarize existing execution evidence. They do not replace execution
records, result envelopes, ticket comments, validation evidence, runtime logs,
or changed-file records.

## Format

The canonical execution report format is Markdown.

The current renderer is side-effect free. It builds Markdown from an execution
record and, when available, the structured execution result envelope.

## Sections

### Summary

Includes:

- execution id;
- project id;
- ticket key;
- ticket path;
- execution status;
- failure kind.

### Times

Includes:

- created time;
- started time;
- finished time;
- updated time.

Times are rendered in UTC using RFC 3339.

Missing optional times render as `none`.

### Runtime

Includes:

- runtime image;
- runtime Dockerfile path;
- container workdir;
- validation command.

These values come from the execution record. They describe the runtime boundary
used for that attempt, even if project runtime configuration changes later.

### Artifacts

Includes:

- execution run directory;
- structured result path;
- runtime log path;
- validation output path.

### Result

When a structured execution result is available, the report includes:

- result status;
- result summary.

When no result is available, the report explicitly says that no execution result
was available.

### Validation

Validation summary includes:

- validation status;
- validation command;
- validation artifact path.

Validation status is derived from the execution record:

- `succeeded` when execution status is `succeeded` and a validation artifact path
  exists;
- `failed` when execution status is `failed` or `blocked` and a validation
  artifact path exists;
- `unknown` when execution is still pending or running and a validation artifact
  path exists;
- `not_run` when no validation artifact path is available.

### Changed Files

Changed files come from the structured execution result envelope.

Each changed file entry includes:

- workspace-relative path;
- optional summary.

Changed files are sorted by path before rendering. Reports render at most 25
changed-file entries and include summary fields showing:

- total changed files;
- rendered changed files;
- whether the list was truncated.

When no changed files are available, the report explicitly says that no changed
files were reported.

### Failure

When the execution record includes a failure reason, the report includes it.

When no failure is recorded, the report explicitly says that no failure was
recorded.

## Current Boundary

Execution reports are written to the execution run artifact directory:

```text
ops/<namespace>/run/<artifact-id>/evidence/execution-report.md
```

`<namespace>` is `default` unless `OPS_USER` is configured.

For ticket result application, `chr ticket apply-result <execution-id>` writes
this report after applying the result to ticket state. Dry-run result
application prints the planned report path but does not write the report.

The writer derives the report path from the execution record run directory. It
creates the target directory when needed. If the target path already exists and
does not look like an execution report, the writer refuses to overwrite it.

Report listing and lookup are documented in `docs/reference/reports.md`.
