# Slice 1: Spec Feature Decoupling

Status: delivered
Slicing: Separate Acceptance Feature Workflow
Plan: `ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md`
Tracker: `ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md`
Branch: `feat/slice-1-spec-feature-decoupling`
PR: `#294`

## Purpose

Separate normal spec generation from Gherkin feature generation so the general
requirements workflow remains use-case centered. This slice removes the previous
spec-plus-feature output path and prepares the codebase for a later dedicated
acceptance feature workflow.

## Delivered Behavior

`GenerateSpec` now writes only the Markdown spec artifact under
`.chc/user/specs/`. It no longer returns feature status or writes
`.chc/user/acceptance/<case>.feature`.

The spec model output contract no longer accepts a `feature` field. Spec
generation remains neutral about the spec body's subject matter.

## Implementation Notes

The old static acceptance feature validation code remains available for later
slices, but spec generation no longer invokes it. Slice 2 can introduce a new
feature-specific schema and service operation at the acceptance feature boundary.

The model payload decoder now treats generated spec output as an exact
`title`, `body`, and `status` object. This rejects accidental or stale model
responses that still include optional feature content.

## Contracts Added or Changed

- `SpecOperationResult` no longer exposes acceptance feature fields.
- Generated spec model output no longer includes `feature`.
- Generated spec model output rejects extra fields.
- Spec generation writes only Markdown specs.

## Files of Interest

- `internal/chcapp/service_spec.go`
- `internal/chc/quiz_model_payload.go`
- `internal/chc/quiz_openai_model.go`
- `internal/chc/spec_generation.go`
- `docs/reference/acceptance-feature-generation.md`

## Validation

- `go test ./internal/chc ./internal/chcapp -run 'Spec|Feature|Gherkin'` (pass)
- `go test ./internal/chc ./internal/chcapp` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- The dedicated acceptance feature service, CLI, feature round, and TUI action
  are still pending in later slices.
- Existing generated feature schema code was removed from the spec model path;
  Slice 2 should add a new schema at the dedicated feature-generation boundary.
