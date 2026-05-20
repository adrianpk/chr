# Architecture

ChrysoCoord separates project-local workflow files from user-level coordination
state.

## Project-Local State

Every managed project has a `.chc/` directory. It stores source workflow assets:

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

## Runtime Boundary

Ticket execution uses Docker as the sandbox boundary.

The host-side `chc` process resolves the project, selects the ticket, plans the
runtime, composes context, writes artifacts, and invokes Docker. The container
runs Codex against the mounted workspace.

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
- SQLite holds derived indexes and local coordination state.
- Docker contains agent execution effects.
- Context artifacts make each agent attempt inspectable.
- Runtime plans and execution records preserve the boundary used for an attempt.
