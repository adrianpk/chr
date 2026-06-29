# Slice 12 Ticket Index Repository Plan

Status: delivered

Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`
Tracker: `.prj/trackers/slice-12-ticket-index-repository-tracker.md`

## Purpose

Persist parsed Markdown tickets into the global SQLite index.

This slice adds the ticket index schema and repository. It does not add a CLI rebuild command or daemon watcher yet.

## Scope

This slice covers:

- SQLite migration for `indexed_tickets` and `indexed_ticket_tags`,
- ticket index repository interface and SQLite implementation,
- upserting parsed `IndexedTicket` values,
- replacing tags for an indexed ticket,
- listing tickets by project,
- deleting stale project tickets before rebuild-style writes,
- focused repository tests.

## Task Breakdown

### T12.1 Ticket Index Migration

Add the SQLite migration for indexed tickets and ticket tags.

Expected behavior:

- store parsed ticket metadata,
- store file state used for stale detection,
- store tags in a separate table,
- keep project/ticket uniqueness explicit.

### T12.2 Ticket Index Repository

Add the SQLite-backed ticket index repository.

Expected behavior:

- run migrations before use,
- upsert indexed tickets,
- replace tags transactionally,
- list indexed tickets for a project in stable order,
- delete indexed tickets for a project.

### T12.3 Validation and Documentation

Document the ticket index repository and validate the slice.

Expected behavior:

- document index schema and repository behavior,
- update tracker,
- run `make verify`.

## Out of Scope

This slice does not cover:

- `chr ticket index rebuild`,
- daemon watcher,
- Telegram integration,
- ticket search CLI,
- stale detection policy beyond persisted file state.

## Validation

Required checks:

```sh
make verify
```
