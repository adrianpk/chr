# Slice 6: Fast Quiz Candidates

Status: delivered
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-6-fast-quiz-candidates`
PR: #188

## Purpose

Add fast-mode candidate buffering while keeping activated questions, raw
capture, and TUI actions inside the existing service-owned quiz flow.

## Delivered Behavior

- Extends the formal quiz model response with validated candidate questions.
- Stores fast candidates in runtime state only.
- Activates buffered candidates as normal questions before making another
  model request.
- Keeps unactivated candidates out of raw capture.
- Clears candidates when switching modes, completing, or cancelling a quiz.
- Adds TUI commands to switch the selected quiz between deep and fast mode.
- Shows candidate count in the quiz detail panel.
- Routes mode changes through the quiz action executor and service operation.

## Implementation Notes

Fast mode remains an internal service/runtime optimization. The TUI exposes the
mode switch but does not render a candidate tree or let the user select from a
batch. Activated candidates become ordinary active questions, so answer and
skip semantics stay unchanged.

If no buffered candidate exists, generation still goes through the formal model
request/response boundary.

## Contracts Added or Changed

- `internal/chr.QuizModelResponse.Candidates`
- `internal/chrapp.QuizView.CandidateCount`
- `internal/chrtui.QuizView.CandidateCount`
- `internal/chrtui.QuizActionDeepMode`
- `internal/chrtui.QuizActionFastMode`
- `internal/chrtui.QuizActionRequest.Mode`

## Files of Interest

- `internal/chr/quiz_model.go`
- `internal/chrapp/service_quiz.go`
- `internal/chrapp/service_quiz_read.go`
- `internal/chrapp/tui_adapters.go`
- `internal/chrtui/quiz_actions.go`
- `internal/chrtui/quizzes.go`
- `internal/chrtui/command_palette.go`
- `internal/chrapp/service_quiz_test.go`
- `internal/chrtui/model_test.go`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chr ./internal/chrapp ./internal/chrtui`
- `make verify`

## Risks and Follow-ups

- Candidate activation is sequential and internal; explicit candidate browsing
  remains out of scope.
- Ariadne trail branch return and richer recovery semantics remain in the next
  slice.
- Curated notes and draft spec synthesis remain in the final curation slice.
