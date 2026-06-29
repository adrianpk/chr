# Reporting and Evidence Slicing Plan

Status: completed

## Purpose

Plan the `Reporting and Evidence` slicing from the feature slicing set map.

This slicing covers the use case:

> A user can understand what `chr` did, what changed, and why a ticket or loop
> is in its current state.

## Source

Primary source:

- `.prj/drafts/chr-feature-slicing-set-map.md`

Related delivered work:

- `.prj/plans/slicing-single-ticket-execution-plan.md`
- `.prj/plans/slicing-ticket-result-application-plan.md`
- `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- `.prj/plans/slicing-codex-runtime-worker-plan.md`
- `.prj/plans/slicing-notification-and-remote-commands-plan.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-result-application.md`
- `docs/reference/ticket-loop.md`

## Current State

Delivered prerequisites:

- execution records persist status, runtime references, artifact paths, and
  failure reasons;
- runtime workers produce structured result envelopes and changed-file
  summaries;
- result application writes ticket comments, validation evidence, changed-file
  notes, blocked reasons, and status updates;
- loop runs persist summary counts and stop outcomes;
- notification audit records persist dispatch and remote-command events.

Missing dedicated UX:

- a stable execution report format;
- a stable loop report format;
- validation, changed-file, blocker, and review handoff summaries;
- report listing;
- report lookup by ticket, execution, or loop.

## Slicing Scope

This slicing includes:

- core report summary models;
- Markdown report rendering for execution and loop outcomes;
- evidence normalization from existing execution, result, ticket, loop, and
  notification records;
- report artifact persistence where reports are produced;
- CLI commands for listing and showing reports;
- reference documentation for report formats and lookup behavior.

## Out of Scope

This slicing does not include:

- a web dashboard;
- hosted report storage;
- forge pull-request automation;
- email or Telegram delivery changes;
- changing the ticket execution model;
- changing the runtime worker contract unless a report needs an already
  available field surfaced.

## Design Direction

Reports should summarize evidence that already exists. They should not become a
second source of truth.

The report model should keep stable identifiers visible:

- project id;
- ticket key;
- execution id;
- loop id where applicable;
- status;
- artifact paths;
- validation path or summary;
- changed files;
- blocker or failure reason;
- generated report path.

Markdown should remain the canonical human-readable report format. JSON can be
added later if a command or integration needs it.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Execution Report Format

Status: delivered.

Expected branch: `feat/slice-1-execution-report-format`
Expected PR title: `feat(slice-1): execution report format`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-1-execution-report-format.md`

Goal:

Define the core execution report model and Markdown rendering contract.

Candidate scope:

- add an execution report summary model;
- render a deterministic Markdown report from execution records and optional
  result data;
- include ticket, execution, status, runtime, artifact, changed-file,
  validation, and failure fields;
- keep report rendering side-effect free;
- add focused model and renderer tests;
- document the execution report format.

Expected outcome:

- later slices can persist, list, and show execution reports without inventing
  report shape in the CLI layer.

### Slice 2: Execution Report Artifact Writer

Status: delivered.

Expected branch: `feat/slice-2-execution-report-writer`
Expected PR title: `feat(slice-2): execution report writer`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-2-execution-report-writer.md`

Goal:

Persist execution reports as project-local artifacts.

Candidate scope:

- define report artifact paths under the existing execution artifact boundary;
- write Markdown reports after execution result application or explicit command
  invocation;
- avoid overwriting unrelated files;
- add writer tests with temporary workspaces;
- document report artifact placement.

Expected outcome:

- execution attempts can leave durable human-readable reports near existing
  execution evidence.

### Slice 3: Validation and Changed-File Summaries

Status: delivered.

Expected branch: `feat/slice-3-evidence-summaries`
Expected PR title: `feat(slice-3): evidence summaries`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-3-evidence-summaries.md`

Goal:

Normalize validation and changed-file evidence for reports.

Candidate scope:

- summarize validation output paths and status;
- summarize changed files from execution results and ticket comments;
- keep summaries bounded and deterministic;
- add tests for empty, successful, failed, and changed-file-heavy cases.

Expected outcome:

- report rendering can show useful evidence without duplicating full logs.

### Slice 4: Loop Report Format

Status: delivered.

Expected branch: `feat/slice-4-loop-report-format`
Expected PR title: `feat(slice-4): loop report format`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-4-loop-report-format.md`

Goal:

Define the loop report model and Markdown rendering contract.

Candidate scope:

- add a loop report summary model;
- include loop id, project id, status, counts, failure reason, timings, and
  related executions when available;
- render deterministic Markdown;
- add focused tests;
- document the loop report format.

Expected outcome:

- users can inspect a loop as one coherent operational unit.

### Slice 5: Report Listing

Status: delivered.

Expected branch: `feat/slice-5-report-listing`
Expected PR title: `feat(slice-5): report listing`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-5-report-listing.md`

Goal:

Expose report discovery through the CLI.

Candidate scope:

- add `chr report` command group;
- add `chr report list`;
- support filtering by kind, ticket key, execution id, loop id, and status;
- print stable tabular or line-oriented output;
- add CLI tests.

Expected outcome:

- users can discover existing reports without knowing artifact paths.

### Slice 6: Report Lookup

Status: delivered.

Expected branch: `feat/slice-6-report-lookup`
Expected PR title: `feat(slice-6): report lookup`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-6-report-lookup.md`

Goal:

Expose direct report lookup through the CLI.

Candidate scope:

- add `chr report show`;
- resolve reports by report id, ticket key, execution id, or loop id;
- print the report body or metadata predictably;
- add clear errors for ambiguous or missing lookups;
- add CLI tests.

Expected outcome:

- users can inspect one report from the command line without searching files.

### Slice 7: Review Handoff Summary

Status: delivered.

Expected branch: `feat/slice-7-review-handoff-summary`
Expected PR title: `feat(slice-7): review handoff summary`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-7-review-handoff-summary.md`

Goal:

Produce concise handoff material for human review.

Candidate scope:

- summarize status, changed files, validation, blocker or failure reason, and
  next recommended action;
- keep summaries derived from existing records;
- add tests for succeeded, failed, and blocked executions;
- document handoff summary semantics.

Expected outcome:

- report output can be used directly by a human reviewer without reading every
  artifact.

### Slice 8: Reporting Documentation

Status: delivered.

Expected branch: `feat/slice-8-reporting-docs`
Expected PR title: `feat(slice-8): reporting docs`
Expected report: `.prj/reports/slices/reporting-and-evidence/slice-8-reporting-docs.md`

Goal:

Document the reporting UX and evidence model.

Candidate scope:

- add reference docs for report formats and commands;
- add how-to documentation for finding execution and loop evidence;
- update related ticket execution and loop docs where needed;
- add the slice implementation report.

Expected outcome:

- reporting behavior is discoverable without reading implementation code.

## Completion Criteria

The slicing is complete when:

- execution reports have a stable Markdown format;
- execution reports can be written as durable artifacts;
- validation and changed-file evidence is summarized consistently;
- loop reports have a stable Markdown format;
- reports can be listed and shown through the CLI;
- review handoff summaries are available;
- documentation describes report formats, commands, and artifact locations;
- `make verify` passes.

## Validation

Required validation for every implementation slice:

```sh
make verify
```

Focused tests should be added for each new model, renderer, writer, repository,
and CLI surface.
