# Loop Reports

Loop reports are human-readable summaries of one foreground ticket loop run.

They summarize existing loop run records and optional related execution records.
They do not replace loop run persistence, execution records, ticket state,
execution reports, runtime logs, or validation artifacts.

## Format

The canonical loop report format is Markdown.

The current renderer is side-effect free. It builds Markdown from a loop run
record and, when available, related execution records.

## Sections

### Summary

Includes:

- loop run id;
- project id;
- loop status.

### Counts

Includes:

- eligible ticket count;
- processed ticket count;
- blocked ticket count;
- failed ticket count.

### Times

Includes:

- created time;
- started time;
- finished time;
- updated time.

Times are rendered in UTC using RFC 3339.

Missing optional times render as `none`.

### Executions

Related executions are optional.

When provided, each execution entry includes:

- execution id;
- ticket key;
- execution status.

Loop reports render at most 25 execution entries and include summary fields
showing:

- total executions;
- rendered executions;
- whether the list was truncated.

When no related executions are available, the report explicitly says that no
executions were recorded.

### Failure

When the loop run includes a failure reason, the report includes it.

When no failure is recorded, the report explicitly says that no failure was
recorded.

## Current Boundary

This reference describes the loop report model and Markdown rendering contract.

Loop report persistence, listing, and lookup belong to later `Reporting and
Evidence` slices.

Report listing is documented in `docs/reference/reports.md`.
