# TUI Navigation and Layout

Status: initial

Sources:

- `.prj/specs/specs/tui-product-surface.md`
- `.prj/drafts/chr-cli-output-polish-notes.md`
- `.prj/plans/slicing-ux-operation-surface-plan.md`

## Purpose

Define the primary navigation model and screen layout for the Chryso TUI.

The design should make project state scannable first and actionable second. It
should be efficient for keyboard use and predictable enough that terminal users
do not need to learn a complex interaction model.

## Layout Model

The default desktop terminal layout should use three stable regions:

- top status bar;
- left navigation rail;
- main content area.

The main content area may split into list and detail panes when width allows.
For narrow terminals, the TUI should switch between list and detail views
instead of rendering cramped panes.

## Top Status Bar

The status bar should show:

- active project name;
- active project path, shortened relative to home when useful;
- current view;
- dirty or warning state when known;
- running operation indicator;
- last refresh or indexing state when relevant.

The status bar should not become a command log. It should summarize current
state and route the user to the right detail view.

## Navigation Rail

Primary navigation sections:

- Overview
- Tickets
- Authoring
- Documents
- Runtime
- Reports
- Notifications
- Settings

The rail should support arrow-key navigation and single-key shortcuts only when
the shortcut is visible in help.

The rail should show counts where they help decision-making:

- open tickets;
- blocked tickets;
- active authoring sessions;
- failed runs;
- unread or pending notifications.

Counts should be derived from indexed state and should tolerate stale indexes by
showing an explicit refresh warning.

## Main Area Patterns

### List View

Lists should have headers.

Rows should expose the fields that disambiguate the record:

- tickets: key, status, priority, severity, title, updated;
- authoring sessions: id or slug, status, target, topic, updated;
- documents: type, status, title/topic, path, updated;
- runs: execution id, ticket, status, validation, started, duration;
- reports: type, related record, status, updated.

The selected row should have a stable visual highlight. Do not rely only on
color when marking selection.

### Detail View

Detail views should prioritize:

1. identity;
2. state;
3. current body or summary;
4. related artifacts;
5. available actions;
6. evidence and history;
7. source path.

The source path belongs near the bottom or in a metadata section unless the path
is the record's primary identity.

### Split View

When terminal width allows, the main area should render:

- left pane: list;
- right pane: selected detail.

When width is constrained, pressing Enter from a list opens detail as a new
screen and Back returns to the list.

## Keyboard Model

Required base keys:

- arrow keys or `j`/`k`: move selection;
- Enter: open selected record;
- Esc or Backspace: go back;
- Tab: move focus between panes or sections;
- `/`: search or filter current list;
- `r`: refresh current view;
- `?`: help for the current view;
- `q`: quit or back out from top-level views.

Mutating actions should not be bound to easy accidental keys without a
confirmation step.

## Search and Filtering

Each list view should support local filtering.

Filtering should be visible as a query line and should not destroy the user's
selection if the filter is cleared.

Ticket filters should support the same concepts as CLI filtering over time:

- status;
- priority;
- severity;
- type;
- tags;
- use case;
- project area;
- text query.

Initial implementation may provide text filtering only, but the model should
not block structured filters later.

## Help

The help view should be contextual.

It should show only keys/actions available in the current view and should
distinguish read-only actions from mutating actions.

The help view should not replace proper labels in the UI. Primary screens
should remain understandable without opening help.

## Error Display

Errors should appear in a bottom message area or modal detail panel depending
on severity.

Transient messages should be short. Detailed errors should be openable as
records, especially when they relate to runtime execution, validation, or
artifact parsing.

## Acceptance Criteria

- Lists have headers.
- The active project and current view are always visible.
- Narrow terminals remain usable.
- Selection and available actions are visible without reading docs.
- Mutating actions are harder to trigger accidentally than read-only actions.
- Current view help is available from every screen.

