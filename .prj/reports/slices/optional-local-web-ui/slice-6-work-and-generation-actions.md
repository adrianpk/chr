# Slice 6: Work And Generation Actions

Status: implementing

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-6-work-and-generation-actions`

Expected PR Title: `feat(slice-6): work and generation actions`

PR: pending

## Action Inventory

Approved first Web UI generation actions:

- `spec.generate`
  - Service: `GenerateSpec`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: generated spec artifact.
- `plan.generate`
  - Service: `GeneratePlan`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: generated plan artifact when no canonical plan exists.
- `plan.regenerate`
  - Service: `RegeneratePlan`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: archives/replaces generated plan.
- `plan.tracking.create`
  - Service: `MaterializePlanTracking`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: tracking artifact and generated tickets.
- `quiz.start`
  - Service: `StartQuiz`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: quiz runtime/raw/curated artifacts.
- `quiz.resume`
  - Service: `ResumeQuiz`.
  - Scope: project-wide or case-scoped.
  - Confirmation: required.
  - Writes: none by itself.

Approved first Web UI work actions:

- `work.preview`
  - Service: `PreviewAgentWorkLoop`.
  - Scope: ready-work queue.
  - Confirmation: required.
  - Writes: none.
- `work.daemon.stop`
  - Service: `RequestDaemonStop`.
  - Scope: current daemon runtime.
  - Confirmation: required.
  - Writes: daemon stop request state.

Deferred Web UI actions:

- `work.start.selected`, `work.start.ready`, and multi-ticket work loops stay
  out of this slice because they can run agent development and apply results.
- Quiz answer, skip, complete, cancel, and mode-switch actions stay out of this
  slice because they require a fuller active-question interaction surface.
- Arbitrary shell commands are not exposed.

## Summary

Pending implementation.

## Changes

Pending implementation.

## Validation

Pending implementation.

## Notes

- The action inventory is committed before endpoint expansion as required by
  the slice tracker.
