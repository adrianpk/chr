# Single Ticket Execution Slicing Plan

Status: completed

## Purpose

Plan the `Single Ticket Execution` slicing from the feature slicing set map.

This slicing covers the use case:

> A user starts one specific ticket and `chr` records what happened.

## Source

Primary source:

- `.prj/drafts/chr-feature-slicing-set-map.md`

Related documents:

- `.prj/drafts/chr-command-surfaces-and-priority.md`
- `.prj/drafts/chr-capability-map.md`
- `.prj/specs/specs/ticket-template.md`
- `.prj/specs/specs/user-daemon-and-global-index.md`
- `docs/reference/project-workspace.md`
- `docs/reference/ticket-reconciliation.md`

Prerequisite delivered work:

- `User Project Context` slicing.
- `Ticket Foundation` slicing.
- `Automatic Reconciliation` slicing.
- `Agent Runtime Boundary` slicing.

## Current State

Already delivered:

- project context resolution for commands;
- ticket creation, parsing, listing, showing, status updates, and comments;
- global ticket index and automatic reconciliation before ticket reads;
- reconciliation blockers for future execution;
- runtime config, template, init, planning, and validation.

Delivered in this slicing:

- durable execution record model;
- execution persistence;
- a `chr ticket start` entrypoint;
- execution preflight that combines ticket, project, reconciliation, and runtime checks;
- an invocation boundary for the first agent harness;
- structured execution result capture;
- validation output capture;
- execution failure recording.

## Slicing Scope

This slicing includes:

- execution record schema and storage;
- execution repository APIs;
- `chr ticket start` command surface;
- single-ticket selection and preflight gating;
- runtime plan reuse before execution;
- first agent harness invocation contract;
- structured execution result and log capture;
- validation output capture for the execution attempt;
- reference documentation for single-ticket execution.

## Out of Scope

This slicing does not include:

- applying execution results to ticket Markdown;
- changing ticket status based on execution outcome;
- writing ticket comments from execution output;
- filtered ticket loops;
- daemon lifecycle;
- retry policies;
- Telegram or email notifications;
- remote user approval flows.

Ticket result application is the next planned slicing after single-ticket
execution.

## Slice Numbering

This slicing uses local slice numbering. The slice number is scoped to this
plan, starts at `Slice 1`, and is the number used in branch names, PR titles,
trackers, reports, and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Execution Record Model

Status: delivered.

Expected branch: `feat/slice-1-execution-record-model`
Expected PR title: `feat(slice-1): execution record model`
Expected report: `.prj/reports/slices/single-ticket-execution/slice-1-execution-record-model.md`

Goal:

Define the durable execution record shape for one ticket execution attempt.

Candidate scope:

- add execution record domain types;
- define execution status values;
- define execution timestamps and ticket/project references;
- define runtime plan/result path references;
- add focused model tests;
- document the execution record contract.

Expected outcome:

- later slices can persist and print one execution attempt using a stable model.

### Slice 2: Execution Repository

Status: delivered.

Expected branch: `feat/slice-2-execution-repository`
Expected PR title: `feat(slice-2): execution repository`
Expected report: `.prj/reports/slices/single-ticket-execution/slice-2-execution-repository.md`

Goal:

Persist execution records in the user-level SQLite store.

Candidate scope:

- add SQLite migration for execution records;
- add repository APIs for create, update, and lookup;
- connect execution records to project ID and ticket key;
- add focused repository tests;
- document persistence layout.

Expected outcome:

- `chr` can durably record execution attempts before invoking any harness.

### Slice 3: Ticket Start Preflight

Status: delivered.

Expected branch: `feat/slice-3-ticket-start-preflight`
Expected PR title: `feat(slice-3): ticket start preflight`
Expected report: `.prj/reports/slices/single-ticket-execution/slice-3-ticket-start-preflight.md`

Goal:

Add a safe `chr ticket start <ticket-key>` preflight command without invoking
an agent.

Candidate scope:

- add `ticket start` command;
- resolve project context through the existing resolver;
- reconcile tickets before selection;
- select one indexed ticket by key;
- run runtime validation;
- create a pending execution record;
- support dry-run output;
- add focused CLI tests;
- document command behavior.

Expected outcome:

- a user can ask `chr` to prepare one ticket execution and see blockers before
  any agent work starts.

### Slice 4: Harness Invocation Contract

Status: delivered.

Expected branch: `feat/slice-4-harness-invocation-contract`
Expected PR title: `feat(slice-4): harness invocation contract`
Expected report: `.prj/reports/slices/single-ticket-execution/slice-4-harness-invocation-contract.md`

Goal:

Define the first command contract for invoking an agent harness inside the
runtime boundary.

Candidate scope:

- add harness invocation model;
- render Docker build/run command arguments from `RuntimePlan`;
- include ticket key and execution ID in artifact paths;
- keep command construction side-effect free;
- add focused command rendering tests;
- document the harness invocation contract.

Expected outcome:

- execution code can construct a deterministic harness invocation without
  embedding Docker argument logic in the CLI.

### Slice 5: Agent Harness Invocation

Status: delivered.

Expected branch: `feat/slice-5-agent-harness-invocation`
Expected PR title: `feat(slice-5): agent harness invocation`
Expected report: `.prj/reports/slices/single-ticket-execution/slice-5-agent-harness-invocation.md`

Goal:

Invoke the configured agent harness for one ticket through the runtime boundary.

Candidate scope:

- add execution runner service;
- invoke the harness command from the runtime plan;
- stream or capture stdout/stderr to execution artifacts;
- update execution record status on start, success, and failure;
- preserve dry-run behavior;
- add focused runner tests with a fake command executor;
- document invocation limits.

Expected outcome:

- `chr ticket start <ticket-key>` can run one bounded execution attempt and
  record the attempt outcome.

### Slice 6: Execution Result Capture

Status: delivered.

Expected branch: `feat/slice-6-execution-result-capture`
Expected PR title: `feat(slice-6): execution result capture`
Expected report: `.prj/reports/slices/single-ticket-execution/slice-6-execution-result-capture.md`

Goal:

Capture structured execution results and validation output as durable artifacts.

Candidate scope:

- define result envelope shape;
- write result and validation artifacts under the runtime run directory;
- record artifact paths in the execution record;
- handle missing or malformed result output;
- add focused result capture tests;
- document execution artifact semantics.

Expected outcome:

- later result-application code can consume one execution attempt without
  parsing terminal output.

### Slice 7: Execution Failure Handling

Status: delivered.

Expected branch: `feat/slice-7-execution-failure-handling`
Expected PR title: `feat(slice-7): execution failure handling`
Expected report: `.prj/reports/slices/single-ticket-execution/slice-7-execution-failure-handling.md`

Goal:

Make single-ticket execution failures explicit and inspectable.

Candidate scope:

- classify preflight, runtime, harness, validation, and result-capture failures;
- record failure reason and diagnostic paths;
- print concise failure output from `ticket start`;
- add focused failure tests;
- document failure semantics.

Expected outcome:

- users and later loop code can distinguish blocked execution from failed
  execution attempts.

## Acceptance Criteria

The slicing is complete when:

- execution records can be modeled and persisted;
- `chr ticket start <ticket-key>` exists;
- execution preflight reconciles tickets and validates runtime readiness;
- one execution attempt can be invoked through the runtime boundary;
- execution output and result artifacts are captured durably;
- execution failure states are recorded clearly;
- every slice has a persistent implementation report;
- validation passes with `make verify`.

## Validation

Required validation for every implementation slice in this slicing:

```sh
make verify
```

Manual smoke checks for the completed slicing:

```sh
make install
tmpdir="$(mktemp -d)"
chr init --dir "$tmpdir"
chr ticket new --dir "$tmpdir" "Smoke ticket"
chr runtime validate --dir "$tmpdir" --dry-run
chr ticket start --dir "$tmpdir" CHR-1 --dry-run
```

Additional smoke checks depend on the harness execution behavior delivered by
later slices in this slicing.

## Notes

- This slicing records execution attempts but does not apply results to ticket
  Markdown.
- Result application is the next slicing in the MVP path.
- Dry-run behavior must remain available before actual harness invocation.
