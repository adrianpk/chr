# Markdown-Only Workspace Model Slicing Plan

Status: active

## Purpose

Plan the transition from duplicated relational persistence toward a
Markdown-only workspace model.

Markdown files under `.chc/` are the source of truth. Runtime code may build an
ephemeral in-memory workspace graph for navigation, diagnostics, and context
assembly, but it must not persist a second authoritative representation in
SQLite for `Project` or `Cases`.

## Source

Primary sources:

- `docs/explanation/tui-interactive-design-notes.md`
- `.prj/specs/specs/tui-project-tab.md`
- `.prj/specs/specs/tui-cases-screen.md`

Conversation decisions:

- Markdown is the only durable source of truth.
- SQLite should not store project/case metadata as a duplicated model.
- The runtime may keep a lightweight mental map in memory.
- The in-memory map stores navigation metadata, relationships, parse status,
  and paths, not full document context bodies.
- LLM context assembly should start from a case when available, or from
  project-wide material when no case applies.
- Manual filesystem edits are valid. The app should rescan/reload from files
  instead of reconciling database state.

## Current State

The current implementation includes:

- `.chc/project.md` as Markdown project metadata;
- `.chc/cases/*.md` as Markdown case metadata;
- SQLite-backed indexed project metadata;
- SQLite-backed indexed case metadata;
- TUI executors that write Markdown and then update relational indexes;
- TUI reload/resync language inherited from the relational index model.

This works, but it keeps two representations for entities whose durable truth
should be the Markdown file.

## Target Model

The target model is:

- Markdown files are the durable source of truth.
- SQLite is not used for `Project` or `Cases`.
- The app scans `.chc/` into an ephemeral `WorkspaceGraph`.
- The graph stores only lightweight metadata:
  - path;
  - kind;
  - slug/key/title/name;
  - status;
  - case association;
  - source references;
  - modified time/hash when useful;
  - parse/diagnostic state.
- Document bodies are read lazily:
  - when a detail view opens;
  - when an editor writes;
  - when context is assembled for an LLM;
  - when validation needs body content.
- User edits through Vim or another editor are accepted as normal filesystem
  changes.
- `r` or equivalent refresh behavior rebuilds the in-memory graph from
  Markdown.

## Navigation Model

The normal operational navigation starts at `Cases`.

```text
Project
  Project-wide
    Dumps
    Quizzes
    Specs
    Plans

  Cases
    <case>
      Dumps
      Quizzes
      Specs
      Plans
```

A case is defined by a canonical case document:

```text
.chc/cases/<slug>.md
```

Other artifact documents may reference the case by slug in frontmatter.

An artifact referencing a non-existent case should not silently create a case.
It should appear as an unresolved case reference in the workspace graph and TUI.

## Out Of Scope

This slicing does not include:

- changing the physical `.chc/` directory layout;
- introducing a daemon or mandatory filesystem watcher;
- adding embeddings or semantic search;
- implementing a full text editor in the TUI;
- removing SQLite from unrelated runtime domains such as tickets, executions,
  loops, notifications, or remote commands;
- changing the ticket execution model;
- changing release/versioning workflow.

## Numbering Rule

Slice numbers are local to this slicing and start at `Slice 1`.

Do not continue from TUI, project metadata, or previous workspace slicing
numbers.

## Slices

### Slice 1: Markdown-Only Contract

Status: delivered.

Expected branch: `feat/slice-1-markdown-only-contract`
Expected PR title: `feat(slice-1): markdown-only contract`
Expected report: `.prj/reports/slices/markdown-only-workspace-model/slice-1-markdown-only-contract.md`

Goal:

Define the product and technical contract before changing runtime behavior.

Candidate scope:

- update Project and Cases specs to declare Markdown as source of truth;
- remove database/index terminology from Project and Cases specs;
- define the ephemeral workspace graph concept;
- define case definition vs case association;
- define unresolved case references;
- define manual filesystem edit behavior;
- define context assembly direction from case/project-wide roots;
- update design notes to match the simplified model.

Expected outcome:

- future implementation slices have a durable contract that removes ambiguity
  around SQLite, resync, and source of truth.

### Slice 2: Remove Project And Case Indexes

Status: delivered.

Expected branch: `feat/slice-2-remove-project-case-indexes`
Expected PR title: `feat(slice-2): remove project case indexes`
Expected report: `.prj/reports/slices/markdown-only-workspace-model/slice-2-remove-project-case-indexes.md`

Goal:

Remove relational persistence for Project and Cases.

Candidate scope:

- delete project metadata repository code;
- delete case metadata repository code;
- stop creating/updating indexed project/case tables;
- remove project/case metadata migrations if safe for unreleased state, or
  leave historical migrations inert only if migration compatibility requires
  it;
- simplify TUI executors so save writes Markdown only;
- keep file parsing/rendering services;
- update tests to validate file-only behavior.

Expected outcome:

- Project and Cases read/write Markdown only, with no SQLite indexing path.

### Slice 3: Workspace Graph Builder

Status: active.

Expected branch: `feat/slice-3-workspace-graph-builder`
Expected PR title: `feat(slice-3): workspace graph builder`
Expected report: `.prj/reports/slices/markdown-only-workspace-model/slice-3-workspace-graph-builder.md`

Goal:

Introduce the in-memory graph used for navigation and context selection.

Candidate scope:

- add a graph package or cohesive internal service;
- scan `.chc/` Markdown files;
- parse frontmatter without loading bodies as primary graph state;
- classify project, case, dump, quiz, spec, plan, ticket, and report where
  supported;
- build case-rooted and project-wide views;
- track unresolved case references;
- track invalid frontmatter diagnostics;
- add tests with fixture workspaces.

Expected outcome:

- the app can reconstruct a lightweight workspace map from Markdown at runtime.

### Slice 4: Filesystem Refresh Semantics

Status: planned.

Expected branch: `feat/slice-4-filesystem-refresh-semantics`
Expected PR title: `feat(slice-4): filesystem refresh semantics`
Expected report: `.prj/reports/slices/markdown-only-workspace-model/slice-4-filesystem-refresh-semantics.md`

Goal:

Make direct filesystem edits a first-class workflow.

Candidate scope:

- define refresh/reload command behavior;
- implement graph rebuild on refresh;
- update changed graph nodes after app writes;
- surface invalid files as diagnostics;
- avoid stale relational terminology;
- decide and implement minimal external-edit protection for active edits, using
  file hash or modified time only as runtime guardrails;
- add tests for external edits, invalid frontmatter, and reload behavior.

Expected outcome:

- edits made through Vim or another editor are visible after refresh without
  database reconciliation.

### Slice 5: TUI Project And Cases Rewire

Status: planned.

Expected branch: `feat/slice-5-tui-project-cases-rewire`
Expected PR title: `feat(slice-5): tui project cases rewire`
Expected report: `.prj/reports/slices/markdown-only-workspace-model/slice-5-tui-project-cases-rewire.md`

Goal:

Point the TUI Project and Cases screens at Markdown services and the workspace
graph.

Candidate scope:

- load Project view from Markdown service;
- load Cases list from workspace graph;
- show case documents and associated artifacts through graph views;
- show unresolved case references in Cases;
- preserve inline editing for project and case documents;
- change `r` behavior to refresh/rebuild graph from Markdown;
- remove index/resync database language from the UI;
- add focused TUI model/render tests.

Expected outcome:

- Project and Cases operate as Markdown-backed screens with graph-assisted
  navigation.

### Slice 6: Context Assembly From Graph

Status: planned.

Expected branch: `feat/slice-6-context-assembly-from-graph`
Expected PR title: `feat(slice-6): context assembly from graph`
Expected report: `.prj/reports/slices/markdown-only-workspace-model/slice-6-context-assembly-from-graph.md`

Goal:

Use the workspace graph to collect LLM context from case or project-wide roots.

Candidate scope:

- add context request model for case-rooted and project-wide assembly;
- collect selected case document;
- collect case-scoped dumps, quizzes, specs, and plans;
- collect project-wide material when no case applies;
- read document bodies lazily during assembly;
- report missing/unresolved references;
- add tests for context selection and ordering.

Expected outcome:

- context assembly can start from a selected case and gather the relevant
  Markdown documents without SQL.

## Validation Strategy

Each slice should run the smallest relevant checks plus `make verify` before PR.

Expected common validation:

```text
make verify
```

Targeted validation should include package-level tests for touched packages.

## Delivery Notes

- Plan and tracker should be committed directly to `dev` before implementation.
- Each slice should use the expected branch and PR title from this plan.
- Each implementation slice should create or update its own slice report before
  opening a PR.
- Slice PRs target `dev`.
