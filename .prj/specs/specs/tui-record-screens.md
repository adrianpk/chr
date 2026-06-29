# TUI Record Screens

Status: initial

Sources:

- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-navigation-and-layout.md`
- `docs/reference/ticket-template.md`
- `docs/reference/authoring-sessions.md`
- `docs/reference/project-workspace.md`

## Purpose

Define the records the TUI must expose and the fields each screen should show.

The TUI should make Chryso's Markdown-first state navigable without forcing
the user to understand every internal path.

## Overview Screen

The overview is the first screen after project resolution.

It should show:

- active project;
- workspace initialization status;
- runtime validation status;
- ticket summary by status;
- blocked or failed work;
- active authoring sessions;
- recent reports;
- latest runtime run;
- next suggested actions.

Suggested next actions should be conservative. They should come from current
state, not from speculative workflow advice.

## Tickets Screen

The tickets screen is the primary work surface.

List columns:

- key;
- status;
- priority;
- severity;
- type;
- title;
- updated.

Detail fields:

- title;
- key;
- status;
- priority;
- severity;
- type;
- tags;
- use case;
- project area;
- branch;
- PR;
- body;
- checklist;
- comments or notes;
- related reports and runs;
- source path.

The checklist should be rendered as checklist state, not as opaque Markdown
text when practical.

Ticket comments should be grouped by timestamp or author/source when metadata
exists.

## Authoring Screen

The authoring screen groups sessions, specs, plans, trackers, dumps, and
quizzes.

The first implementation may focus on sessions and quizzes, but the information
architecture should allow all authoring artifacts to be surfaced.

Session list columns:

- session id or slug;
- status;
- target;
- topic;
- updated.

Session detail fields:

- session identity;
- target artifact type;
- topic;
- status;
- transcript path;
- artifacts path;
- source artifacts;
- generated outputs;
- available promotions;
- source path.

The TUI should avoid making `auth-...` IDs the primary human label when a topic
or slug is available.

## Documents Screen

The documents screen provides a cross-artifact index.

Record families:

- dumps;
- quizzes;
- specs;
- plans;
- trackers;
- reports;
- handoff summaries.

List columns:

- type;
- status;
- title or topic;
- related ticket or plan when available;
- updated;
- path.

Document detail should render Markdown text in a readable pane with metadata
above it.

The TUI does not need to be a full Markdown editor in the first version. Opening
the source file in the user's editor can remain a later integration.

## Runtime Screen

The runtime screen exposes execution capability and evidence.

It should show:

- configured runtime image;
- validation command;
- runtime readiness checks;
- latest validation result;
- latest execution result;
- log and result artifact paths;
- environment readiness warnings when known.

Runtime validation output should use compact sections, not duplicated `ok`
phrasing.

## Reports Screen

Reports should be visible by type and recency.

Useful report families:

- slice reports;
- execution reports;
- loop reports;
- handoff summaries;
- validation evidence.

Reports should link back to their originating ticket, run, loop, slice, plan, or
tracker when metadata allows it.

## Settings Screen

Settings should expose only meaningful choices.

It should not present binary-owned invariants as user-configurable fields.

Show:

- ticket key prefix if supported;
- notification config status;
- runtime image/config where it is actually user-owned;
- project registry state;
- derived index status;
- config file path.

Do not show `.chr/` root or Markdown canonical assets as editable settings.

## Acceptance Criteria

- Each record family has a list and a detail model.
- Lists use headers and field names that match domain vocabulary.
- Source paths are visible but secondary.
- Records can be discovered from semantic fields such as title, topic, status,
  and target.
- The screens do not expose internal invariants as if they were user choices.
