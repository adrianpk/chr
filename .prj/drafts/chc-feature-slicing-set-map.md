# ChrysoCoord Feature Slicing Set Map

Status: draft

## Purpose

Define a meta-level map of the known and inferred ChrysoCoord feature set, grouped into slicings.

In this document:

- a feature set is the broad product capability area;
- a slicing is a connected group of slices that covers one use case, feature, or workflow;
- a slice is one branch and one PR with a concrete, reviewable delivery unit.

## Slicing Size Rule

A slicing should be large enough to justify multiple slices and small enough to remain conceptually coherent.

Target shape:

- one slicing covers a recognizable use case, feature, or workflow;
- one slicing usually contains 4 to 12 slices;
- one slice should deliver one concrete capability in one reviewable PR;
- a slice should not be so small that the planning ceremony is larger than the work;
- a slice should not be so large that it covers most of the application.

## Feature Set Overview

Known and inferred feature areas:

- project initialization and protocol assets;
- user-level project registry and active project context;
- ticket model, indexing, inspection, and lifecycle;
- automatic reconciliation between Markdown and SQLite;
- operational document workflows;
- agent runtime boundary;
- ticket execution;
- filtered ticket loop;
- daemon and asynchronous coordination;
- notification and remote command surfaces;
- reporting and evidence;
- maintenance and recovery tooling.

## Slicing: Managed Project Bootstrap

Use case:

> A user initializes a Go project so it can be managed by `chc`.

Current state: mostly delivered.

Candidate slices:

- CLI bootstrap with `urfave/cli`;
- project-local `.chc/` initialization;
- asset profiles;
- root `AGENTS.md`;
- default protocol assets;
- profile-specific defaults;
- initialization dry-run and safety checks.

## Slicing: User Project Context

Use case:

> A user registers local projects and selects which project `chc` should operate on by default.

Current state: completed for the current MVP scope.

Candidate slices:

- user environment paths;
- project registry persistence;
- `project add`;
- `project list`;
- active project selection;
- active project inspection;
- project removal or disablement;
- project path validation and recovery.

## Slicing: Ticket Foundation

Use case:

> A user creates and inspects durable project tickets that agents can later execute.

Current state: completed for the current MVP scope. Ticket creation, parsing, indexing, listing, showing, comments, status transitions, and lifecycle validation are delivered.

Candidate slices:

- ticket template model;
- `ticket new`;
- ticket Markdown parser;
- ticket index repository;
- `ticket list`;
- `ticket show`;
- ticket comments;
- ticket status transitions;
- ticket lifecycle validation.

## Slicing: Automatic Reconciliation

Use case:

> A user can edit ticket files directly without remembering manual synchronization commands.

Current state: completed for the current MVP scope. Ticket file discovery, index rebuild, index status, manual recovery commands, automatic refresh before ticket reads, and the reconciliation boundary for future execution are delivered.

Candidate slices:

- ticket file fingerprinting;
- stale index detection;
- automatic refresh before ticket reads;
- automatic refresh before ticket execution;
- manual `index status` for debugging;
- manual `index rebuild` for recovery;
- conflict or parse-error reporting;
- cloned-project rehydration.

Manual synchronization commands are maintenance tools, not normal workflow requirements.

## Slicing: Operational Documents

Use case:

> A user captures and promotes working context through drafts, specs, dumps, quizzes, plans, and trackers.

Current state: manually practiced in this repository, not productized in `chc`.

Candidate slices:

- document asset templates;
- draft creation;
- spec creation;
- dump start and curation;
- quiz start and curation;
- draft-to-spec promotion;
- slicing draft generation;
- plan and tracker generation;
- document listing and lookup.

This slicing supports the broader workflow but is not the shortest path to the operational MVP.

## Slicing: Agent Runtime Boundary

Use case:

> A managed project has a Docker-based runtime contract for agent ticket execution.

Current state: completed for the current MVP scope. Runtime config, runtime asset template, runtime init, runtime plan, and runtime validation are delivered.

Candidate slices:

- runtime asset template;
- runtime config location under `.chc/`;
- validation command contract;
- Dockerfile materialization;
- runtime path and mount contract;
- result artifact path contract;
- dry-run runtime validation;
- runtime documentation.

This Dockerfile is for agent work, not production or human development.

## Slicing: Single Ticket Execution

Use case:

> A user starts one specific ticket and `chc` records what happened.

Current state: delivered.

Candidate slices:

- execution record schema;
- execution repository;
- `ticket start` command skeleton;
- ticket and project validation before execution;
- runtime preparation;
- agent harness invocation;
- structured result capture;
- validation output capture;
- execution failure handling.

## Slicing: Ticket Result Application

Use case:

> After an execution attempt, `chc` updates durable project state with status, evidence, and comments.

Current state: completed for the current MVP scope. Result application model, ticket state writing, comments, validation evidence, changed-file summaries, blocked reasons, affected-ticket index refresh, and `ticket apply-result` are delivered.

Candidate slices:

- ticket state writer;
- execution comments;
- validation evidence writer;
- changed-files summary;
- blocked reason writer;
- Markdown and SQLite consistency update;
- result report artifact;
- review-ready state handling.

## Slicing: Filtered Ticket Loop

Use case:

> A user starts a loop and `chc` keeps working through eligible tickets until completion, threshold, or blocker.

Current state: not delivered.

Candidate slices:

- eligible ticket query;
- priority, severity, area, tag, type, use-case, and status filters;
- ordering rules;
- loop command skeleton;
- loop run records;
- repeated single-ticket execution;
- stop conditions;
- retry and iteration thresholds;
- blocked-ticket handling;
- loop status reporting.

## Slicing: Codex Runtime Worker

Use case:

> `chc` runs Codex inside the project runtime container so one ticket can
> produce workspace edits, validation output, and a structured execution result.

Current state: not delivered. The runtime boundary and harness invocation exist,
but the worker that invokes Codex inside the container is not yet explicit.

Candidate slices:

- Codex command contract inside the runtime container;
- ticket and workspace context injection;
- mounted workspace write behavior;
- structured `result.json` production;
- changed-file discovery after the worker run;
- validation command execution after Codex edits;
- Codex failure classification;
- worker documentation and dry-run visibility.

## Slicing: Daemon Coordination

Use case:

> A single user-level `chc` process coordinates multiple registered projects and long-running work.

Current state: not delivered.

Candidate slices:

- daemon command skeleton;
- daemon lifecycle state;
- project registry loading;
- active loops owned by daemon;
- periodic reconciliation;
- event queue;
- daemon status reporting;
- graceful stop;
- stale process recovery.

The daemon should live in the same user-facing binary unless a later constraint forces a split.

## Slicing: Notification and Remote Commands

Use case:

> `chc` notifies the user and accepts remote input when asynchronous execution needs attention.

Current state: not delivered.

Candidate slices:

- notification event model;
- Telegram configuration;
- Telegram test command;
- Telegram blocker notification;
- Telegram reply as ticket comment or decision;
- loop-finished notification;
- email configuration;
- email summaries;
- notification audit records.

Manual notification commands are setup checks. The main value is daemon, loop, and agent-triggered messaging.

## Slicing: Reporting and Evidence

Use case:

> A user can understand what `chc` did, what changed, and why a ticket or loop is in its current state.

Current state: not delivered.

Candidate slices:

- execution report format;
- loop report format;
- validation summaries;
- changed-files summaries;
- blocker summaries;
- review handoff summaries;
- report listing;
- report lookup by ticket or execution.

## Slicing: Maintenance and Recovery

Use case:

> A user can inspect and repair local `chc` state when files, database state, or runtime artifacts drift.

Current state: partially supported through lower-level persistence primitives, not exposed as recovery UX.

Candidate slices:

- database migration status;
- project registry repair;
- index rebuild;
- parse error listing;
- orphaned execution detection;
- missing project path handling;
- runtime artifact cleanup;
- diagnostic bundle creation.

## Operational Cutoff Path

The first useful local system should prioritize slicings in this order:

1. User Project Context.
2. Ticket Foundation.
3. Automatic Reconciliation.
4. Agent Runtime Boundary.
5. Single Ticket Execution.
6. Ticket Result Application.
7. Codex Runtime Worker.
8. Filtered Ticket Loop.

At this cutoff, `chc` can take local tickets, run Codex in the runtime
container, apply workspace edits, validate, record evidence, and update ticket
state.

The loop at this cutoff is a foreground runner. A user starts it from a
terminal, it processes eligible tickets until completion, blocker, or limit,
and then exits. This is enough for local utility, but it is not the final
coordination model.

## Post-Cutoff Path

These slicings are expected after the first useful local system. They are not
required for the operational cutoff, but they remain part of the product
direction and should not be dropped from planning:

1. Daemon Coordination.
2. Notification and Remote Commands.
3. Reporting and Evidence.
4. Maintenance and Recovery.
5. Operational Documents.

Post-cutoff work can be pulled earlier only when it directly blocks local loop
operation, debugging, or evidence quality.

Daemon Coordination turns the foreground runner into a persistent background
coordinator with multi-project ownership, status reporting, remote-control
surfaces, and stale-process recovery.

## Open Questions

- Which slicings are already complete enough to mark as delivered rather than active?
- Should document workflows remain a separate slicing or be split by artifact family?
- Which reporting artifacts are required before the first ticket loop is acceptable?
- Does the first daemon need file watching, or is automatic reconciliation on command boundaries enough?
- Which notification events are required for the first remote-feedback loop?
