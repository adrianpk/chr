# Single Ticket Execution Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`
Priority draft: `.prj/drafts/chc-command-surfaces-and-priority.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm `Agent Runtime Boundary` is completed.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Execution Record Model | Delivered | `feat/slice-1-execution-record-model` | `feat(slice-1): execution record model` | `.prj/reports/slices/single-ticket-execution/slice-1-execution-record-model.md` | Repository PR #34. |
| Slice 2: Execution Repository | Delivered | `feat/slice-2-execution-repository` | `feat(slice-2): execution repository` | `.prj/reports/slices/single-ticket-execution/slice-2-execution-repository.md` | Repository PR #35. |
| Slice 3: Ticket Start Preflight | Delivered | `feat/slice-3-ticket-start-preflight` | `feat(slice-3): ticket start preflight` | `.prj/reports/slices/single-ticket-execution/slice-3-ticket-start-preflight.md` | Repository PR #36. |
| Slice 4: Harness Invocation Contract | Delivered | `feat/slice-4-harness-invocation-contract` | `feat(slice-4): harness invocation contract` | `.prj/reports/slices/single-ticket-execution/slice-4-harness-invocation-contract.md` | Repository PR #37. |
| Slice 5: Agent Harness Invocation | Reviewing | `feat/slice-5-agent-harness-invocation` | `feat(slice-5): agent harness invocation` | `.prj/reports/slices/single-ticket-execution/slice-5-agent-harness-invocation.md` | Run one bounded harness attempt. |
| Slice 6: Execution Result Capture | Planned | `feat/slice-6-execution-result-capture` | `feat(slice-6): execution result capture` | `.prj/reports/slices/single-ticket-execution/slice-6-execution-result-capture.md` | Capture structured result artifacts. |
| Slice 7: Execution Failure Handling | Planned | `feat/slice-7-execution-failure-handling` | `feat(slice-7): execution failure handling` | `.prj/reports/slices/single-ticket-execution/slice-7-execution-failure-handling.md` | Record clear execution failure states. |

## Slice 1: Execution Record Model

Pre-flight:

- Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
- Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- Slicing: `Single Ticket Execution`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-execution-record-model`
- Expected PR title: `feat(slice-1): execution record model`
- Expected report: `.prj/reports/slices/single-ticket-execution/slice-1-execution-record-model.md`

Candidate tasks:

- [x] Add execution record domain types.
- [x] Define execution status values.
- [x] Define ticket, project, runtime, and artifact references.
- [x] Add focused model tests.
- [x] Document execution record contract.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `7c73f87` `feat(execution): add execution record model`
- T1.2: `7c73f87` `feat(execution): add execution record model`
- T1.3: `7c73f87` `feat(execution): add execution record model`
- T1.4: `59c1ec7` `test(execution): cover execution record model`
- T1.5: `6fd8a01` `docs(execution): document execution record contract`
- T1.6: `accbae6` `docs(execution): report execution record model slice`
- T1.7: `docs(execution): record execution model validation`

## Slice 2: Execution Repository

Pre-flight:

- Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
- Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- Slicing: `Single Ticket Execution`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-execution-repository`
- Expected PR title: `feat(slice-2): execution repository`
- Expected report: `.prj/reports/slices/single-ticket-execution/slice-2-execution-repository.md`

Candidate tasks:

- [x] Add execution record migration.
- [x] Add execution repository APIs.
- [x] Store project ID and ticket key references.
- [x] Add focused repository tests.
- [x] Document execution persistence layout.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `69a13bc` `feat(execution): add execution record migration`
- T2.2: `5e8b502` `feat(execution): add execution repository`
- T2.3: `5e8b502` `feat(execution): add execution repository`
- T2.4: `b28cfa2` `test(execution): cover execution repository`
- T2.5: `318c5b5` `docs(execution): document execution persistence`
- T2.6: `9c2f75a` `docs(execution): report execution repository slice`
- T2.7: `docs(execution): record execution repository validation`

## Slice 3: Ticket Start Preflight

Pre-flight:

- Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
- Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- Slicing: `Single Ticket Execution`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-ticket-start-preflight`
- Expected PR title: `feat(slice-3): ticket start preflight`
- Expected report: `.prj/reports/slices/single-ticket-execution/slice-3-ticket-start-preflight.md`

Candidate tasks:

- [x] Add `chc ticket start`.
- [x] Resolve project context through the existing resolver.
- [x] Reconcile tickets before selecting work.
- [x] Select one indexed ticket by key.
- [x] Run runtime validation before execution.
- [x] Create pending execution record.
- [x] Support dry-run output.
- [x] Add focused CLI tests.
- [x] Document command behavior.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `e75bca1` `feat(cli): add ticket start preflight`
- T3.2: `e75bca1` `feat(cli): add ticket start preflight`
- T3.3: `e75bca1` `feat(cli): add ticket start preflight`
- T3.4: `e75bca1` `feat(cli): add ticket start preflight`
- T3.5: `e75bca1` `feat(cli): add ticket start preflight`
- T3.6: `e75bca1` `feat(cli): add ticket start preflight`
- T3.7: `e75bca1` `feat(cli): add ticket start preflight`
- T3.8: `74e6875` `test(cli): cover ticket start preflight`
- T3.9: `7161384` `docs(execution): document ticket start preflight`
- T3.10: `7161384` `docs(execution): document ticket start preflight`
- T3.11: `docs(execution): record ticket start validation`

## Slice 4: Harness Invocation Contract

Pre-flight:

- Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
- Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- Slicing: `Single Ticket Execution`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-harness-invocation-contract`
- Expected PR title: `feat(slice-4): harness invocation contract`
- Expected report: `.prj/reports/slices/single-ticket-execution/slice-4-harness-invocation-contract.md`

Candidate tasks:

- [x] Add harness invocation model.
- [x] Render Docker build/run command arguments from `RuntimePlan`.
- [x] Include ticket key and execution ID in artifact paths.
- [x] Keep command construction side-effect free.
- [x] Add focused command rendering tests.
- [x] Document harness invocation contract.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `4d1d0ed` `feat(execution): add harness invocation contract`
- T4.2: `4d1d0ed` `feat(execution): add harness invocation contract`
- T4.3: `4d1d0ed` `feat(execution): add harness invocation contract`
- T4.4: `4d1d0ed` `feat(execution): add harness invocation contract`
- T4.5: `509aff9` `test(execution): cover harness invocation contract`
- T4.6: `2e34812` `docs(execution): document harness invocation contract`
- T4.7: `2e34812` `docs(execution): document harness invocation contract`
- T4.8: `docs(execution): record harness invocation validation`

## Slice 5: Agent Harness Invocation

Pre-flight:

- Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
- Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- Slicing: `Single Ticket Execution`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-agent-harness-invocation`
- Expected PR title: `feat(slice-5): agent harness invocation`
- Expected report: `.prj/reports/slices/single-ticket-execution/slice-5-agent-harness-invocation.md`

Candidate tasks:

- [x] Add execution runner service.
- [x] Invoke harness command from runtime plan.
- [x] Capture stdout and stderr to execution artifacts.
- [x] Update execution status on start, success, and failure.
- [x] Preserve dry-run behavior.
- [x] Add focused runner tests with a fake executor.
- [x] Document invocation limits.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `594a7f9` `feat(execution): run ticket harness`
- T5.2: `594a7f9` `feat(execution): run ticket harness`
- T5.3: `594a7f9` `feat(execution): run ticket harness`
- T5.4: `594a7f9` `feat(execution): run ticket harness`
- T5.5: `594a7f9` `feat(execution): run ticket harness`
- T5.6: `b640680` `test(execution): cover harness runner`
- T5.7: `8defcdf` `docs(execution): document harness invocation run`
- T5.8: `8defcdf` `docs(execution): document harness invocation run`
- T5.9: `docs(execution): record harness invocation run validation`

## Slice 6: Execution Result Capture

Pre-flight:

- Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
- Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- Slicing: `Single Ticket Execution`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-execution-result-capture`
- Expected PR title: `feat(slice-6): execution result capture`
- Expected report: `.prj/reports/slices/single-ticket-execution/slice-6-execution-result-capture.md`

Candidate tasks:

- [ ] Define result envelope shape.
- [ ] Write result artifacts under runtime run directories.
- [ ] Record artifact paths in execution records.
- [ ] Handle missing or malformed result output.
- [ ] Add focused result capture tests.
- [ ] Document execution artifact semantics.
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

## Slice 7: Execution Failure Handling

Pre-flight:

- Plan: `.prj/plans/slicing-single-ticket-execution-plan.md`
- Tracker: `.prj/trackers/slicing-single-ticket-execution-tracker.md`
- Slicing: `Single Ticket Execution`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-execution-failure-handling`
- Expected PR title: `feat(slice-7): execution failure handling`
- Expected report: `.prj/reports/slices/single-ticket-execution/slice-7-execution-failure-handling.md`

Candidate tasks:

- [ ] Classify preflight, runtime, harness, validation, and result-capture failures.
- [ ] Record failure reason and diagnostic paths.
- [ ] Print concise failure output from `ticket start`.
- [ ] Add focused failure tests.
- [ ] Document failure semantics.
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
- Slice 1 validation: `go test ./internal/chc` pass.
- Slice 1 validation: `make verify` pass.
- Slice 2 validation: `go test ./internal/chc` pass.
- Slice 2 validation: `make verify` pass.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- This slicing records execution attempts but does not apply execution results
  to ticket Markdown.
- Each implementation slice must include a Markdown implementation report
  before PR creation.
