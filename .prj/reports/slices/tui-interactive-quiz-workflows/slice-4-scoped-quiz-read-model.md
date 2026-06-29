# Slice 4: Scoped Quiz Read Model

Status: delivered
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-4-scoped-quiz-read-model`
PR: #186

## Purpose

Add a quiz-specific read model so the TUI can render project-wide and
use-case-scoped quiz state without parsing Markdown or runtime JSON directly.

## Delivered Behavior

- Lists one project-wide quiz scope and one quiz scope per use case.
- Includes runtime-only quiz scopes when runtime state exists without matching
  metadata.
- Distinguishes not-started scopes from active, idle, done, cancelled, and
  blocked runtime states.
- Carries scope, status, mode, artifact paths, budget, counts, updated time,
  active question summary, trail summary, and curated preview.
- Adapts service quiz views into TUI quiz views at startup.
- Renders `F4 Quizzes` from the quiz read model instead of the generic
  document-family list.
- Keeps quiz selection independent from generic document selection.

## Implementation Notes

The read model is built in `internal/chrapp` from service-owned quiz state and
metadata. The TUI receives already-normalized quiz views and only renders them.

Not-started scopes still expose canonical raw, curated, and runtime paths so
later interaction slices can start the correct project-wide or use-case quiz
without recomputing layout in the TUI.

## Contracts Added or Changed

- `internal/chrapp.QuizView`
- `internal/chrapp.QuizViewStatusNotStarted`
- `internal/chrtui.QuizView`
- `internal/chrtui.QuizScope`
- `chrtui.Options.Quizzes`
- `chrtui.Model` scoped quiz selection state

## Files of Interest

- `internal/chrapp/service_quiz_read.go`
- `internal/chrapp/service_quiz.go`
- `internal/chrapp/tui_adapters.go`
- `internal/chrapp/app.go`
- `internal/chrtui/quizzes.go`
- `internal/chrtui/model.go`
- `internal/chrtui/shell.go`
- `internal/chrtui/model_test.go`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chrapp ./internal/chrtui`
- `make verify`

## Risks and Follow-ups

- This slice renders quiz state only. Start, answer, skip, done, cancel, and
  mode actions are connected to the TUI in later slices.
- Curated preview is intentionally compact; full curated note rendering remains
  part of later curation polish.
