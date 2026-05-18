# Ticket Foundation Slicing Plan

Status: active

## Purpose

Plan the `Ticket Foundation` slicing from the feature slicing set map.

This slicing covers the use case:

> A user creates, inspects, comments on, and updates durable project tickets that agents can later execute.

## Source

Primary source:

- `.prj/drafts/chc-feature-slicing-set-map.md`

Related specifications:

- `.prj/specs/specs/ticket-template.md`
- `.prj/specs/specs/ticket-template-model.md`
- `.prj/specs/specs/user-daemon-and-global-index.md`
- `.prj/specs/specs/user-daemon-and-global-index-model.md`

Prerequisite delivered work:

- Repository PR #5: `feat(slice-5): ticket new`
- Repository PR #11: `feat(slice-1): ticket markdown parser`
- Repository PR #12: `feat(slice-2): ticket index repository`
- Repository PR #16: `feat(slice-6): project validation recovery`

The prerequisite PR titles keep their historical review titles. They are not
renumbered by this slicing plan.

## Current State

Already delivered:

- `chc ticket new`;
- ticket Markdown rendering;
- ticket Markdown parsing;
- indexed ticket model;
- SQLite ticket index repository;
- active project context resolution for ticket commands.

Still missing:

- a user-facing ticket list command;
- a user-facing ticket show command;
- ticket comments;
- ticket status transitions;
- ticket lifecycle validation.

## Slicing Scope

This slicing includes:

- ticket listing from the existing index repository;
- ticket inspection from the existing parser/index model;
- comments stored in Markdown ticket files;
- controlled status transitions;
- validation of ticket lifecycle and metadata consistency;
- project context behavior through the existing resolver.

## Out of Scope

This slicing does not include:

- automatic reconciliation between Markdown and SQLite;
- daemon watcher behavior;
- Docker runtime setup;
- agent execution;
- filtered execution loops;
- Telegram or email notifications.

## Slice Numbering

This slicing uses local slice numbering. The slice number is scoped to this
plan, starts at `Slice 1`, and is the number used in branch names, PR titles,
trackers, and reports for this slicing.

Repository PR numbers and previous slicing numbers are review records only.
They do not define this slicing's numbers.

## Slices

### Slice 1: Ticket List

Status: reviewing.

Expected branch: `feat/slice-1-ticket-list`
Expected PR title: `feat(slice-1): ticket list`

Goal:

Expose a user-facing command to list tickets for the resolved project.

Candidate scope:

- resolve project context through the existing resolver;
- read indexed tickets for the selected project;
- support stable default ordering;
- print parseable ticket rows;
- add focused CLI tests.

Expected outcome:

- the user can inspect known tickets without opening the filesystem manually.

### Slice 2: Ticket Show

Status: planned.

Expected branch: `feat/slice-2-ticket-show`
Expected PR title: `feat(slice-2): ticket show`

Goal:

Expose a user-facing command to inspect one ticket.

Candidate scope:

- resolve a ticket by key;
- load ticket data from the parsed/indexed model;
- show metadata and body sections;
- report missing or ambiguous tickets clearly;
- add focused CLI tests.

Expected outcome:

- the user can inspect one ticket by key from the CLI.

### Slice 3: Ticket Comments

Status: planned.

Expected branch: `feat/slice-3-ticket-comments`
Expected PR title: `feat(slice-3): ticket comments`

Goal:

Add durable comments to ticket files.

Candidate scope:

- define comment rendering in the existing Markdown ticket shape;
- append comments without losing existing ticket content;
- preserve created/updated metadata behavior;
- add parser coverage for comments if needed.

Expected outcome:

- tickets can capture review notes, agent notes, and user feedback.

### Slice 4: Ticket Status Transitions

Status: planned.

Expected branch: `feat/slice-4-ticket-status-transitions`
Expected PR title: `feat(slice-4): ticket status transitions`

Goal:

Add controlled status updates for tickets.

Candidate scope:

- define allowed status values;
- update ticket Markdown metadata;
- update `updated_at`;
- reject unknown statuses;
- keep filesystem writes scoped to the resolved project.

Expected outcome:

- tickets can move through the initial lifecycle without manual frontmatter editing.

### Slice 5: Ticket Lifecycle Validation

Status: planned.

Expected branch: `feat/slice-5-ticket-lifecycle-validation`
Expected PR title: `feat(slice-5): ticket lifecycle validation`

Goal:

Validate ticket metadata and lifecycle consistency.

Candidate scope:

- validate required fields;
- validate status, priority, severity, type, and tag shapes;
- report parse or lifecycle errors clearly;
- prepare the boundary for automatic reconciliation.

Expected outcome:

- ticket commands can fail clearly when ticket files are malformed or inconsistent.

## Acceptance Criteria

The slicing is complete when:

- tickets can be listed for the resolved project;
- one ticket can be shown by key;
- comments can be appended durably;
- status can be updated through a controlled command;
- ticket metadata and lifecycle errors are reported clearly;
- all commands use the existing project context resolver;
- validation passes with `make verify`.

## Validation

Required validation for every implementation slice in this slicing:

```sh
make verify
```

Manual smoke checks for the completed slicing:

```sh
make install
tmpdir="$(mktemp -d)"
chc init --dir "$tmpdir"
chc ticket new --dir "$tmpdir" "Do something"
chc ticket list --dir "$tmpdir"
chc ticket show --dir "$tmpdir" CHC-1
```

## Risks

- Listing from SQLite can be stale until automatic reconciliation exists.
- Manual Markdown edits can make parser and index state diverge.
- Comments and status updates must preserve existing ticket body content.
- Ticket lifecycle validation must not become stricter than the current template can satisfy.

## Notes

- Automatic reconciliation is the next slicing after ticket visibility and lifecycle commands are usable.
- Manual index recovery commands remain maintenance tools, not normal workflow requirements.
