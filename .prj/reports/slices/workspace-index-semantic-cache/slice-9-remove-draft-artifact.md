# Slice 9: Draft Artifact Removal

Status: delivered
Slicing: Workspace Index And Semantic Cache
Plan: `.prj/plans/slicing-workspace-index-semantic-cache-plan.md`
Tracker: `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`
Branch: `feat/slice-9-remove-draft-artifact`
PR: #172

Verified merge: `6c56d2e84d56d54ea0f4449ce60457f62c7a5d77`, an ancestor of dev
as checked on 2026-10-20.

## Purpose

Remove `Draft` as a first-class artifact family from active TUI and document
surfaces while preserving `draft` where it represents lifecycle state, such as a
spec or authoring session that still needs review.

## Delivered Behavior

- Remove the command palette action that planned `Create draft`.
- Retarget TUI authoring session creation from draft artifacts to specs.
- Keep authoring session status `draft` as lifecycle metadata.
- Remove `.prj/drafts` and `.chr/drafts` from active TUI document discovery.
- Add tests that draft creation is not exposed by the default command registry.
- Add tests that legacy draft directories are ignored by active document
  discovery.
- Update active product specs to describe dump/quiz/spec workflows instead of a
  first-class draft document family.

## Implementation Notes

Historical files can remain in legacy draft directories. The TUI no longer
classifies those directories as an active document family, which keeps old notes
available on disk without treating draft as a current product concept.

Spec roughness is represented by `status: draft`, not by a separate draft
document type.

## Contracts Added or Changed

- Removed `ArtifactActionCreateDraft`.
- `CreateTUIAuthoringSession` now creates a spec-targeted authoring session.
- Active document roots no longer include draft directories.
- Active workflow documentation now uses dump, quiz, spec, plan, tracker, ticket,
  and report families.

## Files of Interest

- `internal/chrtui/command_palette.go`
- `internal/chrtui/artifact_actions.go`
- `internal/chrtui/documents.go`
- `internal/chrtui/model_test.go`
- `internal/chrapp/service_artifact_ticket.go`
- `internal/chrapp/tui_adapters.go`
- `internal/chrapp/service_test.go`
- `.prj/specs/specs/context-composition-pipeline.md`
- `.prj/specs/specs/tui-authoring-and-quiz-workflow.md`
- `.prj/specs/specs/tui-operational-console.md`
- `.prj/specs/specs/tui-product-surface.md`
- `.prj/specs/specs/tui-record-screens.md`
- `.prj/specs/specs/ticket-resolution-lifecycle.md`
- `.prj/reports/slices/workspace-index-semantic-cache/slice-9-remove-draft-artifact.md`
- `.prj/trackers/slicing-workspace-index-semantic-cache-tracker.md`

## Validation

- `go test ./internal/chrtui ./internal/chrapp` passed on 2026-05-23.
- `make verify` passed on 2026-05-23.

## Risks and Follow-ups

- CLI transition commands with historical draft naming are outside this TUI and
  active-document slice.
- Slice 8 was opened separately and may need merge-order handling if it touches
  the same TUI files before this slice lands.
