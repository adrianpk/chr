# Dashboard Board Ticket Operations Quiz

Date: 2026-05-26
Topic: Dashboard, Board, ticket generation, and ticket operations

## Purpose

Capture pre-spec decisions and tensions for the next Board/Dashboard planning
cycle. The feature is not only a visual board. It also touches ticket creation,
ticket generation contracts, operational status, indexing, and future
multi-operator workspace layout.

## Core Direction

The Board should be the operational status surface for the current project and
agent workflow. It should show how work is evolving and let the user create
new backlog items or urgent hotfix tickets.

The later F8 Dashboard can become the broader metrics-oriented view. This Board
MVP should stay focused on ticket flow, status, attention signals, and
actionable operation.

## Board Modes

The Board should support two mental modes:

- Compact mode: the left pane is a window into one board column.
- Maximized mode: the TUI shows a broader board with multiple columns visible.

Compact mode should navigate normal columns with arrow-like navigation. The
special `attention` view should stay outside the normal cycle and be toggled
with a dedicated shortcut. Repeating the shortcut should return to the previous
normal column.

## Initial Columns

Initial board flow:

- `backlog`
- `ready_for_dev`
- `in_progress`
- `ready_for_review`
- `ready_for_merge`
- `done`

UI labels should use human-readable text without underscores, such as "Ready
for dev".

`done` tickets can later be archived by a retention heuristic based on age and
maximum visible count. Archival is not part of the immediate MVP unless needed
by implementation constraints.

## Attention Model

The first attention model should be conservative. Candidate signals:

- blocked ticket;
- important or new agent comment;
- comment mentioning the user;
- validation failure;
- invalid or unknown status/category.

Older-ticket heuristics, reviewer-specific PR state, ownership anomalies, and
state/evidence conflicts should be treated as later enhancements unless they
already exist in the current ticket model.

The left pane should show a restrained text-oriented hint on attention rows.
The right pane should remain the primary place to inspect details. Attention
reasons should start as plain text. If that becomes too noisy, the display can
be refined during implementation.

## Manual Ticket Creation

Board-created tickets should require enough structure to be useful without
slowing operation:

- title;
- short synopsis;
- priority;
- severity;
- type;
- optional use case;
- optional project area.

Manual tickets should default to `backlog`. The user can edit the Markdown and
set another status explicitly.

Unknown statuses from manual Markdown edits should not silently disappear into
normal backlog. They should be discoverable through the `attention` view.

## Generated Tickets

Generated tickets must not rely only on a source document link. Each generated
ticket should include a persisted short synopsis that explains the work without
opening the source plan/spec/tracking document.

The LLM contract for generated tickets should include synopsis alongside title,
body, acceptance criteria, slice identifier, dependencies, and source
references.

For manually created tickets, the TUI should ask for synopsis. For assisted
manual creation, ChrysoCoord may help generate a synopsis.

External Markdown without synopsis should display poorly but should not
necessarily become attention. A later sync workflow may generate missing
synopses only when the field is empty after trimming whitespace. It must not
overwrite non-empty user text.

## SQLite Role

SQLite should be considered an operational index for Board metadata, not a
replacement for Markdown bodies.

SQLite is part of the MVP. The implementation should inspect repository
history because the project previously had model/persistence work that may
inform the new design, even if it should not be copied mechanically.

The Board index can store:

- internal row id;
- user-facing key such as `CHC-123`;
- title;
- synopsis;
- status or board column;
- priority;
- severity;
- type;
- use case;
- project area;
- attention flags;
- source kind;
- source references;
- ticket Markdown path;
- created and updated timestamps.

The ticket body, comments, and rich context should remain in Markdown. The
Board can list/filter/sort from SQLite and load the selected ticket Markdown
for the right pane.

## Sync Direction

Manual sync is preferred over a daemon watcher.

Possible post-MVP sync responsibilities:

- reconcile Markdown tickets not present in the Board index;
- detect duplicate or invalid keys;
- detect invalid statuses;
- optionally generate missing empty synopses from ticket body with LLM;
- preserve non-empty user-authored synopsis text.

Daemon watching should be mentioned as evaluated but not preferred. The
project should avoid adding always-running processes when a stateless/manual
operation is enough.

## Multi-Operator Insight

The Board discussion exposed a broader `.chc` collision issue. In a
multi-user repository, collisions can happen for tickets and for generated
dumps, quizzes, specs, plans, tracking files, and context artifacts.

MVP can assume a monouser repository/operator workflow, but the implementation
should still prepare the userspace shape so the later multi-operator model is
not blocked. The initial shape can be a single active user directory under
`.chc/user/`, selected by a required environment variable. A `.envrc` can set
that variable for users who use direnv, while users without direnv must set it
another way.

The spec must also record the future direction:

- generated ChrysoCoord artifacts should eventually live under an operator
  namespace;
- a likely future shape is `.chc/operators/<operator-id>/...`;
- MVP may start with `.chc/user/` and a single configured operator id;
- operator id can be resolved from local config, an environment variable, or
  `.envrc`;
- this approach avoids collisions without requiring a daemon or a shared
  binary database;
- the issue should be revisited soon after MVP, not forgotten.

## MVP Boundary

MVP should include:

- compact Board column navigation;
- maximized Board view concept;
- manual ticket creation from Board;
- mandatory synopsis for ChrysoCoord-created tickets;
- generated ticket contract with synopsis;
- conservative attention signals;
- right-pane ticket detail inspection;
- monouser/operator assumption documented;
- SQLite operational ticket metadata index;
- initial userspace structure for a single configured operator.

Post-MVP should include:

- Markdown sync/reconciliation;
- LLM-generated synopsis for empty external Markdown synopsis;
- archive/retention heuristics for old done tickets;
- richer attention signals;
- full multi-operator userspace under `.chc/operators/<operator-id>/...`.

## Open Questions

- How much of the previous SQLite/model history can be reused or adapted?
- How should the right pane evolve if plain-text attention reasons become too
  noisy?
- Which existing status values need migration or aliases?
- How much of the current `.chc` layout should be redesigned before Board
  writes new artifacts?
