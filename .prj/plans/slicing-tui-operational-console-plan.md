# TUI Operational Console Slicing Plan

Status: planned

## Purpose

Plan the `TUI Operational Console` implementation as its own slicing set.

This set starts at `Slice 1` and replaces the previously planned TUI
implementation slices in `UX Operation Surface`.

## Source

Primary sources:

- `.prj/specs/specs/tui-operational-console.md`
- `/home/adrian/Downloads/bubble_tea_tui_spec_draft.md`
- `/home/adrian/Downloads/agent_requirements_draft.md`

Reference sources:

- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-navigation-and-layout.md`
- `.prj/specs/specs/tui-record-screens.md`
- `.prj/specs/specs/tui-authoring-and-quiz-workflow.md`
- `.prj/specs/specs/tui-actions-and-safety.md`
- `.prj/specs/specs/tui-ticket-execution-and-evidence.md`
- `.prj/specs/specs/tui-implementation-boundary.md`
- `.prj/plans/slicing-ux-operation-surface-plan.md`
- `.prj/trackers/slicing-ux-operation-surface-tracker.md`

## Current State

The CLI and domain layers already provide:

- project resolution;
- workspace initialization;
- ticket creation, listing, status transitions, execution, loop, comments, and
  reports;
- authoring sessions, contextual quizzes, and promotion;
- runtime validation and execution evidence;
- maintenance, notification, and remote command records.

The TUI implementation has not started.

## Design Direction

The TUI is the normal human operating surface.

The CLI remains the automation and expert interface.

`chc` with no subcommand should open the TUI.

`chc tui` should open the same TUI explicitly.

The implementation should use a dedicated internal TUI package, not a large
inline command implementation.

Tickets are the operational task resolution unit. Plans and slicing group and
order tickets. Tracking documents are visible artifacts, not a competing manual
progress source.

## Out Of Scope

This slicing does not include:

- browser UI;
- multi-user workflows;
- multi-agent concurrent repository mutation;
- a full internal Markdown editor;
- replacing Markdown-backed project state;
- redesigning the ticket/slice/tracking domain model;
- live Telegram/email UI beyond showing existing records where available.

## Numbering Rule

Slice numbers are local to this slicing and start at `Slice 1`.

Do not continue from `UX Operation Surface` slice numbers.

## Slices

### Slice 1: TUI App Foundation

Status: planned.

Expected branch: `feat/slice-1-tui-app-foundation`
Expected PR title: `feat(slice-1): tui app foundation`
Expected report: `.prj/reports/slices/tui-operational-console/slice-1-tui-app-foundation.md`

Goal:

Create the Bubble Tea application boundary and entry behavior without adding
domain mutations.

Candidate scope:

- add Bubble Tea dependencies;
- add `internal/chctui` package;
- add `chc tui` command;
- make `chc` with no subcommand open the TUI;
- keep `chc --help` behavior;
- implement base model, update, and view;
- show active project and navigation shell;
- include placeholder read-only views for Dashboard, Board, Tickets, Documents,
  Authoring, Agent, Runtime, Reports, and Help;
- add model/update tests where practical.

Expected outcome:

- users can launch the TUI through `chc` or `chc tui` and see a stable
  project-aware operational shell.

### Slice 2: Read-Only Board And Ticket Detail

Status: planned.

Expected branch: `feat/slice-2-tui-board-ticket-detail`
Expected PR title: `feat(slice-2): tui board ticket detail`
Expected report: `.prj/reports/slices/tui-operational-console/slice-2-tui-board-ticket-detail.md`

Goal:

Make tickets the first useful operational surface.

Candidate scope:

- load tickets through existing ticket repositories/indexes;
- group tickets by status for a board view;
- list key, status, priority, severity, type, and title;
- select tickets and show detail;
- render source path, body summary, checklist summary, comments/notes summary,
  branch and PR fields when available;
- show plan/slice/tracker references only when represented in existing ticket
  metadata or text;
- add filtering by text;
- keep all ticket behavior read-only.

Expected outcome:

- users can inspect implementation state from a board-like TUI without copying
  ticket IDs between commands.

### Slice 3: Document And Authoring Explorer

Status: planned.

Expected branch: `feat/slice-3-tui-document-authoring-explorer`
Expected PR title: `feat(slice-3): tui document authoring explorer`
Expected report: `.prj/reports/slices/tui-operational-console/slice-3-tui-document-authoring-explorer.md`

Goal:

Expose project-local documents and authoring sessions as navigable artifacts.

Candidate scope:

- list dumps, quizzes, drafts, specs, plans, slicing plans, trackers, reports,
  and handoff material;
- list authoring sessions with handle/id, status, target, topic, updated time;
- show session inputs, outputs, provenance, transcript path, and artifacts path;
- render selected Markdown artifact read-only;
- show tracker documents as artifacts, not as a progress authority;
- show plan/slicing documents as grouping/intention artifacts;
- add tests for document discovery and view model construction.

Expected outcome:

- users can move from ticket state into supporting project knowledge without
  chasing paths manually.

### Slice 4: Agent And Runtime Observability

Status: planned.

Expected branch: `feat/slice-4-tui-agent-runtime-observability`
Expected PR title: `feat(slice-4): tui agent runtime observability`
Expected report: `.prj/reports/slices/tui-operational-console/slice-4-tui-agent-runtime-observability.md`

Goal:

Make agent, loop, runtime, and evidence state visible before adding control
actions.

Candidate scope:

- show daemon status when available;
- show latest loop records and execution records;
- show latest validation status and runtime readiness summary;
- show blocked reasons and failure summaries when available;
- show relevant artifact paths for logs, results, reports, validation evidence,
  and handoff material;
- keep all views read-only;
- avoid starting or stopping work in this slice.

Expected outcome:

- users can see what the agent did, what failed, and where evidence lives from
  the TUI.

### Slice 5: Command Palette Foundation

Status: planned.

Expected branch: `feat/slice-5-tui-command-palette`
Expected PR title: `feat(slice-5): tui command palette`
Expected report: `.prj/reports/slices/tui-operational-console/slice-5-tui-command-palette.md`

Goal:

Add a discoverable command palette without wiring high-risk mutations.

Candidate scope:

- add `:` command palette overlay;
- add command registry model;
- add filtering by typed text;
- expose navigation commands;
- expose refresh and help commands;
- expose open selected source path as an explicit read-only action if safe;
- show command descriptions and shortcut hints;
- keep mutating commands disabled or absent until later slices.

Expected outcome:

- users can discover navigation and safe operations without memorizing keyboard
  shortcuts.

### Slice 6: Safe Artifact Creation Actions

Status: planned.

Expected branch: `feat/slice-6-tui-artifact-creation-actions`
Expected PR title: `feat(slice-6): tui artifact creation actions`
Expected report: `.prj/reports/slices/tui-operational-console/slice-6-tui-artifact-creation-actions.md`

Goal:

Wire low-risk creation flows through existing domain operations.

Candidate scope:

- create ticket;
- create authoring session;
- create context-aware quiz;
- create dump or draft where existing operations support it;
- show confirmation/review before durable writes;
- reload state after successful writes;
- show validation errors clearly;
- preserve CLI parity and auditability.

Expected outcome:

- users can start common project artifacts from the TUI without using long CLI
  commands.

### Slice 7: Ticket Mutation Actions

Status: planned.

Expected branch: `feat/slice-7-tui-ticket-mutation-actions`
Expected PR title: `feat(slice-7): tui ticket mutation actions`
Expected report: `.prj/reports/slices/tui-operational-console/slice-7-tui-ticket-mutation-actions.md`

Goal:

Expose explicit ticket operations instead of free-form opaque Markdown edits.

Candidate scope:

- add ticket comment;
- move ticket state;
- toggle checklist item when the parser supports it safely;
- add checklist item if an existing operation exists or is introduced
  narrowly;
- validate ticket after mutation;
- show operation result and reload board/detail state;
- keep destructive operations out of scope unless explicitly confirmed and
  supported.

Expected outcome:

- users can maintain ticket state through structured operations that match the
  agent workflow.

### Slice 8: Agent Work Control

Status: planned.

Expected branch: `feat/slice-8-tui-agent-work-control`
Expected PR title: `feat(slice-8): tui agent work control`
Expected report: `.prj/reports/slices/tui-operational-console/slice-8-tui-agent-work-control.md`

Goal:

Wire agent execution controls after read-only observability and safe local
actions exist.

Candidate scope:

- start selected ticket;
- start loop with explicit filters;
- stop loop or request stop where supported;
- show confirmation for long-running operations;
- show operation progress or completion state;
- surface failure reasons and evidence paths after execution;
- preserve ticket as the operational resolution unit.

Expected outcome:

- users can delegate work to the agent from the TUI while retaining visibility
  into state, evidence, and blockers.
