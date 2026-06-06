# Slice 5: Acceptance Workflow Docs

Status: delivered
Slicing: Gherkin Acceptance Generation
Plan: `ops/plans/2026-06-06-gherkin-acceptance-generation-plan.md`
Tracker: `ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md`
Branch: `feat/slice-5-acceptance-workflow-docs`
PR: `#293`

## Purpose

Document the completed acceptance feature generation workflow and validate the
full implementation before final review.

## Delivered Behavior

Reference documentation now describes acceptance feature generation as an
optional extension of spec generation. It explains readiness decisions,
readiness slots, generated `.feature` artifact location, static checks, and the
deferred `godog` execution path.

The command reference and user guide now tell users that `chc spec generate`
can write `.chc/user/acceptance/<scope>.feature` when context is suitable and
static checks pass.

The document generation model now records that feature files are supporting
acceptance artifacts, while Markdown specs remain canonical.

The root agent contract now includes the state-alignment rule used during this
sliced workflow.

## Implementation Notes

No new command was added in this slice. The implemented command surface remains
`chc spec generate`; the feature artifact is a conditional output of that
workflow.

`godog` remains explicitly documented as deferred. Generated feature files are
not described as executable validation evidence until a later execution path
exists.

## Contracts Added or Changed

- Added `docs/reference/acceptance-feature-generation.md`.
- Updated command reference for spec generation side effects.
- Updated user guide Specs section.
- Updated document generation model explanation.
- Updated project workspace layout reference with `acceptance/`.
- Updated documentation index.
- Updated root agent operating contract with state-alignment rules.

## Files of Interest

- `docs/reference/acceptance-feature-generation.md`
- `docs/reference/commands.md`
- `docs/how-to/user-guide/index.md`
- `docs/explanation/document-generation-model.md`
- `docs/reference/project-workspace.md`
- `AGENTS.md`

## Validation

- `make verify` (pass)

## Risks and Follow-ups

- `godog` execution remains a future workflow; this slice only documents the
  deferred path.
- Final slice closure should mark the tracker delivered after this PR merges.
