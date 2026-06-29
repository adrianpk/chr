# Slice 6: Acceptance Workflow Docs And Fixture

Status: delivered
Slicing: Separate Acceptance Feature Workflow
Plan: `ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md`
Tracker: `ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md`
Branch: `feat/slice-6-acceptance-workflow-docs-and-fixture`
PR: `#299`

## Purpose

Document the separated acceptance feature workflow and validate that the dialer
fixture remains usable as a resettable testbed.

## Delivered Behavior

The user guide now describes the Specs view `f` action and the CLI command for
acceptance feature generation. Reference docs now identify both generated
feature artifacts and feature-round notes:

- `.chr/user/acceptance/<case>.feature`
- `.chr/user/acceptance/<case>-feature-round.md`

The dialer fixture still resets cleanly and contains the `call-outcome-capture`
use case with metadata and dump context ready for quiz/spec generation.

## Implementation Notes

No runtime behavior changed in this slice. The dialer validation intentionally
stops before model-backed generation. It verifies the resettable fixture state
and the pre-spec guard for acceptance feature generation.

## Contracts Added or Changed

- User Guide documents Specs view `f generate feature`.
- Command reference documents feature-round note paths.
- Workspace layout reference includes acceptance feature artifacts and
  feature-round notes.
- Explanation docs describe feature-round notes as the post-spec clarification
  artifact.

## Files of Interest

- `docs/how-to/user-guide/index.md`
- `docs/reference/commands.md`
- `docs/reference/acceptance-feature-generation.md`
- `docs/reference/project-workspace.md`
- `docs/explanation/document-generation-model.md`

## Validation

- `make qa-reset QA_CASE=dialer QA_REF=002` (pass)
- `go run ./cmd/chr spec list --dir qa/project/dialer` (pass)
- `go run ./cmd/chr acceptance feature generate --dir qa/project/dialer --use-case call-outcome-capture` (expected failure: generated spec is required)
- `go -C qa/project/dialer test ./...` (pass)
- `make verify` (pass)

## Risks and Follow-ups

- `make qa-smoke QA_CASE=dialer QA_REF=002` currently expects the baseline
  dialer fixture to fail, but the fixture tests now pass. That smoke expectation
  is separate from the acceptance feature workflow and was not changed here.
- `godog` execution remains deferred.
