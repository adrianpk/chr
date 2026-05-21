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
| Slice 1: Execution Report Format | Reviewing | `feat/slice-1-execution-report-format` | `feat(slice-1): execution report format` | `.prj/reports/slices/reporting-and-evidence/slice-1-execution-report-format.md` | Repository PR #118 open. |
| Slice 2: Execution Report Artifact Writer | Planned | `feat/slice-2-execution-report-writer` | `feat(slice-2): execution report writer` | `.prj/reports/slices/reporting-and-evidence/slice-2-execution-report-writer.md` | Pending. |
| Slice 3: Validation and Changed-File Summaries | Planned | `feat/slice-3-evidence-summaries` | `feat(slice-3): evidence summaries` | `.prj/reports/slices/reporting-and-evidence/slice-3-evidence-summaries.md` | Pending. |
| Slice 4: Loop Report Format | Planned | `feat/slice-4-loop-report-format` | `feat(slice-4): loop report format` | `.prj/reports/slices/reporting-and-evidence/slice-4-loop-report-format.md` | Pending. |
| Slice 5: Report Listing | Planned | `feat/slice-5-report-listing` | `feat(slice-5): report listing` | `.prj/reports/slices/reporting-and-evidence/slice-5-report-listing.md` | Pending. |
| Slice 6: Report Lookup | Planned | `feat/slice-6-report-lookup` | `feat(slice-6): report lookup` | `.prj/reports/slices/reporting-and-evidence/slice-6-report-lookup.md` | Pending. |
| Slice 7: Review Handoff Summary | Planned | `feat/slice-7-review-handoff-summary` | `feat(slice-7): review handoff summary` | `.prj/reports/slices/reporting-and-evidence/slice-7-review-handoff-summary.md` | Pending. |
| Slice 8: Reporting Documentation | Planned | `feat/slice-8-reporting-docs` | `feat(slice-8): reporting docs` | `.prj/reports/slices/reporting-and-evidence/slice-8-reporting-docs.md` | Pending. |

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

- [ ] Define report artifact paths.
- [ ] Write Markdown reports from existing execution evidence.
- [ ] Avoid overwriting unrelated files.
- [ ] Add writer tests.
- [ ] Document artifact placement.
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

- [ ] Summarize validation output paths and status.
- [ ] Summarize changed files from existing evidence.
- [ ] Keep summaries bounded and deterministic.
- [ ] Add focused tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending
- T3.2: pending
- T3.3: pending
- T3.4: pending
- T3.5: pending
- T3.6: pending

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

- [ ] Add loop report summary model.
- [ ] Render deterministic Markdown.
- [ ] Include loop status, counts, failure reason, timings, and related executions when available.
- [ ] Add focused tests.
- [ ] Document the loop report format.
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

- [ ] Add `chc report` command group.
- [ ] Add `chc report list`.
- [ ] Support report filters.
- [ ] Print stable output.
- [ ] Add CLI tests.
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

- [ ] Add `chc report show`.
- [ ] Resolve reports by report id, ticket key, execution id, or loop id.
- [ ] Print report body or metadata predictably.
- [ ] Add clear errors for ambiguous or missing lookups.
- [ ] Add CLI tests.
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

- [ ] Summarize status, changed files, validation, blocker or failure reason, and next action.
- [ ] Keep summaries derived from existing records.
- [ ] Add tests for succeeded, failed, and blocked executions.
- [ ] Document handoff summary semantics.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending
- T7.6: pending

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

- [ ] Add reference docs for report formats and commands.
- [ ] Add how-to documentation for finding execution and loop evidence.
- [ ] Update related ticket execution and loop docs.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1: pending
- T8.2: pending
- T8.3: pending
- T8.4: pending
- T8.5: pending

## Validation Record

- Slice 1 focused tests: `go test ./internal/chc -run 'Test(NewExecutionReport|RenderExecutionReport)'` pass.
- Slice 1 full validation: `make verify` pass.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- Reports summarize existing evidence and must not become a second source of
  truth.
