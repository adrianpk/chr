# Codex Runtime Worker Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm prerequisite execution and result-application slicings are completed.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Worker Contract Model | Delivered | `feat/codex-worker-slice-1-contract-model` | `feat(codex-worker-slice-1): contract model` | `.prj/reports/slices/codex-runtime-worker/slice-1-contract-model.md` | Repository PR #49. Naming corrected for later slices. |
| Slice 2: Worker Context Materialization | Delivered | `feat/slice-2-worker-context-materialization` | `feat(slice-2): worker context materialization` | `.prj/reports/slices/codex-runtime-worker/slice-2-context-materialization.md` | Repository PR #50. |
| Slice 3: Harness Worker Invocation | Delivered | `feat/slice-3-harness-worker-invocation` | `feat(slice-3): harness worker invocation` | `.prj/reports/slices/codex-runtime-worker/slice-3-harness-invocation.md` | Repository PR #51. |
| Slice 4: Result Envelope Changed Files | Delivered | `feat/slice-4-result-changed-files` | `feat(slice-4): result changed files` | `.prj/reports/slices/codex-runtime-worker/slice-4-result-changed-files.md` | Repository PR #52. |
| Slice 5: Changed File Discovery | Planned | `feat/slice-5-changed-file-discovery` | `feat(slice-5): changed file discovery` | `.prj/reports/slices/codex-runtime-worker/slice-5-changed-file-discovery.md` | Discover actual workspace changes. |
| Slice 6: Post-Worker Validation | Planned | `feat/slice-6-post-worker-validation` | `feat(slice-6): post worker validation` | `.prj/reports/slices/codex-runtime-worker/slice-6-post-worker-validation.md` | Validate after Codex edits. |
| Slice 7: Worker Failure Classification and Docs | Planned | `feat/slice-7-worker-failure-docs` | `feat(slice-7): worker failure classification and docs` | `.prj/reports/slices/codex-runtime-worker/slice-7-failure-docs.md` | Document and classify worker failures. |

## Slice 1: Worker Contract Model

Pre-flight:

- Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
- Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
- Slicing: `Codex Runtime Worker`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-worker-contract-model`
- Expected PR title: `feat(slice-1): worker contract model`
- Expected report: `.prj/reports/slices/codex-runtime-worker/slice-1-contract-model.md`

Candidate tasks:

- [x] Add worker command/domain types.
- [x] Define ticket, execution, result, log, and validation paths.
- [x] Render the shell command that invokes Codex inside the container.
- [x] Validate required command inputs.
- [x] Add focused command-rendering tests.
- [x] Document the worker contract.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `87d430a` `feat(codex-worker): add worker command contract`
- T1.2: `87d430a` `feat(codex-worker): add worker command contract`
- T1.3: `87d430a` `feat(codex-worker): add worker command contract`
- T1.4: `87d430a` `feat(codex-worker): add worker command contract`
- T1.5: `67b32c9` `test(codex-worker): cover worker command contract`
- T1.6: `docs(codex-worker): document worker contract`
- T1.7: `docs(codex-worker): document worker contract`
- T1.8: `docs(codex-worker): record worker contract validation`

## Slice 2: Worker Context Materialization

Pre-flight:

- Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
- Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
- Slicing: `Codex Runtime Worker`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-worker-context-materialization`
- Expected PR title: `feat(slice-2): worker context materialization`
- Expected report: `.prj/reports/slices/codex-runtime-worker/slice-2-context-materialization.md`

Candidate tasks:

- [x] Create a worker prompt/context artifact path.
- [x] Render ticket key, ticket path, execution ID, and result contract.
- [x] Keep the context deterministic and reviewable.
- [x] Add focused materialization tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `9553dcd` `feat(codex-worker): add worker context materializer`
- T2.2: `9553dcd` `feat(codex-worker): add worker context materializer`
- T2.3: `9553dcd` `feat(codex-worker): add worker context materializer`
- T2.4: `379702d` `test(codex-worker): cover worker context materializer`
- T2.5: `docs(codex-worker): document worker context materialization`
- T2.6: `docs(codex-worker): record worker context validation`

## Slice 3: Harness Worker Invocation

Pre-flight:

- Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
- Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
- Slicing: `Codex Runtime Worker`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-harness-worker-invocation`
- Expected PR title: `feat(slice-3): harness worker invocation`
- Expected report: `.prj/reports/slices/codex-runtime-worker/slice-3-harness-invocation.md`

Candidate tasks:

- [x] Add worker command into harness invocation rendering.
- [x] Preserve Docker build/run behavior.
- [x] Pass worker context and artifact paths through environment or arguments.
- [x] Keep command construction side-effect free.
- [x] Add focused invocation tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `2803e96` `feat(codex-worker): invoke worker through harness`
- T3.2: `2803e96` `feat(codex-worker): invoke worker through harness`
- T3.3: `2803e96` `feat(codex-worker): invoke worker through harness`
- T3.4: `2803e96` `feat(codex-worker): invoke worker through harness`
- T3.5: `42f8037` `test(codex-worker): cover harness worker invocation`
- T3.6: `docs(codex-worker): document harness worker invocation`
- T3.7: `docs(codex-worker): record harness worker invocation validation`

## Slice 4: Result Envelope Changed Files

Pre-flight:

- Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
- Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
- Slicing: `Codex Runtime Worker`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-result-changed-files`
- Expected PR title: `feat(slice-4): result changed files`
- Expected report: `.prj/reports/slices/codex-runtime-worker/slice-4-result-changed-files.md`

Candidate tasks:

- [x] Extend execution result types with changed files.
- [x] Preserve strict JSON decoding.
- [x] Validate changed-file paths and summaries.
- [x] Map result changed files into result application.
- [x] Add focused result-loading tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `2305bc2` `feat(codex-worker): carry result changed files`
- T4.2: `2305bc2` `feat(codex-worker): carry result changed files`
- T4.3: `2305bc2` `feat(codex-worker): carry result changed files`
- T4.4: `2305bc2` `feat(codex-worker): carry result changed files`
- T4.5: `16e8c5f` `test(codex-worker): cover result changed files`
- T4.6: `docs(codex-worker): document result changed files`
- T4.7: `docs(codex-worker): record result changed files validation`

## Slice 5: Changed File Discovery

Pre-flight:

- Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
- Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
- Slicing: `Codex Runtime Worker`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-changed-file-discovery`
- Expected PR title: `feat(slice-5): changed file discovery`
- Expected report: `.prj/reports/slices/codex-runtime-worker/slice-5-changed-file-discovery.md`

Candidate tasks:

- [ ] Inspect git changed files after execution.
- [ ] Scope discovery to the project workspace.
- [ ] Handle clean worktrees.
- [ ] Avoid reverting or mutating changes.
- [ ] Add focused discovery tests.
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

## Slice 6: Post-Worker Validation

Pre-flight:

- Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
- Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
- Slicing: `Codex Runtime Worker`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-post-worker-validation`
- Expected PR title: `feat(slice-6): post worker validation`
- Expected report: `.prj/reports/slices/codex-runtime-worker/slice-6-post-worker-validation.md`

Candidate tasks:

- [ ] Run the configured validation command after the worker command.
- [ ] Preserve validation output artifact paths.
- [ ] Classify validation failures distinctly.
- [ ] Ensure result capture sees validation failure correctly.
- [ ] Add focused execution runner tests.
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

## Slice 7: Worker Failure Classification and Docs

Pre-flight:

- Plan: `.prj/plans/slicing-codex-runtime-worker-plan.md`
- Tracker: `.prj/trackers/slicing-codex-runtime-worker-tracker.md`
- Slicing: `Codex Runtime Worker`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-worker-failure-docs`
- Expected PR title: `feat(slice-7): worker failure classification and docs`
- Expected report: `.prj/reports/slices/codex-runtime-worker/slice-7-failure-docs.md`

Candidate tasks:

- [ ] Add Codex worker failure classification.
- [ ] Surface worker diagnostics in execution output.
- [ ] Document container prerequisites and result contract.
- [ ] Document dry-run visibility.
- [ ] Add focused failure tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending
- T7.6: pending
- T7.7: pending

## Validation Record

- Planning validation: documentation diff reviewed.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- This slicing makes the real coding worker explicit before the foreground loop
  runner.
