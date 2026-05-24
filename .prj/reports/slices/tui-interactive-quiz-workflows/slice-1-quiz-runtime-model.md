# Slice 1: Quiz Runtime Model

Status: reviewing
Slicing: TUI Interactive Quiz Workflows
Plan: `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
Tracker: `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`
Branch: `feat/slice-1-quiz-runtime-model`
PR: pending

## Purpose

Establish the file-backed interactive quiz runtime model before adding
repositories, service operations, or TUI behavior.

## Delivered Behavior

- Adds quiz runtime types for scope, status, mode, active question, entries,
  coverage topics, Ariadne trail nodes, fast candidates, and runtime state.
- Adds validation for runtime identity, project scope, status, mode, budget,
  active-question lifecycle, question depth, answer/skip invariants, coverage
  counts, trail nodes, candidates, paths, and timestamps.
- Documents early refinements for augmented LLM context, dump plus complete quiz
  promotion to a draft spec, and one formal model request/response boundary.

## Implementation Notes

The slice is intentionally model-only. It does not introduce persistence,
service orchestration, model calls, TUI rendering, or file writes.

The runtime model keeps fast candidates separate from durable quiz entries so
later slices can prevent speculative questions from polluting raw capture,
curated notes, tags, or context retrieval.

## Contracts Added or Changed

- `internal/chc.QuizRuntimeState`
- `internal/chc.QuizScope`
- `internal/chc.QuizQuestion`
- `internal/chc.QuizEntry`
- `internal/chc.QuizCoverageTopic`
- `internal/chc.QuizTrailNode`
- `internal/chc.QuizCandidate`
- `internal/chc.ValidateQuizRuntimeState`
- `internal/chc.ValidateQuizScope`
- `internal/chc.ValidateQuizQuestion`
- `internal/chc.ValidateQuizEntry`
- `internal/chc.ValidateQuizCoverageTopic`
- `internal/chc.ValidateQuizTrailNode`
- `internal/chc.ValidateQuizCandidate`
- `DefaultQuizQuestionBudget`
- `MaxQuizFollowUpDepth`

## Files of Interest

- `internal/chc/quiz_runtime.go`
- `internal/chc/quiz_runtime_test.go`
- `.prj/specs/specs/tui-interactive-quiz-workflow.md`
- `.prj/specs/specs/tui-interactive-quiz-workflow-model.md`
- `.prj/plans/slicing-tui-interactive-quiz-workflows-plan.md`
- `.prj/trackers/slicing-tui-interactive-quiz-workflows-tracker.md`

## Validation

- `go test ./internal/chc`
- `make verify`

## Risks and Follow-ups

- Persistence, service operations, model-boundary payload execution, and TUI
  behavior are intentionally deferred to later slices.
