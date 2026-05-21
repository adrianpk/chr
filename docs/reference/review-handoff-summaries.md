# Review Handoff Summaries

Review handoff summaries are compact Markdown summaries for a completed or
blocked ticket execution.

They are derived from existing execution report data. They do not replace
execution reports, ticket comments, validation evidence, changed-file summaries,
runtime logs, or result artifacts.

## Purpose

A review handoff summary gives the user the minimum information needed to decide
what to inspect next after an execution attempt:

- execution and ticket identity;
- execution and result status;
- result summary;
- validation status, command, and artifact path;
- changed-file summary;
- blocker or failure reason when present;
- result, log, and report artifact paths;
- next action.

## Format

The canonical handoff summary format is Markdown.

The current renderer is side-effect free. It builds Markdown from an
`ExecutionReport` value and does not read runtime logs, validation logs, result
JSON, ticket files, or Git state.

## Sections

### Overview

Includes:

- execution id;
- project id;
- ticket key;
- ticket path;
- execution status;
- result status;
- result summary.

### Validation

Includes:

- validation status;
- validation command;
- validation artifact path.

Validation status comes from the execution report.

### Changed Files

Includes:

- total changed files;
- rendered changed files;
- truncation marker;
- changed-file paths and optional summaries.

Changed-file data comes from the execution report.

### Blocker

Includes blocker or failure details when the execution recorded a failure
reason.

When no failure is recorded, the summary explicitly says that no blocker or
failure was recorded.

### Artifacts

Includes:

- result artifact path;
- runtime log path;
- execution report path.

### Next Action

The next action is derived from execution and validation status:

- succeeded execution with succeeded validation: review changed files and
  validation evidence;
- succeeded execution with other validation status: review the result and
  confirm validation evidence;
- failed or blocked execution: inspect blocker details, runtime log, validation
  output, and result artifact;
- pending or running execution: wait for execution to finish before review.

## Boundary

Review handoff summaries are currently a rendering contract. They are intended
for user-facing review surfaces and later commands that need a concise handoff
view.

Persistence and notification delivery are outside this reference.
