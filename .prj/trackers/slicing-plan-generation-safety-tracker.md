# Plan Generation Safety Slicing Tracker

Status: delivered

Plan: `.prj/plans/slicing-plan-generation-safety-plan.md`

## Active Slice

Slicing name: Plan Generation Safety

Active slice number: none

Active slice short name: none

Expected branch name: none

Expected PR title: none

Expected report path:
none

Prerequisite PRs or delivered work:

- TUI use-case-based Dumps, Quizzes, Specs, and Plans surfaces.
- Model-backed spec generation flow.
- Model-backed plan generation flow.
- Project-wide planning baseline and readiness context.

## Slicing Checklist

- [x] Capture plan generation safety risks.
- [x] Capture generate versus regenerate distinction.
- [x] Capture archive-before-replace requirement.
- [x] Capture archive context exclusion requirement.
- [x] Capture long-running generation timeout and concurrency risks.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [x] Complete all planned slices.
- [x] Validate completed slicing with `make verify`.
- [x] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Plan Action Semantics | Delivered | `feat/slice-1-plan-action-semantics` | `feat(slice-1): plan action semantics` | `.prj/reports/slices/plan-generation-safety/slice-1-plan-action-semantics.md` | Implemented; `go test ./...` passes; `make verify` blocked by lint debt. |
| Slice 2: Archive Before Replace | Delivered | `feat/slice-2-archive-before-replace` | `feat(slice-2): archive before replace` | `.prj/reports/slices/plan-generation-safety/slice-2-archive-before-replace.md` | Implemented; `go test ./...` passes; `make verify` blocked by lint debt. |
| Slice 3: Archive Context Exclusion | Delivered | `feat/slice-3-archive-context-exclusion` | `feat(slice-3): archive context exclusion` | `.prj/reports/slices/plan-generation-safety/slice-3-archive-context-exclusion.md` | Implemented; `go test ./...` passes; `make verify` blocked by lint debt. |
| Slice 4: Generation Runtime Safety | Delivered | `feat/slice-4-generation-runtime-safety` | `feat(slice-4): generation runtime safety` | `.prj/reports/slices/plan-generation-safety/slice-4-generation-runtime-safety.md` | Implemented; `go test ./...` passes; `make verify` blocked by lint debt. |

## Slice 1: Plan Action Semantics

Candidate tasks:

- [x] T1.1 Add plan action IDs for generate-missing, regenerate, and refresh.
- [x] T1.2 Make `g` no-op with an explanatory status when a canonical plan
      already exists.
- [x] T1.3 Add explicit regenerate command/key path that can be confirmed
      separately from create-missing.
- [x] T1.4 Ensure plan busy state blocks repeated keypresses and command
      palette dispatch for plan actions.
- [x] T1.5 Add TUI and adapter tests for missing-plan generation, existing-plan
      generate refusal, refresh, and regenerate intent dispatch.
- [x] T1.6 Add slice implementation report.
- [x] T1.7 Run targeted tests and `make verify`.

Task commits:

- `9d9db1f feat(tui): separate plan generate and regenerate actions`
- `09f21f7 test(tui): cover plan action semantics`

Validation:

- `go test ./...`: pass.
- `make verify`: fail in `lint-strict` because of existing lint debt across
  plan, spec, and TUI code.

## Slice 2: Archive Before Replace

Candidate tasks:

- [x] T2.1 Add canonical plan archive path layout with timestamped filenames.
- [x] T2.2 Archive existing canonical plans before regenerate writes.
- [x] T2.3 Preserve canonical plans when archive fails.
- [x] T2.4 Add archive metadata for source path and archive time.
- [x] T2.5 Add service tests for regenerate with existing plan, regenerate with
      missing plan, and archive failure.
- [x] T2.6 Add slice implementation report.
- [x] T2.7 Run targeted tests and `make verify`.

Task commits:

- `183a730 docs(plan): select slice 2 archive before replace`
- `e38f1e6 feat(plan): add generated plan archive layout`
- `3da3dc0 feat(plan): archive before regenerating plans`
- `27a4f61 test(plan): cover regenerate archive behavior`

Validation:

- `go test ./internal/chr`: pass.
- `go test ./internal/chrapp`: pass.
- `go test ./...`: pass.
- `make verify`: fail in `lint-strict` because of existing lint debt across
  plan, spec, and TUI code.

## Slice 3: Archive Context Exclusion

Candidate tasks:

- [x] T3.1 Exclude `.chr/plans/archive/**` from active plan context.
- [x] T3.2 Exclude `.chr/specs/archive/**` from active generation context when
      present.
- [x] T3.3 Ensure TUI plan lists ignore archived plans.
- [x] T3.4 Ensure generated plan provenance does not include archived paths.
- [x] T3.5 Add tests for workspace/context exclusion and TUI listing behavior.
- [x] T3.6 Add slice implementation report.
- [x] T3.7 Run targeted tests and `make verify`.

Task commits:

- `0f97588 docs(plan): select slice 3 archive context exclusion`
- `c8605aa fix(index): exclude archived generated documents`
- `f554174 test(plan): verify archived documents stay out of generation`
- `8d658e9 fix(tui): hide archived generated documents`

Validation:

- `go test ./internal/chr`: pass.
- `go test ./internal/chrapp`: pass.
- `go test ./internal/chrtui`: pass.
- `go test ./...`: pass.
- `make verify`: fail in `lint-strict` because of existing lint debt across
  plan, spec, and TUI code.

## Slicing Closure

All planned slices are delivered:

- Slice 1: PR `#198`
- Slice 2: PR `#199`
- Slice 3: PR `#200`
- Slice 4: PR `#201`

Slicing validation:

- `go test ./...`: pass.
- `make verify`: fail in `lint-strict` because of existing lint debt across
  plan, spec, and TUI code.

Tuning decision: no additional tuning slice is planned for this slicing. The
original validation was blocked by repository-wide lint debt, not unfinished
plan generation safety behavior. Historical failures below remain unchanged.

Closure audit on 2026-10-20 verified PR #201 merged at
`6def9bf24e102d9de82b3c552b9d42725ef964d2`, an ancestor of dev. The current
development gate, `make verify`, passed on
`de2ebb48cb6000670bd2973b98b8f86dafe93f95` with zero lint issues. This is current
development evidence, not a replacement for historical results or nightly proof.

## Slice 4: Generation Runtime Safety

Candidate tasks:

- [x] T4.1 Add longer timeout behavior for plan/spec generation than quiz
      next-question generation.
- [x] T4.2 Add service-level same-scope generation guard.
- [x] T4.3 Improve failure status when generation fails while an older plan is
      still visible.
- [x] T4.4 Extend diagnostic logging for guard, archive, success, and failure
      events.
- [x] T4.5 Add tests for guard behavior and failure messaging.
- [x] T4.6 Add slice implementation report.
- [x] T4.7 Run targeted tests and `make verify`.

Task commits:

- `1189b57 docs(plan): select slice 4 generation runtime safety`
- `6a58219 feat(generation): guard long-running spec and plan requests`
- `27dc933 fix(tui): clarify generation failure state`
- `505f092 test(generation): cover timeout and guard behavior`
- `07ca43b feat(generation): log spec generation lifecycle`
- `ba67767 fix(openai): extend spec and plan generation timeout`
- `6634da0 test(tui): fix generation failure assertion flow`

Validation:

- `go test ./internal/chr`: pass.
- `go test ./internal/chrapp`: pass.
- `go test ./internal/chrtui`: pass.
- `go test ./...`: pass.
- `make verify`: fail in `lint-strict` because of existing lint debt across
  plan, spec, and TUI code.
