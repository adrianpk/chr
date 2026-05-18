# Slice 5 Ticket New Plan

Status: active

Spec: `.prj/specs/specs/ticket-template.md`
Model: `.prj/specs/specs/ticket-template-model.md`
Tracker: `.prj/trackers/slice-5-ticket-new-tracker.md`

## Purpose

Add the first usable ticket primitive: `chc ticket new`.

The command should create a Markdown ticket file under `.chc/tickets/` using the initial ticket template model.

## Scope

This slice covers:

- ticket metadata and template rendering in `internal/chc`,
- generated ticket keys using the configured prefix and existing ticket files,
- project-wise ticket creation under `.chc/tickets/`,
- use-case ticket creation under `.chc/tickets/<use_case>/`,
- compact Markdown ticket body generation,
- `chc ticket new` CLI wiring,
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

- create `.chc/tickets/` when needed,
- derive the next key from existing ticket filenames or metadata,
- use the configured key prefix, defaulting to `CHC`,
- write project-wise tickets directly under `.chc/tickets/`,
- write use-case tickets under `.chc/tickets/<use_case>/`,
- generate filenames as `key-truncated-name.md`.

### T5.3 CLI Command

Expose the behavior through `chc ticket new`.

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
go run ./cmd/chc init --dir "$tmpdir"
go run ./cmd/chc ticket new --dir "$tmpdir" "Do something"
test -f "$tmpdir/.chc/tickets/CHC-1-Do-something.md"
go run ./cmd/chc ticket new --dir "$tmpdir" --use-case onboarding "Add setup flow"
test -f "$tmpdir/.chc/tickets/onboarding/CHC-2-Add-setup-flow.md"
```
