# User Daemon and Global Index

Status: initial

Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Purpose

Define the user-centric runtime model for `chr`.

`chr` is a local tool for one user managing multiple local Go projects. It is not a SaaS architecture and should not introduce account, tenant, or organization concepts.

## One Binary

`chr` remains a single binary.

The binary supports two execution modes:

- one-shot CLI commands, such as `chr init` and `chr ticket new`;
- a resident daemon command, `chr daemon`.

There should not be a separate `chrd` binary.

## User Daemon

`chr daemon` is a user daemon.

The daemon may observe multiple registered local projects at the same time. This avoids running one `chr` process and one Telegram bot per project.

The daemon is responsible for long-running behavior:

- filesystem watching for registered project ticket directories;
- keeping the global ticket index fresh;
- handling Telegram integration;
- coordinating ticket loops;
- maintaining runtime locks, runs, and traces.

## Project Workspace

Each managed project owns a project-local `.chr/` directory.

The `.chr/` directory contains durable project artifacts such as:

- project config;
- Markdown tickets;
- project protocol assets;
- project-local runtime files when needed.

Markdown ticket files are the durable canonical ticket representation. A cloned project containing `.chr/tickets/*.md` should contain enough information to rehydrate the ticket index on another machine.

## User Environment

`chr` needs an abstraction that resolves user-level paths.

This abstraction must support normal mode and development mode, so developing `chr` does not accidentally write to real user config or data paths.

Normal mode should resolve paths from platform conventions, for example:

```text
config -> ~/.config/chr
data   -> ~/.local/share/chr
cache  -> ~/.cache/chr
state  -> ~/.local/state/chr
```

Development mode should resolve those paths into a local sandbox, for example:

```text
config -> ._workspace/config/chr
data   -> ._workspace/data/chr
cache  -> ._workspace/cache/chr
state  -> ._workspace/state/chr
```

Exact directory names may change during implementation, but the separation must remain explicit.

## Project Registry

The daemon needs a user-local registry of projects it manages.

The registry is not product data and does not replace project-local `.chr/` state. It only records which local projects this user's daemon should know about.

Future commands may include:

```text
chr project add .
chr project list
chr project reindex <project>
```

Registering a cloned project should read its `.chr/` directory and rehydrate the global index from Markdown tickets.

## Global Ticket Index

The ticket index is global to the user, not per project.

The index should live under the user data directory, for example:

```text
~/.local/share/chr/chr.db
```

The index is derived from registered projects and their Markdown ticket files. It is an operational index for fast search, filtering, Telegram responses, and daemon loops.

The index is not the canonical ticket source.

The global index requires a schema migration system before registry or ticket index tables become durable behavior.

Migration files should be embedded SQL assets with deterministic ordering and explicit up/down sections:

```text
assets/migration/sqlite/20260518000001-create_projects.sql
```

```sql
-- +migrate Up
CREATE TABLE projects (...);

-- +migrate Down
DROP TABLE IF EXISTS projects;
```

The migrator should be a small core primitive owned by `chr`, not an external framework dependency. It should apply pending migrations in order, record applied versions, and make schema setup deterministic for CLI and daemon code.

The current implementation package direction is `internal/chr`. If a `pkg/core` package is introduced for reusable core primitives, that move should be explicit in its own slice rather than incidental to persistence work.

## Project Model

The user daemon needs a project entity.

A registered project represents a local project known by the user daemon. It does not replace project-local `.chr/config.yml`; it is the user-level record that allows one daemon to observe and operate across multiple local projects.

The initial project attributes are:

- stable local project id;
- human-readable project name;
- absolute project path;
- enabled flag;
- registered timestamp;
- last indexed timestamp.

The project entity must have a relational shape because it will be persisted in the global user database.

## Markdown and Index Contract

Markdown ticket files and the global index must match.

The index must store enough file state to detect stale rows, such as:

- project identity;
- ticket path;
- file modification time;
- file size;
- content hash;
- indexed timestamp.

When a ticket file changes externally, the daemon should detect the change and refresh the corresponding index rows.

When the daemon is not running, one-shot CLI commands should still work and should be able to refresh or rebuild the index when needed.

## Runtime State

Runtime state is separate from ticket data.

Locks, active runs, traces, daemon process state, and transient Telegram state are not canonical ticket content. They may live under user state paths or project-local runtime paths depending on the implementation need.

## Package Direction

Core behavior should live under `internal/chr`.

The current `internal/workspace` package name creates conceptual tension because the project-local `.chr/` workspace is one entity among several `chr` concepts. The core package should expose names such as:

```go
chr.Workspace
chr.NewWorkspace(...)
chr.UserEnvironment
chr.ProjectRegistry
chr.TicketIndex
```

The CLI package should wire commands and delegate behavior to `internal/chr`.

## Open Questions

- What exact development-mode activation mechanism should be used?
- Should the project registry live in YAML first, SQLite first, or both?
- Which runtime files belong under user state versus project-local `.chr/runtime/`?
- What exact filesystem watcher behavior is required for recursive ticket directories?
