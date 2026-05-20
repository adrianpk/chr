# Codex Runtime Worker Slicing Plan

Status: complete

## Purpose

Plan the `Codex Runtime Worker` slicing from the feature slicing set map.

This slicing covers the use case:

> `chc` runs Codex inside the project runtime container so one ticket can
> produce workspace edits, validation output, and a structured execution result.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related documents:

- `.prj/plans/slicing-agent-runtime-boundary-plan.md`
- `.prj/plans/slicing-single-ticket-execution-plan.md`
- `.prj/plans/slicing-ticket-result-application-plan.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-result-application.md`

Prerequisite delivered work:

- `Agent Runtime Boundary` slicing.
- `Single Ticket Execution` slicing.
- `Ticket Result Application` slicing.

## Current State

Delivered:

- project runtime configuration and Docker runtime planning;
- Docker build/run command rendering;
- ticket execution records and artifact paths;
- execution result loading and validation;
- result application back into ticket state;
- explicit Codex command contract inside the runtime container;
- ticket and workspace context passed to the worker;
- worker-owned structured result production;
- changed-file discovery from workspace edits;
- validation after Codex edits;
- Codex-specific failure classification;
- worker dry-run and documentation.

## Slicing Scope

This slicing includes:

- Codex worker command modeling;
- prompt and context materialization for one ticket execution;
- harness invocation changes needed to call the worker;
- structured result envelope extensions for changed files;
- changed-file discovery after the worker run;
- validation command execution after Codex edits;
- Codex failure classification;
- user-facing documentation for the worker contract.

## Out of Scope

This slicing does not include:

- foreground filtered ticket loops;
- daemon coordination;
- notification or remote command surfaces;
- multi-project scheduling;
- PR creation or forge integration;
- a hosted Codex service.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Worker Contract Model

Status: delivered.

Expected branch: `feat/slice-1-worker-contract-model`
Expected PR title: `feat(slice-1): worker contract model`
Expected report: `.prj/reports/slices/codex-runtime-worker/slice-1-contract-model.md`

Goal:

Define the side-effect free Codex worker command and context contract.

Candidate scope:

- add worker command/domain types;
- define ticket, execution, result, log, and validation paths;
- render the shell command that invokes Codex inside the container;
- validate required command inputs;
- add focused command-rendering tests;
- document the worker contract.

Expected outcome:

- later slices can route harness execution through a stable worker contract.

### Slice 2: Worker Context Materialization

Status: delivered.

Expected branch: `feat/slice-2-worker-context-materialization`
Expected PR title: `feat(slice-2): worker context materialization`
Expected report: `.prj/reports/slices/codex-runtime-worker/slice-2-context-materialization.md`

Goal:

Write the ticket execution context consumed by Codex inside the container.

Candidate scope:

- create a worker prompt/context artifact path;
- render ticket key, ticket path, execution ID, and result contract;
- keep the context deterministic and reviewable;
- add focused materialization tests.

Expected outcome:

- each execution has a durable Codex input artifact.

### Slice 3: Harness Worker Invocation

Status: delivered.

Expected branch: `feat/slice-3-harness-worker-invocation`
Expected PR title: `feat(slice-3): harness worker invocation`
Expected report: `.prj/reports/slices/codex-runtime-worker/slice-3-harness-invocation.md`

Goal:

Run the Codex worker command through the existing Docker harness invocation.

Candidate scope:

- add worker command into harness invocation rendering;
- preserve Docker build/run behavior;
- pass worker context and artifact paths through environment or arguments;
- keep command construction side-effect free;
- add focused invocation tests.

Expected outcome:

- `ticket start` can invoke the worker instead of running only validation.

### Slice 4: Result Envelope Changed Files

Status: delivered.

Expected branch: `feat/slice-4-result-changed-files`
Expected PR title: `feat(slice-4): result changed files`
Expected report: `.prj/reports/slices/codex-runtime-worker/slice-4-result-changed-files.md`

Goal:

Carry changed-file summaries in the structured execution result envelope.

Candidate scope:

- extend execution result types with changed files;
- preserve strict JSON decoding;
- validate changed-file paths and summaries;
- map result changed files into result application;
- add focused result-loading tests.

Expected outcome:

- changed files can flow from the worker result into ticket result application.

### Slice 5: Changed File Discovery

Status: delivered.

Expected branch: `feat/slice-5-changed-file-discovery`
Expected PR title: `feat(slice-5): changed file discovery`
Expected report: `.prj/reports/slices/codex-runtime-worker/slice-5-changed-file-discovery.md`

Goal:

Discover workspace file changes caused by the worker run.

Candidate scope:

- inspect git changed files after execution;
- scope discovery to the project workspace;
- handle clean worktrees;
- avoid reverting or mutating changes;
- add focused discovery tests.

Expected outcome:

- worker executions can report actual changed paths even when summaries are
  minimal.

### Slice 6: Post-Worker Validation

Status: delivered.

Expected branch: `feat/slice-6-post-worker-validation`
Expected PR title: `feat(slice-6): post worker validation`
Expected report: `.prj/reports/slices/codex-runtime-worker/slice-6-post-worker-validation.md`

Goal:

Run validation after Codex edits and capture validation outcome separately from
worker execution.

Candidate scope:

- run the configured validation command after the worker command;
- preserve validation output artifact paths;
- classify validation failures distinctly;
- ensure result capture sees validation failure correctly;
- add focused execution runner tests.

Expected outcome:

- execution success means Codex ran and validation passed.

### Slice 7: Worker Failure Classification and Docs

Status: delivered.

Expected branch: `feat/slice-7-worker-failure-docs`
Expected PR title: `feat(slice-7): worker failure classification and docs`
Expected report: `.prj/reports/slices/codex-runtime-worker/slice-7-failure-docs.md`

Goal:

Make Codex worker failures explicit and document the end-to-end worker contract.

Candidate scope:

- add Codex worker failure classification;
- surface worker diagnostics in execution output;
- document container prerequisites and result contract;
- document dry-run visibility;
- add focused failure tests;
- run `make verify`.

Expected outcome:

- users can distinguish Codex worker failure, harness failure, result capture
  failure, and validation failure.

## Validation

Each implementation slice ran targeted tests and `make verify`.

Completed slicing validation after the final slice merge:

- `make verify`: pass.

## Risks

- Codex CLI availability and authentication inside the container must be
  explicit, not assumed from the host environment.
- The worker must not hide workspace mutations behind opaque output.
- The result envelope must remain strict enough for automation but simple
  enough for worker-side generation.
