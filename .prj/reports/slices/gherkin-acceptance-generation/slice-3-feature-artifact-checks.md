# Slice 3: Feature Artifact Checks

Status: delivered
Slicing: Gherkin Acceptance Generation
Plan: `ops/plans/2026-06-06-gherkin-acceptance-generation-plan.md`
Tracker: `ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md`
Branch: `feat/slice-3-feature-artifact-checks`
PR: `#291`

## Purpose

Add local contracts and deterministic static checks for generated Gherkin
feature artifacts before later slices write them alongside specs.

## Delivered Behavior

ChrysoCoord now has a canonical acceptance feature artifact layout under
`.chc/user/acceptance/<scope>.feature`.

Generated feature artifacts can be checked without executing `godog`. Static
checks validate Feature/Scenario structure, require each scenario to include
Given, When, and Then steps, require traceability tags, reject open-question
content, reject vague steps, and reject internal implementation details.

## Implementation Notes

The validator is intentionally standard-library only. It performs deterministic
line-oriented checks rather than executing feature files or generating step
definitions.

Feature-level traceability tags are inherited by scenarios. Scenario-level tags
can also satisfy traceability when they use supported prefixes.

## Contracts Added or Changed

- Added `AcceptanceFeatureArtifact`.
- Added `AcceptanceFeaturePaths`.
- Added `AcceptanceFeatureLayout`.
- Added `ValidateAcceptanceFeatureArtifact`.
- Added `AcceptanceFeatureCheckResult`.
- Added `AcceptanceFeatureCheckFinding`.
- Added `CheckAcceptanceFeatureArtifact`.

## Files of Interest

- `internal/chc/acceptance_feature.go`
- `internal/chc/acceptance_feature_test.go`

## Validation

- `go test ./internal/chc -run 'AcceptanceFeatureLayout|ValidateAcceptanceFeatureArtifact'` (pass)
- `go test ./internal/chc -run 'AcceptanceFeature|CheckAcceptanceFeature'` (pass)
- `go test ./internal/chc` (pass)

## Risks and Follow-ups

- The static checker is intentionally conservative and does not replace a future
  `godog` execution path.
- Later slices still need to connect checked feature artifacts to spec
  generation and workspace writes.
