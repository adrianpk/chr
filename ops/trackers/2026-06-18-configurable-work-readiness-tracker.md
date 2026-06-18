# Configurable Work Readiness Tracker

Status: reviewing

Plan: `ops/plans/2026-06-18-configurable-work-readiness-plan.md`

Spec: `ops/specs/2026-06-18-configurable-work-readiness.md`

Model: `ops/specs/2026-06-18-configurable-work-readiness-model.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | work policy config | delivered | `feat/slice-1-work-policy-config` | `#313` | `.prj/reports/slices/configurable-work-readiness/slice-1-work-policy-config.md` |
| Slice 2 | work doctor diagnostics | delivered | `feat/slice-2-work-doctor-diagnostics` | `#314` | `.prj/reports/slices/configurable-work-readiness/slice-2-work-doctor-diagnostics.md` |
| Slice 3 | safe readiness repair | delivered | `feat/slice-3-safe-readiness-repair` | `#315` | `.prj/reports/slices/configurable-work-readiness/slice-3-safe-readiness-repair.md` |
| Slice 4 | dash readiness guidance | reviewing | `feat/slice-4-dash-readiness-guidance` | pending | `.prj/reports/slices/configurable-work-readiness/slice-4-dash-readiness-guidance.md` |

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
| T2.1 | Add work readiness diagnostic and status models. | done | `68f76c2` | Include policy values in diagnostics. |
| T2.2 | Add a core readiness service that evaluates policy and Git state. | done | `a3c89d0` | Classifies known Git failures. |
| T2.3 | Add `chc work doctor` output. | done | `1ead8e0` | No agent work starts. |
| T2.4 | Add CLI/core tests for classified diagnostics. | done | `a94b9c5` | Covers missing remote ref, branch-behind, runtime artifacts, and no-Git mode. |
| T2.5 | Add the slice implementation report and validation evidence. | done | this commit | Report created with reviewing status. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Add repair planning for known readiness diagnostics. | done | `6772370` | Repairs are attached to diagnostics and respect policy gates. |
| T3.2 | Implement safe runtime-artifact ignore repair. | done | `852df8f` | Uses local `.git/info/exclude`, not project `.gitignore`. |
| T3.3 | Implement safe branch/remote repair primitives. | done | `5fe1b1a` | Adds fetch, switch, create, and fast-forward repair without reset, stash, force-push, or conflict resolution. |
| T3.4 | Add repair tests for idempotence and unsafe-state refusal. | done | `6b13d8d` | Covers idempotence, user-change refusal, branch creation, fetch plus fast-forward, and CLI `--fix-safe`. |
| T3.5 | Add the slice implementation report and validation evidence. | done | this commit | Report status is `reviewing` before PR. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add a TUI read model for work readiness diagnostics. | done | `9f6de9a` | Reuses core diagnostics in TUI options, refresh, and action results. |
| T4.2 | Render blocked readiness state and repair guidance in Dash. | done | `9f6de9a` | Dash Current panel includes readiness status, diagnostics, and repair guidance. |
| T4.3 | Wire safe repair and retry into Dash work actions where safe. | done | `9f6de9a` | `x` confirms safe repair; blocked work actions stop before execution. |
| T4.4 | Update docs and QA smoke coverage. | done | `43f502a` | QA reset now creates a ready `dev` plus local `origin/dev`; smoke checks `work doctor`. |
| T4.5 | Add the slice implementation report and validation evidence. | done | this commit | Report status is `reviewing` before PR. |

## Validation Log

- Slice 1: `go test ./internal/core -run 'WorkspaceConfig|WorkConfig|ReviewWorkflow|NotificationConfig'` (pass)
- Slice 1: `go test ./internal/core -run 'WorkSelectionPreflight'` (pass)
- Slice 1: `go test ./internal/app -run 'Work|TicketLoop|TicketStart'` (pass)
- Slice 1: `go test ./internal/core ./internal/app` (pass)
- Slice 1: `make lint-strict` (pass)
- Slice 1: `make verify` (pass)
- Slice 2: `go test ./internal/core -run 'WorkReadiness|WorkSelectionPreflight'` (pass)
- Slice 2: `go test ./internal/app -run 'WorkDoctor'` (pass)
- Slice 2: `go test ./internal/core ./internal/app` (pass)
- Slice 2: `make lint-strict` (pass)
- Slice 2: `make verify` (pass)
- Slice 3: `go test ./internal/core -run 'WorkReadiness|WorkSelectionPreflight'` (pass)
- Slice 3: `go test ./internal/app -run 'WorkDoctor'` (pass)
- Slice 3: `go test ./internal/core ./internal/app` (pass)
- Slice 3: `make lint-strict` (pass)
- Slice 3: `make verify` (pass)
- Slice 4: `go test ./internal/tui ./internal/app` (pass)
- Slice 4: `make qa-smoke-tictactoe` (pass)
- Slice 4: `make qa-smoke-blog` (pass)
- Slice 4: `make verify` (pass)

## Notes

- Defaults intentionally keep ChrysoCoord's own `dev` plus `origin` workflow.
- The feature should make those defaults configurable instead of hardcoded.
- Repair behavior must stay non-destructive.
