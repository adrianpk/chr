# TUI Product Surface

Status: initial

Sources:

- `.prj/drafts/chc-cli-output-polish-notes.md`
- `.prj/plans/slicing-ux-operation-surface-plan.md`
- `.prj/tickets/open/20260522071859-align-cli-authoring-and-tui-operation-surface.md`
- `docs/how-to/full-system-lab-runbook.md`
- `docs/how-to/user-guide/index.md`

## Purpose

Define the role of the Bubble Tea TUI in ChrysoCoord.

The TUI is the primary local human operation surface over the existing CLI and
domain services. It is not a replacement for command-line capabilities. Every
operation that matters for automation, runbooks, testing, and expert use should
continue to have a command surface.

The TUI exists because normal operation should not require copying opaque IDs,
remembering long command flags, reading walls of absolute paths, or manually
cross-referencing generated Markdown files.

## User Model

The primary user is one local operator working on one managed project at a time.

The operator may be:

- bootstrapping a new Go project into ChrysoCoord;
- reviewing project-local tickets, documents, reports, and runtime evidence;
- creating or refining authoring artifacts;
- running a ticket directly;
- starting or observing a filtered ticket loop;
- recovering from a blocked execution;
- preparing the next implementation slice.

The TUI should assume technical literacy, but it should not assume that the user
remembers ChrysoCoord's internal path layout or every command option.

## Product Role

The TUI should make common work visible and selectable:

- records are selected from lists;
- details are shown in panes;
- actions run against the selected record;
- generated files remain visible as project-local Markdown artifacts;
- long-running operations show state, not just command output;
- errors and blockers remain inspectable after the screen changes.

The TUI should preserve the same source of truth as the CLI:

- `.chc/` remains the project-local workspace;
- Markdown remains the durable record for tickets, authoring artifacts, plans,
  specs, trackers, reports, quizzes, dumps, and handoff notes;
- runtime evidence remains under the runtime artifact paths;
- user-level indexes remain derived operational state.

## Current Scope

The first TUI generation should support local foreground operation.

Required first-class areas:

- project overview;
- ticket board and ticket details;
- authoring sessions;
- drafts, specs, plans, trackers, quizzes, and dumps;
- runtime validation;
- ticket execution and loop evidence;
- reports and handoff summaries;
- notification and remote-command status as read-only observability.

The first useful TUI should not need to implement every mutating operation. It
must still be coherent when read-only areas link to CLI commands or mark actions
as unavailable.

## Non-Goals

The TUI is not:

- a browser application;
- a multi-user coordination surface;
- a multi-agent concurrent control plane;
- a replacement for Markdown files;
- a hidden database editor;
- a generic terminal file manager;
- a chat transcript viewer;
- a direct replacement for version control review.

The TUI should not introduce concepts that are absent from the CLI/domain model
only to make screens look richer.

## Human Operation Principles

### Selection Over Copying

The user should select a ticket, session, document, run, or report from a list
and trigger actions against that selection. Copying IDs remains possible, but it
should not be the normal path.

### State Before Paths

The TUI should show status, target, topic, title, updated time, validation
state, and next useful action before showing file paths.

Paths are secondary. They are useful for opening artifacts and debugging, but
they should not dominate primary views.

### Relative Paths by Default

When paths appear in the TUI, they should be relative to the managed project
root unless the absolute path is required for disambiguation.

### Project-Local Auditability

Every generated or changed artifact should be traceable back to project-local
files. The TUI should expose where records live without making the path the main
mental model.

### No Magic Conversations

Interactive authoring, quiz, execution, and recovery flows must persist useful
material. The TUI may feel conversational, but the durable outcome must be
stored as Markdown artifacts, ticket comments, reports, or runtime evidence.

## Required Current/Future Labels

The TUI must clearly distinguish:

- implemented actions;
- planned actions;
- unavailable actions due to missing configuration;
- unavailable actions due to invalid project state;
- blocked or failed actions with recorded evidence.

Planned features such as Telegram or email interaction may be visible as
configuration/status records, but they must not be presented as already working
unless implemented.

## Fit With CLI

The TUI should call internal domain services directly where practical, not
shell out to `chc` commands as its primary architecture.

The command surface remains the behavioral contract for users. The TUI should
match command behavior and output semantics, and tests should keep both paths
aligned.

When a TUI action has an equivalent CLI command, the detail or confirmation
screen may show the equivalent command in a compact form. This helps learning
and preserves the scriptable foundation.

## Acceptance Criteria

- A user can understand what project is active from the first screen.
- A user can find tickets, authoring sessions, reports, and runtime evidence
  without copying IDs.
- A user can inspect a record and see its status, source file, and next relevant
  actions.
- A user can distinguish durable artifacts from derived indexes and runtime
  records.
- A user can leave the TUI and continue through CLI commands without state loss.
- The TUI does not require or imply a web server, account, tenant, or
  multi-user model.

