# Slice 5 Ticket New Plan

Status: delivered

Spec: `.prj/specs/specs/ticket-template.md`
Model: `.prj/specs/specs/ticket-template-model.md`
Tracker: `.prj/trackers/slice-5-ticket-new-tracker.md`

## Purpose

Add the first usable ticket primitive: `chr ticket new`.

The command should create a Markdown ticket file under `.chr/tickets/` using the initial ticket template model.

## Scope

This slice covers:

- ticket metadata and template rendering in `internal/chr`,
- generated ticket keys using the configured prefix and existing ticket files,
- project-wise ticket creation under `.chr/tickets/`,
- use-case ticket creation under `.chr/tickets/<use_case>/`,
- compact Markdown ticket body generation,
- `chr ticket new` CLI wiring,
- focused tests for key generation, path selection, and rendered content.

## Task Breakdown

### T5.1 Ticket Model and Renderer

Create the internal ticket model and Markdown renderer.

Expected behavior:

- render YAML frontmatter with `snake_case` fields,
- render compact body headings,
- support generated timestamps supplied by the caller for deterministic tests,
- keep optional metadata fields empty when not provided.

### T5.2 Ticket Repository and Key Generation

Create the file-based ticket repository behavior.

Expected behavior:

- create `.chr/tickets/` when needed,
- derive the next key from existing ticket filenames or metadata,
- use the configured key prefix, defaulting to `CHR`,
- write project-wise tickets directly under `.chr/tickets/`,
- write use-case tickets under `.chr/tickets/<use_case>/`,
- generate filenames as `key-truncated-name.md`.

### T5.3 CLI Command

Expose the behavior through `chr ticket new`.

Expected behavior:

- accept a required ticket name/title,
- accept optional flags for `priority`, `severity`, `type`, `tags`, `use-case`, and `project-area`,
- create the ticket in the current workspace by default,
- support `--dir` consistently with existing workspace behavior,
- print the created ticket path.

## Out of Scope

This slice does not cover:

- ticket status transitions,
- ticket listing,
- ticket loop execution,
- Docker agent runtime,
- SQLite persistence,
- messaging integrations,
- moving tickets between directories.

## Validation

Required checks:

```sh
make verify
```

Manual behavior checks:

```sh
tmpdir="$(mktemp -d)"
go run ./cmd/chr init --dir "$tmpdir"
go run ./cmd/chr ticket new --dir "$tmpdir" "Do something"
test -f "$tmpdir/.chr/tickets/CHR-1-Do-something.md"
go run ./cmd/chr ticket new --dir "$tmpdir" --use-case onboarding "Add setup flow"
test -f "$tmpdir/.chr/tickets/onboarding/CHR-2-Add-setup-flow.md"
```
