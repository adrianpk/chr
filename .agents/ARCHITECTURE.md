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
