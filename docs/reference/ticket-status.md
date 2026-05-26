# Ticket Status

## Purpose

Ticket status commands update the lifecycle status stored in Markdown ticket
frontmatter.

Status lives in ticket metadata, not in the file path.

## CLI

Update a ticket status with:

```sh
chc ticket status <key> <status>
```

The command accepts the standard ticket project flags:

- `--dir` to resolve a project directly;
- `--dev-root` for development-mode user state.

## Status Values

Allowed status values are:

```text
backlog
ready_for_dev
in_progress
ready_for_review
ready_for_merge
done
```

Legacy status values are still accepted for compatibility:

```text
open
ready
reviewing
blocked
solved
merged
closed
```

Board column mapping:

| Status | Board behavior |
| --- | --- |
| `backlog` | Backlog column. |
| `ready_for_dev` | Ready For Dev column. |
| `in_progress` | In Progress column. |
| `ready_for_review` | Ready For Review column. |
| `ready_for_merge` | Ready For Merge column. |
| `done` | Done column. |
| `open` | Legacy alias for Backlog. |
| `ready` | Legacy alias for Ready For Dev. |
| `reviewing` | Legacy alias for Ready For Review. |
| `solved` | Legacy alias for Ready For Merge. |
| `merged`, `closed` | Legacy aliases for Done. |
| `blocked` | Attention view; not part of the normal Board column cycle. |

## Behavior

The command:

- resolves the project through the command project context resolver;
- finds the ticket by key in the user-level ticket index;
- updates `status` in Markdown frontmatter;
- updates `updated_at`;
- writes the modified Markdown ticket file.

## Scope

The command does not rebuild the ticket index.

The command does not enforce transition rules between statuses yet.

The command does not move ticket files between directories.
