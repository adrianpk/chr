# Index Status CLI

## Purpose

`chc index status` reports whether the resolved project's derived ticket index
matches canonical Markdown ticket files.

This is a diagnostic command for users and agents. Normal ticket read workflows
should refresh derived state automatically once reconciliation is wired into
those commands.

## Usage

```sh
chc index status
```

The command accepts the standard project resolution flags:

```sh
chc index status --dir /path/to/project
chc index status --dev-root _workspace
```

## Behavior

The command:

- resolves the project through `--dir`, active project, or current directory;
- opens the user-level SQLite ticket index;
- discovers Markdown ticket files under `.chc/tickets`;
- compares discovered files against indexed ticket rows;
- reports missing, stale, orphaned, and invalid ticket state;
- exits with a non-zero status when any index drift exists.

## Output

Clean output includes:

```text
missing: 0
stale: 0
orphaned: 0
invalid: 0
```

When drift exists, the command prints summary counts and detail rows:

```text
missing: 1
stale: 1
orphaned: 1
invalid: 1
missing_ticket: CHC-1-Do-something.md
stale_ticket: CHC-2-Update-flow.md
orphaned_ticket: CHC-3-Removed.md
invalid_ticket: CHC-4-Bad.md ticket frontmatter opening delimiter is required
```

## Exit Status

The command exits successfully only when all counts are zero.

The command exits with an error when any missing, stale, orphaned, or invalid
ticket state is detected.

## Scope

The command does not mutate Markdown ticket files.

The command does not rebuild the index.

The command does not watch the filesystem.
