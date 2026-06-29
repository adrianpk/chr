# Index Rebuild CLI

## Purpose

`chr index rebuild` rebuilds the resolved project's derived ticket index from
canonical Markdown ticket files.

This is a recovery and rehydration command. Normal ticket read workflows should
not depend on the user remembering to run it.

## Usage

```sh
chr index rebuild
```

The command accepts the standard project resolution flags:

```sh
chr index rebuild --dir /path/to/project
chr index rebuild --dev-root _workspace
```

## Behavior

The command:

- resolves the project through `--dir`, active project, or current directory;
- opens the user-level SQLite ticket index;
- discovers Markdown ticket files under `.chr/user/tickets`;
- deletes existing indexed ticket rows for the resolved project;
- indexes successfully parsed tickets;
- reports parse errors without failing the whole rebuild.

## Output

Successful output includes:

```text
indexed: 1
parse_errors: 0
```

When parse errors exist, the command prints one line per invalid ticket:

```text
parse_error: CHR-1-Bad.md ticket frontmatter opening delimiter is required
```

## Scope

The command does not mutate Markdown ticket files.

The command does not watch the filesystem.

The command does not run agents or ticket loops.
