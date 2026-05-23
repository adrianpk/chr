# Slice 2: Remove Project And Case Indexes

Status: reviewing
Slicing: Markdown-Only Workspace Model
Plan: `.prj/plans/slicing-markdown-only-workspace-model-plan.md`
Tracker: `.prj/trackers/slicing-markdown-only-workspace-model-tracker.md`
Branch: `feat/slice-2-remove-project-case-indexes`
PR: pending

## Purpose

Remove duplicated SQLite persistence for Project and Cases.

After this slice, Project and Case metadata are persisted only as Markdown files
under `.chc/`. SQLite remains available for unrelated runtime domains such as
tickets, executions, loops, daemon state, notifications, and remote commands.

## Delivered Behavior

- Project metadata saves write `.chc/project.md` only.
- Case metadata saves write `.chc/cases/*.md` only.
- Project and Case refresh actions reload Markdown files instead of updating a
  relational index.
- The project/case indexed tables are no longer created by migrations.
- Project/Case repository APIs for SQLite metadata indexes are removed.
- Runtime source hashes remain as in-process write-conflict guardrails.

## Implementation Notes

The Markdown parsing, rendering, validation, and atomic write services remain in
place.

The TUI executors now call `ProjectMetadataService` and `CaseMetadataService`
directly. They no longer open the user SQLite database after Project/Case saves
or refreshes.

The migration set drops the unreleased indexed Project/Case metadata
migrations. Migration-count tests were updated from nine available migrations to
seven.

## Contracts Added or Changed

- Removed `ProjectMetadataRepository` and `IndexedProjectMetadata`.
- Removed `CaseMetadataRepository` and `IndexedCaseMetadata`.
- Removed project/case metadata index migrations.
- Renamed Project/Case TUI reload language from resync to refresh.
- Kept source-hash stale-write detection in the Markdown services.

## Files of Interest

- `internal/chc/project_metadata.go`
- `internal/chc/case_metadata.go`
- `internal/chcapp/app.go`
- `internal/chctui/project_metadata_actions.go`
- `internal/chctui/case_metadata_actions.go`
- `cmd/chc/assets/migration/sqlite/`
- `internal/chc/migration_test.go`
- `internal/chc/migration_status_test.go`

## Validation

- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- Slice 3 must introduce the runtime `WorkspaceGraph` so navigation and context
  selection can use a structured in-memory view instead of direct per-screen
  file loading.
