# Ticket Index Rebuild

## Purpose

Ticket index rebuild reconstructs one project's derived SQLite ticket rows from
canonical Markdown ticket files.

Markdown ticket files remain canonical. Rebuild exists to recover or rehydrate
derived index state.

## Input

The rebuild service receives:

- project id;
- project directory;
- optional indexed timestamp.

The service discovers Markdown tickets under:

```text
<project>/.chr/user/tickets
```

## Behavior

Rebuild:

- discovers Markdown ticket files;
- parses each discovered ticket;
- deletes existing indexed rows for the project;
- upserts each successfully parsed ticket;
- keeps parse errors in the rebuild result.

Parse errors do not stop the whole rebuild. Valid tickets are still indexed.

Infrastructure errors, such as discovery or database failures, stop the rebuild.

## Output

The rebuild result includes:

- `IndexedTickets`: number of successfully indexed tickets;
- `ParseErrors`: ticket files that could not be parsed or validated.

Each parse error includes:

- absolute path;
- relative path under `.chr/user/tickets`;
- underlying error.

## Scope

Rebuild does not mutate Markdown files.

Rebuild does not watch the filesystem.

Rebuild does not decide whether automatic refresh is needed before a command.
