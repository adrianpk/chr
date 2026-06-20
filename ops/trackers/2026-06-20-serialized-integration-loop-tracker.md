# Serialized Integration Loop Tracker

Status: active

Plan: `ops/plans/2026-06-20-serialized-integration-loop-plan.md`

Spec: `ops/specs/2026-06-20-serialized-integration-loop.md`

Model: `ops/specs/2026-06-20-serialized-integration-loop-model.md`

## Slice Status

| Slice | Short Name | Status | Branch | PR | Report |
| --- | --- | --- | --- | --- | --- |
| Slice 1 | cycle policy and state model | delivered | `feat/slice-1-cycle-policy-and-state-model` | `#336` | `.prj/reports/slices/serialized-integration-loop/slice-1-cycle-policy-and-state-model.md` |
| Slice 2 | one-ticket cycle runner | delivered | `feat/slice-2-one-ticket-cycle-runner` | `#337` | `.prj/reports/slices/serialized-integration-loop/slice-2-one-ticket-cycle-runner.md` |
| Slice 3 | high-level agentic review | delivered | `feat/slice-3-high-level-agentic-review` | `#338` | `.prj/reports/slices/serialized-integration-loop/slice-3-high-level-agentic-review.md` |
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
| T2.1 | Introduce a cycle runner that selects one ticket per iteration. | done | `1ee364c`, `57a06e7` | Added work-run integration cycle runner state and per-cycle ticket selection. |
| T2.2 | Refactor current loop execution to delegate development to the cycle runner. | done | `1ee364c`, `57a06e7` | `work run-all` now uses the integration cycle runner instead of a stale multi-ticket queue. |
| T2.3 | Refresh ticket index and config after each cycle. | done | `1ee364c` | Each continuation path reselects from refreshed ticket index and workspace config. |
| T2.4 | Stop `human_review` mode after `ready_for_review` handoff. | done | `1ee364c` | Development success now records a loop barrier when review or merge is still required. |
| T2.5 | Update dry-run output to describe first selected cycle and remaining eligible tickets. | done | `1ee364c` | Dry-run prints `cycle_ticket` and `remaining_eligible`. |
| T2.6 | Replace tests that expect two development-only processed tickets with tests that expect a barrier after the first unfinished integration cycle. | done | `1ee364c`, `57a06e7` | Updated service and CLI expectations for one processed ticket plus blocked loop status. |
| T2.7 | Add tests proving no second ticket starts until the first reaches a terminal cycle state. | done | `1ee364c`, `57a06e7` | Tests assert CHC-2 has no execution record when CHC-1 is still waiting for integration. |
| T2.8 | Add the slice implementation report and validation evidence. | done | this commit | Report added with final validation evidence. |

## Slice 3 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T3.1 | Define the reviewer context artifact and prompt contract. | done | `4ccb3fb` | Added local review context rendering and Codex reviewer prompt. |
| T3.2 | Add structured review result parsing and validation. | done | `4ccb3fb` | Added strict JSON parsing, schema, decision validation, and finding validation. |
| T3.3 | Persist local review artifacts under `.chc/`. | done | `4ccb3fb` | Reviews write context, structured result, summary, and Codex run artifacts under runtime review directories. |
| T3.4 | Apply `pass`, `needs_changes`, `insufficient_evidence`, and `human_required` outcomes to ticket state. | done | `f838c0e`, `7a4fa6b` | Core maps review decisions to ticket review metadata; app applies them during `work run-all`. |
| T3.5 | Add non-blocking note handling with existing comment limits. | done | `7a4fa6b` | Agentic review non-blocking notes become limited local ticket comments. |
| T3.6 | Add fake-reviewer tests for clean pass and high-level blocking findings. | done | `7a4fa6b` | Service tests cover pass and `needs_changes` decisions through a fake reviewer. |
| T3.7 | Add tests proving cosmetic findings do not block when validation passed. | done | `7a4fa6b` | Cosmetic notes are written without preventing `ready_for_merge`. |
| T3.8 | Add the slice implementation report and validation evidence. | done | this commit | Report added with final validation evidence. |

## Slice 4 Tasks

| Task | Description | Status | Commit | Notes |
| --- | --- | --- | --- | --- |
| T4.1 | Add merge phase orchestration to the cycle runner. | done | `b78afb3`, `76883ad` | The work loop now runs merge after review pass before selecting later tickets. |
| T4.2 | Reuse existing ticket merge command/preflight behavior instead of duplicating provider logic. | done | `b78afb3` | Loop merge delegates to `core.RunTicketMergeCommand` with configured work and review settings. |
| T4.3 | Ensure `trusted_auto_merge` still requires merge before continuing. | done | `79f4c6a`, `b78afb3`, `6f6460e` | Trusted mode writes pass review handoff metadata and still merges before continuing. |
| T4.4 | Add tests proving ticket 2 starts only after ticket 1 has merge metadata and status `done`. | done | `6f6460e`, `76883ad` | Service tests cover two-ticket continuation only after merge success. |
| T4.5 | Add tests for merge preflight blocked, provider merge failed, and cleanup after merge. | done | `6f6460e` | Service tests cover blocked preflight and provider failure; cleanup remains covered by existing merge command tests. |
| T4.6 | Update loop report output with merge phase and merge result. | done | `b78afb3` | `work run-all` execution output now includes merge result rows. |
| T4.7 | Add the slice implementation report and validation evidence. | done | this commit | Report added with final validation evidence. |

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
- Slice 2: `go test ./internal/app -run 'AgentWorkLoop|WorkRunAll'` (pass)
- Slice 2: `go test ./internal/app ./internal/core` (pass)
- Slice 2: `make lint-strict` (pass)
- Slice 2: `make verify` (pass)
- Slice 3: `go test ./internal/core -run AgenticReview` (pass)
- Slice 3: `go test ./internal/app -run 'AgenticReview|StartAllReadyAgentWork|WorkRunAll'` (pass)
- Slice 3: `make lint-strict` (pass)
- Slice 3: `go test ./internal/app ./internal/core` (pass)
- Slice 3: `make verify` (pass)
- Slice 4: `go test ./internal/app -run 'StartAllReadyAgentWork|WorkRunAll'` (pass)
- Slice 4: `go test ./internal/core -run 'TrustedAutoMerge|TicketMergeCommand|TicketMergePreflight|IntegrationCycle'` (pass)
- Slice 4: `go test ./internal/app` (pass)
- Slice 4: `make lint-strict` (pass)
- Slice 4: `make verify` (pass)

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
