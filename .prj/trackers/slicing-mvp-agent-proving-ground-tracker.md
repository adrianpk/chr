# MVP Agent Proving Ground Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify temporary MVP proving-ground need.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Todolist Fixture Project | Delivered | `feat/slice-1-proving-ground-fixture` | `feat(slice-1): proving ground fixture` | `.prj/reports/slices/mvp-agent-proving-ground/slice-1-proving-ground-fixture.md` | Repository PR #72. |
| Slice 2: Direct Proving Objective Contract | Planned | `feat/slice-2-proving-objective-contract` | `feat(slice-2): proving objective contract` | `.prj/reports/slices/mvp-agent-proving-ground/slice-2-proving-objective-contract.md` | Direct objective input. |
| Slice 3: Sandbox Codex Runner | Planned | `feat/slice-3-proving-sandbox-runner` | `feat(slice-3): proving sandbox runner` | `.prj/reports/slices/mvp-agent-proving-ground/slice-3-proving-sandbox-runner.md` | Codex in sandbox. |
| Slice 4: Validation Runner | Planned | `feat/slice-4-proving-validation-runner` | `feat(slice-4): proving validation runner` | `.prj/reports/slices/mvp-agent-proving-ground/slice-4-proving-validation-runner.md` | Compile, lint, tests. |
| Slice 5: Minimal Ralph Loop | Planned | `feat/slice-5-proving-ralph-loop` | `feat(slice-5): proving ralph loop` | `.prj/reports/slices/mvp-agent-proving-ground/slice-5-proving-ralph-loop.md` | Bounded retries. |
| Slice 6: End-to-End Agent Proof | Planned | `feat/slice-6-proving-e2e-agent-proof` | `feat(slice-6): proving e2e agent proof` | `.prj/reports/slices/mvp-agent-proving-ground/slice-6-proving-e2e-agent-proof.md` | Final proof. |

## Slice 1: Todolist Fixture Project

Pre-flight:

- Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
- Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
- Slicing: `MVP Agent Proving Ground`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-proving-ground-fixture`
- Expected PR title: `feat(slice-1): proving ground fixture`
- Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-1-proving-ground-fixture.md`

Candidate tasks:

- [x] Add fixture source files.
- [x] Include a deliberate backend bug.
- [x] Include a missing "mark task completed" feature.
- [x] Include `.chc/` runtime config and hand-authored sandbox Dockerfile.
- [x] Include compile, lint, and test validation commands.
- [x] Include fixture materialization documentation or helper.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `3b8e4de`
- T1.2: `3b8e4de`
- T1.3: `3b8e4de`
- T1.4: `3b8e4de`
- T1.5: `3b8e4de`
- T1.6: `3b8e4de`
- T1.7: `3b8e4de`
- T1.8: `3b8e4de`

## Slice 2: Direct Proving Objective Contract

Pre-flight:

- Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
- Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
- Slicing: `MVP Agent Proving Ground`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-proving-objective-contract`
- Expected PR title: `feat(slice-2): proving objective contract`
- Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-2-proving-objective-contract.md`

Candidate tasks:

- [ ] Define bugfix objective.
- [ ] Define feature objective.
- [ ] Define expected validation commands.
- [ ] Define artifact paths.
- [ ] Define success, failed, and blocked outcomes.
- [ ] Add focused validation if practical.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T2.1: pending
- T2.2: pending
- T2.3: pending
- T2.4: pending
- T2.5: pending
- T2.6: pending
- T2.7: pending
- T2.8: pending

## Slice 3: Sandbox Codex Runner

Pre-flight:

- Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
- Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
- Slicing: `MVP Agent Proving Ground`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-proving-sandbox-runner`
- Expected PR title: `feat(slice-3): proving sandbox runner`
- Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-3-proving-sandbox-runner.md`

Candidate tasks:

- [ ] Build or reuse fixture agent image.
- [ ] Materialize execution context.
- [ ] Run `codex exec` in the sandbox.
- [ ] Capture logs and result artifacts.
- [ ] Avoid clone and push behavior.
- [ ] Add dry-run or command-rendering tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending
- T3.2: pending
- T3.3: pending
- T3.4: pending
- T3.5: pending
- T3.6: pending
- T3.7: pending
- T3.8: pending

## Slice 4: Validation Runner

Pre-flight:

- Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
- Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
- Slicing: `MVP Agent Proving Ground`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-proving-validation-runner`
- Expected PR title: `feat(slice-4): proving validation runner`
- Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-4-proving-validation-runner.md`

Candidate tasks:

- [ ] Run fixture validation commands.
- [ ] Capture stdout, stderr, and exit codes.
- [ ] Write validation artifacts.
- [ ] Classify validation failure for Ralph Loop input.
- [ ] Add success and failure tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending
- T4.2: pending
- T4.3: pending
- T4.4: pending
- T4.5: pending
- T4.6: pending
- T4.7: pending

## Slice 5: Minimal Ralph Loop

Pre-flight:

- Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
- Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
- Slicing: `MVP Agent Proving Ground`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-proving-ralph-loop`
- Expected PR title: `feat(slice-5): proving ralph loop`
- Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-5-proving-ralph-loop.md`

Candidate tasks:

- [ ] Define max attempts.
- [ ] Append validation failure context to the next prompt.
- [ ] Stop on pass, repeated failure, blocked state, or max attempts.
- [ ] Preserve attempt evidence.
- [ ] Keep loop fixture-scoped.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending
- T5.2: pending
- T5.3: pending
- T5.4: pending
- T5.5: pending
- T5.6: pending
- T5.7: pending

## Slice 6: End-to-End Agent Proof

Pre-flight:

- Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
- Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
- Slicing: `MVP Agent Proving Ground`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-proving-e2e-agent-proof`
- Expected PR title: `feat(slice-6): proving e2e agent proof`
- Expected report: `.prj/reports/slices/mvp-agent-proving-ground/slice-6-proving-e2e-agent-proof.md`

Candidate tasks:

- [ ] Materialize the fixture under `_workspace/testing/projects/todolist`.
- [ ] Run the bugfix objective.
- [ ] Run the feature objective.
- [ ] Verify compile, lint, and tests pass.
- [ ] Capture final evidence.
- [ ] Document limitations and follow-up generalization.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending
- T6.2: pending
- T6.3: pending
- T6.4: pending
- T6.5: pending
- T6.6: pending
- T6.7: pending
- T6.8: pending

## Validation Record

- Planning validation: documentation diff reviewed.

## Notes

- This slicing intentionally opens a temporary direct execution window.
- Hardcoding is acceptable inside this slicing when it proves the agent can
  solve code tasks through sandboxed Codex and bounded validation retries.
