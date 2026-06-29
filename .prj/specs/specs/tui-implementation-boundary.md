# TUI Implementation Boundary

Status: initial

Sources:

- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-navigation-and-layout.md`
- `.prj/specs/specs/tui-actions-and-safety.md`
- `.agents/CONVENTIONS.md`
- `.agents/ARCHITECTURE.md`

## Purpose

Define implementation boundaries for adding a Bubble Tea TUI without distorting
the existing CLI/domain architecture.

This is a technical product spec, not a package design finalization.

## Command Entry Point

The expected user command is:

```sh
chr tui
```

The command should resolve the active project the same way normal user-facing
commands do.

Developer-only flags may exist for tests and isolated labs, but the main user
path should not require them.

## Domain Service Reuse

The TUI should reuse the same domain services as the CLI.

It should not parse command output as its normal integration boundary.

The CLI and TUI should share:

- project resolution;
- workspace loading;
- ticket repository/index access;
- authoring session repository;
- quiz generation services;
- runtime validation;
- execution/loop state access;
- report discovery;
- notification status access.

Where services currently exist only behind command handlers, an implementation
slice should extract the minimal domain boundary needed by both CLI and TUI.

## Model Boundaries

The TUI should keep display models separate from persistence models.

Display models may include:

- shortened paths;
- derived status labels;
- merged counts;
- selected-row state;
- view-local filters;
- action availability flags.

Persistence models should remain owned by the existing packages.

The TUI should not write Markdown by string manipulation when a domain writer
already exists.

## Bubble Tea Architecture

The TUI implementation should follow a simple Bubble Tea model:

- root model owns active project and current route;
- views own local selection/filter/input state;
- commands trigger domain operations;
- operation results are delivered as messages;
- long-running operations report progress through messages when available.

The first implementation should prioritize boring state management over clever
composition.

## Package Placement

Expected package direction:

- CLI command wiring remains under existing command packages;
- TUI model/view/update code should live under an internal package;
- shared domain logic should not move into the TUI package;
- rendering helpers should be local to the TUI unless useful elsewhere.

The package layout should follow the existing architecture cleanup rule:
`main.go` remains assembly-only and should not accumulate TUI logic.

## Testing Strategy

Test the TUI in layers:

- pure model update tests for selection, routing, filtering, and action state;
- rendering snapshot/string tests only for stable compact outputs;
- service adapter tests with fakes;
- command wiring smoke test for `chr tui`;
- no fragile terminal-size golden tests unless the output is intentionally
  stable.

Runtime execution should be tested through existing runtime/domain tests, not
duplicated as terminal UI tests.

## First Implementation Slice

The first TUI implementation should be read-only.

Minimum useful scope:

- `chr tui` starts inside a managed project;
- overview screen loads;
- navigation rail works;
- ticket list/detail works;
- authoring session list/detail works;
- runtime status/validation history is visible where available;
- reports list/detail works;
- contextual help works;
- refresh works.

Mutating actions can be visible as disabled actions when useful, but they should
not pretend to work.

## Second Implementation Slice

The next TUI implementation should wire selected low-risk actions:

- runtime validate;
- create ticket;
- update ticket status;
- add ticket comment;
- create authoring session;
- create context-aware quiz;
- show equivalent command;
- refresh/rebuild index where safe.

Ticket execution and loops may follow once
the confirmation and operation record model is in place.

## Acceptance Criteria

- `chr tui` is the only required user entry point for local TUI operation.
- The TUI uses domain services rather than scraping CLI output.
- Display models do not leak into persistence contracts.
- The first TUI can be useful without mutating files.
- TUI code does not add logic to `main.go`.
- Tests cover routing, selection, and action availability without brittle full
  terminal snapshots.
