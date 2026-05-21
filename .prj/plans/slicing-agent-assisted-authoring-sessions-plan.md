# Agent-Assisted Authoring Sessions Slicing Plan

Status: delivered

## Purpose

Plan the `Agent-Assisted Authoring Sessions` slicing.

This slicing adds a model-backed authoring loop on top of the Markdown-first
session backend. The goal is to let `chc` conduct bounded discovery sessions
for quizzes, dumps, drafts, specs, plans, slice sets, and tickets while
persisting transcript, curated answers, generated artifacts, and provenance.

This is the product-owned version of the current manual conversation workflow:
the agent asks focused follow-up questions, records answers, curates the result,
and writes Markdown artifacts that the user can review and edit.

## Source

Primary sources:

- `.prj/drafts/chc-feature-slicing-set-map.md`
- `.prj/drafts/chc-command-surfaces-and-priority.md`
- `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
- `docs/reference/authoring-sessions.md`
- `docs/reference/ticket-execution.md`

Prerequisite delivered work:

- `Markdown-First Authoring Sessions` slicing.
- `Ticket Execution Context Hardening` slicing.
- `Codex Runtime Worker` slicing.

## Current State

Delivered:

- authoring session model;
- Markdown session storage;
- authoring artifact layout;
- static quiz generation;
- authoring session CLI;
- quiz-to-draft promotion;
- authoring reference documentation.

Remaining gap:

- static quizzes ask all questions at once;
- no host-side model runner exists for authoring;
- transcripts are placeholder files, not structured session evidence;
- answers are not curated into reusable summaries;
- generated artifacts are not synthesized by a bounded assistant workflow;
- assisted authoring has no pause, resume, or finalize command shape.

## Product Boundary

Agent-assisted authoring prepares documents. It does not execute tickets, edit
code, run the Docker ticket worker, or create concurrent agents.

The operating model remains one local operator and one active authoring session
or execution workflow per project. A later daemon may initiate or resume
authoring, but this slicing keeps the first implementation CLI-driven.

## Execution Model

The assisted authoring runner is host-side coordination.

`chc` owns:

- session state and transcript storage;
- authoring target and topic;
- context package for the authoring agent;
- model invocation request and response recording;
- turn limits and follow-up depth;
- answer curation;
- generated Markdown artifacts;
- provenance links.

The model owns:

- proposing the next question;
- deciding when enough information has been collected;
- drafting or refining Markdown artifact content.

The Docker ticket worker owns:

- resolving executable ticket work;
- mutating code;
- producing execution results.

The authoring runner must not reuse conversation history unless that history has
been materialized into session files.

## Slicing Scope

This slicing includes:

- assisted authoring session contracts;
- transcript event model and persistence;
- authoring context package construction;
- model invocation adapter boundary;
- adaptive quiz loop;
- answer curation and session summary;
- artifact synthesis into Markdown;
- CLI commands for start, answer, resume, and finalize;
- docs for the assisted authoring workflow.

## Out of Scope

This slicing does not include:

- Bubble Tea TUI;
- Telegram or email authoring surfaces;
- daemon-owned authoring loops;
- ticket execution changes;
- Docker worker changes;
- multi-agent authoring;
- multi-user collaboration;
- automatic promotion for every possible document transition.

## Numbering Rule

Slice numbers below are local to this slicing. They start at `Slice 1` for this
plan, and are the numbers used in branch names, PR titles, trackers, reports,
and delivery notes for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Assisted Session Contract

Status: delivered.

Expected branch: `feat/slice-1-assisted-session-contract`
Expected PR title: `feat(slice-1): assisted session contract`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-1-assisted-session-contract.md`

Goal:

Define side-effect-free contracts for assisted authoring turns, intents, states,
and limits.

Candidate scope:

- add assisted session configuration;
- define turn state and turn result models;
- define user answer and assistant question models;
- define completion and blocked reasons;
- add focused model validation tests.

Expected outcome:

- later slices can model an assisted authoring loop without invoking a model.

### Slice 2: Transcript Event Storage

Status: delivered.

Expected branch: `feat/slice-2-assisted-transcript-storage`
Expected PR title: `feat(slice-2): assisted transcript storage`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-2-assisted-transcript-storage.md`

Goal:

Persist assisted authoring transcript events into the existing session layout.

Candidate scope:

- define transcript event types;
- render transcript events to Markdown;
- parse stored transcript events;
- append events without rewriting unrelated session artifacts;
- add repository tests.

Expected outcome:

- every assisted turn leaves durable Markdown evidence.

### Slice 3: Authoring Context Package

Status: delivered.

Expected branch: `feat/slice-3-assisted-context-package`
Expected PR title: `feat(slice-3): assisted context package`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-3-assisted-context-package.md`

Goal:

Build the host-side context consumed by the authoring model.

Candidate scope:

- include agent identity for authoring analyst behavior;
- include project and authoring session metadata;
- include relevant `.chc/` artifacts and selected docs;
- include transcript summary and unanswered gaps;
- keep context deterministic and materialized as Markdown.

Expected outcome:

- model turns can be grounded in explicit session and project context.

### Slice 4: Model Invocation Adapter

Status: delivered.

Expected branch: `feat/slice-4-assisted-model-adapter`
Expected PR title: `feat(slice-4): assisted model adapter`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-4-assisted-model-adapter.md`

Goal:

Define and implement the first host-side adapter boundary for model-backed
authoring.

Candidate scope:

- define model request and response envelopes;
- support dry-run or fake adapter for tests;
- record request and response artifacts;
- keep provider details behind an interface;
- classify adapter failures.

Expected outcome:

- assisted authoring can call a model through a replaceable boundary.

### Slice 5: Adaptive Quiz Loop

Status: delivered.

Expected branch: `feat/slice-5-assisted-quiz-loop`
Expected PR title: `feat(slice-5): assisted quiz loop`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-5-assisted-quiz-loop.md`

Goal:

Run a bounded adaptive question-and-answer loop for one authoring session.

Candidate scope:

- ask the next question;
- accept one user answer;
- allow limited follow-up depth;
- stop on completion, user stop, blocked state, or turn limit;
- persist each turn.

Expected outcome:

- `chc` can conduct a durable assisted quiz rather than generating all
  questions upfront.

### Slice 6: Answer Curation

Status: delivered.

Expected branch: `feat/slice-6-assisted-answer-curation`
Expected PR title: `feat(slice-6): assisted answer curation`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-6-assisted-answer-curation.md`

Goal:

Curate transcript answers into a stable session summary.

Candidate scope:

- generate curated answers from transcript events;
- preserve explicit decisions and open questions;
- write a curated summary artifact;
- update session provenance;
- add deterministic tests with fake model output.

Expected outcome:

- later artifact synthesis can consume a compact durable summary instead of the
  raw transcript only.

### Slice 7: Assisted Artifact Synthesis

Status: delivered.

Expected branch: `feat/slice-7-assisted-artifact-synthesis`
Expected PR title: `feat(slice-7): assisted artifact synthesis`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-7-assisted-artifact-synthesis.md`

Goal:

Generate a Markdown artifact from curated assisted authoring output.

Candidate scope:

- synthesize one conservative artifact path, such as curated quiz to draft or
  draft to spec;
- preserve source session and summary provenance;
- write an editable target artifact;
- avoid overwriting without an explicit flag;
- add synthesis tests.

Expected outcome:

- assisted authoring can produce a durable Markdown artifact, not just a
  transcript.

### Slice 8: Assisted Authoring CLI

Status: delivered.

Expected branch: `feat/slice-8-assisted-authoring-cli`
Expected PR title: `feat(slice-8): assisted authoring cli`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-8-assisted-authoring-cli.md`

Goal:

Expose assisted authoring through CLI commands.

Candidate scope:

- add assisted session start;
- add answer submission;
- add resume/status display;
- add finalize command;
- include dry-run and fake-adapter test paths;
- add CLI tests.

Expected outcome:

- a user can run a bounded assisted authoring session from the terminal.

### Slice 9: Assisted Authoring Docs

Status: delivered.

Expected branch: `feat/slice-9-assisted-authoring-docs`
Expected PR title: `feat(slice-9): assisted authoring docs`
Expected report: `.prj/reports/slices/agent-assisted-authoring-sessions/slice-9-assisted-authoring-docs.md`

Goal:

Document assisted authoring contracts and operator workflow.

Candidate scope:

- update authoring reference;
- update user guide;
- document transcript, context, and generated artifact paths;
- document model-boundary and persistence guarantees;
- add slice report.

Expected outcome:

- users can understand when to use static authoring versus assisted authoring.

## Open Questions

- Which provider adapter should be first-class in the initial implementation:
  Codex CLI, OpenAI API, or a generic command adapter?
- Should assisted sessions be fully synchronous in the CLI, or should pause and
  resume be the default interaction model?
- Which artifact synthesis path should be delivered first?
- How much project context should be included before the authoring context
  package becomes too noisy?
