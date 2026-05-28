# Ticket Index Refresh

## Purpose

Ticket read commands refresh the derived ticket index before reading indexed
ticket rows.

This keeps normal CLI usage aligned with canonical Markdown ticket files without
requiring users to remember `chc index rebuild`.

## Refreshed Commands

The following commands refresh the resolved project's ticket index before they
read from the index:

- `chc ticket list`
- `chc ticket show`
- `chc ticket comment`
- `chc ticket status`

## Behavior

For the resolved project, each command:

- opens the user-level SQLite ticket index;
- discovers Markdown tickets under `.chc/user/tickets`;
- rebuilds indexed rows for that project from canonical files;
- stops before reading ticket rows when reconciliation blockers exist.

Refresh is scoped to the resolved project only. It does not scan or mutate other
registered projects.

## Blockers

If one or more ticket files cannot be parsed, reconciliation returns blockers
and the command exits with an error.

The error includes the blocker count and the first invalid ticket path:

```text
refresh ticket index: 1 blockers: CHC-1-Bad.md ticket frontmatter opening delimiter is required
```

## Scope

Automatic refresh does not watch the filesystem.

Automatic refresh does not run agents or ticket loops.

Manual `chc index rebuild` remains available as an explicit recovery command.
