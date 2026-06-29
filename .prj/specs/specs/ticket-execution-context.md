# Ticket Execution Context

Status: initial

Sources:

- `/home/adrian/Downloads/context_composition_and_document_transitions.md`
- `/home/adrian/Downloads/chr_ticket_execution_context_and_ralph_loop.md`
- `docs/reference/ticket-execution.md`

## Purpose

Define the layered context model used when `chr` asks an agent to resolve one
ticket.

The current execution flow already has project resolution, tickets, runtime
planning, harness invocation, execution records, artifacts, result contracts,
and Codex invocation. This spec formalizes how those pieces compose into a
reproducible execution context.

## Context Layers

Ticket execution context is built from stable layers to specific layers.

## Context Source Inventory

The context composer should treat these sources as the initial full inventory
before pruning:

- agent identity and role definition;
- root `AGENTS.md` and nested `AGENTS.md` files that apply to the execution
  path;
- project-local `.chr/` configuration, tickets, runtime assets, and workflow
  documents;
- project-local documentation and conventions;
- ticket use-case documents, including related specs, drafts, plans, trackers,
  dumps, and quizzes;
- filesystem tree output for orientation;
- text search over source, tests, docs, and `.chr/`;
- selected `gograph` query output for Go code context;
- validation commands and expected completion signals;
- current execution attempt metadata;
- previous Ralph Loop attempt artifacts when retrying;
- output contract and evidence paths.

The inventory is intentionally broader than the prompt budget. The composer
should collect enough source material to decide what matters, then prune,
summarize, and materialize only the context required for the current ticket.

Conversation history is not an implicit Ralph Loop input. If prior discussion
is relevant, it must first be captured into a durable dump, draft, spec, ticket
comment, or execution artifact and then included through the normal source
inventory.

## Docker Runtime Compatibility

The context composer must remain compatible with the existing Docker runtime
boundary.

`chr` should compose context on the host side before Docker execution. The
runtime container should receive the result through the existing mounted
workspace and artifact paths:

- `CHR_WORKER_CONTEXT_PATH` points to the materialized `context.md`;
- `CHR_TICKET_PATH` points to the ticket or direct objective;
- `CHR_RESULT_PATH` points to the result envelope;
- `CHR_LOG_PATH` points to the agent log;
- `CHR_VALIDATION_PATH` points to validation evidence.

Additional composer outputs should live under the execution artifact directory,
for example:

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

The worker should be able to solve the ticket by reading `context.md` first and
following references to sibling context artifacts when needed.

The first implementation should not require the container to run `chr` or build
context internally. Running context composition outside the container keeps
workflow policy, pruning, source selection, and secret handling under `chr`
control.

## Per-Execution Context Package

The default model is to generate a fresh context package for each execution or
Ralph Loop attempt.

Each Docker run should receive context tailored to the ticket, project state,
attempt number, validation history, and selected code context it is about to
work on. This may repeat some work across executions, but it keeps the behavior
simple, auditable, and easy to reason about.

Invariant context may later be cached or precomputed as appendable partials,
for example:

- stable agent identity;
- stable project conventions;
- applicable `AGENTS.md` hierarchy;
- stable `.chr/` workflow descriptions;
- stable project identity and ownership metadata.

Those partials are an optimization, not the initial contract. The first
implementation should materialize the full execution context package explicitly
for each run and record enough provenance to show which sources were included.

### Invariant and Fresh Context

The composer should distinguish between context that is expected to be stable
across many executions and context that should be regenerated for each
execution.

Invariant or cacheable context includes:

- agent identity and operating stance;
- project identity;
- project conventions;
- applicable `AGENTS.md` hierarchy;
- stable `.chr/` workflow descriptions;
- default validation policy;
- repository-level documentation that changes rarely.

Fresh per-execution context includes:

- selected ticket or objective;
- ticket use-case documents and references;
- current filesystem tree;
- current text-search results;
- current `gograph` build and selected query output;
- current git status and dirty-worktree state;
- current validation commands and recent validation failures;
- current Ralph Loop attempt number and previous attempt evidence.

The initial implementation may recompute everything for clarity. Later
implementations can cache invariant partials and append them to fresh execution
context when their source fingerprints still match.

### Agent Runtime Layer

This layer defines how the agent runs.

It includes:

- runtime image;
- runtime Dockerfile;
- container workdir;
- mounts;
- Codex worker command;
- execution environment variables;
- artifact paths;
- result contract;
- runtime plan.

This layer answers: how does the agent execute?

### Agent Identity Layer

This layer defines who the agent is expected to be during execution.

It includes:

- role and operating stance;
- domain expectations;
- engineering quality bar;
- language and documentation rules;
- allowed autonomy;
- reporting expectations;
- forbidden behaviors.

For Go project execution, the default agent identity should present the worker
as a senior software engineer with strong Go, testing, infrastructure, and
repository hygiene skills. The identity should be specific enough to guide
behavior but should not replace ticket acceptance criteria or project-local
instructions.

### Project Layer

This layer defines where the agent works.

It includes:

- project ID;
- absolute project path;
- project-local `.chr/` directory;
- local project config;
- runtime config;
- user-level project registry;
- ticket index;
- workspace mounting rules.
- applicable `AGENTS.md` instructions;
- project-local `.chr/` workflow and runtime assets;
- project conventions, docs, and validation commands.

This layer answers: which managed project and workspace are being operated on?

### Ticket Layer

This layer defines the unit of work.

It includes:

- ticket key;
- ticket path;
- title and body;
- metadata;
- status;
- priority;
- severity;
- tags;
- use case;
- project area;
- acceptance criteria;
- checklist;
- comments;
- blockers;
- document references.

This layer answers: what concrete work should be attempted?

Ticket references may point to specs, drafts, plans, trackers, dumps, quizzes,
or other project documents. The composer should include these references when
they define the use case, acceptance criteria, constraints, or domain language
needed to resolve the ticket.

### Execution Layer

This layer defines the specific attempt.

It includes:

- execution ID;
- execution status;
- timestamps;
- run artifact directory;
- result artifact path;
- runtime log path;
- validation artifact path.

This separates the durable ticket from one attempt to solve it.

### Materialized Agent Context Layer

This layer is the `context.md` artifact consumed by Codex inside the runtime
container.

The materialized context must be deterministic, reviewable, and scoped to one
execution. It should contain the minimum required runtime, project, ticket, and
execution information plus references or summaries for richer semantic context.

### Code and Abstraction Layer

This layer describes relevant repository code.

It may include:

- selected filesystem tree output;
- selected text search results;
- selected `gograph` query output;
- files and directories likely related to the ticket;
- packages and symbols;
- APIs and command surfaces;
- persistence models and migrations;
- tests;
- validation commands;
- architecture and convention references.

This layer should be selected, not dumped wholesale.

## Required Properties

Execution context must be:

- reproducible for a given execution attempt;
- auditable through artifacts;
- tied to one project and one ticket;
- explicit about assumptions and missing information;
- bounded enough to avoid drowning the agent in irrelevant context;
- rich enough to avoid forcing the agent to rediscover stable project facts.

## Suggested Materialized Context Shape

A materialized context artifact may contain:

- objective;
- agent role;
- project;
- operational protocol;
- use case;
- ticket;
- checklist;
- related documents;
- repository tree summary;
- code graph context;
- validation commands;
- Ralph Loop policy;
- sandbox constraints;
- output contract.
- previous attempt summary when retrying;
- validation failure details when retrying.

The context builder may support verbosity modes:

- `verbose`: include broad summaries and selected source excerpts;
- `compact`: include summaries, paths, and high-signal excerpts;
- `minimal`: include only required execution fields and ticket prompt;
- `debug`: include extra provenance, readiness checks, and selection reasons.

The default should prefer `compact`.

## Current Implementation Fit

Already delivered:

- project context resolution;
- ticket selection through the index;
- execution records;
- runtime planning;
- harness invocation;
- worker context materialization;
- result artifact contracts.

Still to formalize:

- semantic context selection from code and documents;
- context budget and pruning rules;
- richer ticket/use-case document inclusion;
- readiness checks before invoking the agent.

## Open Questions

- Which fields are mandatory in every materialized context artifact?
- Should the context artifact embed full document content or references plus
  summaries?
- Which context layers are always present, and which are conditional?
- How should context completeness be scored before execution starts?
- Which verbosity modes should become user-facing options?
