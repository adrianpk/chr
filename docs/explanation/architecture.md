# Architecture

chr separates project-local workflow files from user-level coordination
state.

## Project-Local State

Every managed project has a `.chr/` directory. It stores source workflow assets:

- tickets;
- comments;
- project protocol files;
- runtime config;
- runtime Dockerfile;
- context artifacts;
- local reports and operational project documents.

These files belong to the managed project. They are meant to travel with the
project when that is useful.

## User-Level State

User-level state stores derived and coordinating data:

- registered projects;
- active project selection;
- ticket indexes;
- execution records;
- loop run records;
- daemon lifecycle state.

This state exists outside the managed project because it represents the local
user's coordination view, not the project's portable source of truth.

Project and case metadata are not duplicated into SQLite as authoritative
records. They remain Markdown in the managed project. Runtime code may build a
structural workspace index in memory, and may cache discardable index files
under `.chr/default/tmp/`, but those caches are derived from project files.

## Runtime Boundary

Ticket execution uses Docker as the sandbox boundary.

The host-side `chr` process resolves the project, selects the ticket, plans the
runtime, composes context, writes artifacts, and invokes Docker. The container
runs Codex against the mounted core.

The container does not own context selection. It receives explicit artifacts and
paths.

## Context Boundary

The context composer builds a bounded package before Codex starts:

```text
context.md
context/
  agent.md
  project.md
  protocols.md
  documents.md
  code-context.md
  previous-attempt.md
  source-inventory.md
```

This package makes agent execution reviewable. The worker should not depend on
implicit conversation history.

## Operation Surfaces

The CLI and TUI are two operation surfaces over the same project files and
service boundaries.

- The TUI is optimized for interactive navigation, inspection, editing,
  confirmations, and focused work modes.
- The CLI is optimized for scripted operations and agent workflows.
- Shared actions route through application services instead of separate
  implementations.

Project metadata, use-case metadata, dumps, runtime quizzes, specs, plans, plan
tracking, tickets, and work actions are expected to preserve this shape. The
[CLI TUI action matrix](../reference/cli-tui-action-matrix.md) records the
current parity status and remaining intentional gaps.

## Ticket Execution Boundary

A ticket execution record represents one attempt to work on one ticket. It is
separate from the Markdown ticket and from result application.

Execution records preserve:

- execution identity;
- selected ticket;
- runtime references;
- artifact paths;
- lifecycle status;
- failure reason when present.

Ticket result application reads execution evidence and updates ticket Markdown.

## Loop and Daemon Boundary

The ticket loop is a foreground runner over eligible tickets in one project.

The daemon is a local foreground coordinator over registered projects. It owns
coordination state and project refresh behavior. It is not a hosted service and
not a notification transport.

## Why This Shape

The design keeps the project workflow auditable while avoiding hidden state:

- Markdown remains readable and portable.
- SQLite holds user-level coordination state and derived indexes where those
  domains still need persistence.
- The workspace structural index is derived from Markdown and can be rebuilt.
- Docker contains agent execution effects.
- Context artifacts make each agent attempt inspectable.
- Runtime plans and execution records preserve the boundary used for an attempt.
