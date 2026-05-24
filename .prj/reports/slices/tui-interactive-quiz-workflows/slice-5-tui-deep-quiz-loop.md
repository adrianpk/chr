# Slice 5: TUI Deep Quiz Loop

Status: reviewing
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-5-tui-deep-quiz-loop`
PR: TBD

## Purpose

Connect the TUI Quizzes screen to the service-owned quiz operations so users
can drive the one-question-at-a-time deep quiz loop from inside CHC.

## Delivered Behavior

- Adds TUI quiz actions for start, resume, answer, skip, complete, cancel, and
  refresh.
- Adds a TUI-native answer editor for the active selected quiz question.
- Dispatches quiz intents through a `QuizActionExecutor` instead of mutating
  runtime state in the TUI.
- Maps TUI quiz actions to application service operations.
- Keeps selected quiz scope after action results reload the quiz read model.
- Renders answer input inside the selected quiz detail panel.
- Adds command-palette entries for quiz loop operations.
- Updates selected source path handling for quiz views.

## Implementation Notes

The TUI still acts as an adapter. It owns transient confirmation and answer
input state, then sends user intent to `internal/chcapp.service` through the
TUI executor adapter.

Start uses deep mode explicitly. Fast candidate buffering remains a later
slice.

## Contracts Added or Changed

- `internal/chctui.QuizActionID`
- `internal/chctui.QuizActionRequest`
- `internal/chctui.QuizActionResult`
- `internal/chctui.QuizActionExecutor`
- `internal/chctui.QuizActionState`
- `internal/chctui.QuizAnswerState`
- `chctui.Options.QuizExecutor`

## Files of Interest

- `internal/chctui/quiz_actions.go`
- `internal/chctui/quizzes.go`
- `internal/chctui/command_palette.go`
- `internal/chctui/model.go`
- `internal/chctui/shell.go`
- `internal/chcapp/tui_adapters.go`
- `internal/chcapp/app.go`
- `internal/chctui/model_test.go`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chctui ./internal/chcapp`
- `make verify`

## Risks and Follow-ups

- Fast mode candidate generation remains in the next slice.
- Ariadne trail recovery and compact branch rendering remain in a later slice.
- Curation and draft spec synthesis remain in the final curation slice.
