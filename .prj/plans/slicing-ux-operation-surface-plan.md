# UX Operation Surface Slicing Plan

Status: planned

## Purpose

Plan the `UX Operation Surface` slicing created from full-system lab feedback.

This slicing covers the use case:

> A user can operate ChrysoCoord efficiently from a readable CLI and a local TUI
> without memorizing opaque IDs, copying absolute paths, or configuring internal
> invariants.

## Source

Primary sources:

- `.prj/drafts/chc-cli-output-polish-notes.md`
- `.prj/tickets/open/20260522071859-align-cli-authoring-and-tui-operation-surface.md`
- `docs/how-to/full-system-lab-runbook.md`

Reference context:

- `docs/how-to/user-guide/index.md`
- `docs/reference/authoring-sessions.md`
- `docs/reference/project-workspace.md`
- `docs/reference/runtime-config.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-loop.md`

## Current State

Delivered foundations:

- project-local `.chc/` workspace initialization;
- Markdown-first authoring sessions, quiz promotion, and assisted
  authoring artifacts;
- ticket creation, listing, status changes, execution, loop, result
  application, and reports;
- runtime validation and Docker execution boundary;
- notification, remote command intake, reporting, and maintenance commands;
- full-system lab fixture and runbook.

Observed gaps:

- CLI output is overly verbose for humans;
- list output lacks headers;
- authoring session IDs are opaque;
- useful quizzes require model-generated questions from project context;
- bootstrap config/manifest expose or repeat internal invariants;
- protocol docs blur project documentation conventions and operational artifact
  contracts;
- no TUI specification exists for daily operation.

## Slicing Scope

This slicing includes:

- readable CLI output defaults;
- consistent relative-path output for human mode;
- table/header conventions for list commands;
- authoring session naming and lookup ergonomics;
- workspace config and manifest simplification;
- protocol doc boundary cleanup;
- context-aware quiz generation foundation;
- detailed Bubble Tea TUI specifications;
- a first read-only TUI shell and navigation model;
- selected TUI actions wired through existing command/domain operations.

## Out of Scope

This slicing does not include:

- a browser/web app;
- multi-user workflows;
- multi-agent concurrent repository mutation;
- replacing the CLI;
- implementing every possible TUI action in the first pass;
- live Telegram or email changes;
- changing the `.chc/` canonical root;
- broad runtime worker redesign beyond UX-facing contracts.

## Design Direction

The CLI remains the executable foundation. Every operation that needs to be
exposed should remain available through commands.

The TUI becomes the normal human workflow surface:

- records are selected, not copied;
- lists are navigated, not memorized;
- actions run against selected entities;
- details and evidence are visible without chasing paths;
- all state remains project-local and auditable under `.chc/`.

Configuration should expose only meaningful user decisions. Invariants such as
`.chc/`, Markdown as canonical artifact format, runtime artifact layout, and
internal workspace paths should be owned by the binary.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: CLI Output Readability

Status: delivered.

Expected branch: `feat/slice-1-cli-output-readability`
Expected PR title: `feat(slice-1): cli output readability`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-1-cli-output-readability.md`

Goal:

Make common command output readable without losing scriptability.

Candidate scope:

- introduce output formatting helpers for human list/detail output;
- prefer project-relative paths in human output;
- add headers to list commands where output has multiple columns;
- remove duplicated status words such as `check: ok ... ok`;
- make authoring quiz/session creation output compact;
- keep structured/full detail available through existing show commands or a
  future `--json`/`--verbose` path;
- add focused CLI tests.

Expected outcome:

- lab output is readable enough that users can identify IDs, statuses, targets,
  paths, and next actions without parsing walls of absolute paths.

### Slice 2: Workspace Bootstrap Contract Cleanup

Status: delivered.

Expected branch: `feat/slice-2-workspace-bootstrap-contract`
Expected PR title: `feat(slice-2): workspace bootstrap contract cleanup`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-2-workspace-bootstrap-contract.md`

Goal:

Remove pseudo-configurable internal invariants from generated workspace
surfaces.

Candidate scope:

- review `.chc/config.yml` generated fields;
- keep only meaningful user decisions in config;
- treat `.chc/`, Markdown canonical artifacts, runtime workdir, and runtime
  artifact layout as binary-owned invariants unless a real user choice exists;
- rewrite or remove `.chc/manifest.md` as an authoritative-looking config
  surface;
- move workspace orientation into a human guide if it remains useful;
- add migration/compatibility handling for existing generated config fields;
- update docs and tests.

Expected outcome:

- a new managed project does not ask the user to configure internals that should
  be product invariants.

### Slice 3: Protocol Document Boundary Cleanup

Status: delivered.

Expected branch: `feat/slice-3-protocol-doc-boundaries`
Expected PR title: `feat(slice-3): protocol doc boundaries`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-3-protocol-doc-boundaries.md`

Goal:

Separate project documentation guidance from ChrysoCoord operational artifact
contracts.

Candidate scope:

- review generated `.chc/agents/*.md` files;
- keep `DOCUMENTING.md` focused on project durable documentation;
- remove `.chc/` operational layout rules from project documentation guidance;
- move ChrysoCoord artifact layout explanation into docs/reference or a clear
  workspace guide;
- ensure operational artifact layout is enforced by code, not convention text;
- update context composition if manifest/protocol sources change;
- add tests for generated assets.

Expected outcome:

- generated protocol docs guide project work without implying that the project
  can redefine ChrysoCoord's internal artifact contracts.

### Slice 4: Authoring Session Ergonomics

Status: delivered.

Expected branch: `feat/slice-4-authoring-session-ergonomics`
Expected PR title: `feat(slice-4): authoring session ergonomics`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-4-authoring-session-ergonomics.md`

Goal:

Make authoring sessions understandable and discoverable without memorizing
opaque IDs.

Candidate scope:

- evaluate replacing `auth-` prefix with a less ambiguous prefix for new
  sessions;
- add or expose session slug/topic lookup where safe;
- improve `author session list` output with headers and updated timestamps;
- improve `author session show` as the canonical detailed view;
- ensure promote/resume/finalize commands can resolve from artifact metadata
  where possible;
- preserve backward compatibility for existing `auth-...` sessions;
- add tests.

Expected outcome:

- users can recover the right authoring session from topic/status/target and
  proceed without treating IDs as the main mental model.

### Slice 5: Context-Aware Quiz Generation Foundation

Status: delivered.

Expected branch: `feat/slice-5-context-aware-quiz-generation`
Expected PR title: `feat(slice-5): context-aware quiz generation`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-5-context-aware-quiz-generation.md`

Goal:

Add a ChrysoCoord-native question generation path that can produce useful
questions for a target artifact.

Candidate scope:

- define a quiz generation request model with project context, target artifact,
  topic, prior answers, and output contract;
- define a provider interface and Codex-backed implementation;
- add an initial fake provider for tests;
- make `spec` target questions cover problem, actors, workflows, acceptance
  criteria, constraints, non-goals, validation evidence, impacted code/docs,
  and open risks;
- do not claim requested `--question-count` was satisfied when fewer questions
  are generated;
- add docs for the structured request/response contract.

Expected outcome:

- `chc` has a clear internal boundary for useful quiz generation, even before
  every provider integration is complete.

### Slice 6: TUI Specification Set

Status: delivered.

Expected branch: `docs/slice-6-tui-specification-set`
Expected PR title: `docs(slice-6): tui specification set`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-6-tui-specification-set.md`

Goal:

Create granular UX/UI specifications for a Bubble Tea TUI before implementation.

Candidate scope:

- specify the TUI product role and scope;
- specify navigation, layout, keybindings, and command palette behavior;
- specify project, ticket, authoring, quiz, report, runtime, and maintenance
  screens;
- specify action confirmation and destructive-action rules;
- specify status, loading, error, and empty states;
- specify CLI/TUI contract boundaries;
- specify accessibility, terminal width behavior, and no-color fallback.

Expected outcome:

- TUI implementation can proceed from concrete UX/UI specifications rather than
  from vague "add a TUI" intent.

### Slice 7: TUI Read-Only Shell

Status: planned.

Expected branch: `feat/slice-7-tui-read-only-shell`
Expected PR title: `feat(slice-7): tui read-only shell`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-7-tui-read-only-shell.md`

Goal:

Add the first Bubble Tea TUI surface with read-only navigation over existing
state.

Candidate scope:

- add `chc tui` command;
- add project resolution and current project display;
- add navigation frame/sidebar or equivalent;
- list tickets, authoring sessions, reports, and runtime status;
- show selected record details;
- open file paths through an explicit key/action where appropriate;
- keep all views read-only in this slice;
- add tests for model/state functions where practical.

Expected outcome:

- users can inspect core ChrysoCoord state without copying IDs between CLI
  commands.

### Slice 8: TUI Action Wiring

Status: planned.

Expected branch: `feat/slice-8-tui-action-wiring`
Expected PR title: `feat(slice-8): tui action wiring`
Expected report: `.prj/reports/slices/ux-operation-surface/slice-8-tui-action-wiring.md`

Goal:

Wire safe common actions into the TUI through existing command/domain
operations.

Candidate scope:

- run runtime validation;
- create and update tickets;
- start a ticket dry-run;
- open authoring session actions;
- generate or continue quiz flows where supported;
- show reports and execution evidence;
- require confirmation for mutating or long-running operations;
- preserve CLI parity and auditability.

Expected outcome:

- the TUI becomes a usable local cockpit for common work while the CLI remains
  the automation surface.
