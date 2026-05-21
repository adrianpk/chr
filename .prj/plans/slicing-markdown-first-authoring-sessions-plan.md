# Markdown-First Authoring Sessions Slicing Plan

Status: delivered

## Purpose

Plan the `Markdown-First Authoring Sessions` slicing.

This slicing productizes the first authoring layer for operational documents:
dumps, quizzes, drafts, specs, plans, slice sets, slices, and tickets. It keeps
Markdown as the canonical editing and review surface while adding enough
structure for `chc` to initialize, track, validate, and promote authoring work.

The goal is to move the current manual conversation-driven workflow into a
durable product-owned workflow without starting with an adaptive model session
or a TUI.

## Source

Primary sources:

- `.prj/drafts/chc-feature-slicing-set-map.md`
- `.prj/drafts/chc-command-surfaces-and-priority.md`
- `.prj/specs/specs/context-composition-pipeline.md`
- `.agents/DOCUMENTING.md`

Prerequisite delivered work:

- `Codex Runtime Worker` slicing.
- `MVP Agent Proving Ground` slicing.
- `Ticket Execution Context Hardening` slicing.
- `docs(authoring): capture post-cutoff roadmap`.
- `refactor(cli): isolate command package`.

## Current State

Delivered:

- project-local `.chc/` initialization;
- Markdown ticket creation, parsing, indexing, status changes, and comments;
- automatic reconciliation for hand-edited ticket Markdown;
- Docker-based Codex ticket execution;
- context composition for ticket execution;
- user documentation for the current implemented workflow.

Remaining gap:

- dumps, quizzes, drafts, specs, plans, and slicing artifacts are still created
  manually in this repository;
- there is no product-owned authoring session identity, status, storage, or
  provenance model;
- there is no command surface for starting a structured questionnaire or
  document promotion;
- there is no shared backend that a later agent-assisted session or TUI can
  reuse.

## Product Boundary

This slicing supports one local operator and one active agent per project. It
does not introduce multi-user collaboration or concurrent agents editing the
same repository.

Authoring prepares project knowledge and executable tickets. Ticket execution
remains the responsibility of the existing Docker worker path.

## Authoring Paths

This slicing implements the base path and the first generated path:

1. Markdown direct.
   Users can create or edit artifacts by hand. Commands must tolerate those
   files and preserve Markdown as the durable source.
2. Static generated authoring.
   `chc` creates structured Markdown questionnaires, drafts, specs, plans, or
   slice-set skeletons that the user edits manually.

Later slicings should build on the same model:

- agent-assisted authoring sessions backed by Codex or another model;
- a Bubble Tea TUI over the same authoring-session backend.

## Execution Model

The authoring session runner is host-side coordination.

`chc` owns:

- authoring session identity and status;
- target document type and topic;
- input and output artifact paths;
- Markdown templates;
- provenance links;
- validation and promotion rules.

The Docker ticket worker owns:

- resolving executable ticket work;
- mutating code during ticket execution;
- producing execution result artifacts.

This slicing does not call Codex for adaptive questioning. It creates the
session and artifact foundation required before assisted sessions are added.

## Artifact Model

The planned model should support artifacts under project-local operational
paths, including:

```text
.chc/authoring/sessions/<session-id>/
  session.md
  transcript.md
  artifacts.md
```

Generated or promoted project artifacts remain in their canonical locations,
for example:

```text
.chc/dumps/
.chc/quizzes/
.chc/drafts/
.chc/specs/
.chc/plans/
.chc/trackers/
.chc/tickets/
```

Exact path names may be refined by implementation slices, but the invariant is
that session evidence and generated Markdown are durable, inspectable, and
editable.

## Slicing Scope

This slicing includes:

- authoring session domain model;
- authoring session storage contract;
- artifact inventory and canonical path resolution;
- Markdown templates for static questionnaires and document skeletons;
- static quiz/questionnaire generation;
- CLI commands for starting, listing, showing, and finalizing authoring
  sessions;
- manual promotion with provenance;
- user-facing reference documentation for the Markdown-first authoring
  workflow.

## Out of Scope

This slicing does not include:

- adaptive LLM questioning;
- Codex-backed authoring sessions;
- Bubble Tea TUI;
- Telegram or email authoring surfaces;
- daemon-owned authoring;
- ticket execution changes;
- concurrent authoring agents;
- automatic conversion from every document type to every other document type.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Authoring Session Model

Status: delivered.

Expected branch: `feat/slice-1-authoring-session-model`
Expected PR title: `feat(slice-1): authoring session model`
Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-1-authoring-session-model.md`

Goal:

Define the side-effect-free model for authoring sessions.

Candidate scope:

- add authoring session ID, status, mode, target, topic, and timestamps;
- define input and output artifact references;
- define provenance links;
- define validation errors for invalid session inputs;
- add focused model tests.

Expected outcome:

- later slices can reason about authoring sessions without touching the
  filesystem.

### Slice 2: Authoring Session Storage

Status: delivered.

Expected branch: `feat/slice-2-authoring-session-storage`
Expected PR title: `feat(slice-2): authoring session storage`
Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-2-authoring-session-storage.md`

Goal:

Persist authoring sessions as project-local Markdown artifacts.

Candidate scope:

- define the session directory layout;
- render and parse `session.md`;
- record transcript and artifact index paths;
- preserve hand-edited Markdown where possible;
- add repository tests.

Expected outcome:

- sessions can be created, loaded, listed, and inspected as durable project
  files.

### Slice 3: Authoring Artifact Layout

Status: delivered.

Expected branch: `feat/slice-3-authoring-artifact-layout`
Expected PR title: `feat(slice-3): authoring artifact layout`
Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-3-authoring-artifact-layout.md`

Goal:

Define canonical authoring artifact paths and inventory behavior.

Candidate scope:

- model artifact kinds for dumps, quizzes, drafts, specs, plans, slice sets,
  slices, trackers, and tickets;
- resolve project-local output paths;
- detect existing artifacts and collision behavior;
- render artifact inventory entries;
- add path and inventory tests.

Expected outcome:

- generated and promoted artifacts have predictable destinations and provenance.

### Slice 4: Static Quiz Generation

Status: delivered.

Expected branch: `feat/slice-4-static-quiz-generation`
Expected PR title: `feat(slice-4): static quiz generation`
Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-4-static-quiz-generation.md`

Goal:

Generate a static Markdown questionnaire for a selected authoring target.

Candidate scope:

- define target-specific question templates;
- generate editable Markdown quiz artifacts;
- link the generated quiz to its authoring session;
- support dry-run rendering;
- add deterministic rendering tests.

Expected outcome:

- the user can start with a useful questionnaire without an interactive model
  session.

### Slice 5: Authoring Session CLI

Status: delivered.

Expected branch: `feat/slice-5-authoring-session-cli`
Expected PR title: `feat(slice-5): authoring session cli`
Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-5-authoring-session-cli.md`

Goal:

Expose the Markdown-first authoring workflow through the CLI.

Candidate scope:

- add `author session new`;
- add `author session list`;
- add `author session show`;
- add `author quiz new` or equivalent target-specific entrypoint;
- wire commands through the existing `internal/chcapp` command package;
- add CLI tests.

Expected outcome:

- authoring sessions can be started and inspected without manually creating all
  session files.

### Slice 6: Manual Promotion and Provenance

Status: delivered.

Expected branch: `feat/slice-6-manual-promotion-provenance`
Expected PR title: `feat(slice-6): manual promotion provenance`
Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-6-manual-promotion-provenance.md`

Goal:

Promote selected Markdown authoring inputs into the next durable artifact while
preserving provenance.

Candidate scope:

- support one conservative promotion path, such as quiz to draft or draft to
  spec;
- record source session and input artifact paths;
- generate an editable target artifact;
- avoid overwriting existing artifacts without an explicit flag;
- add promotion tests.

Expected outcome:

- authoring can produce a durable next-step artifact without depending on
  conversation memory.

### Slice 7: Authoring Docs Reference

Status: delivered.

Expected branch: `feat/slice-7-authoring-docs-reference`
Expected PR title: `feat(slice-7): authoring docs reference`
Expected report: `.prj/reports/slices/markdown-first-authoring-sessions/slice-7-authoring-docs-reference.md`

Goal:

Document the Markdown-first authoring contract and operator workflow.

Candidate scope:

- add reference documentation for session fields and artifact layout;
- add user-guide coverage for static quiz generation and promotion;
- document hand-edited Markdown tolerance;
- document future assisted-session and TUI compatibility boundaries;
- add slice report.

Expected outcome:

- users can understand the current authoring workflow and how it relates to
  later assisted sessions.

## Open Questions

- Which first promotion path is most useful: quiz to draft, dump plus quiz to
  spec, or draft to spec?
- Should authoring artifacts live only under `.chc/`, or should repository
  operational workspaces such as `.prj/` remain supported as an advanced mode?
- Which metadata fields are required for every authoring artifact versus only
  for generated artifacts?
- Should static question templates be embedded in Go code, stored as project
  assets, or both?
