# Go Architecture

This document describes Go repository structure and ownership boundaries.

## Objective

Keep the Go codebase explicit, package-oriented, and easy to reason about without hiding workflows in oversized packages or command entrypoints.

## Baseline structure

```text
cmd/          # executable entrypoints
internal/     # application code private to this module
pkg/          # public/shared contracts only when external reuse is intentional
assets/       # embedded non-code assets, schemas, templates, and manifests
docs/         # user-facing and technical documentation
ops/          # specs, plans, tickets, quiz notes, dumps, and delivery artifacts
testdata/     # test fixtures owned by package tests
.agents/      # collaboration and execution rules for this repository
AGENTS.md     # top-level agent operating contract
```

## Boundary rules

- Command packages under `cmd/` should assemble dependencies, parse process-level configuration, start lifecycle components, and stay thin.
- Application behavior should live in `internal/` by default.
- ChrysoCoord core behavior should live under `internal/chc`.
- ChrysoCoord application operations should live under `internal/chcapp` behind
  the central operation service boundary.
- Use `pkg/` only for contracts, envelope types, or libraries that are intentionally shared outside the module.
- Keep package boundaries aligned to behavior and ownership, not generic utility buckets.
- Isolate IO adapters, persistence, transport, and external service integration from domain decisions.
- Long-running components should expose explicit lifecycle methods when startup, readiness, or shutdown behavior matters.
- Embedded assets should be grouped under `assets/` or a clearly owned package-local `testdata/` directory.
- Tests should live beside the package behavior they validate.
- Delivery planning, specs, tickets, quiz notes, dumps, and cycle records should live in `ops/`.

## Application Operation Service

`internal/chcapp` owns the application operation service boundary.

The operation service is the shared implementation path for behaviors that can
be invoked from more than one adapter, including the CLI, TUI, and future
Telegram or email adapters. Adapters should translate input, handle transport or
interaction concerns, and map operation results to their own output models.
They should not own reusable workflow logic.

The service implementation may be split across files by capability, such as:

- `service_workspace.go`;
- `service_metadata.go`;
- `service_artifact_ticket.go`;
- `service_work.go`.

That file split does not imply separate service implementations. Prefer one
central operation service type with methods grouped by file over separate
entity services such as ProjectService, TicketService, or AuthoringService.

Adapter-facing wrappers belong in adapter-specific files such as
`tui_adapters.go`. They may depend on adapter result types. Production
`service*.go` files must stay adapter-neutral and must not import
`internal/chctui`.

Narrow interfaces are acceptable when an adapter needs a small contract, but
they should usually be implemented by the central operation service or a thin
adapter wrapper over it.

## Workspace Index

`internal/chc` owns the workspace index model.

The workspace index is a local graph/read-model over project Markdown files. It
is not a database and it must not become the durable source of truth. The
filesystem remains authoritative, and the index must be rebuildable from
workspace files.

The structural index should be deterministic and local. It stores lightweight
facts needed for navigation, relationship traversal, suggestions, diagnostics,
and context selection:

- node kind, slug, key, title, description, status, and case association;
- source path, source hash, and modified time;
- explicit edges such as ownership, reference, and case association;
- lookup maps by kind, path, case slug, and ticket key;
- diagnostics for invalid files or unresolved references.

Document bodies should not be loaded into the structural index by default.
Bodies are read lazily when a detail view, context assembler, writer, validator,
or semantic indexing worker needs them.

Documents such as dumps, quizzes, specs, and plans may be associated with a use
case by canonical case slug. When no use case is declared, the document is
project-wide. Missing case slugs are unresolved references; they do not create
implicit cases.

## Workspace Cache And Semantic Digests

Workspace cache files are discardable local acceleration artifacts. They may
live under a project-local cache path such as:

```text
.chc/user/tmp/cache/workspace-index/
```

Cache files must never be treated as authoritative. Missing, stale, corrupt, or
schema-incompatible cache files should trigger a rebuild from Markdown files.

The semantic index is separate from the structural index. It may use an LLM to
generate compact agent-oriented retrieval hints such as search terms, aliases,
concepts, decisions, constraints, and related slugs. These hints are not
human-facing labels and should not replace source documents.

Semantic digest cache entries must be invalidated by at least:

- digest schema version;
- semantic prompt hash;
- model identifier;
- source hash.

Semantic digest generation must be optional and background-capable. TUI startup
and normal structural refresh must not require network access or block on LLM
calls. Missing or stale semantic digests should appear as index health status,
not as hard failures.

## Document Artifact Model

`Draft` is not a first-class document kind in the active architecture.

Rough or exploratory material belongs in dumps and quizzes. Formalized behavior
belongs in specs. A spec may use `status: draft` when it needs review before
acceptance, but that status does not imply a separate draft artifact family.

The preferred authoring flow is:

```text
dump
quiz
dump + quiz -> spec(status=draft)
spec(status=accepted/final) -> plan
```

## Runtime assembly

- Runtime startup should follow a deterministic assembly order.
- Constructors should store dependencies and lightweight configuration only.
- Failable initialization belongs in explicit setup/start paths, not constructors.
- Readiness and health checks should be visible at package boundaries when the component is operationally significant.
- Shutdown should be explicit and context-aware.

## Change discipline

- Structural changes must be explicit in commit scope.
- Do not introduce additional top-level directories or shared packages without clear ownership need.
- Prefer extending existing Go package boundaries before creating new framework layers.
