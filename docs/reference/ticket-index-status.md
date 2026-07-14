# Ticket Index Status

## Purpose

Ticket index status compares canonical Markdown ticket files with the derived
SQLite ticket index for one project.

It reports drift without mutating Markdown files or database rows.

## Input

The status checker receives:

- project id;
- project directory;
- optional indexed timestamp for parsing discovered files.

## Status Categories

`Missing` means a Markdown ticket file exists but no indexed row exists for its
relative path.

`Stale` means a Markdown ticket file exists and has an indexed row, but the
indexed file size or content hash no longer matches the parsed file.

`Orphaned` means an indexed row exists but the Markdown file no longer exists
under `.chr/default/tickets`.

`Invalid` means a Markdown ticket file exists but cannot be parsed or validated.

## Fingerprint Comparison

Status uses deterministic file fingerprints:

- file size;
- content hash.

File modification time is still stored in the index, but it is not used as the
primary stale check because timestamp precision can vary across filesystems and
SQLite serialization.

## Scope

Status does not rebuild the index.

Status does not mutate ticket files.

Status does not watch the filesystem.
