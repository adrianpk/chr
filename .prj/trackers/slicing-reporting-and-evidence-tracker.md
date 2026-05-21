# Reporting and Evidence Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm execution, result application, loop, worker, and notification prerequisites are delivered.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Execution Report Format | Delivered | `feat/slice-1-execution-report-format` | `feat(slice-1): execution report format` | `.prj/reports/slices/reporting-and-evidence/slice-1-execution-report-format.md` | Repository PR #118 merged. |
| Slice 2: Execution Report Artifact Writer | Delivered | `feat/slice-2-execution-report-writer` | `feat(slice-2): execution report writer` | `.prj/reports/slices/reporting-and-evidence/slice-2-execution-report-writer.md` | Repository PR #119 merged. |
| Slice 3: Validation and Changed-File Summaries | Delivered | `feat/slice-3-evidence-summaries` | `feat(slice-3): evidence summaries` | `.prj/reports/slices/reporting-and-evidence/slice-3-evidence-summaries.md` | Repository PR #120 merged. |
| Slice 4: Loop Report Format | Delivered | `feat/slice-4-loop-report-format` | `feat(slice-4): loop report format` | `.prj/reports/slices/reporting-and-evidence/slice-4-loop-report-format.md` | Repository PR #121 merged. |
| Slice 5: Report Listing | Delivered | `feat/slice-5-report-listing` | `feat(slice-5): report listing` | `.prj/reports/slices/reporting-and-evidence/slice-5-report-listing.md` | Repository PR #122 merged. |
| Slice 6: Report Lookup | Delivered | `feat/slice-6-report-lookup` | `feat(slice-6): report lookup` | `.prj/reports/slices/reporting-and-evidence/slice-6-report-lookup.md` | Repository PR #123 merged. |
| Slice 7: Review Handoff Summary | Delivered | `feat/slice-7-review-handoff-summary` | `feat(slice-7): review handoff summary` | `.prj/reports/slices/reporting-and-evidence/slice-7-review-handoff-summary.md` | Repository PR #124 merged. |
| Slice 8: Reporting Documentation | Reviewing | `feat/slice-8-reporting-docs` | `feat(slice-8): reporting docs` | `.prj/reports/slices/reporting-and-evidence/slice-8-reporting-docs.md` | PR pending. |

## Slice 1: Execution Report Format

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-execution-report-format`
- Expected PR title: `feat(slice-1): execution report format`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-1-execution-report-format.md`

Candidate tasks:

- [x] Add execution report summary model.
- [x] Add deterministic Markdown renderer.
- [x] Include execution, ticket, runtime, artifact, changed-file, validation, and failure fields.
- [x] Add focused model and renderer tests.
- [x] Document the execution report format.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `71ff968`
- T1.2: `71ff968`
- T1.3: `71ff968`
- T1.4: `71ff968`
- T1.5: `4d62ce9`
- T1.6: `26fcc64`
- T1.7: `make verify`

## Slice 2: Execution Report Artifact Writer

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-execution-report-writer`
- Expected PR title: `feat(slice-2): execution report writer`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-2-execution-report-writer.md`

Candidate tasks:

- [x] Define report artifact paths.
- [x] Write Markdown reports from existing execution evidence.
- [x] Avoid overwriting unrelated files.
- [x] Add writer tests.
- [x] Document artifact placement.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `f82dc2f`
- T2.2: `f82dc2f`
- T2.3: `f82dc2f`
- T2.4: `f82dc2f`
- T2.5: `561d4e7`
- T2.6: `docs(slice-2): record execution report writer`
- T2.7: `make verify`

## Slice 3: Validation and Changed-File Summaries

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-evidence-summaries`
- Expected PR title: `feat(slice-3): evidence summaries`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-3-evidence-summaries.md`

Candidate tasks:

- [x] Summarize validation output paths and status.
- [x] Summarize changed files from existing evidence.
- [x] Keep summaries bounded and deterministic.
- [x] Add focused tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `c2e483d`
- T3.2: `c2e483d`
- T3.3: `c2e483d`
- T3.4: `c2e483d`
- T3.5: `bc82aa8`
- T3.6: `make verify`

## Slice 4: Loop Report Format

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-loop-report-format`
- Expected PR title: `feat(slice-4): loop report format`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-4-loop-report-format.md`

Candidate tasks:

- [x] Add loop report summary model.
- [x] Render deterministic Markdown.
- [x] Include loop status, counts, failure reason, timings, and related executions when available.
- [x] Add focused tests.
- [x] Document the loop report format.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T4.1: `9d38525`
- T4.2: `9d38525`
- T4.3: `9d38525`
- T4.4: `9d38525`
- T4.5: `b10ab17`
- T4.6: `bdb7781`
- T4.7: `make verify`

## Slice 5: Report Listing

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-report-listing`
- Expected PR title: `feat(slice-5): report listing`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-5-report-listing.md`

Candidate tasks:

- [x] Add `chc report` command group.
- [x] Add `chc report list`.
- [x] Support report filters.
- [x] Print stable output.
- [x] Add CLI tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T5.1: `baa4010`
- T5.2: `baa4010`
- T5.3: `baa4010`
- T5.4: `baa4010`
- T5.5: `baa4010`
- T5.6: `b597a8e`
- T5.7: `make verify`

## Slice 6: Report Lookup

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-report-lookup`
- Expected PR title: `feat(slice-6): report lookup`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-6-report-lookup.md`

Candidate tasks:

- [x] Add `chc report show`.
- [x] Resolve reports by report id, ticket key, execution id, or loop id.
- [x] Print report body or metadata predictably.
- [x] Add clear errors for ambiguous or missing lookups.
- [x] Add CLI tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T6.1: `1fe77da`
- T6.2: `1fe77da`
- T6.3: `1fe77da`
- T6.4: `1fe77da`
- T6.5: `1fe77da`
- T6.6: `docs(slice-6): record report lookup`
- T6.7: `make verify`

## Slice 7: Review Handoff Summary

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-review-handoff-summary`
- Expected PR title: `feat(slice-7): review handoff summary`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-7-review-handoff-summary.md`

Candidate tasks:

- [x] Summarize status, changed files, validation, blocker or failure reason, and next action.
- [x] Keep summaries derived from existing records.
- [x] Add tests for succeeded, failed, and blocked executions.
- [x] Document handoff summary semantics.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T7.1: `a385fd5`
- T7.2: `a385fd5`
- T7.3: `a385fd5`
- T7.4: `f0235b6`
- T7.5: `docs(slice-7): record review handoff summary`
- T7.6: `make verify`

## Slice 8: Reporting Documentation

Pre-flight:

- Plan: `.prj/plans/slicing-reporting-and-evidence-plan.md`
- Tracker: `.prj/trackers/slicing-reporting-and-evidence-tracker.md`
- Slicing: `Reporting and Evidence`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-reporting-docs`
- Expected PR title: `feat(slice-8): reporting docs`
- Expected report: `.prj/reports/slices/reporting-and-evidence/slice-8-reporting-docs.md`

Candidate tasks:

- [x] Add reference docs for report formats and commands.
- [x] Add how-to documentation for finding execution and loop evidence.
- [x] Update related ticket execution and loop docs.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T8.1: `335370e`
- T8.2: `335370e`
- T8.3: `335370e`
- T8.4: `docs(slice-8): record reporting docs`
- T8.5: `make verify`

## Validation Record

- Slice 1 focused tests: `go test ./internal/chc -run 'Test(NewExecutionReport|RenderExecutionReport)'` pass.
- Slice 1 full validation: `make verify` pass.
- Slice 2 focused tests: `go test ./internal/chc -run 'Test(ExecutionReportWriter|NewExecutionReport|RenderExecutionReport)'` pass.
- Slice 2 focused CLI tests: `go test ./internal/chcapp -run 'TestTicketApplyResult'` pass.
- Slice 2 full validation: `make verify` pass.
- Slice 3 focused tests: `go test ./internal/chc -run 'Test(NewExecutionReport|RenderExecutionReport|ExecutionReportWriter)'` pass.
- Slice 3 full validation: `make verify` pass.
- Slice 4 focused tests: `go test ./internal/chc -run 'Test(NewLoopReport|RenderLoopReport)'` pass.
- Slice 4 full validation: `make verify` pass.
- Slice 5 focused core tests: `go test ./internal/chc -run 'TestReportLister'` pass.
- Slice 5 focused CLI tests: `go test ./internal/chcapp -run 'TestReportList'` pass.
- Slice 5 full validation: `make verify` pass.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- Reports summarize existing evidence and must not become a second source of
  truth.
