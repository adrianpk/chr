# ChrysoCoord Feature Slicing Set Map

Status: draft

## Purpose

Define a meta-level map of the known and inferred ChrysoCoord feature set, grouped into slicings.

In this document:

- a feature set is the broad product capability area;
- a slicing is a connected group of slices that covers one use case, feature, or workflow;
- a slice is one branch and one PR with a concrete, reviewable delivery unit.

## Product Boundary

ChrysoCoord is a simple orchestration tool for working on one local project with
minimal operator intervention after bootstrap.

The steady-state workflow is:

```text
tickets -> one active agent -> validation -> result application
tickets -> serial loop -> repeated single-agent execution
```

The operator may command or observe the system from the CLI, Telegram, or email.
Remote command surfaces exist to unblock work, inspect status, and provide
decisions when a running loop or daemon needs input.

ChrysoCoord is not a multi-user collaboration system and not a concurrent
multi-agent editing system. The assumed operating model is one local operator
and at most one active agent mutating a workspace at a time. A daemon may
coordinate registered projects and loops, but it should serialize work for a
project rather than run multiple agents against the same repository.

Authoring exists to turn rough project knowledge into executable tickets. It is
not a standalone editorial product.

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
- authoring sessions for dumps, quizzes, drafts, specs, plans, and slices;
- agent runtime boundary;
- ticket execution;
- ticket execution context hardening;
- MVP agent proving ground;
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

## Slicing: Operational Documents and Authoring

Use case:

> A user captures and promotes working context through dumps, quizzes,
> drafts, specs, plans, slice sets, slices, tickets, and trackers.

Current state: manually practiced in this repository, not productized in `chc`.
Markdown files already act as the durable editing surface. The missing product
layer is ergonomic creation, validation, promotion, and optional agent-assisted
authoring.

This slicing is not a replacement for direct Markdown editing. Advanced users
must remain able to create and edit artifacts by hand. Commands should make
common starts and transitions easier, validate structure, preserve provenance,
and connect artifacts without making the CLI the primary long-form editor.

The likely product shape has three complementary paths:

1. Markdown direct.
   Users create or edit artifacts directly as Markdown. This is the base
   capability and the review layer for every other path.
2. Static generated authoring.
   `chc` generates a questionnaire, draft, spec, plan, or slice skeleton in
   Markdown. The user fills or edits it manually. This is simple, reproducible,
   and useful without an adaptive session.
3. Agent-assisted authoring.
   `chc` starts a bounded authoring session backed by Codex or another model.
   The agent acts as an analyst: asks adaptive follow-up questions, captures
   answers, curates them, and produces Markdown artifacts. This is the product
   equivalent of the current manual quiz/dump/spec workflow.

A later TUI can sit on the same authoring-session model. It should not define a
separate workflow. Its job would be ergonomic interaction: ask, answer, pause,
resume, inspect transcript, edit generated Markdown, and finalize artifacts.

Shared model candidates:

- authoring session identity and status;
- mode: static, assisted, or TUI-backed;
- target: dump, quiz, draft, spec, plan, slice set, slice, or ticket;
- topic and initial prompt;
- context inputs;
- question and answer transcript;
- curated answers and decisions;
- generated artifact paths;
- provenance links between inputs and outputs.

Candidate transitions:

- dump to draft;
- quiz to draft;
- dump plus quiz to spec;
- draft to spec;
- spec to plan;
- plan to slice set;
- slice to ticket;
- execution result to report or tracker update.

Command design constraints:

- Markdown remains the canonical durable format.
- Commands must tolerate hand-authored documents.
- Commands should support dry-run when they create or transform artifacts.
- Commands should preserve enough provenance to explain generated outputs.
- Agent-assisted authoring is host-side coordination, not ticket execution.
- The Docker worker remains responsible for resolving tickets, not for deciding
  product planning structure.

Candidate slices:

- authoring artifact inventory and storage contract;
- Markdown templates for dumps, quizzes, drafts, specs, plans, and slice sets;
- static questionnaire generation;
- authoring session model and repository;
- agent-assisted quiz session runner;
- agent-assisted dump curation and synthesis;
- draft-to-spec promotion;
- spec-to-plan promotion;
- plan-to-slice-set generation;
- slice-to-ticket generation;
- document listing, lookup, validation, and relationship graph;
- TUI authoring surface over the same session backend.

This slicing supports the broader workflow but is not the shortest path to the
ticket-execution MVP. After Docker execution and context composition exist, it
becomes important because otherwise dumps, quizzes, specs, and plans remain
dependent on external conversations rather than product-owned workflows. The
purpose is to move from rough knowledge to executable tickets; once a project is
bootstrapped, the normal workflow should be ticket creation, ticket editing, and
serial agent or loop execution.

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

Current state: completed for the current operational cutoff scope. Eligible
ticket selection, loop filters, ordering, foreground command behavior, loop run
records, repeated execution, stop conditions, and loop status reporting are
delivered.

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

Current state: completed for the current operational cutoff scope. The runtime
worker command, context materialization, Docker invocation, result envelope,
changed-file discovery, post-worker validation, and worker failure
classification are delivered.

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

Current state: completed for the planned post-cutoff daemon coordination scope.
Daemon command surface, lifecycle state, state persistence, status reporting,
single-instance guard, registered project loading, foreground run loop, graceful
stop, and stale recovery are delivered.

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

## Slicing: MVP Agent Proving Ground

Use case:

> A controlled project proves that `chc` can run a sandboxed Codex agent that
> fixes code, validates, retries through Ralph Loop, and leaves the project
> working.

Current state: completed for the temporary post-cutoff MVP hardening window.
The proving-ground fixture, objective contract, sandbox Codex runner,
validation runner, Ralph Loop, and end-to-end proof are delivered.

Candidate slices:

- todolist fixture project;
- direct proving objective contract;
- sandbox Codex runner;
- validation runner;
- minimal Ralph Loop;
- end-to-end agent proof.

This slicing explicitly allows hardcoded fixture paths, a hand-authored sandbox
Dockerfile, and a direct execution path before the official workflow is fully
generalized.

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
2. MVP Agent Proving Ground.
3. Ticket Execution Context Hardening.
4. Operational Documents and Authoring.
5. Notification and Remote Commands.
6. Reporting and Evidence.
7. Maintenance and Recovery.

Post-cutoff work can be pulled earlier only when it directly blocks local loop
operation, debugging, or evidence quality.

Daemon Coordination turns the foreground runner into a persistent background
coordinator with multi-project ownership, status reporting, remote-control
surfaces, and stale-process recovery.

Ticket Execution Context Hardening formalizes how ticket, project, runtime,
document, protocol, and code-derived context are materialized before Codex
execution. It also captures Ralph Loop retry policy and execution handoff
evidence. This work strengthens the post-cutoff MVP before remote notification
surfaces.

MVP Agent Proving Ground temporarily prioritizes a direct proof over generality:
given a controlled fixture project, `chc` should run Codex in the agent sandbox,
validate compile, lint, and tests, retry with Ralph Loop, and leave a
non-trivial bugfix or feature working.

Operational Documents and Authoring should be revisited after the Docker worker
and context-composition path because the same primitives now make assisted
authoring plausible. The recommended sequence is:

1. document Markdown-first authoring and hand-edited artifact tolerance;
2. add static generated questionnaires and templates;
3. add host-side agent-assisted authoring sessions;
4. add a TUI over the same session backend.

The first two steps are ergonomic and structural. The third step is the core
capability: `chc` should be able to command Codex or another model from outside
the ticket worker to conduct a bounded quiz/dump/spec session, persist the
transcript, curate answers, and produce Markdown. The TUI should come after the
session model is stable enough to avoid designing a separate workflow.

## Open Questions

- Should document workflows remain one authoring slicing or split into static
  authoring, assisted sessions, and TUI surfaces?
- Which proving-ground behavior should be generalized before Notification and Remote Commands?
- Which execution-context specs should be implemented before or during the proving ground?
- Which reporting artifacts are required before remote execution feedback is acceptable?
- Does the next daemon phase need file watching, or is automatic reconciliation on command boundaries enough?
- Which notification events are required for the first remote-feedback loop?
- What minimum authoring session state is required for pause, resume, and
  artifact regeneration?
- Should assisted authoring run in the host process, the same Docker runtime
  family as ticket execution, or a separate authoring runtime?
