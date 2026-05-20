# Agent Runtime Boundary Slicing Plan

Status: completed

## Purpose

Plan the `Agent Runtime Boundary` slicing from the feature slicing set map.

This slicing covers the use case:

> A managed project has a Docker-based runtime contract for agent ticket execution.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related documents:

- `.prj/drafts/chc-command-surfaces-and-priority.md`
- `.prj/drafts/chc-capability-map.md`
- `.prj/drafts/chc-slicing-map.md`
- `.prj/specs/specs/user-daemon-and-global-index.md`
- `docs/reference/project-workspace.md`
- `docs/reference/ticket-reconciliation.md`

Prerequisite delivered work:

- `User Project Context` slicing.
- `Ticket Foundation` slicing.
- `Automatic Reconciliation` slicing.

## Current State

Already delivered:

- project-local `.chc/` bootstrap;
- project config file creation;
- project registry and active context;
- ticket model and global ticket index;
- ticket reconciliation before execution selection.

Delivered by this slicing:

- a runtime config contract under `.chc/config.yml`;
- an agent-runtime Dockerfile template;
- a command to materialize runtime assets safely;
- a computed runtime boundary plan for future execution;
- a validation command that checks runtime readiness without starting ticket execution.

## Slicing Scope

This slicing includes:

- runtime config model and defaults;
- Dockerfile template for agent ticket work;
- runtime asset materialization;
- runtime paths, mounts, and artifact directory contract;
- dry-run runtime validation;
- reference documentation for the runtime boundary.

## Out of Scope

This slicing does not include:

- invoking Codex or any agent harness;
- executing a ticket;
- creating execution records;
- applying ticket results;
- filtered loops;
- daemon lifecycle;
- Telegram or email notifications.

The runtime boundary prepares the contract that later execution slicings will use.

## Slice Numbering

This slicing uses local slice numbering. The slice number is scoped to this
plan, starts at `Slice 1`, and is the number used in branch names, PR titles,
trackers, reports, and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Runtime Config Model

Status: delivered.

Expected branch: `feat/slice-1-runtime-config-model`
Expected PR title: `feat(slice-1): runtime config model`
Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-1-runtime-config-model.md`

Goal:

Add a typed runtime configuration model loaded from `.chc/config.yml`.

Candidate scope:

- extend `WorkspaceConfig` with runtime fields;
- define safe defaults when config keys are absent;
- parse runtime image, dockerfile path, workdir, artifact directory, and validation command;
- add focused config tests;
- document the config fields.

Expected outcome:

- runtime code can read one explicit config contract instead of hard-coding paths in commands.

### Slice 2: Runtime Asset Template

Status: delivered.

Expected branch: `feat/slice-2-runtime-asset-template`
Expected PR title: `feat(slice-2): runtime asset template`
Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-2-runtime-asset-template.md`

Goal:

Provide default agent-runtime assets for managed projects.

Candidate scope:

- add an agent-runtime Dockerfile template to embedded assets;
- add default runtime config fields to profile config files;
- ensure `chc init` creates the runtime template without overwriting existing user files;
- add focused init tests;
- document the template purpose and ownership.

Expected outcome:

- new managed projects receive a tunable Dockerfile for agent ticket work.

### Slice 3: Runtime Init CLI

Status: delivered.

Expected branch: `feat/slice-3-runtime-init-cli`
Expected PR title: `feat(slice-3): runtime init cli`
Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-3-runtime-init-cli.md`

Goal:

Expose safe runtime asset materialization as a CLI command.

Candidate scope:

- add a `runtime` command group if needed;
- add `chc runtime init`;
- resolve project context through the existing resolver;
- support dry-run output;
- skip existing runtime files by default;
- add focused CLI tests;
- document command behavior.

Expected outcome:

- existing managed projects can install or inspect runtime defaults without re-running full workspace initialization.

### Slice 4: Runtime Plan Contract

Status: delivered.

Expected branch: `feat/slice-4-runtime-plan-contract`
Expected PR title: `feat(slice-4): runtime plan contract`
Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-4-runtime-plan-contract.md`

Goal:

Compute the Docker runtime boundary plan that future execution code will use.

Candidate scope:

- add a core runtime planner;
- resolve project root, workspace root, Dockerfile path, container workdir, mount paths, and result artifact paths;
- keep the planner side-effect free;
- add focused tests for defaults and custom config;
- document path and mount semantics.

Expected outcome:

- execution code can request a deterministic runtime boundary without constructing Docker paths itself.

### Slice 5: Runtime Validate CLI

Status: delivered.

Expected branch: `feat/slice-5-runtime-validate-cli`
Expected PR title: `feat(slice-5): runtime validate cli`
Expected report: `.prj/reports/slices/agent-runtime-boundary/slice-5-runtime-validate-cli.md`

Goal:

Expose runtime readiness checks without running an agent.

Candidate scope:

- add `chc runtime validate`;
- resolve project context and runtime config;
- verify required runtime files and directories;
- print the computed Docker build/run plan;
- support dry-run validation without requiring Docker execution;
- add focused CLI tests;
- document validation output and limits.

Expected outcome:

- a user can verify the runtime boundary before ticket execution exists.

## Acceptance Criteria

The slicing is complete when:

- runtime config can be loaded from `.chc/config.yml`;
- managed projects can receive a default agent-runtime Dockerfile template;
- existing projects can materialize runtime defaults through CLI;
- runtime paths and mounts can be computed through one core contract;
- runtime readiness can be inspected without invoking an agent;
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
chc init --dir "$tmpdir"
chc runtime init --dir "$tmpdir" --dry-run
chc runtime init --dir "$tmpdir"
chc runtime validate --dir "$tmpdir" --dry-run
```

## Risks

- Runtime defaults must not overwrite user-tuned Dockerfiles.
- Runtime commands must not execute agent work during this slicing.
- Docker is the preferred boundary, but tests must not require a Docker daemon.
- The runtime Dockerfile must be clearly distinct from production or human development Dockerfiles.

## Notes

- This slicing prepares ticket execution but does not perform it.
- Markdown reports are required for every slice in this slicing.
- HTML rendering is not part of this slicing; Markdown reports remain canonical.
