# Slice 1: Plan Tracking Codex Backend

Status: reviewing
Slicing: Codex Artifact Generation
Plan: `.prj/plans/slicing-codex-artifact-generation-plan.md`
Tracker: `.prj/trackers/slicing-codex-artifact-generation-tracker.md`
Branch: `feat/slice-1-plan-tracking-codex-backend`
PR: pending

## Purpose

Route plan tracking synthesis through the shared Codex runner while preserving
the existing Go-owned tracking and ticket materialization flow.

## Delivered Behavior

- Added Codex output schema support for `QuizModelActionMaterializePlanTracking`.
- Added `CHR_PLAN_TRACKING_GENERATION_BACKEND=codex` routing for plan tracking
  model calls.
- Preserved existing `MaterializePlanTracking` ownership of tracking writes and
  ticket creation.
- Added tests that plan tracking can use the Codex model adapter.
- Added tests that invalid generated tracking payloads do not write tracking or
  ticket artifacts.

## Implementation Notes

The Codex adapter now accepts the existing plan tracking strict JSON schema and
parses the response through the same `quizModelResponseFromOutput` path as the
direct model adapter.

The service routing remains opt-in for this slice. Setting
`CHR_PLAN_TRACKING_GENERATION_BACKEND=codex` routes only
`QuizModelActionMaterializePlanTracking` to Codex; other actions continue to use
their configured adapters.

Codex is still limited to synthesis. It returns structured tracking JSON. Go
validates the response, writes `.chr/user/plans/*-tracking.md`, and creates
ticket files.

## Contracts Added or Changed

- `CodexQuizModelAdapter` supports `QuizModelActionMaterializePlanTracking`.
- `newQuizModelAdapter` recognizes `CHR_PLAN_TRACKING_GENERATION_BACKEND=codex`.

## Files of Interest

- `internal/chr/quiz_codex_model.go`
- `internal/chr/quiz_codex_model_test.go`
- `internal/chrapp/service.go`
- `internal/chrapp/service_test.go`
- `internal/chrapp/service_plan_test.go`

## Validation

- `gofmt -w internal/chr/quiz_codex_model.go internal/chr/quiz_codex_model_test.go internal/chrapp/service.go internal/chrapp/service_test.go internal/chrapp/service_plan_test.go` - pass
- `go test ./internal/chr ./internal/chrapp` - pass
- `make verify` - pass

## Risks and Follow-ups

- Slice 2 should harden the generated ticket payload boundary further.
- Default routing remains deferred to Slice 3.
