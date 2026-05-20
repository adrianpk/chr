# Ticket Execution Context

Status: initial

Sources:

- `/home/adrian/Downloads/context_composition_and_document_transitions.md`
- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`
- `docs/reference/ticket-execution.md`

## Purpose

Define the layered context model used when `chc` asks an agent to resolve one
ticket.

The current execution flow already has project resolution, tickets, runtime
planning, harness invocation, execution records, artifacts, result contracts,
and Codex invocation. This spec formalizes how those pieces compose into a
reproducible execution context.

## Context Layers

Ticket execution context is built from stable layers to specific layers.

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

### Project Layer

This layer defines where the agent works.

It includes:

- project ID;
- absolute project path;
- project-local `.chc/` directory;
- local project config;
- runtime config;
- user-level project registry;
- ticket index;
- workspace mounting rules.

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
