# Board Ticket Operations

Status: initial

Model: `.prj/specs/specs/board-ticket-operations-model.md`
Source quiz: `.prj/quizzes/curated/dashboard-board-ticket-operations-quiz.md`

## Purpose

Define the Board MVP as the operational ticket-flow surface for chr.

The Board is not the later metrics dashboard. The later `F8` Dashboard may
become a broader Grafana-like view for activity, throughput, validation, and
runtime signals. This spec focuses on ticket creation, generated ticket
metadata, board status flow, attention signals, SQLite-backed ticket metadata,
and the first userspace shape needed to avoid blocking later multi-operator
work.

## Product Boundary

The Board should let an operator inspect the current project state, see work
advance in near real time as agents update tickets, and create new backlog or
hotfix tickets without leaving the TUI.

The MVP includes:

- compact Board column navigation;
- maximized Board view;
- manual ticket creation from the Board;
- generated ticket synopsis support;
- a SQLite operational metadata index;
- conservative attention detection;
- right-pane ticket detail inspection;
- a single active project userspace under `.chr/default/`;
- CLI and TUI routing through shared services.

The MVP excludes:

- full multi-operator `.chr/operators/<operator-id>/...`;
- daemon-based filesystem watching;
- LLM-assisted Markdown sync for external tickets;
- automatic archival of old done tickets;
- rich dashboard metrics reserved for `F8` Dash.

## Board Views

Compact mode treats the left pane as a window into one normal Board column.
The selected normal column controls which tickets appear in the master list.
The right pane shows detail for the selected ticket.

Maximized mode shows a wider multi-column board. It should remain dense and
keyboard-first, not a decorative dashboard.

Normal columns are:

- `backlog`
- `ready_for_dev`
- `in_progress`
- `ready_for_review`
- `ready_for_merge`
- `done`

UI labels should be human-readable and avoid underscores:

- Backlog
- Ready for dev
- In progress
- Ready for review
- Ready for merge
- Done

Existing ticket status values should be migrated or aliased deliberately. The
current historical values include `open`, `ready`, `reviewing`, `solved`,
`merged`, and `closed`. The Board implementation must preserve readability for
existing tickets while moving new chr-created tickets to the Board
status vocabulary.

## Attention View

`attention` is a virtual view outside the normal column cycle. Compact column
navigation should move only through normal columns. A dedicated keybinding may
toggle attention mode and return to the previous normal column when repeated.

Attention should start conservative. MVP attention reasons are:

- blocked ticket;
- important or new agent comment when detectable from existing Markdown;
- comment mentioning the operator when detectable from existing Markdown;
- validation failure when represented in ticket or execution evidence;
- invalid or unknown status/category.

Older-ticket heuristics, reviewer-specific PR state, ownership anomalies, and
state/evidence conflicts are later enhancements unless an existing model makes
them cheap and reliable.

Attention hints in the list should be restrained and text-oriented. The detail
pane should show plain-text attention reasons first. If the text becomes noisy,
a later slice can refine placement or rendering.

## Ticket Creation

Manual Board ticket creation should collect:

- title;
- short synopsis;
- priority;
- severity;
- type;
- optional use case;
- optional project area.

Manual Board tickets default to `backlog`. A user who edits Markdown manually
may set another status, but unknown statuses must become discoverable through
the attention view.

Generated tickets from plan tracking materialization must include a persisted
short synopsis. A source document reference is traceability, not a replacement
for a readable ticket summary.

## Ticket Body and Synopsis

Markdown remains the durable ticket body and comment source. SQLite stores
searchable and selectable metadata. It must not duplicate large ticket bodies.

`synopsis` is part of the ticket metadata contract. It should be short enough
to display in a Board row or detail header and useful enough to understand the
ticket without opening the source plan or spec.

For generated tickets, the LLM output contract must include `synopsis` beside
title, body, acceptance criteria, slice identifier, dependencies, and source
references.

For manual TUI ticket creation, the operator should provide synopsis as plain
text. Assisted synopsis generation can be added later, but empty synopsis
handling must not block the MVP.

If a future sync operation inspects external Markdown and the synopsis field is
empty after trimming whitespace, an LLM may propose a synopsis from the body.
It must not overwrite non-empty user text, even if the text is low quality.

## SQLite Role

SQLite is part of the MVP. It is an operational metadata index, not the durable
body store.

The project already has:

- a migration runner;
- a user-level `chr.db`;
- `indexed_tickets`;
- `indexed_ticket_tags`;
- ticket discovery, parsing, rebuilding, and reconciliation;
- CLI commands for index rebuild and status.

The new Board work should extend or adapt that foundation instead of adding a
parallel persistence stack.

The Board index should support:

- stable ticket lookup by project and key;
- deterministic Board column queries;
- synopsis display;
- attention filtering;
- source traceability;
- monotonic key allocation that does not require scanning the filesystem for
  every new ticket;
- reload after ticket mutation.

Markdown body, comments, checklist text, and rich context stay in Markdown.
The Board may load the selected ticket Markdown for detail rendering.

## Userspace Direction

The MVP is a single-operator project workflow, but it should prepare a
project-local active userspace:

```text
.chr/
  user/
```

The active operator id should be resolved from a required environment variable
for workspace-mutating workflows that need userspace ownership. A `.envrc` may
set that variable for users who use direnv. Users without direnv must set it in
their shell or automation environment.

The MVP does not need to fully implement:

```text
.chr/
  operators/
    <operator-id>/
```

The spec records that as the preferred future shape for multi-operator
repositories. It avoids collisions between artifacts created by different
operators while still allowing users to version or ignore their own `.chr`
workspace according to project policy.

## Sync Direction

Manual sync is the preferred future approach for reconciling external Markdown
tickets with the SQLite index. A daemon watcher was evaluated and should remain
out of the MVP because always-running processes add operational weight that is
not needed for the first Board.

Future sync may:

- find Markdown tickets missing from SQLite;
- detect duplicate or invalid keys;
- detect invalid statuses;
- validate required metadata;
- fill empty synopsis through LLM assistance when the body is useful;
- preserve non-empty synopsis text.

For the MVP, tickets created through chr should update the SQLite index
through service boundaries. Existing index rebuild and status commands remain
the recovery path.

## CLI and TUI Parity

Board behavior must respect the CLI/TUI parity direction:

- durable operations should have CLI equivalents when meaningful outside an
  interactive terminal;
- TUI actions should route through shared services;
- CLI commands should not become legacy paths;
- tests should prove shared behavior where practical.

Pure layout navigation can remain TUI-only. Ticket creation, listing, status
movement, index rebuild/status, and future Board inspection should remain
scriptable.

## Acceptance Criteria

- New chr-created tickets include `synopsis`.
- New ticket statuses use the Board vocabulary.
- Existing legacy statuses are mapped or surfaced without hiding tickets.
- The SQLite ticket index stores Board metadata needed for list/detail views.
- The Board compact view can navigate normal columns.
- The attention view can be toggled outside normal column navigation.
- The selected ticket detail shows synopsis, source references when present,
  and plain-text attention reasons.
- Manual Board ticket creation can create a backlog ticket with metadata.
- Plan tracking materialization creates tickets with meaningful synopsis.
- Existing CLI ticket commands and TUI Board behavior route through shared
  service logic where the action is durable.
- Documentation states what is MVP, what is future sync, and what is future
  multi-operator userspace.

## Open Questions

- Which exact environment variable name should select the active operator for
  `.chr/default/`?
- Should legacy statuses be migrated on write, mapped only for display, or both?
- Should monotonic ticket allocation use a new SQLite sequence table or derive
  from existing indexed rows with a transaction lock?
- Which attention reasons can be detected reliably from current Markdown and
  execution artifacts without adding fragile parsing?
