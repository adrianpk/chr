# Slice 3: Quiz Service Operations

Status: delivered
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-3-quiz-service-operations`
PR: #185

## Purpose

Expose service-owned quiz operations over the application service boundary so
later adapters can drive project-wide and use-case-scoped quiz flows without
owning persistence, scope resolution, or model request assembly.

## Delivered Behavior

- Adds a formal quiz model request and response boundary for next-question
  generation.
- Adds service operations for list, start, resume, answer, skip, complete,
  cancel, and mode switching.
- Defaults new quizzes to deep mode when no explicit mode is requested.
- Resolves project-wide and use-case-scoped quiz runtime and artifact paths
  through the file-backed quiz repository.
- Appends raw capture when answers or skips are recorded.
- Saves runtime state after service mutations.
- Uses the service-owned active question for answer and skip operations.
- Builds model-backed question-generation context from agent protocol,
  project metadata, scoped use-case metadata, project dumps, and matching
  use-case dumps.

## Implementation Notes

The service remains the orchestration boundary. TUI and future adapters should
send quiz intent to `internal/chrapp.service` and render the returned state.
They should not write quiz artifacts, parse dump context, or construct model
requests directly.

The model boundary is intentionally adapter-based. This slice adds semantic
request and response payload validation, but it does not bind quiz generation
to a concrete provider.

## Contracts Added or Changed

- `internal/chr.QuizModelRequest`
- `internal/chr.QuizModelResponse`
- `internal/chr.QuizModelAdapter`
- `internal/chr.InvokeQuizModel`
- `internal/chrapp.QuizListResult`
- `internal/chrapp.QuizOperationResult`
- `internal/chrapp.QuizStartRequest`
- `internal/chrapp.QuizResumeRequest`
- `internal/chrapp.QuizAnswerRequest`
- `internal/chrapp.QuizSkipRequest`
- `internal/chrapp.QuizCompleteRequest`
- `internal/chrapp.QuizCancelRequest`
- `internal/chrapp.QuizModeRequest`

## Files of Interest

- `internal/chr/quiz_model.go`
- `internal/chr/quiz_model_test.go`
- `internal/chrapp/service.go`
- `internal/chrapp/service_quiz.go`
- `internal/chrapp/service_quiz_test.go`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chr ./internal/chrapp`
- `make verify`

## Risks and Follow-ups

- Fast mode currently records the mode preference; candidate buffering remains
  in a later slice.
- Curated quiz notes and draft spec synthesis remain in a later slice.
- The concrete model provider adapter remains outside this slice.
