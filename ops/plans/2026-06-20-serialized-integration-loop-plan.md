# Serialized Integration Loop Plan

Status: delivered

Active Slice: none

Tracker: `ops/trackers/2026-06-20-serialized-integration-loop-tracker.md`

Spec: `ops/specs/2026-06-20-serialized-integration-loop.md`

Model: `ops/specs/2026-06-20-serialized-integration-loop-model.md`

## Purpose

Convert automatic ready-ticket work from a development-only loop into a
serialized integration loop.

The loop must complete one ticket's development, review policy, and merge before
starting the next normal ticket. If review requests repair, the same ticket is
repaired before later work. If review or merge blocks, the loop stops.

## Delivery Rules

- One branch and one PR per slice.
- PRs target `dev`.
- Slice numbering is scoped to this plan and starts at `Slice 1`.
- Each slice updates its implementation report under
  `.prj/reports/slices/serialized-integration-loop/`.
- Each task gets one conventional commit.
- Plan and tracker must be committed before runtime code changes begin.
- Start the next slice only after the previous slice is merged into `dev`.
- Do not introduce parallel execution in this slicing.
- Do not loosen validation, review, or merge gates to keep the loop moving.

## Slice Map

| Slice | Short Name | Branch | PR Title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | cycle policy and state model | `feat/slice-1-cycle-policy-and-state-model` | `feat(slice-1): cycle policy and state model` | `.prj/reports/slices/serialized-integration-loop/slice-1-cycle-policy-and-state-model.md` |
| Slice 2 | one-ticket cycle runner | `feat/slice-2-one-ticket-cycle-runner` | `feat(slice-2): one-ticket cycle runner` | `.prj/reports/slices/serialized-integration-loop/slice-2-one-ticket-cycle-runner.md` |
| Slice 3 | high-level agentic review | `feat/slice-3-high-level-agentic-review` | `feat(slice-3): high-level agentic review` | `.prj/reports/slices/serialized-integration-loop/slice-3-high-level-agentic-review.md` |
| Slice 4 | merge-before-next-ticket | `feat/slice-4-merge-before-next-ticket` | `feat(slice-4): merge before next ticket` | `.prj/reports/slices/serialized-integration-loop/slice-4-merge-before-next-ticket.md` |
| Slice 5 | finite repair cycle | `feat/slice-5-finite-repair-cycle` | `feat(slice-5): finite repair cycle` | `.prj/reports/slices/serialized-integration-loop/slice-5-finite-repair-cycle.md` |
| Slice 6 | observability docs and QA | `feat/slice-6-observability-docs-and-qa` | `feat(slice-6): observability docs and QA` | `.prj/reports/slices/serialized-integration-loop/slice-6-observability-docs-and-qa.md` |

## Slice 1: Cycle Policy and State Model

Goal: define the integration-cycle service boundary and durable state decisions
without changing runtime behavior.

Behavior:

- Core has an explicit integration cycle policy model.
- The model maps existing `review.mode` values to loop behavior:
  `human_review`, `agentic_review`, and `trusted_auto_merge`.
- The loop can describe one selected ticket as one integration cycle.
- The code has stable enums for cycle phase, cycle status, review decision, and
  blocker categories.
- The implementation decision for dedicated cycle persistence vs report-only
  reconstruction is made and documented.

Tasks:

- T1.1 Add core value types for integration cycle mode, phase, status, and
  blocker.
- T1.2 Add policy resolution from existing review workflow configuration.
- T1.3 Add a side-effect-free cycle planning function for one selected ticket.
- T1.4 Decide whether Slice 1 introduces a dedicated `integration_cycles` table
  or defers it.
- T1.5 If a table is introduced, add an additive migration and repository tests.
- T1.6 Add tests for policy mapping across `human_review`, `agentic_review`,
  and `trusted_auto_merge`.
- T1.7 Add docs or reference notes for cycle mode semantics.
- T1.8 Add the slice implementation report and validation evidence.

## Slice 2: One-Ticket Cycle Runner

Goal: make `run-all` execute one ticket cycle at a time and rehydrate state
after each cycle.

Behavior:

- `run-all` no longer processes a stale initial multi-ticket queue directly.
- Each cycle selects exactly one ticket.
- After development result application, the loop rehydrates ticket and policy
  state before deciding whether to continue.
- `human_review` mode stops after producing a review handoff.
- Existing explicit `chc work ticket <ticket>` remains available.
- Existing `--max-iterations` limits integration cycles, not raw development
  executions.

Tasks:

- T2.1 Introduce a cycle runner that selects one ticket per iteration.
- T2.2 Refactor current loop execution to delegate development to the cycle
  runner.
- T2.3 Refresh ticket index and config after each cycle.
- T2.4 Stop `human_review` mode after `ready_for_review` handoff.
- T2.5 Update dry-run output to describe first selected cycle and remaining
  eligible tickets.
- T2.6 Replace tests that expect two development-only processed tickets with
  tests that expect a barrier after the first unfinished integration cycle.
- T2.7 Add tests proving no second ticket starts until the first reaches a
  terminal cycle state.
- T2.8 Add the slice implementation report and validation evidence.

## Slice 3: High-Level Agentic Review

Goal: add the high-level reviewer phase for `agentic_review` mode.

Behavior:

- Review receives ticket, diff, PR/reference, validation, execution report, and
  relevant spec/plan context.
- Review output is structured and persisted as a local review artifact.
- Review blocks only on objective correctness, scope, evidence, contract,
  safety, or mergeability issues.
- Review does not block on cosmetic style, optional refactors, or preferences
  that passed configured validation.
- Clean pass moves the ticket to `ready_for_merge`.
- Insufficient evidence or human-required outcomes block the cycle.

Tasks:

- T3.1 Define the reviewer context artifact and prompt contract.
- T3.2 Add structured review result parsing and validation.
- T3.3 Persist local review artifacts under `.chc/`.
- T3.4 Apply `pass`, `needs_changes`, `insufficient_evidence`, and
  `human_required` outcomes to ticket state.
- T3.5 Add non-blocking note handling with existing comment limits.
- T3.6 Add fake-reviewer tests for clean pass and high-level blocking findings.
- T3.7 Add tests proving cosmetic findings do not block when validation passed.
- T3.8 Add the slice implementation report and validation evidence.

## Slice 4: Merge Before Next Ticket

Goal: require successful merge before the outer loop selects later normal work.

Behavior:

- `agentic_review` mode invokes existing merge preflight and merge execution
  after review pass.
- `trusted_auto_merge` mode invokes merge preflight and merge execution after
  successful development evidence.
- Merge success marks the cycle done.
- Merge failure blocks the cycle and stops the loop.
- Worktree cleanup after merge continues to run according to existing policy.

Tasks:

- T4.1 Add merge phase orchestration to the cycle runner.
- T4.2 Reuse existing ticket merge command/preflight behavior instead of
  duplicating provider logic.
- T4.3 Ensure `trusted_auto_merge` still requires merge before continuing.
- T4.4 Add tests proving ticket 2 starts only after ticket 1 has merge metadata
  and status `done`.
- T4.5 Add tests for merge preflight blocked, provider merge failed, and cleanup
  after merge.
- T4.6 Update loop report output with merge phase and merge result.
- T4.7 Add the slice implementation report and validation evidence.

## Slice 5: Finite Repair Cycle

Goal: repair the current ticket when review requests changes without allowing
unbounded dev-review loops.

Behavior:

- `needs_changes` returns the same ticket to development with repair handoff
  metadata.
- Repair reuses the same branch, pull request, and worktree.
- The default maximum repair cycle count is one.
- If repair fails review again, the ticket blocks and later work does not start.
- Repair count is durable.

Tasks:

- T5.1 Add durable repair count handling.
- T5.2 Feed review findings into the repair worker context.
- T5.3 Ensure repair selection outranks all later normal work.
- T5.4 Re-run review after repair.
- T5.5 Block after the configured/default repair limit.
- T5.6 Add tests for one successful repair, repeated review failure, missing
  repair handoff, and later-ticket blocking.
- T5.7 Add the slice implementation report and validation evidence.

## Slice 6: Observability Docs and QA

Goal: make the serialized integration loop inspectable and prove it in QA.

Behavior:

- Event streams distinguish outer loop events from inner cycle events.
- Dash shows selected ticket, phase, review outcome, merge outcome, and blocked
  reason.
- Reports explain why the loop stopped or continued.
- Documentation describes the loop as `dev -> review policy -> merge` per
  ticket.
- QA scenarios prove the barrier with at least two tickets.

Tasks:

- T6.1 Update agent event types/messages for cycle selection, review, repair,
  merge, and rehydration.
- T6.2 Update Dash rendering for active cycle phase and stop reason.
- T6.3 Update loop and execution reports with cycle details.
- T6.4 Update user guide, ticket loop reference, review workflow reference, and
  changelog.
- T6.5 Add or update resettable QA fixture with two dependent tickets.
- T6.6 Add smoke or manual runbook proving ticket 2 does not start until ticket
  1 is merged.
- T6.7 Run `make verify` and relevant QA smoke commands.
- T6.8 Add the slice implementation report and final cycle report.

## Validation Strategy

- Unit tests for policy mapping, cycle planning, review result validation, and
  repair limits.
- Service tests for one-cycle-at-a-time loop behavior.
- CLI tests for dry-run and run-all output.
- Fake provider tests for review/merge pass and blocked cases.
- Worktree-mode tests proving the same branch/worktree is reused for repair.
- QA runbook proving the two-ticket barrier end to end.
- `make lint-strict`
- `make verify`

## Open Decisions

- Whether Slice 1 introduces a dedicated `integration_cycles` table or relies on
  report reconstruction from existing durable state.
- Whether `chc work ticket <ticket>` should remain development-only or gain a
  separate full-cycle single-ticket command.
- Whether project templates should eventually default to `agentic_review`,
  `trusted_auto_merge`, or remain `human_review`.
