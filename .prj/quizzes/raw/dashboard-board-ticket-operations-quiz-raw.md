# Dashboard Board Ticket Operations Quiz Raw Capture

Date: 2026-05-26
Topic: Dashboard, Board, ticket generation, and ticket operations

## Raw Capture

The user wants the Board/Dashboard work to cover more than visualization. It
must also account for ticket creation, generated ticket quality, operational
status, and future multi-operator workspace concerns.

The Board should primarily expose project status and real-time evolution of
agent execution. It should also allow the user to create tickets manually for
backlog intake or urgent hotfixes.

The broader Graphana-like dashboard is deferred to a later F8 Dashboard view.
This Board work should focus on status and operation rather than high-level
metrics. Counters can exist, but visual state should be obvious without turning
the view into a metrics dashboard.

Tickets requiring attention should have a restrained visual hint in the row.
Initial attention signals should be conservative: blocked tickets, important
comments, comments mentioning the user, validation failures, and clearly
invalid categorization. PR reviewer and old-ticket heuristics are likely later
work.

The right pane already shows full available ticket detail. If the left pane row
contains an attention hint, moving the cursor to that ticket should make the
right pane sufficient to inspect the reason. The exact layout for attention
metadata can be refined later.

The compact Board view should treat the left pane as a window into one board
column. Users should be able to navigate normal board columns with arrow-like
navigation. Maximized mode should show a fuller TUI board resembling a classic
Kanban dashboard.

Initial column flow should be:

- backlog
- ready for dev
- in progress
- ready for review
- ready for merge
- done

The underlying status expression should be explicit, such as `ready_for_dev`,
to avoid ambiguity with review and merge readiness. UI labels should not show
underscores.

An `attention` column should exist as a virtual/special column. It should stay
outside the normal compact column cycle and be toggled with a dedicated
shortcut. Repeating the shortcut should return to the previously selected
normal column.

Done tickets can accumulate. A future retention/archive heuristic may move old
done tickets to archived based on done/complete age and maximum visible count.

Manual ticket creation from Board should ask for title, synopsis, priority,
severity, type, optional use case, and optional project area. New manual
tickets should default to backlog, but the user can edit the status if needed.

Invalid or unknown statuses from manually edited Markdown should be surfaced.
The preferred first model is an `attention` virtual column rather than silently
burying such tickets in backlog.

Tickets generated from plan/spec/tracking currently link to source docs but
need a visible synopsis. The synopsis should be persisted, visible in the Board
list/detail, and not force users to open the source plan/spec to understand the
work.

Ticket identity raised a larger question. A user-facing key such as `CHR-123`
is useful, but UUIDs in Markdown are not enough if users can copy and duplicate
them. A service would need to validate uniqueness. SQLite can help with
constraints, monotonic key assignment, search, filtering, board columns, and
status presentation.

SQLite is part of the Board MVP. The user explicitly pointed to repository
history because the project previously had model/persistence work that may be
useful to inspect. The preferred MVP direction is not to store the full ticket
body in SQLite. SQLite should store operational metadata and references:

- internal row id
- key
- title
- synopsis
- status or board column
- priority
- severity
- type
- use case
- project area
- attention flags
- source kind
- source references
- ticket Markdown path
- timestamps

Long body, comments, and full context remain in Markdown. The Board can list
and filter from SQLite while the right pane can read Markdown by path.

Synopsis should be mandatory in Chryso-created flows:

- LLM-generated tickets must include synopsis.
- manually created tickets from Board should require synopsis.
- imported or externally edited Markdown without synopsis should display
  poorly but not necessarily become attention.

A post-MVP sync could inspect external Markdown. If the synopsis is empty after
trimming whitespace and the body contains enough information, an LLM can
generate a synopsis. It must not overwrite non-empty user text, even if the text
is low quality.

Sync was discussed as post-MVP. MVP should not require daemon watchers. A
manual keybinding/command for sync is preferable if reconciliation becomes
necessary. Daemon watching should be mentioned as evaluated but not preferred
because the project avoids keeping extra processes running when a stateless
workflow is enough.

The conversation opened a broader `.chr` multi-user concern. If several users
generate artifacts in the same repo, collisions can happen not only for tickets
but also dumps, quizzes, specs, plans, context, and tracking. A future direction
is operator userspace under `.chr/operators/<operator-id>/...`. The operator id
could be resolved from local config, an environment variable, or `.envrc`.
This should remain outside the MVP but recorded so the design does not need to
rediscover the issue later.

For MVP, assume a monouser repository/operator workflow, while documenting
that multi-operator support should use namespaced operator artifacts soon.
Even in MVP, it may be useful to prepare a single-user userspace shape such as
`.chr/default/`, with the active user selected through a required environment
variable. Users can set that through `.envrc` when using direnv, or by any
other environment mechanism. Full `.chr/operators/<operator-id>/...`
multi-operator behavior remains outside MVP.

Attention details should start with plain text in the UI. If that consumes too
much space or becomes visually noisy, the presentation can be refined during
implementation.
