# Slice 7: Ariadne Trail Recovery

Status: delivered
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-7-ariadne-trail-recovery`
PR: #189

## Purpose

Persist and render Ariadne-style quiz branch state so the service can recover
active quiz paths after resume and return to broader coverage after a bounded
follow-up branch closes.

## Delivered Behavior

- Adds a runtime branch return reason for max-depth branch closure.
- Rebuilds the active trail whenever a question becomes active.
- Keeps trunk questions as a one-node trail.
- Extends follow-up questions by parent id when the parent exists in the trail.
- Clears branch trail after answering a maximum-depth follow-up.
- Activates the next trunk question with a fresh trail after branch return.
- Exposes branch return reason through service and TUI quiz read models.
- Renders branch return reason in the selected quiz detail panel.

## Implementation Notes

The service remains responsible for trail recovery. The TUI only renders
service-provided trail summary and branch return state.

Depth enforcement continues to use the existing quiz validation contract:
questions and candidates are valid only between depth 1 and
`MaxQuizFollowUpDepth`.

## Contracts Added or Changed

- `internal/chc.QuizBranchReturnReasonMaxDepth`
- `internal/chc.QuizRuntimeState.BranchReturnReason`
- `internal/chcapp.QuizView.BranchReturnReason`
- `internal/chctui.QuizView.BranchReturnReason`

## Files of Interest

- `internal/chc/quiz_runtime.go`
- `internal/chcapp/service_quiz.go`
- `internal/chcapp/service_quiz_read.go`
- `internal/chcapp/service_quiz_test.go`
- `internal/chctui/quizzes.go`
- `internal/chctui/model_test.go`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chc ./internal/chcapp ./internal/chctui`
- `make verify`

## Risks and Follow-ups

- This slice records and displays branch recovery state; curated notes and
  draft spec synthesis remain in the curation slice.
- The service uses parent id to rebuild existing trails. If a provider returns
  an unknown parent id, the active trail falls back to the current question.
