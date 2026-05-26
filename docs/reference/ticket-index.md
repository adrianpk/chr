# Ticket Index

## Purpose

The ticket index stores parsed Markdown tickets in the global user SQLite database.

Markdown ticket files remain canonical. The index is derived state used for fast listing, search, daemon loops, and messaging workflows.

## Schema

Indexed tickets are stored in `indexed_tickets`.

Tags are stored separately in `indexed_ticket_tags`.

The ticket row key is:

```text
project_id + path
```

The ticket key is also unique per project:

```text
project_id + key
```

## Repository

The core API is exposed through `chc.TicketIndex`.

Initial operations:

- upsert one indexed ticket;
- list indexed tickets for a project;
- get one indexed ticket by project id and ticket key;
- delete all indexed tickets for a project.

Upserting a ticket replaces its tag rows transactionally.

Indexed ticket rows include Board-facing metadata:

- `synopsis`;
- `board_column`;
- `source_kind`;
- `source_ref`;
- `slice`;
- `use_case`;
- `project_area`;
- classification fields such as status, priority, severity, and type.

The Board uses this derived metadata for column selection, attention detection,
filtering, and selected-ticket detail. Markdown ticket files remain the durable
ticket body.

## Scope

The repository does not parse Markdown files.

The repository does not decide stale detection policy.

The repository does not implement rebuild commands or daemon watcher behavior.

Project-level rebuild behavior is documented in
`docs/reference/ticket-index-rebuild.md`.
