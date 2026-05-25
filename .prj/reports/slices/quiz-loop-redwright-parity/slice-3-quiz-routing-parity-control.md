# Slice 3: Quiz Routing Parity Control

Status: delivered
Slicing: Quiz Loop RedWright Parity
Plan: `.prj/plans/slicing-quiz-loop-redwright-parity-plan.md`
Tracker: `.prj/trackers/slicing-quiz-loop-redwright-parity-tracker.md`
Branch: `feat/slice-3-quiz-routing-parity-control`
PR: #195

## Purpose

Align quiz branch routing with the intended one-question-at-a-time behavior.

This slice focuses on branch closure semantics after skipped questions and on
preserving trunk recovery when runtime state has already decided a branch should
not continue.

## Delivered Behavior

- Skipping an active quiz question now closes the current branch.
- The next generated question after a skip starts from trunk coverage instead
  of continuing the skipped follow-up path.
- Runtime depth calculation keeps trunk depth when the state has no active
  branch trail.
- Existing max-depth branch recovery remains unchanged.

## Implementation Notes

`updateQuizStateAfterEntry` now treats `QuizEntryStatusSkipped` as a branch
return condition. It clears the branch trail and records a skipped branch-return
reason before the next model request is generated.

`activateQuizQuestion` still clears branch return reason after the next active
question is installed, so the visible active state does not remain stuck in a
transient routing reason.

## Contracts Added or Changed

- `QuizBranchReturnReasonSkipped` records skip-driven branch closure.
- A skipped follow-up question must not force the next active question to stay
  on that follow-up branch.
- A branch-return state with no active trail must resolve model follow-up depth
  to runtime trunk depth.

## Files of Interest

- `internal/chc/quiz_runtime.go`
  - adds skipped branch return reason
- `internal/chcapp/service_quiz.go`
  - returns skipped branches to trunk coverage
- `internal/chcapp/service_quiz_test.go`
  - covers skip-driven trunk recovery
- `internal/chc/quiz_model_test.go`
  - covers runtime depth after branch return

## Validation

- `go test ./internal/chc ./internal/chcapp`
- `git diff --check`
- `make verify`

## Risks and Follow-ups

- Slice 4 still needs manual lab validation against representative project and
  use-case quiz sessions.
