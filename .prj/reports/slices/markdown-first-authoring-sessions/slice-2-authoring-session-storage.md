# Slice 2: Authoring Session Storage

Status: delivered
Slicing: Markdown-First Authoring Sessions
Plan: `.prj/plans/slicing-markdown-first-authoring-sessions-plan.md`
Tracker: `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`
Branch: `feat/slice-2-authoring-session-storage`
PR: #92

## Purpose

Persist authoring sessions as project-local Markdown artifacts after the
side-effect-free session model from Slice 1.

## Delivered Behavior

The repository now has a filesystem-backed authoring session repository that
can:

- derive the session directory layout under `.chc/authoring/sessions/<id>/`;
- create `session.md`, `transcript.md`, and `artifacts.md`;
- render `session.md` with frontmatter plus Markdown tables for inputs,
  outputs, and provenance;
- parse `session.md` back into an `AuthoringSession`;
- load one session by ID;
- list sessions from the project-local authoring directory;
- reject duplicate session creation without overwriting existing files.

## Implementation Notes

Storage remains project-local and Markdown-first. The repository writes durable
files that a user can inspect and edit. Parsing currently preserves the session
contract fields needed by the model: scalar session metadata, artifact
references, and provenance links.

`transcript.md` and `artifacts.md` are created as stable sibling files for later
slices. This slice does not populate adaptive transcripts or generated artifact
inventories beyond the initial placeholders.

## Contracts Added or Changed

New internal contracts:

- `AuthoringSessionRepository`
- `AuthoringSessionPaths`
- `AuthoringSessionLayout`
- `RenderAuthoringSession`
- `ParseAuthoringSession`
- `ErrAuthoringSessionNotFound`

## Files of Interest

- `internal/chc/authoring_session_repository.go`
- `internal/chc/authoring_session_repository_test.go`
- `.prj/trackers/slicing-markdown-first-authoring-sessions-tracker.md`

## Validation

- `go test ./internal/chc -run 'TestAuthoringSessionRepository|TestRenderAndParseAuthoringSession'`: pass
- `make verify`: pass

## Risks and Follow-ups

This slice stores the session contract only. It does not define canonical
artifact destinations, generate quizzes, expose CLI commands, or promote
artifacts.
