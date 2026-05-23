# TUI Operational Console

Status: draft

Sources:

- `/home/adrian/Downloads/bubble_tea_tui_spec_draft.md`
- `/home/adrian/Downloads/agent_requirements_draft.md`
- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-navigation-and-layout.md`
- `.prj/specs/specs/tui-record-screens.md`
- `.prj/specs/specs/tui-authoring-and-quiz-workflow.md`
- `.prj/specs/specs/tui-actions-and-safety.md`
- `.prj/specs/specs/tui-ticket-execution-and-evidence.md`
- `.prj/specs/specs/tui-implementation-boundary.md`

## Purpose

Define the ChrysoCoord Bubble Tea terminal interface as a first-class local
operational console.

The TUI is not a decorative wrapper around commands. It is the main human
surface for operating a local ChrysoCoord project from a desktop development
environment while preserving the CLI as the automation and expert surface.

The TUI should make the system feel like one coherent local agent workflow
console rather than a loose set of Markdown files, commands, IDs, runtime logs,
and implementation reports.

## Product Position

ChrysoCoord remains a single local tool.

The default human entry point is:

```sh
chc
```

The explicit equivalent remains:

```sh
chc tui
```

`chc --help` keeps command help behavior.

The TUI is designed for an interactive desktop development environment. CI,
cron, pipes, non-interactive terminals, and remote automation are not the target
for this surface. Those remain CLI concerns.

## Core Principle

The TUI is ticket-centered, document-aware, and agent-observable.

- Tickets are the operational unit of task resolution.
- Documents provide knowledge, planning, and refinement context.
- The board exposes ticket state.
- The agent operates through tickets and leaves observable evidence.
- Plans and slicing organize work and order tickets.
- Tracking should not become a competing manual source of truth.

## Mental Model

The user should understand the TUI through five primary spaces:

- Workspace
- Board
- Ticket
- Documents
- Agent

### Workspace

Workspace is the selected project context.

It shows:

- project name and path;
- `.chc/` initialization status;
- current registry status;
- runtime readiness summary;
- warning state when project-local assets are missing or inconsistent.

### Board

Board is the primary operational overview.

It groups tickets by lifecycle state and answers:

- what is pending;
- what is ready;
- what is being worked;
- what is blocked;
- what is under review;
- what is done.

Board state comes from tickets.

### Ticket

Ticket is the concrete resolution unit.

It shows:

- identity;
- status;
- priority, severity, type, labels, use case, and project area when present;
- description;
- checklist;
- comments or notes;
- linked docs, plan, slice, branch, PR, reports, runs, and evidence when
  represented;
- source path.

### Documents

Documents are project-local knowledge and planning artifacts.

This includes:

- dumps;
- quizzes;
- specs;
- plans;
- slicing plans;
- trackers;
- reports;
- handoff material.

Documents can be viewed and later opened in the user's editor. The first TUI
implementation does not need an internal Markdown editor.

### Agent

Agent is the observability and control space for ticket execution.

It shows:

- current agent status;
- active ticket when known;
- latest execution record;
- latest loop record;
- latest validation evidence;
- blocked reasons;
- runtime logs and result paths;
- available agent actions in later slices.

## Source Of Truth

The TUI must not create a second state model.

It reads and writes through the same repositories, domain services, Markdown
artifacts, runtime records, indexes, and config that the CLI and agent use.

When a durable operation exists in the CLI, the TUI should call the same domain
operation or repository boundary rather than parsing command output.

TUI-specific state may include only ephemeral UI state:

- selected view;
- selected row;
- focused pane;
- active filter;
- help overlay state;
- transient messages.

TUI-specific state must not become project workflow state.

## Ticket, Plan, Slice, And Tracking Model

The ticket is the task resolution unit.

Plans and slicing define grouping, ordering, and intent. They answer why a set
of tickets exists and in what order the work should proceed.

Tracking should not compete with tickets as a second manually maintained source
of progress. If a tracking Markdown artifact exists, the TUI should display it
as a document. It should not treat it as more authoritative than ticket state.

Preferred direction:

- tickets carry executable state;
- plans and slicing group and order tickets;
- tracking becomes a derived or aggregate view where possible;
- existing tracker documents remain visible and inspectable;
- the TUI avoids hardening a duplicate plan/tracker/ticket model until the
  domain model is explicit.

The TUI may show a future relationship like:

```text
Plan
  Slice 1
    Ticket CHC-12
    Ticket CHC-13
  Slice 2
    Ticket CHC-14
```

The status shown beside each ticket must come from the ticket.

## Agent Workflow Model

The agent workflow has two conceptual loops.

The outer loop selects eligible tickets and processes them sequentially.

The inner loop resolves one ticket by reading context, performing work, running
validation, updating state, and recording evidence.

The TUI should make this understandable without exposing unnecessary machinery.

For read-only slices, show:

- loop status;
- selected or active ticket when known;
- latest execution status;
- blocked or waiting state;
- evidence paths and summaries.

For later action slices, prefer explicit operations over unconstrained Markdown
editing:

- add ticket comment;
- complete checklist item;
- move ticket state;
- attach artifact reference;
- append execution note;
- mark blocked;
- start ticket execution;
- start or stop a loop.

## Command Entry Behavior

`chc` with no subcommand opens the TUI.

`chc tui` opens the same TUI.

`chc --help` prints help.

Subcommands keep their existing behavior.

The default `chc` behavior should be implemented in the command layer, not by
adding logic to `main.go` beyond assembly.

## Package Boundary

The TUI implementation should use a separate package boundary.

Expected shape:

- `cmd/chc`: main assembly only;
- `internal/chcapp`: command construction and CLI/TUI entrypoints;
- `internal/chctui`: Bubble Tea model, update, view, data loading, and UI
  state;
- `internal/chc`: domain models, repositories, and services.

The TUI should not import CLI command output formatting as a data contract.

## Top-Level Navigation

Primary views:

- Dashboard
- Board
- Tickets
- Documents
- Authoring
- Agent
- Runtime
- Reports
- Help

The first implementation can include placeholder read-only views where data
loading is not complete, but the navigation model should be stable enough to
grow without replacing the mental model.

## Global Keys

Required baseline keys:

- Arrow keys or `j`/`k`: move selection.
- Enter: open selected record or enter detail.
- Esc or Backspace: go back or close overlay.
- Tab: cycle focus when multiple panes exist.
- `/`: filter current list.
- `r`: refresh current view.
- `?`: help for current view.
- `q`: back or quit depending on depth.
- `ctrl+c`: quit.

Future keys:

- `:` opens command palette.
- `g` opens global navigation overlay.
- `d` toggles details on narrow layouts.

Important actions must be discoverable through visible help or a command
palette. They should not depend on memorized hidden shortcuts.

## Command Palette Direction

The command palette is a required direction, but it does not need to be complete
in the first read-only slice.

It should eventually support:

- create ticket;
- create dump;
- start quiz;
- generate spec;
- create plan;
- open board;
- open document;
- open ticket by key;
- refresh workspace;
- validate workspace;
- start one ticket;
- start loop;
- stop loop;
- show agent status;
- show help.

The palette filters by typed text and shows command name, short description,
and shortcut when available.

## Layout

Default layout for comfortable terminals:

```text
┌────────────────────────────────────────────────────────────────────────────┐
│ Workspace: project-name                 View: Board        Agent: Idle      │
├──────────────────────┬───────────────────────────────────────┬─────────────┤
│ Navigation           │ Main Content                          │ Context     │
│ Dashboard            │ Board / Ticket / Document / Agent     │ Details     │
│ Board                │                                       │ Related     │
│ Tickets              │                                       │ Actions     │
│ Documents            │                                       │ Evidence    │
│ Authoring            │                                       │             │
│ Agent                │                                       │             │
│ Runtime              │                                       │             │
│ Reports              │                                       │             │
├──────────────────────┴───────────────────────────────────────┴─────────────┤
│ ? help   / filter   r refresh   enter open   q back/quit                   │
└────────────────────────────────────────────────────────────────────────────┘
```

The left pane is stable navigation.

The center pane is the current work surface.

The right pane is contextual detail, related records, evidence, or available
actions.

## Responsive Behavior

The TUI should target desktop terminal use.

Comfortable target:

- 100 columns by 30 rows.

Minimum usable target:

- 80 columns by 24 rows.

At narrow sizes:

- hide left navigation;
- show breadcrumb or current view label;
- hide right context pane by default;
- allow detail toggle later;
- avoid rendering overlapping text.

If the terminal is too small, show a clear message rather than a broken layout.

## Visual Style

The style should be sober, dense, and operational.

Use color for state, not decoration.

Every color-coded state must also have text.

Avoid decorative panels, oversized headings, and marketing-style screens.

Use stable widths for columns and rows where practical so selection and dynamic
messages do not shift the layout unexpectedly.

## Dashboard

Dashboard is the first informational view when the TUI opens.

It should show:

- active project;
- workspace status;
- ticket counts by state;
- blocked ticket count;
- latest agent or loop status;
- latest execution status;
- recent artifacts or reports;
- current warning state.

Dashboard should not become a dense document browser. It should orient the user
and provide paths into Board, Tickets, Documents, and Agent.

## Board

Board is the first major operational view.

It should group tickets by state.

Initial states should come from existing ticket status values. Do not invent a
separate board status model in the TUI.

Board row/card fields:

- ticket key;
- title;
- status;
- priority when present;
- checklist ratio when available;
- blocked marker when available;
- linked plan/slice hint when available.

Read-only board actions:

- move selection;
- open ticket detail;
- filter;
- refresh;
- open help.

Mutating board actions are deferred to later slices.

## Ticket Detail

Ticket detail should show:

- key and title;
- status;
- priority, severity, type;
- labels, use case, project area;
- source path;
- description/body summary;
- checklist;
- comments or notes;
- branch and PR fields when present;
- linked reports/runs/evidence when available.

Checklist and comments should be rendered as first-class sections when the
parser exposes them. If parsing is not yet structured enough, render the source
Markdown section with a clear label and avoid pretending it is structured.

## Documents

Documents view should index:

- dumps;
- quizzes;
- specs;
- plans;
- slicing plans;
- trackers;
- reports;
- handoff summaries.

List fields:

- type;
- title or topic;
- status when known;
- related ticket or plan when known;
- updated time when known;
- path.

Detail should render Markdown content read-only with metadata above it.

Opening in `$EDITOR` is deferred to an action slice.

## Authoring

Authoring view should expose:

- authoring sessions;
- quizzes;
- specs;
- plans;
- trackers.

Session list fields:

- handle or id;
- status;
- target;
- topic;
- updated time.

Session detail should show:

- session metadata;
- inputs;
- outputs;
- provenance count or links;
- transcript path;
- artifacts path;
- generated quiz/spec paths.

The TUI should not force users to treat session IDs as the main mental model
when topic or slug is available.

## Agent

Agent view should show:

- daemon status when available;
- loop status when available;
- latest selected ticket;
- latest execution record;
- latest validation result;
- blocked reasons;
- latest notification events when relevant.

Slice 1 of the TUI implementation can show placeholders for unavailable agent
data, but the view should exist.

## Runtime

Runtime view should show:

- runtime readiness summary;
- Docker image;
- validation command;
- latest runtime validation result;
- latest execution artifact paths;
- warnings.

It should reuse existing compact runtime output semantics.

## Reports

Reports view should show:

- slice reports;
- execution reports;
- loop reports;
- handoff summaries;
- validation evidence.

Reports should link back to related ticket, run, loop, slice, plan, or tracker
when metadata exists.

## Actions And Mutations

The first implementation set must separate read-only rendering from mutation.

Action slices should add one safe group at a time:

- create artifacts;
- mutate tickets;
- open editor;
- start agent work;
- control loops.

Mutating actions require confirmation unless they are trivially reversible or
already have an explicit editor save boundary.

Long-running actions require visible progress or an operation result message.

## External Editor

Long text should be edited through the user's configured editor.

Candidate editor actions:

- edit dump;
- edit spec;
- edit plan;
- edit ticket;
- write long comment.

After editor exit, the TUI should reload and validate the artifact before
presenting success.

## Error Handling

Errors should be visible, concise, and inspectable.

Use a bottom message area for short transient errors.

Use detail panels or records for longer parse/runtime/validation errors.

Do not silently ignore malformed Markdown or missing files when that affects
state shown to the user.

## Testing Direction

Prefer testing:

- data loading functions;
- view model construction;
- navigation update behavior;
- filtering;
- action planning;
- confirmation state;
- render snapshots only where stable enough.

Avoid brittle full-terminal snapshot tests for every screen.

## Initial Slicing Direction

The TUI implementation should be its own slicing set with local slice numbers
starting at `Slice 1`.

The old `UX Operation Surface` planned TUI implementation slices should be
treated as extracted into this new TUI-specific set.

## Acceptance Criteria

- `chc` opens the TUI by default.
- `chc tui` opens the same TUI.
- The TUI has a dedicated package boundary.
- The first screen or navigation model makes Board, Tickets, Documents, and
  Agent visible.
- Tickets are treated as the operational progress source.
- Plans/slicing group and order tickets without becoming a competing progress
  state in the TUI.
- Tracking documents are shown as artifacts unless or until a derived tracking
  model exists.
- Read-only and mutating capabilities are implemented in separate slices.
- The TUI does not create a second source of truth.
