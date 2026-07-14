# Slice 2: Acceptance Feature Service

Status: delivered
Slicing: Separate Acceptance Feature Workflow
Plan: `ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md`
Tracker: `ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md`
Branch: `feat/slice-2-acceptance-feature-service`
PR: `#295`

## Purpose

Add a dedicated service operation for acceptance feature generation, separated
from spec generation.

## Delivered Behavior

The service can now generate an acceptance feature for a selected scope through
`GenerateAcceptanceFeature`.

The operation builds a model request with project, use-case, dump, quiz, and
spec context. It requires an existing spec before feature generation can run.

The feature model contract returns `generated`, `needs_more_info`, `rejected`,
or `not_applicable`. Generated feature bodies are written to
`.chr/default/acceptance/<case>.feature` only after deterministic static checks
pass. Non-generated statuses return service results without writing a feature
file.

## Implementation Notes

The feature-generation action uses the existing `QuizModelAdapter` boundary with
a new action, output schema, and provider instructions. Codex and OpenAI routing
both understand the action.

Static check failures are converted into a rejected operation result rather than
a hard service error. Contract failures and missing spec context remain errors.

## Contracts Added or Changed

- Added `QuizModelActionGenerateAcceptanceFeature`.
- Added `AcceptanceFeatureStatusNeedsMoreInfo`.
- Added feature output decoding and schema support.
- Added `GenerateAcceptanceFeature` service operation.
- Added service result statuses for generated, needs-more-info, rejected, and
  not-applicable outcomes.

## Files of Interest

- `internal/chr/acceptance_feature.go`
- `internal/chr/quiz_model.go`
- `internal/chr/quiz_model_payload.go`
- `internal/chr/quiz_openai_model.go`
- `internal/chrapp/service_acceptance_feature.go`
- `internal/chrapp/service_acceptance_feature_test.go`

## Validation

- `go test ./internal/chr ./internal/chrapp` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- CLI wiring was delivered in Slice 3.
- Feature-round default confirmation was delivered in Slice 4.
- TUI action wiring was delivered in Slice 5.
