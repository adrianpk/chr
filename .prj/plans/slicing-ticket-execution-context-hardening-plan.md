# Ticket Execution Context Hardening Slicing Plan

Status: delivered

## Purpose

Plan the `Ticket Execution Context Hardening` slicing.

This slicing defines and implements the Context Composer: the host-side
component that builds the execution context consumed by Codex inside the Docker
runtime.

The goal is to move from a small worker prompt to an auditable context package
that includes agent identity, project protocol, ticket/use-case documents,
repository orientation, semantic Go context, validation expectations, and
Ralph Loop attempt evidence.

## Source

Primary sources:

- `.prj/specs/specs/ticket-execution-context.md`
- `.prj/specs/specs/ticket-execution-context-hardening.md`
- `.prj/specs/specs/semantic-code-context.md`
- `.prj/specs/specs/agent-sandbox-and-runtime.md`
- `.prj/specs/specs/ralph-loop.md`
- `.prj/specs/specs/execution-evidence-and-handoff.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/mvp-agent-proving-ground.md`

## Current State

Delivered mechanics:

- Docker runtime boundary;
- runtime path and mount planning;
- Codex worker command rendering;
- worker context materialization;
- result and validation artifact contracts;
- bounded Ralph Loop mechanics in the proving ground;
- end-to-end proof that Codex can fix a mounted Go project through Docker.

Remaining gap:

- context composition is still minimal and not source-aware enough for larger
  repositories;
- project protocol, `.chr/`, project docs, ticket references, tree/search
  context, semantic Go context, and previous attempts are not yet composed into
  one durable execution context package;
- invariant context and fresh per-execution context are not yet separated;
- context readiness is not checked before invoking Codex.

## Execution Model

The Context Composer runs outside Docker.

`chr` owns:

- selecting context sources;
- pruning and summarizing;
- materializing context artifacts;
- deciding readiness;
- invoking Docker.

Docker owns:

- running Codex;
- exposing the mounted workspace;
- letting Codex read `context.md`;
- letting Codex edit code and write result/log artifacts.

Docker does not clone the repository in the initial model. It mounts an
existing host workspace. The same model remains compatible with a later
ephemeral worktree mode because the composer writes artifacts into the runtime
workspace before Docker starts.

## Context Package Shape

Each execution or Ralph Loop attempt receives its own context package:

```text
<run_artifact_dir>/context.md
<run_artifact_dir>/context/
  agent.md
  project.md
  protocols.md
  documents.md
  code-context.md
  previous-attempt.md
  source-inventory.md
```

`context.md` is the main Codex input referenced by
`CHR_WORKER_CONTEXT_PATH`. Sibling context files are referenced from
`context.md` and are visible inside Docker through the existing workspace
mount.

## Invariant and Fresh Context

The first implementation may recompute everything for clarity.

Later slices may introduce cacheable partials for stable context:

- agent identity;
- project identity;
- project conventions;
- applicable `AGENTS.md` hierarchy;
- stable `.chr/` workflow descriptions;
- default validation policy.

Fresh context should be rebuilt per execution:

- selected ticket or objective;
- use-case and referenced docs;
- filesystem tree;
- text search results;
- `gograph` build and selected query output;
- git status;
- validation failures;
- Ralph Loop attempt evidence.

## Slicing Scope

This slicing includes:

- Context Composer core model;
- source inventory and provenance;
- agent identity context;
- applicable `AGENTS.md` loading;
- selected `.chr/` workflow and runtime context;
- project and use-case document context;
- filesystem tree and text-search context;
- `gograph` CLI context provider for Go projects;
- Ralph Loop previous-attempt context;
- Docker-compatible context materialization;
- execution readiness checks and handoff evidence.

## Out of Scope

This slicing does not include:

- notification or remote command surfaces;
- hosted execution;
- multi-user execution;
- Tree-sitter integration;
- `gopls` daemon integration;
- clone, commit, push, or PR creation from inside the container;
- replacing the existing Docker runtime boundary.

Ephemeral worktree mode and PAT-backed remote push flows are future runtime
workspace concerns. This slicing keeps the context package compatible with
those future modes but does not implement them.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Context Composer Contract

Status: delivered.

Expected branch: `feat/slice-1-context-composer-contract`
Expected PR title: `feat(slice-1): context composer contract`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-1-context-composer-contract.md`

Goal:

Define the side-effect-free core model for composing execution context before
Docker starts.

Candidate scope:

- add composer input and output models;
- define source inventory entries;
- define materialized artifact paths;
- define readiness status and warnings;
- define invariant versus fresh source classification;
- add focused tests.

Expected outcome:

- context sources can be modeled without changing Docker invocation.

### Slice 2: Agent Identity and Protocol Context

Status: delivered.

Expected branch: `feat/slice-2-context-agent-protocols`
Expected PR title: `feat(slice-2): context agent protocols`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-2-context-agent-protocols.md`

Goal:

Materialize stable agent identity and applicable project protocol context.

Candidate scope:

- define default Go/software/infrastructure agent identity;
- load root and nested `AGENTS.md` files that apply to the execution path;
- load selected `.chr/` workflow and runtime context;
- render `agent.md`;
- render `protocols.md`;
- record source provenance.

Expected outcome:

- Codex receives stable role and protocol context before reading project code.

### Slice 3: Project and Use-Case Document Context

Status: delivered.

Expected branch: `feat/slice-3-context-project-documents`
Expected PR title: `feat(slice-3): context project documents`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-3-context-project-documents.md`

Goal:

Collect bounded project and ticket-related document context.

Candidate scope:

- summarize project identity and validation commands;
- follow ticket references to specs, drafts, plans, trackers, dumps, and
  quizzes;
- render `project.md`;
- render `documents.md`;
- record included and skipped sources in `source-inventory.md`;
- avoid implicit conversation history.

Expected outcome:

- use-case context reaches Codex through durable artifacts instead of hidden
  conversation state.

### Slice 4: Tree and Search Context

Status: delivered.

Expected branch: `feat/slice-4-context-tree-search`
Expected PR title: `feat(slice-4): context tree search`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-4-context-tree-search.md`

Goal:

Add cheap repository orientation before semantic graph integration.

Candidate scope:

- capture bounded filesystem tree output;
- derive text-search terms from ticket and validation context;
- run targeted searches over source, tests, docs, and `.chr/`;
- render tree and search sections into `code-context.md`;
- keep output deterministic enough for tests.

Expected outcome:

- Codex starts with repository orientation without scanning from zero.

### Slice 5: Go Graph Context Provider

Status: delivered.

Expected branch: `feat/slice-5-context-gograph-provider`
Expected PR title: `feat(slice-5): context gograph provider`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-5-context-gograph-provider.md`

Goal:

Integrate `gograph` CLI as the first semantic code context provider for Go
projects.

Candidate scope:

- detect whether `gograph` is available;
- run `gograph build .` by default;
- support optional `gograph build . --precise`;
- query selected source, callers, callees, tests, deps, changes, and hotspots;
- render selected output into `code-context.md`;
- degrade gracefully when `gograph` is unavailable.

Expected outcome:

- larger Go repositories get reusable semantic orientation without dumping a
  whole graph into the prompt.

### Slice 6: Ralph Attempt Context

Status: delivered.

Expected branch: `feat/slice-6-context-ralph-attempts`
Expected PR title: `feat(slice-6): context ralph attempts`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-6-context-ralph-attempts.md`

Goal:

Feed previous attempt evidence into retries through explicit artifacts.

Candidate scope:

- load previous validation artifact;
- summarize failed commands and repeated failure signatures;
- include previous changed files and result status;
- render `previous-attempt.md`;
- link previous-attempt context from the next `context.md`;
- keep conversation history out unless it was materialized as an artifact.

Expected outcome:

- Ralph Loop retries use explicit attempt evidence, not hidden memory.

### Slice 7: Docker-Compatible Context Materialization

Status: delivered.

Expected branch: `feat/slice-7-context-docker-materialization`
Expected PR title: `feat(slice-7): context docker materialization`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-7-context-docker-materialization.md`

Goal:

Write the composed context package into the artifact layout consumed by Docker.

Candidate scope:

- write `context.md` under `run_artifact_dir`;
- write sibling context artifacts under `run_artifact_dir/context/`;
- ensure paths are visible through the runtime mount;
- keep `CHR_WORKER_CONTEXT_PATH` as the main Codex input;
- add dry-run visibility for included sources;
- add tests for path rendering and artifact layout.

Expected outcome:

- the richer context package works with the existing Docker runtime.

### Slice 8: Execution Readiness and Handoff

Status: delivered.

Expected branch: `feat/slice-8-context-readiness-handoff`
Expected PR title: `feat(slice-8): context readiness handoff`
Expected report: `.prj/reports/slices/ticket-execution-context-hardening/slice-8-context-readiness-handoff.md`

Goal:

Block low-quality executions before Codex starts and expose context evidence
after execution.

Candidate scope:

- define required context fields;
- report missing ticket, protocol, validation, or runtime context;
- record context completeness and warnings;
- link context artifacts from execution records or reports;
- document remaining worktree and PAT-backed push requirements;
- add focused tests.

Expected outcome:

- ticket execution has a reviewable context handoff before agent work begins.

## Acceptance Criteria

The slicing is complete when:

- `chr` can compose a deterministic context package before Docker starts;
- `context.md` remains the main Codex input;
- AGENTS, `.chr/`, project docs, ticket docs, tree/search, `gograph`, and
  previous attempts are represented as bounded artifacts;
- the Docker runtime contract consumes richer artifacts without becoming the
  context composer;
- readiness failures are reported before invoking Codex;
- every slice has a persistent implementation report;
- validation passes with `make verify`.

## Validation

Required validation for every implementation slice in this slicing:

```sh
make verify
```

Additional smoke checks should include a dry-run context build once the CLI is
available.

## Risks

- Context can become too large unless pruning is part of the contract.
- Generated context can become stale unless provenance and refresh rules are
  explicit.
- `gograph build . --precise` may fail in broken repositories and must not
  block the default fast path.
- Secret values must never be rendered into durable context artifacts.
- Mounted workspace execution requires users to avoid concurrent edits while
  the agent is running.

## Notes

- This slicing generalizes the proving-ground context path.
- It does not change the fact that Docker is the runtime boundary.
- It does not implement clone, commit, push, or PR behavior.
