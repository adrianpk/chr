# Configurable Work Readiness Tracker

Status: reviewing

Plan: `ops/plans/2026-06-18-configurable-work-readiness-plan.md`

Spec: `ops/specs/2026-06-18-configurable-work-readiness.md`

Model: `ops/specs/2026-06-18-configurable-work-readiness-model.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | work policy config | reviewing | `feat/slice-1-work-policy-config` | pending | `.prj/reports/slices/configurable-work-readiness/slice-1-work-policy-config.md` |
| Slice 2 | work doctor diagnostics | planned | `feat/slice-2-work-doctor-diagnostics` | pending | `.prj/reports/slices/configurable-work-readiness/slice-2-work-doctor-diagnostics.md` |
| Slice 3 | safe readiness repair | planned | `feat/slice-3-safe-readiness-repair` | pending | `.prj/reports/slices/configurable-work-readiness/slice-3-safe-readiness-repair.md` |
| Slice 4 | dash readiness guidance | planned | `feat/slice-4-dash-readiness-guidance` | pending | `.prj/reports/slices/configurable-work-readiness/slice-4-dash-readiness-guidance.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add work branch policy config model, defaults, and parser coverage. | done | `25c7a47` | Preserve default `dev` plus `origin`. |
| T1.2 | Update work-selection preflight to evaluate the configured policy. | done | `49277d3` | Removes hardcoded branch and remote names from policy logic. |
| T1.3 | Update app/service call sites to pass workspace work policy. | done | `281fdb6` | Keep review-repair ordering unchanged. |
| T1.4 | Add tests for default, custom branch, current-branch, and no-Git modes. | done | `1e4ce9b` | Includes core policy coverage. |
| T1.5 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Add work readiness diagnostic and status models. | pending | pending | Include policy values in diagnostics. |
| T2.2 | Add a core readiness service that evaluates policy and Git state. | pending | pending | Classify known Git failures. |
| T2.3 | Add `chc work doctor` output. | pending | pending | No agent work should start. |
| T2.4 | Add CLI/core tests for classified diagnostics. | pending | pending | Cover missing remote ref and branch-behind states. |
| T2.5 | Add the slice implementation report and validation evidence. | pending | pending | Report status should be `reviewing` before PR. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Add repair planning for known readiness diagnostics. | pending | pending | Respect config and safety boundaries. |
| T3.2 | Implement safe runtime-artifact ignore repair. | pending | pending | Do not hide user-owned changes. |
| T3.3 | Implement safe branch/remote repair primitives. | pending | pending | No reset, stash, force-push, or conflict resolution. |
| T3.4 | Add repair tests for idempotence and unsafe-state refusal. | pending | pending | Include negative tests for user changes. |
| T3.5 | Add the slice implementation report and validation evidence. | pending | pending | Report status should be `reviewing` before PR. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add a TUI read model for work readiness diagnostics. | pending | pending | Reuse core diagnostics. |
| T4.2 | Render blocked readiness state and repair guidance in Dash. | pending | pending | Replace opaque work-action failures. |
| T4.3 | Wire safe repair and retry into Dash work actions where safe. | pending | pending | Avoid unsafe automatic repairs. |
| T4.4 | Update docs and QA smoke coverage. | pending | pending | Catch readiness blockers before manual TUI runs. |
| T4.5 | Add the slice implementation report and validation evidence. | pending | pending | Report status should be `reviewing` before PR. |

## Validation Log

- Slice 1: `go test ./internal/core -run 'WorkspaceConfig|WorkConfig|ReviewWorkflow|NotificationConfig'` (pass)
- Slice 1: `go test ./internal/core -run 'WorkSelectionPreflight'` (pass)
- Slice 1: `go test ./internal/app -run 'Work|TicketLoop|TicketStart'` (pass)
- Slice 1: `go test ./internal/core ./internal/app` (pass)
- Slice 1: `make lint-strict` (pass)
- Slice 1: `make verify` (pass)

## Notes

- Defaults intentionally keep ChrysoCoord's own `dev` plus `origin` workflow.
- The feature should make those defaults configurable instead of hardcoded.
- Repair behavior must stay non-destructive.
