# Plan Generation Safety Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-plan-generation-safety-plan.md`

## Active Slice

Slicing name: Plan Generation Safety

Active slice number: Slice 1

Active slice short name: plan action semantics

Expected branch name: `feat/slice-1-plan-action-semantics`

Expected PR title: `feat(slice-1): plan action semantics`

Expected report path:
`.prj/reports/slices/plan-generation-safety/slice-1-plan-action-semantics.md`

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
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Plan Action Semantics | Delivered | `feat/slice-1-plan-action-semantics` | `feat(slice-1): plan action semantics` | `.prj/reports/slices/plan-generation-safety/slice-1-plan-action-semantics.md` | Implemented; `go test ./...` passes; `make verify` blocked by lint debt. |
| Slice 2: Archive Before Replace | Planned | `feat/slice-2-archive-before-replace` | `feat(slice-2): archive before replace` | `.prj/reports/slices/plan-generation-safety/slice-2-archive-before-replace.md` | Depends on Slice 1. |
| Slice 3: Archive Context Exclusion | Planned | `feat/slice-3-archive-context-exclusion` | `feat(slice-3): archive context exclusion` | `.prj/reports/slices/plan-generation-safety/slice-3-archive-context-exclusion.md` | Depends on Slice 2. |
| Slice 4: Generation Runtime Safety | Planned | `feat/slice-4-generation-runtime-safety` | `feat(slice-4): generation runtime safety` | `.prj/reports/slices/plan-generation-safety/slice-4-generation-runtime-safety.md` | Depends on Slices 1-3. |

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

- [ ] T2.1 Add canonical plan archive path layout with timestamped filenames.
- [ ] T2.2 Archive existing canonical plans before regenerate writes.
- [ ] T2.3 Preserve canonical plans when archive fails.
- [ ] T2.4 Add archive metadata for source path and archive time.
- [ ] T2.5 Add service tests for regenerate with existing plan, regenerate with
      missing plan, and archive failure.
- [ ] T2.6 Add slice implementation report.
- [ ] T2.7 Run targeted tests and `make verify`.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 3: Archive Context Exclusion

Candidate tasks:

- [ ] T3.1 Exclude `.chc/plans/archive/**` from active plan context.
- [ ] T3.2 Exclude `.chc/specs/archive/**` from active generation context when
      present.
- [ ] T3.3 Ensure TUI plan lists ignore archived plans.
- [ ] T3.4 Ensure generated plan provenance does not include archived paths.
- [ ] T3.5 Add tests for workspace/context exclusion and TUI listing behavior.
- [ ] T3.6 Add slice implementation report.
- [ ] T3.7 Run targeted tests and `make verify`.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 4: Generation Runtime Safety

Candidate tasks:

- [ ] T4.1 Add longer timeout behavior for plan/spec generation than quiz
      next-question generation.
- [ ] T4.2 Add service-level same-scope generation guard.
- [ ] T4.3 Improve failure status when generation fails while an older plan is
      still visible.
- [ ] T4.4 Extend diagnostic logging for guard, archive, success, and failure
      events.
- [ ] T4.5 Add tests for guard behavior and failure messaging.
- [ ] T4.6 Add slice implementation report.
- [ ] T4.7 Run targeted tests and `make verify`.

Task commits:

- Pending.

Validation:

- Pending.
