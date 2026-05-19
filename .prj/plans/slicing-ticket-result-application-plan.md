# Ticket Result Application Slicing Plan

Status: active

## Purpose

Plan the `Ticket Result Application` slicing from the feature slicing set map.

This slicing covers the use case:

> After an execution attempt, `chc` updates durable project state with status,
> evidence, and comments.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related documents:

- `.prj/drafts/chc-command-surfaces-and-priority.md`
- `.prj/plans/slicing-single-ticket-execution-plan.md`
- `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-validation.md`

Prerequisite delivered work:

- `Ticket Foundation` slicing.
- `Automatic Reconciliation` slicing.
- `Single Ticket Execution` slicing.

## Current State

Already delivered:

- Markdown ticket parsing and rendering;
- ticket status updates and comments;
- ticket index reconciliation before reads;
- durable execution records;
- harness result capture and classified execution failures.

Still missing:

- applying execution result state to ticket Markdown;
- writing execution comments to tickets;
- recording validation evidence in ticket-visible state;
- summarizing changed files from execution output;
- recording blocked/failure reasons in tickets;
- preserving Markdown and SQLite consistency after result application;
- producing a result application report artifact;
- marking review-ready state when an execution succeeds.

## Slicing Scope

This slicing includes:

- result application model and command boundary;
- ticket state updates from execution outcomes;
- ticket comments generated from execution records;
- validation evidence writing;
- changed-files summary writing;
- blocked/failure reason writing;
- reconciliation after ticket mutation;
- result application report artifact;
- review-ready state handling.

## Out of Scope

This slicing does not include:

- filtered ticket loops;
- daemon coordination;
- Telegram or email notifications;
- retry policy;
- PR creation or forge integration;
- applying arbitrary agent output that is not represented in the structured
  execution result contract.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Result Application Model

Status: planned.

Expected branch: `feat/slice-1-result-application-model`
Expected PR title: `feat(slice-1): result application model`
Expected report: `.prj/reports/slices/ticket-result-application/slice-1-result-application-model.md`

Goal:

Define a side-effect free model for applying one execution result to one ticket.

Candidate scope:

- add result application domain types;
- map execution statuses and result outcomes to ticket-facing outcomes;
- define comment, evidence, changed-file, and blocked-reason fields;
- validate required references;
- add focused model tests;
- document result application contract.

Expected outcome:

- later slices can write ticket files using one stable application contract.

### Slice 2: Ticket State Writer

Status: planned.

Expected branch: `feat/slice-2-ticket-state-writer`
Expected PR title: `feat(slice-2): ticket state writer`
Expected report: `.prj/reports/slices/ticket-result-application/slice-2-ticket-state-writer.md`

Goal:

Update ticket Markdown status from one result application.

Candidate scope:

- add a ticket result state writer;
- map success to review-ready ticket state;
- map execution failure to blocked or failed ticket state;
- preserve existing ticket metadata and comments;
- add focused writer tests.

Expected outcome:

- execution outcomes can update the canonical ticket file status.

### Slice 3: Execution Comment Writer

Status: planned.

Expected branch: `feat/slice-3-execution-comment-writer`
Expected PR title: `feat(slice-3): execution comment writer`
Expected report: `.prj/reports/slices/ticket-result-application/slice-3-execution-comment-writer.md`

Goal:

Append execution summary comments to ticket Markdown.

Candidate scope:

- write execution ID, status, summary, and failure reason comments;
- include artifact paths in the comment;
- preserve existing comments;
- add focused comment writer tests.

Expected outcome:

- ticket readers can inspect execution outcome without querying SQLite.

### Slice 4: Validation Evidence Writer

Status: planned.

Expected branch: `feat/slice-4-validation-evidence-writer`
Expected PR title: `feat(slice-4): validation evidence writer`
Expected report: `.prj/reports/slices/ticket-result-application/slice-4-validation-evidence-writer.md`

Goal:

Record validation evidence from execution artifacts in ticket-visible state.

Candidate scope:

- define validation evidence fields;
- write validation command and artifact path;
- preserve validation history;
- add focused evidence writer tests.

Expected outcome:

- tickets show which validation evidence supports an execution result.

### Slice 5: Changed Files Summary

Status: planned.

Expected branch: `feat/slice-5-changed-files-summary`
Expected PR title: `feat(slice-5): changed files summary`
Expected report: `.prj/reports/slices/ticket-result-application/slice-5-changed-files-summary.md`

Goal:

Record a concise changed-files summary from structured execution output.

Candidate scope:

- extend result application data with changed files;
- write changed-file summary comments or section content;
- handle empty changed-file lists;
- add focused summary tests.

Expected outcome:

- reviewers can see what files the execution claims to have changed.

### Slice 6: Blocked Reason Writer

Status: planned.

Expected branch: `feat/slice-6-blocked-reason-writer`
Expected PR title: `feat(slice-6): blocked reason writer`
Expected report: `.prj/reports/slices/ticket-result-application/slice-6-blocked-reason-writer.md`

Goal:

Record clear blocked or failure reasons in ticket-visible state.

Candidate scope:

- write blocked reason from execution failure classification;
- include diagnostic artifact paths;
- keep failure details concise;
- add focused blocked reason tests.

Expected outcome:

- blocked tickets explain why execution stopped and where diagnostics live.

### Slice 7: Markdown Index Consistency

Status: planned.

Expected branch: `feat/slice-7-markdown-index-consistency`
Expected PR title: `feat(slice-7): markdown index consistency`
Expected report: `.prj/reports/slices/ticket-result-application/slice-7-markdown-index-consistency.md`

Goal:

Reconcile the ticket index after result application mutates ticket Markdown.

Candidate scope:

- refresh the affected ticket after writes;
- ensure SQLite index reflects updated ticket status and metadata;
- surface reconciliation blockers;
- add focused consistency tests.

Expected outcome:

- canonical Markdown and derived SQLite state remain aligned after applying a
  result.

### Slice 8: Result Apply CLI

Status: planned.

Expected branch: `feat/slice-8-result-apply-cli`
Expected PR title: `feat(slice-8): result apply cli`
Expected report: `.prj/reports/slices/ticket-result-application/slice-8-result-apply-cli.md`

Goal:

Expose result application through a CLI command.

Candidate scope:

- add `chc ticket apply-result <execution-id>`;
- resolve project context;
- load one execution record;
- apply ticket status, comments, evidence, and diagnostics;
- write an application report artifact;
- support dry-run output;
- add focused CLI tests;
- document command behavior.

Expected outcome:

- a user can apply one completed execution attempt to the durable ticket file
  and refreshed index.

## Validation

Each implementation slice must run targeted tests and `make verify`.

The completed slicing must run `make verify` after the final slice is merged.

## Risks

- Ticket template sections may need incremental extension. Keep additions
  backward-compatible with existing ticket parsing.
- Result application should not mutate tickets when execution records are
  incomplete or ambiguous.
- The command must preserve canonical Markdown as the source of truth and keep
  SQLite as derived state.
