# Slice 2: Archive Before Replace

Status: delivered
Slicing: Plan Generation Safety
Plan: `.prj/plans/slicing-plan-generation-safety-plan.md`
Tracker: `.prj/trackers/slicing-plan-generation-safety-tracker.md`
Branch: `feat/slice-2-archive-before-replace`
PR: `#199`

## Purpose

Preserve an existing canonical generated plan before explicit regeneration
replaces it.

## Delivered Behavior

- Regenerating an existing plan archives the current canonical file before
  invoking the model.
- Archived plans are written under `.chc/plans/archive/`.
- Archive filenames include the scope ID and UTC archive timestamp.
- Archive content records `archived_from` and `archived_at` metadata.
- Regenerating a missing plan behaves like create and does not create an
  archive directory.
- Archive failure stops regeneration before model invocation and leaves the
  canonical plan untouched.

## Implementation Notes

- `GeneratePlan` remains the direct create/write operation.
- `RegeneratePlan` owns archive-before-replace semantics in the application
  service.
- The TUI plan adapter now dispatches regenerate intent to `RegeneratePlan`
  instead of reusing `GeneratePlan`.
- Archived plans remain visible to the workspace index until Slice 3 excludes
  archive paths from active context.

## Contracts Added or Changed

- Added `ArchivedPlanPaths`.
- Added `ArchivedPlanLayout`.
- Added `ArchiveGeneratedPlan`.
- Added `RenderArchivedPlan`.
- Added `PlanRegenerateRequest`.
- Added `service.RegeneratePlan`.

## Files of Interest

- `internal/chc/plan_generation.go`
- `internal/chc/plan_generation_test.go`
- `internal/chcapp/service_plan.go`
- `internal/chcapp/service_plan_test.go`
- `internal/chcapp/tui_adapters.go`

## Validation

- `go test ./internal/chc`: pass
- `go test ./internal/chcapp`: pass
- `go test ./...`: pass
- `make verify`: fail

`make verify` fails in `lint-strict` on existing lint debt across the plan,
spec, and TUI areas. The current failure set includes pre-existing complexity,
function length, unused function, errcheck, formatting, and whitespace rules.

## Commits

- `183a730 docs(plan): select slice 2 archive before replace`
- `e38f1e6 feat(plan): add generated plan archive layout`
- `3da3dc0 feat(plan): archive before regenerating plans`
- `27a4f61 test(plan): cover regenerate archive behavior`

## Risks and Follow-ups

- Archived plans are recoverable history but are not excluded from active
  generation context until Slice 3.
- Service-level same-scope generation locking remains planned for Slice 4.
