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
open
ready
in_progress
reviewing
blocked
solved
merged
closed
```

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
