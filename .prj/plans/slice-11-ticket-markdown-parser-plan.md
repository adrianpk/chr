# Slice 11 Ticket Markdown Parser Plan

Status: delivered

Spec: `.prj/specs/specs/ticket-template.md`
Index Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`
Tracker: `.prj/trackers/slice-11-ticket-markdown-parser-tracker.md`

## Purpose

Add the core parser that converts Markdown ticket files into an indexable model.

The parser is required before building the SQLite ticket index and index rebuild commands.

## Scope

This slice covers:

- `chr.IndexedTicket` model,
- Markdown frontmatter parsing for ticket metadata,
- file stat capture for path, relative path, size, mtime, and content hash,
- parser tests with temporary ticket files,
- no ticket index persistence yet.

## Task Breakdown

### T11.1 Indexed Ticket Model

Add the core model representing a ticket row derived from Markdown.

Expected behavior:

- include project id,
- include absolute and relative paths,
- include ticket metadata fields,
- include file mtime, size, content hash, and indexed timestamp.

### T11.2 Ticket Markdown Parser

Add the parser that reads a Markdown ticket file and returns `IndexedTicket`.

Expected behavior:

- require YAML frontmatter delimiters,
- parse string metadata fields,
- parse `tags: [...]` lists produced by the current renderer,
- parse created/updated timestamps when present,
- compute a content hash,
- preserve enough file metadata for stale detection.

### T11.3 Validation and Documentation

Document ticket parsing behavior and validate the slice.

Expected behavior:

- document parser contract,
- update tracker,
- run `make verify`.

## Out of Scope

This slice does not cover:

- SQLite ticket index persistence,
- index rebuild command,
- daemon watcher,
- Telegram integration,
- status mutation commands.

## Validation

Required checks:

```sh
make verify
```
