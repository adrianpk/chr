# Serialized Integration Loop Tracker

Status: active

Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`

Spec: `ops/specs/2026-06-20-serialized-integration-loop.md`

Model: `ops/specs/2026-06-20-serialized-integration-loop-model.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | cycle policy and state model | reviewing | `feat/slice-1-cycle-policy-and-state-model` | pending | `.prj/reports/slices/serialized-integration-loop/slice-1-cycle-policy-and-state-model.md` |
| Slice 2 | one-ticket cycle runner | planned | `feat/slice-2-one-ticket-cycle-runner` | pending | `.prj/reports/slices/serialized-integration-loop/slice-2-one-ticket-cycle-runner.md` |
| Slice 3 | high-level agentic review | planned | `feat/slice-3-high-level-agentic-review` | pending | `.prj/reports/slices/serialized-integration-loop/slice-3-high-level-agentic-review.md` |
| Slice 4 | merge-before-next-ticket | planned | `feat/slice-4-merge-before-next-ticket` | pending | `.prj/reports/slices/serialized-integration-loop/slice-4-merge-before-next-ticket.md` |
| Slice 5 | finite repair cycle | planned | `feat/slice-5-finite-repair-cycle` | pending | `.prj/reports/slices/serialized-integration-loop/slice-5-finite-repair-cycle.md` |
| Slice 6 | observability docs and QA | planned | `feat/slice-6-observability-docs-and-qa` | pending | `.prj/reports/slices/serialized-integration-loop/slice-6-observability-docs-and-qa.md` |

## Slice 1 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T1.1 | Add core value types for integration cycle mode, phase, status, and blocker. | done | `a774f32` | Added package-level integration cycle enums and validation helpers. |
| T1.2 | Add policy resolution from existing review workflow configuration. | done | `b9a694c` | Mapped `review.mode` to cycle policy without changing defaults or runtime behavior. |
| T1.3 | Add a side-effect-free cycle planning function for one selected ticket. | done | `f5009c1` | Planner returns selected-ticket cycle metadata with no IO or ticket mutation. |
| T1.4 | Decide whether Slice 1 introduces a dedicated `integration_cycles` table or defers it. | done | `02a435b` | Deferred table creation; Slice 1 uses existing durable state plus side-effect-free planning contracts. |
| T1.5 | If a table is introduced, add an additive migration and repository tests. | done | `02a435b` | Not applicable for Slice 1 because the table was explicitly deferred. |
| T1.6 | Add tests for policy mapping across `human_review`, `agentic_review`, and `trusted_auto_merge`. | done | `61daf4e`, `7eb0cba` | Added table-driven tests for policy mapping, planner validation, value validation, and strict-lint shape. |
| T1.7 | Add docs or reference notes for cycle mode semantics. | done | `84c44d6` | Added ticket loop and review workflow notes for serialized cycle mode semantics. |
| T1.8 | Add the slice implementation report and validation evidence. | done | this commit | Report added with final validation evidence. |

## Slice 2 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T2.1 | Introduce a cycle runner that selects one ticket per iteration. | pending | pending | |
| T2.2 | Refactor current loop execution to delegate development to the cycle runner. | pending | pending | |
| T2.3 | Refresh ticket index and config after each cycle. | pending | pending | |
| T2.4 | Stop `human_review` mode after `ready_for_review` handoff. | pending | pending | |
| T2.5 | Update dry-run output to describe first selected cycle and remaining eligible tickets. | pending | pending | |
| T2.6 | Replace tests that expect two development-only processed tickets with tests that expect a barrier after the first unfinished integration cycle. | pending | pending | |
| T2.7 | Add tests proving no second ticket starts until the first reaches a terminal cycle state. | pending | pending | |
| T2.8 | Add the slice implementation report and validation evidence. | pending | pending | |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Define the reviewer context artifact and prompt contract. | pending | pending | |
| T3.2 | Add structured review result parsing and validation. | pending | pending | |
| T3.3 | Persist local review artifacts under `.chc/`. | pending | pending | |
| T3.4 | Apply `pass`, `needs_changes`, `insufficient_evidence`, and `human_required` outcomes to ticket state. | pending | pending | |
| T3.5 | Add non-blocking note handling with existing comment limits. | pending | pending | |
| T3.6 | Add fake-reviewer tests for clean pass and high-level blocking findings. | pending | pending | |
| T3.7 | Add tests proving cosmetic findings do not block when validation passed. | pending | pending | |
| T3.8 | Add the slice implementation report and validation evidence. | pending | pending | |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add merge phase orchestration to the cycle runner. | pending | pending | |
| T4.2 | Reuse existing ticket merge command/preflight behavior instead of duplicating provider logic. | pending | pending | |
| T4.3 | Ensure `trusted_auto_merge` still requires merge before continuing. | pending | pending | |
| T4.4 | Add tests proving ticket 2 starts only after ticket 1 has merge metadata and status `done`. | pending | pending | |
| T4.5 | Add tests for merge preflight blocked, provider merge failed, and cleanup after merge. | pending | pending | |
| T4.6 | Update loop report output with merge phase and merge result. | pending | pending | |
| T4.7 | Add the slice implementation report and validation evidence. | pending | pending | |

## Slice 5 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T5.1 | Add durable repair count handling. | pending | pending | |
| T5.2 | Feed review findings into the repair worker context. | pending | pending | |
| T5.3 | Ensure repair selection outranks all later normal work. | pending | pending | |
| T5.4 | Re-run review after repair. | pending | pending | |
| T5.5 | Block after the configured/default repair limit. | pending | pending | |
| T5.6 | Add tests for one successful repair, repeated review failure, missing repair handoff, and later-ticket blocking. | pending | pending | |
| T5.7 | Add the slice implementation report and validation evidence. | pending | pending | |

## Slice 6 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T6.1 | Update agent event types/messages for cycle selection, review, repair, merge, and rehydration. | pending | pending | |
| T6.2 | Update Dash rendering for active cycle phase and stop reason. | pending | pending | |
| T6.3 | Update loop and execution reports with cycle details. | pending | pending | |
| T6.4 | Update user guide, ticket loop reference, review workflow reference, and changelog. | pending | pending | |
| T6.5 | Add or update resettable QA fixture with two dependent tickets. | pending | pending | |
| T6.6 | Add smoke or manual runbook proving ticket 2 does not start until ticket 1 is merged. | pending | pending | |
| T6.7 | Run `make verify` and relevant QA smoke commands. | pending | pending | |
| T6.8 | Add the slice implementation report and final cycle report. | pending | pending | |

## Validation Log

- Planning: `git diff --check` (pass)
- Slice 1 T1.1: `go test ./internal/core` (pass)
- Slice 1 T1.2: `go test ./internal/core -run 'ReviewWorkflow|IntegrationCycle'` (pass)
- Slice 1 T1.3: `go test ./internal/core -run IntegrationCycle` (pass)
- Slice 1 T1.6: `go test ./internal/core -run 'IntegrationCycle'` (pass)
- Slice 1 T1.8: `make lint-strict` (pass)
- Slice 1 T1.8: `make verify` (pass)

## Open Decisions

- Slice 1 defers a dedicated `integration_cycles` table and relies on existing
  ticket, execution, review, merge, and loop state until the runner needs a
  persistence table.
- Whether `chc work ticket <ticket>` should remain development-only or gain a
  separate full-cycle single-ticket command.
- Whether project templates should eventually default to `agentic_review`,
  `trusted_auto_merge`, or remain `human_review`.

## Notes

- The core requirement is the per-ticket integration barrier:
  later normal work starts only after the selected ticket is done or blocked.
- Review must remain high-level and objective. Cosmetic findings are
  non-blocking when configured validation passed.
- This slicing intentionally does not introduce parallel execution. It creates
  the integration discipline needed before parallel work can be safe.
