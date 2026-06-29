# Slice 2: Slice Detail Expansion

Status: delivered
Slicing: Staged Plan Generation
Plan: `.prj/plans/slicing-staged-plan-generation-plan.md`
Tracker: `.prj/trackers/slicing-staged-plan-generation-tracker.md`
Branch: `feat/slice-2-slice-detail-expansion`
PR: #238

## Purpose

Introduce the one-slice detail expansion contract without connecting staged
generation to the final `GeneratePlan` runtime path.

## Delivered Behavior

- Added a `plan_slice_detail_generation` model action.
- Added structured slice detail response models and validation.
- Added strict payload decoding for one-slice detail responses.
- Added an OpenAI JSON schema and provider instructions for slice detail
  expansion.
- Added service request construction that includes the full outline and target
  slice context sections.

## Implementation Notes

Slice detail validation accepts only details for the target outline slice. It
rejects unknown slice IDs, renamed slices, and branch metadata that diverges
from the authoritative outline.

The service request builder reuses the same planning context used by current
plan generation, then appends `plan_outline` and `target_slice` sections for
the detail expansion call.

## Contracts Added or Changed

- `QuizModelActionGeneratePlanSliceDetail`
- `QuizModelResponse.SliceDetail`
- `GeneratedPlanSliceDetail`
- `GeneratedPlanSliceTask`
- `GeneratedPlanSliceTicketMetadata`
- `GeneratedPlanSliceBranchMetadata`
- `QuizModelPlanSliceDetailPayload`
- OpenAI response format name: `generated_plan_slice_detail`

## Files of Interest

- `internal/chr/plan_slice_detail.go`
- `internal/chr/plan_outline.go`
- `internal/chr/quiz_model.go`
- `internal/chr/quiz_model_payload.go`
- `internal/chr/quiz_openai_model.go`
- `internal/chrapp/service_plan.go`
- `internal/chr/quiz_model_test.go`
- `internal/chrapp/service_plan_test.go`

## Validation

- `gofmt -w internal/chr/quiz_model.go internal/chr/plan_outline.go internal/chr/plan_slice_detail.go internal/chr/quiz_model_payload.go internal/chr/quiz_openai_model.go internal/chr/quiz_model_test.go internal/chrapp/service_plan.go internal/chrapp/service_plan_test.go` - pass
- `go test ./internal/chr ./internal/chrapp` - pass
- `go test ./...` - pass
- `git diff --check` - pass
- `make verify` - fail: `lint-strict` reports pre-existing lint debt in existing test, trace, service, and TUI files.

## Risks and Follow-ups

- The complete repository gate remains blocked by existing lint-strict debt.
- Slice 3 must connect outline plus detail responses into deterministic final
  Markdown assembly without writing partial plans on staged generation failure.
