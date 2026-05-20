# Filtered Ticket Loop Slicing Plan

Status: completed

## Purpose

Plan the `Filtered Ticket Loop` slicing from the feature slicing set map.

This slicing covers the use case:

> A user starts a loop and `chc` keeps working through eligible tickets until
> completion, threshold, or blocker.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related documents:

- `.prj/plans/slicing-automatic-reconciliation-plan.md`
- `.prj/plans/slicing-single-ticket-execution-plan.md`
- `.prj/plans/slicing-ticket-result-application-plan.md`
- `.prj/plans/slicing-codex-runtime-worker-plan.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-result-application.md`

Prerequisite delivered work:

- `Automatic Reconciliation` slicing.
- `Single Ticket Execution` slicing.
- `Ticket Result Application` slicing.
- `Codex Runtime Worker` slicing.

## Current State

Delivered prerequisites:

- tickets are indexed and refreshed before ticket reads;
- one ticket can be selected and executed;
- Codex can run through the runtime container worker path;
- execution result artifacts can be captured;
- execution results can be applied back into ticket state.

Delivered by this slicing:

- eligible ticket selection for loop execution;
- user-facing filters for the loop;
- deterministic ticket ordering;
- durable loop run records;
- repeated single-ticket execution;
- stop conditions and thresholds;
- blocked-ticket handling across iterations;
- loop status and reporting output.

## Slicing Scope

This slicing includes:

- loop eligibility and filter models;
- eligible ticket query behavior;
- ordering rules for loop execution;
- foreground loop command skeleton;
- loop run persistence;
- iteration execution using the existing single-ticket execution path;
- stop conditions and iteration thresholds;
- blocked-ticket handling;
- loop status reporting.

## Out of Scope

This slicing does not include:

- daemon coordination;
- notification or remote command surfaces;
- multi-project scheduling;
- background process ownership;
- forge or PR creation automation;
- hosted execution services.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Eligible Ticket Query

Status: delivered.

Expected branch: `feat/slice-1-eligible-ticket-query`
Expected PR title: `feat(slice-1): eligible ticket query`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-1-eligible-ticket-query.md`

Goal:

Define the core loop ticket selection model and query eligible tickets from the
existing ticket index.

Candidate scope:

- add loop eligibility option types;
- select tickets by project and eligible statuses;
- keep selection read-only;
- return deterministic results;
- add focused query tests;
- document the eligibility contract.

Expected outcome:

- later slices can build filters, ordering, and loop execution on one query
  boundary.

### Slice 2: Loop Filters

Status: delivered.

Expected branch: `feat/slice-2-loop-filters`
Expected PR title: `feat(slice-2): loop filters`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-2-loop-filters.md`

Goal:

Allow loop ticket selection to be narrowed by ticket metadata.

Candidate scope:

- add priority, severity, area, tag, type, use-case, and status filters;
- validate filter values where existing ticket validation rules apply;
- keep filtering independent from CLI parsing;
- add focused filter tests.

Expected outcome:

- loop selection can target a useful subset of indexed tickets.

### Slice 3: Loop Ordering

Status: delivered.

Expected branch: `feat/slice-3-loop-ordering`
Expected PR title: `feat(slice-3): loop ordering`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-3-loop-ordering.md`

Goal:

Make ticket loop ordering deterministic and reviewable.

Candidate scope:

- define default ordering;
- support priority and severity ordering;
- tie-break by ticket key or created time;
- add focused ordering tests.

Expected outcome:

- repeated loop runs pick tickets in predictable order.

### Slice 4: Loop Command Skeleton

Status: delivered.

Expected branch: `feat/slice-4-loop-command-skeleton`
Expected PR title: `feat(slice-4): loop command skeleton`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-4-loop-command-skeleton.md`

Goal:

Add a foreground `ticket loop` command that resolves project context, refreshes
the ticket index, and prints the planned loop selection.

Candidate scope:

- add CLI command and flags;
- resolve project and refresh index like ticket reads;
- call the eligible ticket query;
- support dry-run output;
- add focused CLI tests.

Expected outcome:

- users can preview which tickets a loop would process.

### Slice 5: Loop Run Records

Status: delivered.

Expected branch: `feat/slice-5-loop-run-records`
Expected PR title: `feat(slice-5): loop run records`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-5-loop-run-records.md`

Goal:

Persist loop run metadata separately from individual execution records.

Candidate scope:

- add loop run model;
- add SQLite migration and repository;
- track lifecycle status and timestamps;
- link loop runs to project IDs;
- add repository tests.

Expected outcome:

- loop runs have durable state for status and later reporting.

### Slice 6: Repeated Ticket Execution

Status: delivered.

Expected branch: `feat/slice-6-repeated-ticket-execution`
Expected PR title: `feat(slice-6): repeated ticket execution`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-6-repeated-ticket-execution.md`

Goal:

Execute selected tickets one at a time using the existing single-ticket
execution path.

Candidate scope:

- run one execution per selected ticket;
- refresh ticket state between iterations;
- apply successful or blocked results through existing result application;
- record iteration summaries;
- add focused runner tests.

Expected outcome:

- a foreground loop can process more than one ticket without duplicating the
  single-ticket execution implementation.

### Slice 7: Stop Conditions and Thresholds

Status: delivered.

Expected branch: `feat/slice-7-stop-conditions`
Expected PR title: `feat(slice-7): stop conditions`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-7-stop-conditions.md`

Goal:

Stop the loop predictably at completion, blocker, failure, or configured limits.

Candidate scope:

- add max-iteration threshold;
- add no-eligible-ticket completion;
- stop on blocker or execution failure;
- avoid repeated attempts on the same ticket in one run;
- add focused stop-condition tests.

Expected outcome:

- foreground loop runs terminate clearly and safely.

### Slice 8: Loop Status Reporting

Status: delivered.

Expected branch: `feat/slice-8-loop-status-reporting`
Expected PR title: `feat(slice-8): loop status reporting`
Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-8-loop-status-reporting.md`

Goal:

Expose loop outcome and iteration evidence to the user.

Candidate scope:

- print loop run summary;
- print processed, skipped, blocked, and failed counts;
- expose execution IDs per iteration;
- add loop report artifact or lookup boundary if needed;
- document foreground loop behavior.

Expected outcome:

- users can understand what the loop did and what needs attention.

## Validation

Each implementation slice must run targeted tests and `make verify`.

The completed slicing must run `make verify` after the final slice is merged.

Completed slicing validation passed on `dev` after Repository PR #63 was merged.

## Risks

- The foreground loop must not become a daemon by accident.
- Loop execution must reuse single-ticket execution and result application
  rather than duplicating behavior.
- Filtering and ordering must be deterministic enough for repeated local use.
- Stop conditions must prevent accidental infinite loops or repeated attempts on
  the same blocked ticket.
