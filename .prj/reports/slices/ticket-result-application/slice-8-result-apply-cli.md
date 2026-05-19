# Slice 8: Result Apply CLI

Status: delivered
Slicing: Ticket Result Application
Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
Branch: `feat/slice-8-result-apply-cli`
PR: `#48`

## Purpose

Expose result application through a user-facing command.

## Delivered Behavior

The CLI now exposes `chc ticket apply-result <execution-id>`.

The command:

- resolves project context;
- loads one execution record;
- loads the execution result artifact;
- applies ticket status, comment, validation evidence, changed files, and
  blocked reason writers;
- refreshes the affected ticket in the derived SQLite index;
- writes a result application report artifact;
- supports `--dry-run`.

## Implementation Notes

The command uses the result application model and writers delivered by previous
slices. Dry-run builds the application plan and prints it without mutating
ticket Markdown, SQLite index state, or report artifacts.

## Contracts Added or Changed

New CLI command:

- `chc ticket apply-result <execution-id>`

New flag:

- `--dry-run`

## Files of Interest

- `cmd/chc/main.go`
- `cmd/chc/main_test.go`
- `docs/reference/ticket-result-application.md`

## Validation

- `go test ./cmd/chc ./internal/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

Changed-file details are applied when present in the result application model.
The current execution result envelope does not yet carry changed-file payloads.
