# Ticket Parser

## Purpose

The ticket parser converts Markdown ticket files into an indexable model.

Markdown remains the canonical ticket source. Parsed tickets are used by future index rebuild and daemon workflows.

## Input

The parser expects a Markdown file with YAML frontmatter delimiters:

```md
---
key: "CHC-1"
title: "Do something"
synopsis: "Implement one focused change."
status: "backlog"
tags: ["uxui"]
---

# Description
```

Missing frontmatter delimiters are rejected.

Ticket metadata is validated before the parser returns an indexable model. See
`docs/reference/ticket-validation.md` for the validation contract.

## Output

The parser returns `chc.IndexedTicket`.

The model includes:

- project id;
- absolute ticket path;
- relative ticket path;
- ticket metadata;
- tags;
- synopsis, Board column, source metadata, and slice metadata when present;
- created and updated timestamps when present;
- file modification time;
- file size;
- content hash;
- indexed timestamp.

## Scope

The parser does not write to SQLite.

The parser does not mutate ticket files.

The parser exists so later index rebuild and watcher code can convert Markdown files into database rows.
