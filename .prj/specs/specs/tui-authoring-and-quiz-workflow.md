# TUI Authoring and Quiz Workflow

Status: initial

Sources:

- `.prj/specs/specs/context-composition-pipeline.md`
- `.prj/specs/specs/tui-product-surface.md`
- `.prj/drafts/chc-cli-output-polish-notes.md`
- `docs/explanation/document-generation-model.md`
- `docs/reference/authoring-sessions.md`

## Purpose

Define how the TUI should support document authoring, quiz-driven knowledge
extraction, and promotion between artifacts.

The TUI should make authoring workflows navigable without hiding the underlying
Markdown artifacts.

## Authoring Mental Model

An authoring session is a durable workspace for producing or refining a target
artifact.

The user should understand a session by:

- topic;
- target artifact type;
- status;
- recent activity;
- source artifacts;
- generated outputs.

The session ID is a stable technical identifier. It should not be the only
human-recognizable handle.

## Session Lifecycle

The TUI should expose these session states:

- draft: session exists and can receive content;
- active: user or agent is adding content;
- ready: enough material exists for promotion;
- promoted: output artifact has been created;
- blocked: session needs user input or has a failed generation;
- archived: session is no longer active but remains inspectable.

The exact persisted values may differ in early implementation. The UI should
still map existing states into clear human labels.

## Session Creation

Creating a session in the TUI should ask for:

- target artifact type;
- topic;
- optional source artifact;
- optional mode or workflow when more than one is available.

The initial form should avoid developer-only fields. It should not ask for
paths that can be derived from the active project.

After creation, the TUI should show the session detail and the next useful
actions.

## Context-Aware Quiz Mode

Context-aware quiz mode generates questions from the project context, authoring
session, target artifact, topic, and a structured model response contract.

It is useful when:

- the user wants the system to extract project-specific knowledge;
- the target artifact needs sharper inputs than a generic checklist can provide;
- the available project docs, tickets, code inventory, and `.chc/` artifacts can
  shape better questions.

The TUI must not present generic template questions as context-aware
interrogation.

If the user requests eight questions and the model returns fewer valid
questions, the UI should state the requested and generated counts separately.

## Context-Aware Quiz Mode

Context-aware quiz mode should use project and session context to produce useful
questions for the target artifact.

Inputs:

- active project identity;
- target artifact type;
- topic;
- source dump, spec, plan, tracker, ticket, or selected document;
- current session transcript;
- previous answers;
- relevant `.chc/` artifacts;
- relevant docs;
- file tree summary;
- optional Go code context summaries.

Outputs:

- candidate questions;
- selected questions;
- answered entries;
- curated notes;
- unresolved issues;
- promotion-ready material.

The TUI should let the user run:

- one-question-at-a-time mode for deep exploration;
- batch mode for a bounded set of questions;
- continue mode for follow-up questions derived from prior answers.

Each mode should persist the generated questions and user answers.

## Quiz Screen

The quiz screen should show:

- session topic;
- target artifact;
- mode;
- current question number;
- maximum question threshold;
- answered count;
- open questions;
- previous answer summary;
- save/persist state.

For one-question-at-a-time mode, the answer input should be focused and the
next action should be explicit.

For batch mode, questions should be shown as a list with answer fields or as a
sequence that can be resumed.

## Promotion

Promotion turns authoring material into another artifact.

Supported conceptual transitions:

- dump to quiz;
- quiz to spec;
- dump plus quiz to spec;
- spec with `status: draft` to accepted or final spec;
- spec to plan;
- plan to tracker;
- plan/tracker to tickets.

The TUI should show available promotions from the selected session or document.

Promotion confirmation should show:

- source artifact;
- target artifact;
- output path;
- overwrite or conflict behavior;
- provenance that will be recorded.

The TUI should not silently overwrite an existing durable artifact.

## Provenance Display

The TUI should show provenance in human terms:

- created from which source artifact;
- generated during which session;
- promoted into which output;
- related ticket, plan, or tracker;
- timestamp when available.

The term "provenance" may appear in metadata/detail sections, but primary UI
copy should use clearer labels such as "created from" and "related artifacts."

## Acceptance Criteria

- A user can find the right authoring session by topic, target, status, and
  updated time.
- A user can create a session without developer-only flags or path knowledge.
- Static and context-aware quiz modes are clearly distinguished.
- Generated questions and answers are persisted as project-local artifacts.
- Promotion previews show source, target, output, and conflict behavior.
- The TUI can explain why a promotion is unavailable.
