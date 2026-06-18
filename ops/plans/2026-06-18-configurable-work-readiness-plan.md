# Configurable Work Readiness Plan

Status: active

Active Slice: Slice 2 - work doctor diagnostics

Tracker: `ops/trackers/2026-06-18-configurable-work-readiness-tracker.md`

Spec: `ops/specs/2026-06-18-configurable-work-readiness.md`

Model: `ops/specs/2026-06-18-configurable-work-readiness-model.md`

## Purpose

Make work execution readiness configurable, diagnosable, and safely repairable
without weakening ChrysoCoord's default `dev` plus `origin` workflow.

The delivery should turn the rigid preflight exposed by the dialer exercise into
a reusable product capability for any project that uses ChrysoCoord to automate
development.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice updates its implementation report under
  `.prj/reports/slices/configurable-work-readiness/`.
- Each task gets one conventional commit.
- After each merged slice, return to `dev` before follow-up work.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | work policy config | `feat/slice-1-work-policy-config` | `feat(slice-1): work policy config` | `.prj/reports/slices/configurable-work-readiness/slice-1-work-policy-config.md` |
| Slice 2 | work doctor diagnostics | `feat/slice-2-work-doctor-diagnostics` | `feat(slice-2): work doctor diagnostics` | `.prj/reports/slices/configurable-work-readiness/slice-2-work-doctor-diagnostics.md` |
| Slice 3 | safe readiness repair | `feat/slice-3-safe-readiness-repair` | `feat(slice-3): safe readiness repair` | `.prj/reports/slices/configurable-work-readiness/slice-3-safe-readiness-repair.md` |
| Slice 4 | dash readiness guidance | `feat/slice-4-dash-readiness-guidance` | `feat(slice-4): dash readiness guidance` | `.prj/reports/slices/configurable-work-readiness/slice-4-dash-readiness-guidance.md` |

## Slice 1: Work Policy Config

Goal: move work branch assumptions into workspace config and make preflight use
that policy.

Behavior:

- Workspace config loads a `work` section with default `dev` plus `origin`
  values.
- Work-selection preflight no longer hardcodes branch or remote names.
- `integration_branch`, `current_branch`, and `no_git_enforcement` modes are
  recognized.
- Existing review-repair ordering still applies independently of branch policy.

Tasks:

- T1.1 Add work branch policy config model, defaults, and parser coverage.
- T1.2 Update work-selection preflight to evaluate the configured policy.
- T1.3 Update app/service call sites to pass workspace work policy.
- T1.4 Add tests for default, custom branch, current-branch, and no-Git modes.
- T1.5 Add the slice implementation report and validation evidence.

## Slice 2: Work Doctor Diagnostics

Goal: expose readiness diagnostics without starting agent work.

Behavior:

- `chc work doctor` reports whether work can start.
- Known Git failures are classified into readiness diagnostic codes.
- Missing remote refs and branch-behind-remote states do not surface raw
  `exit status 128` as the primary message.
- CLI output includes relevant configured policy values.

Tasks:

- T2.1 Add work readiness diagnostic and status models.
- T2.2 Add a core readiness service that evaluates policy and Git state.
- T2.3 Add `chc work doctor` output.
- T2.4 Add CLI/core tests for classified diagnostics.
- T2.5 Add the slice implementation report and validation evidence.

## Slice 3: Safe Readiness Repair

Goal: repair non-destructive readiness problems when policy allows it.

Behavior:

- `chc work doctor --fix-safe` applies safe repairs.
- ChrysoCoord runtime artifacts can be excluded from Git dirtiness without
  editing user code.
- Existing configured integration branches can be switched to when safe.
- Remote fetch and fast-forward repair are available when clean and
  non-divergent.
- Unsafe states remain blocking with manual guidance.

Tasks:

- T3.1 Add repair planning for known readiness diagnostics.
- T3.2 Implement safe runtime-artifact ignore repair.
- T3.3 Implement safe branch/remote repair primitives.
- T3.4 Add repair tests for idempotence and unsafe-state refusal.
- T3.5 Add the slice implementation report and validation evidence.

## Slice 4: Dash Readiness Guidance

Goal: make Dash show and use readiness diagnostics before running work.

Behavior:

- Dash displays structured readiness diagnostics when work cannot start.
- Dash can trigger safe repair and retry when the service reports it is
  available.
- Command references explain branch policy config and doctor usage.
- QA reset validation catches readiness blockers before manual TUI runs.

Tasks:

- T4.1 Add a TUI read model for work readiness diagnostics.
- T4.2 Render blocked readiness state and repair guidance in Dash.
- T4.3 Wire safe repair and retry into Dash work actions where safe.
- T4.4 Update docs and QA smoke coverage.
- T4.5 Add the slice implementation report and validation evidence.
