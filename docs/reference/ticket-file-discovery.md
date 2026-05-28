# Ticket File Discovery

## Purpose

Ticket file discovery enumerates canonical Markdown ticket files for one managed
project.

Discovery is a core primitive for automatic reconciliation between project-local
Markdown files and the derived user-level SQLite ticket index.

## Input

Discovery receives a project directory.

The ticket directory is:

```text
<project>/.chc/user/tickets
```

## Output

Discovery returns `chc.DiscoveredTicketFile` values.

Each discovered file includes:

- `Path`: absolute or caller-provided filesystem path to the Markdown file;
- `RelativePath`: slash-separated path relative to `.chc/user/tickets`.

Results are sorted by `RelativePath`.

## Included Files

Discovery includes files with the `.md` extension below `.chc/user/tickets`,
including nested use-case directories.

Discovery ignores directories and non-Markdown files.

## Missing Directory

Missing `.chc/user/tickets` is an error.

Managed projects are expected to contain the ticket directory created by
`chc init`.

## Scope

Discovery does not parse ticket files.

Discovery does not write to SQLite.

Discovery does not decide whether indexed ticket rows are stale.
