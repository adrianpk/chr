# Application Operation Service Consolidation Slicing Plan

Status: active

## Purpose

Consolidate adapter-facing operations behind a shared application service
boundary.

The TUI already exposes operations that are useful beyond an interactive
terminal: workspace refresh, Project and Case saves, ticket creation and
mutation, authoring session creation, ticket execution, ready-loop execution,
and daemon stop requests.

Those behaviors should not live as TUI-specific executor implementations.
They should be available through one operational implementation that can be
called from the TUI, CLI, and future adapters such as Telegram or email.

## Source

Primary sources:

- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `internal/chcapp/app.go`
- `internal/chctui/*_actions.go`
- `internal/chctui/workspace_refresh.go`
- `internal/chctui/project_metadata_actions.go`
- `internal/chctui/case_metadata_actions.go`

Conversation decisions:

- Use a central application service boundary for operations.
- Avoid one implementation service per entity.
- Allow several narrow interfaces when useful for adapters.
- Keep the implementation centered in one service type, split across files.
- Keep adapters responsible for UX, transport, confirmation, and rendering.
- Keep business and workflow logic out of `internal/chctui`.
- Future Telegram and email adapters should reuse the same operational path.

## Current State

The current implementation has `internal/chcapp` constructing several
TUI-specific executor structs:

- workspace refresh executor;
- Project metadata executor;
- Case metadata executor;
- artifact action executor;
- ticket action executor;
- work action executor.

Those executors already live in `internal/chcapp`, but they are named and
typed around TUI contracts and return `internal/chctui` result types. This
keeps operational flows coupled to the TUI adapter boundary.

## Target Model

The target model is:

- `internal/chcapp` owns a central operation service implementation.
- The service may be named `Service` if exported across packages, or `service`
  if it stays internal to `internal/chcapp`.
- The service can be split across files by capability:
  - workspace operations;
  - metadata operations;
  - artifact operations;
  - ticket operations;
  - work-control operations.
- The split is by files and methods, not by independent service implementations.
- Narrow interfaces can describe adapter needs, but the central service remains
  the implementation.
- `internal/chcapp` operation request and result types are adapter-neutral.
- `internal/chctui` owns only TUI state, command planning, confirmation,
  keyboard handling, rendering, and mapping between TUI state and operation
  requests.
- The operation service does not import `internal/chctui`.

## Out Of Scope

This slicing does not include:

- adding Telegram or email adapters;
- changing TUI behavior or visual design;
- changing ticket execution semantics;
- replacing the CLI framework;
- removing SQLite from ticket, execution, loop, notification, or daemon
  domains;
- introducing dependency injection frameworks;
- creating separate `ProjectService`, `TicketService`, `AuthoringService`, or
  similar service universes.

## Numbering Rule

Slice numbers are local to this slicing and start at `Slice 1`.

Do not continue from Markdown-only workspace model, TUI Operational Console, or
any previous slicing number.

## Slices

### Slice 1: Operation Service Boundary

Status: planned.

Expected branch: `feat/slice-1-operation-service-boundary`
Expected PR title: `feat(slice-1): operation service boundary`
Expected report: `.prj/reports/slices/application-operation-service-consolidation/slice-1-operation-service-boundary.md`

Goal:

Introduce the central operation service boundary without changing behavior.

Candidate scope:

- add the central service type in `internal/chcapp`;
- centralize shared dependencies needed by operation methods;
- add adapter-neutral request/result types for the first migrated boundary;
- keep existing TUI options and behavior stable;
- delegate at least one low-risk TUI executor path through the service;
- add focused tests for service construction and delegated behavior where
  practical.

Expected outcome:

- future slices have a concrete service boundary to move operations into.

### Slice 2: Workspace And Metadata Operations

Status: planned.

Expected branch: `feat/slice-2-workspace-metadata-operations`
Expected PR title: `feat(slice-2): workspace metadata operations`
Expected report: `.prj/reports/slices/application-operation-service-consolidation/slice-2-workspace-metadata-operations.md`

Goal:

Move workspace refresh, Project save, and Case save behavior into the shared
operation service.

Candidate scope:

- move workspace graph refresh into service methods;
- move Project metadata save into service methods;
- move Case metadata save into service methods;
- keep TUI executor interfaces as adapter wrappers when needed;
- map service results to TUI results outside the service;
- add focused tests for refresh and metadata save operations.

Expected outcome:

- Project, Case, and workspace refresh flows no longer depend on TUI-specific
  executor implementations.

### Slice 3: Artifact And Ticket Operations

Status: planned.

Expected branch: `feat/slice-3-artifact-ticket-operations`
Expected PR title: `feat(slice-3): artifact ticket operations`
Expected report: `.prj/reports/slices/application-operation-service-consolidation/slice-3-artifact-ticket-operations.md`

Goal:

Move artifact creation and ticket mutation behavior into the shared operation
service.

Candidate scope:

- move default ticket creation into service methods;
- move authoring session creation into service methods;
- move selected-ticket comment addition into service methods;
- move selected-ticket status updates into service methods;
- centralize selected ticket lookup and ticket reload paths;
- keep unsupported action errors as adapter-neutral service results or errors;
- add focused tests for artifact and ticket operations.

Expected outcome:

- TUI, CLI, and future adapters can share ticket creation and mutation flows.

### Slice 4: Work Control Operations

Status: planned.

Expected branch: `feat/slice-4-work-control-operations`
Expected PR title: `feat(slice-4): work control operations`
Expected report: `.prj/reports/slices/application-operation-service-consolidation/slice-4-work-control-operations.md`

Goal:

Move execution and daemon work controls into the shared operation service.

Candidate scope:

- move selected-ticket execution into service methods;
- move one-ticket ready-loop execution into service methods;
- move daemon stop requests into service methods;
- centralize observability reload after work actions;
- preserve existing execution failure handling;
- add focused tests for work-control operation outcomes where practical.

Expected outcome:

- work execution and stop flows are no longer TUI-specific.

### Slice 5: Adapter Interface Cleanup

Status: planned.

Expected branch: `feat/slice-5-adapter-interface-cleanup`
Expected PR title: `feat(slice-5): adapter interface cleanup`
Expected report: `.prj/reports/slices/application-operation-service-consolidation/slice-5-adapter-interface-cleanup.md`

Goal:

Clean up adapter-facing interfaces after service migration.

Candidate scope:

- reduce TUI executor wrappers to thin adapter mappings;
- remove obsolete TUI-named operation structs from `internal/chcapp`;
- keep adapter-neutral operation types in `internal/chcapp`;
- ensure operation service files do not import `internal/chctui`;
- organize service methods across capability files without splitting the
  implementation into separate entity services;
- add tests or static checks for the intended import boundary where practical.

Expected outcome:

- the shared service boundary is explicit and adapters do not own operation
  workflows.

### Slice 6: CLI Convergence Pass

Status: planned.

Expected branch: `feat/slice-6-cli-convergence-pass`
Expected PR title: `feat(slice-6): cli convergence pass`
Expected report: `.prj/reports/slices/application-operation-service-consolidation/slice-6-cli-convergence-pass.md`

Goal:

Route obvious overlapping CLI operations through the shared service.

Candidate scope:

- identify CLI command paths that duplicate migrated TUI operation behavior;
- migrate the lowest-risk overlapping commands to service calls;
- preserve current CLI output contracts;
- avoid broad CLI rewrites;
- add focused CLI tests for migrated commands.

Expected outcome:

- the service is proven as a common operation layer, not only a TUI extraction.

### Slice 7: Architecture Documentation

Status: planned.

Expected branch: `docs/slice-7-architecture-documentation`
Expected PR title: `feat(slice-7): architecture documentation`
Expected report: `.prj/reports/slices/application-operation-service-consolidation/slice-7-architecture-documentation.md`

Goal:

Document the operation service boundary and adapter responsibilities.

Candidate scope:

- update workspace architecture documentation with the service boundary;
- document that adapters do not implement use cases;
- document that one central service may be split across files;
- document when narrow interfaces are appropriate;
- add final slicing report context.

Expected outcome:

- future work has durable guidance for keeping operations centralized.

## Validation Strategy

Each implementation slice should run the smallest relevant package tests plus
`make verify` before PR.

Expected common validation:

```text
make verify
```

Targeted validation should include package-level tests for touched packages.

## Delivery Notes

- Plan and tracker are committed directly to `dev` before implementation.
- Each implementation slice uses the expected branch and PR title from this
  plan.
- Each implementation slice creates or updates its own slice report before
  opening a PR.
- Slice PRs target `dev`.
