# Slice 1: Outline Generation Contract

Status: reviewing
Slicing: Staged Plan Generation
Plan: `.prj/plans/slicing-staged-plan-generation-plan.md`
Tracker: `.prj/trackers/slicing-staged-plan-generation-tracker.md`
Branch: `feat/slice-1-outline-generation-contract`
PR: pending

## Purpose

Introduce the staged plan generation outline contract without changing the
current `GeneratePlan` runtime behavior.

## Delivered Behavior

- Added a `plan_outline_generation` model action.
- Added a structured plan outline response model and validation.
- Added strict payload decoding for outline responses.
- Added an OpenAI JSON schema and provider instructions for outline generation.
- Added service request construction that reuses plan generation context.

## Implementation Notes

The outline is validated as the authoritative slice map for later staged
generation work. Slice IDs must be sequential from `Slice 1`, and every slice
must include branch, PR, report, complexity, rationale, objective, dependency,
and implementation-boundary metadata.

The service request builder now shares plan-generation request assembly between
the existing full-plan action and the new outline action. The existing full-plan
request ID, strategy, and runtime path remain unchanged.

## Contracts Added or Changed

- `QuizModelActionGeneratePlanOutline`
- `QuizModelResponse.PlanOutline`
- `GeneratedPlanOutline`
- `GeneratedPlanOutlineSlice`
- `GeneratedPlanOutlineDependency`
- `QuizModelPlanOutlinePayload`
- OpenAI response format name: `generated_plan_outline`

## Files of Interest

- `internal/chc/plan_outline.go`
- `internal/chc/quiz_model.go`
- `internal/chc/quiz_model_payload.go`
- `internal/chc/quiz_openai_model.go`
- `internal/chcapp/service_plan.go`
- `internal/chc/quiz_model_test.go`
- `internal/chcapp/service_plan_test.go`

## Validation

- `gofmt -w internal/chc/quiz_model.go internal/chc/quiz_model_payload.go internal/chc/quiz_openai_model.go internal/chc/plan_outline.go internal/chc/quiz_model_test.go internal/chcapp/service_plan.go internal/chcapp/service_plan_test.go` - pass
- `go test ./internal/chc ./internal/chcapp` - pass
- `go test ./...` - pass
- `git diff --check` - pass
- `make verify` - fail: `lint-strict` reports pre-existing lint debt in `internal/chc/authoring_contextual_quiz_test.go`, `internal/chc/llm_trace.go`, `internal/chc/quiz_model_test.go`, `internal/chc/workspace_context_test.go`, `internal/chcapp/service_spec_test.go`, `internal/chcapp/service_test.go`, `internal/chctui/model.go`, and `internal/chctui/shell.go`.

## Risks and Follow-ups

- The complete repository gate remains blocked by existing lint-strict debt.
- Slice 2 must consume this outline contract without allowing detail generation
  to reorder or rename outline slices.
