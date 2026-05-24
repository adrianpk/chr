# Slice 2: Quiz Artifact Repository

Status: reviewing
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-2-quiz-artifact-repository`
PR: pending

## Purpose

Add the file-backed artifact repository needed by later service and TUI slices
to persist scoped quiz raw capture, curated quiz notes, and runtime state.

## Delivered Behavior

- Resolves canonical project-wide and use-case-scoped quiz artifact paths under
  `.chc/quizzes/`.
- Resolves canonical runtime state cache paths under `.chc/tmp/quiz-runtime/`.
- Appends raw quiz capture entries for answered or skipped activated
  questions.
- Writes curated quiz notes from distilled decisions, constraints, preferences,
  open questions, and implementation implications.
- Saves and loads runtime JSON state.
- Keeps speculative candidates in runtime state only, out of raw and curated
  Markdown artifacts.

## Implementation Notes

The repository is file-backed and lives in `internal/chc`. It does not call the
model boundary, perform service orchestration, or expose TUI behavior.

The raw and curated renderers are intentionally deterministic so later slices
can reuse them from service operations without duplicating Markdown formatting
in adapters.

## Contracts Added or Changed

- `internal/chc.QuizArtifactRepository`
- `internal/chc.NewQuizArtifactRepository`
- `internal/chc.QuizArtifactLayout`
- `internal/chc.QuizArtifactPaths`
- `internal/chc.QuizCuratedNotes`
- `internal/chc.RenderQuizRawCapture`
- `internal/chc.RenderQuizCuratedNotes`
- `QuizArtifactsDirectory`
- `QuizRuntimeDirectory`

## Files of Interest

- `internal/chc/quiz_artifact_repository.go`
- `internal/chc/quiz_artifact_repository_test.go`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chc`
- `make verify`

## Risks and Follow-ups

- Service operations still need to decide when to append raw capture, refresh
  curated notes, and save runtime state.
- Runtime JSON schema versioning remains deferred unless a later slice needs
  migration behavior.
