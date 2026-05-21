# Slice 5: Authoring Session CLI

Status: delivered
Slicing: Markdown-First Authoring Sessions
Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
Branch: `feat/slice-5-authoring-session-cli`
PR: #95

## Purpose

Expose the Markdown-first authoring session workflow through the CLI.

## Delivered Behavior

The CLI now includes an `author` command family with:

- `author session new`;
- `author session list`;
- `author session show`;
- `author quiz new`;
- `author quiz new --dry-run`.

The commands resolve the current project using the same project-context rules as
other project-local commands. Static quiz creation writes the generated quiz
artifact and stores a session with output and provenance metadata.

## Implementation Notes

The CLI wiring lives in `internal/chcapp`, keeping `cmd/chc/main.go` unchanged
as a dependency assembly entrypoint.

This slice intentionally keeps the command surface compact. It exposes the
foundation needed to create sessions and static quizzes, while leaving artifact
promotion for Slice 6.

## Contracts Added or Changed

New user-facing commands:

```text
chc author session new
chc author session list
chc author session show
chc author quiz new
```

## Files of Interest

- `internal/chcapp/author.go`
- `internal/chcapp/app.go`
- `internal/chcapp/cli_test.go`
- `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chcapp -run 'TestAuthor'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice does not implement promotion from quiz answers into drafts or specs.
It also does not add assisted model-backed authoring or a TUI.
