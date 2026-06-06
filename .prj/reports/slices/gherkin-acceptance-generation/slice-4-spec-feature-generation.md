# Slice 4: Spec Feature Generation

Status: reviewing
Slicing: Gherkin Acceptance Generation
Plan: `ops/plans/2026-06-06-gherkin-acceptance-generation-plan.md`
Tracker: `ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md`
Branch: `feat/slice-4-spec-feature-generation`
PR: pending

## Purpose

Generate and persist a linked Gherkin feature artifact alongside generated
specs when readiness and static checks allow it.

## Delivered Behavior

Spec generation model output now supports an optional `feature` object with
`generated`, `rejected`, or `not_applicable` status.

When the model returns a generated feature, the service writes
`.chc/user/acceptance/<scope>.feature` only after deterministic static checks
pass. Invalid generated features do not fail spec generation and are reported
as rejected.

Spec generation results and generation logs now include feature generation
status and the feature path when a feature is written.

## Implementation Notes

The Markdown spec remains canonical and is always written when spec generation
succeeds. Feature generation is an additional artifact path, not a replacement
for the spec.

The service writes features only after the spec path exists, so the feature
artifact can carry spec traceability and provenance.

## Contracts Added or Changed

- Added optional `feature` to generated spec model output.
- Added `AcceptanceFeatureGeneration`.
- Added `WriteAcceptanceFeature`.
- Added feature status and path to `SpecOperationResult`.
- Added feature generation status logs for generated, rejected, skipped, and
  final reported statuses.

## Files of Interest

- `internal/chc/acceptance_feature.go`
- `internal/chc/quiz_model_payload.go`
- `internal/chc/quiz_openai_model.go`
- `internal/chc/quiz_draft_spec.go`
- `internal/chcapp/service_spec.go`

## Validation

- `go test ./internal/chc ./internal/chcapp -run 'GeneratedSpecFeatureSchema|SpecGenerationInstructionsDescribeOptionalFeature|GeneratedSpecOutputCarriesFeature|WriteAcceptanceFeature|ServiceGenerateSpec.*Feature|ServiceGenerateSpecRejectsInvalidGeneratedFeature|ServiceGenerateSpecReportsNotApplicableFeature'` (pass)
- `go test ./internal/chc` (pass)
- `go test ./internal/chcapp` (pass)

## Risks and Follow-ups

- Feature files are statically checked but not executed; `godog` integration
  remains deferred.
- The model may still produce rejected feature content; the service preserves
  the spec and reports rejection rather than treating it as acceptance evidence.
