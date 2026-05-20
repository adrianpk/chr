# MVP Agent Proving Ground Slicing Tracker

Status: delivered

Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify temporary MVP proving-ground need.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [x] Complete all planned slices.
- [x] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Todolist Fixture Project | Delivered | `feat/slice-1-proving-ground-fixture` | `feat(slice-1): proving ground fixture` | `.prj/reports/slices/mvp-agent-proving-ground/slice-1-proving-ground-fixture.md` | Repository PR #72. |
| Slice 2: Direct Proving Objective Contract | Delivered | `feat/slice-2-proving-objective-contract` | `feat(slice-2): proving objective contract` | `.prj/reports/slices/mvp-agent-proving-ground/slice-2-proving-objective-contract.md` | Repository PR #73. |
| Slice 3: Sandbox Codex Runner | Delivered | `feat/slice-3-proving-sandbox-runner` | `feat(slice-3): proving sandbox runner` | `.prj/reports/slices/mvp-agent-proving-ground/slice-3-proving-sandbox-runner.md` | Repository PR #74. |
| Slice 4: Validation Runner | Delivered | `feat/slice-4-proving-validation-runner` | `feat(slice-4): proving validation runner` | `.prj/reports/slices/mvp-agent-proving-ground/slice-4-proving-validation-runner.md` | Repository PR #75 merged. |
| Slice 5: Minimal Ralph Loop | Delivered | `feat/slice-5-proving-ralph-loop` | `feat(slice-5): proving ralph loop` | `.prj/reports/slices/mvp-agent-proving-ground/slice-5-proving-ralph-loop.md` | Repository PR #76 merged. |
| Slice 6: End-to-End Agent Proof | Delivered | `feat/slice-6-proving-e2e-agent-proof` | `feat(slice-6): proving e2e agent proof` | `.prj/reports/slices/mvp-agent-proving-ground/slice-6-proving-e2e-agent-proof.md` | Repository PR #77 merged. |

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

- [x] Define bugfix objective.
- [x] Define feature objective.
- [x] Define expected validation commands.
- [x] Define artifact paths.
- [x] Define success, failed, and blocked outcomes.
- [x] Add focused validation if practical.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `564189c`
- T2.2: `564189c`
- T2.3: `564189c`
- T2.4: `564189c`
- T2.5: `564189c`
- T2.6: `564189c`
- T2.7: `564189c`
- T2.8: `564189c`

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

- [x] Build or reuse fixture agent image.
- [x] Materialize execution context.
- [x] Run `codex exec` in the sandbox.
- [x] Capture logs and result artifacts.
- [x] Avoid clone and push behavior.
- [x] Add dry-run or command-rendering tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `021879b`
- T3.2: `021879b`
- T3.3: `021879b`
- T3.4: `021879b`
- T3.5: `021879b`
- T3.6: `021879b`
- T3.7: `021879b`
- T3.8: `021879b`

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

- [x] Run fixture validation commands.
- [x] Capture stdout, stderr, and exit codes.
- [x] Write validation artifacts.
- [x] Classify validation failure for Ralph Loop input.
- [x] Add success and failure tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `f4dcd53`
- T4.2: `f4dcd53`
- T4.3: `f4dcd53`
- T4.4: `f4dcd53`
- T4.5: `f4dcd53`
- T4.6: `f4dcd53`
- T4.7: `f4dcd53`

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

- [x] Define max attempts.
- [x] Append validation failure context to the next prompt.
- [x] Stop on pass, repeated failure, blocked state, or max attempts.
- [x] Preserve attempt evidence.
- [x] Keep loop fixture-scoped.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `b41d3ca`
- T5.2: `b41d3ca`
- T5.3: `b41d3ca`
- T5.4: `b41d3ca`
- T5.5: `b41d3ca`
- T5.6: `b41d3ca`
- T5.7: `b41d3ca`

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

- [x] Materialize the fixture under `_workspace/testing/projects/todolist`.
- [x] Run the bugfix objective.
- [x] Run the feature objective.
- [x] Verify compile, lint, and tests pass.
- [x] Capture final evidence.
- [x] Document limitations and follow-up generalization.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: `9c374da`
- T6.2: `9c374da`
- T6.3: `9c374da`
- T6.4: `9c374da`
- T6.5: `9c374da`
- T6.6: `9c374da`
- T6.7: `9c374da`
- T6.8: `9c374da`

## Validation Record

- Planning validation: documentation diff reviewed.

## Notes

- This slicing intentionally opens a temporary direct execution window.
- Hardcoding is acceptable inside this slicing when it proves the agent can
  solve code tasks through sandboxed Codex and bounded validation retries.
