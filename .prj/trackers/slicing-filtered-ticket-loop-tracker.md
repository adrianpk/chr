# Filtered Ticket Loop Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
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
| Slice 1: Eligible Ticket Query | Delivered | `feat/slice-1-eligible-ticket-query` | `feat(slice-1): eligible ticket query` | `.prj/reports/slices/filtered-ticket-loop/slice-1-eligible-ticket-query.md` | Repository PR #56. |
| Slice 2: Loop Filters | Delivered | `feat/slice-2-loop-filters` | `feat(slice-2): loop filters` | `.prj/reports/slices/filtered-ticket-loop/slice-2-loop-filters.md` | Repository PR #57. |
| Slice 3: Loop Ordering | In Review | `feat/slice-3-loop-ordering` | `feat(slice-3): loop ordering` | `.prj/reports/slices/filtered-ticket-loop/slice-3-loop-ordering.md` | Repository PR #58. |
| Slice 4: Loop Command Skeleton | Planned | `feat/slice-4-loop-command-skeleton` | `feat(slice-4): loop command skeleton` | `.prj/reports/slices/filtered-ticket-loop/slice-4-loop-command-skeleton.md` | Foreground command preview. |
| Slice 5: Loop Run Records | Planned | `feat/slice-5-loop-run-records` | `feat(slice-5): loop run records` | `.prj/reports/slices/filtered-ticket-loop/slice-5-loop-run-records.md` | Persist loop metadata. |
| Slice 6: Repeated Ticket Execution | Planned | `feat/slice-6-repeated-ticket-execution` | `feat(slice-6): repeated ticket execution` | `.prj/reports/slices/filtered-ticket-loop/slice-6-repeated-ticket-execution.md` | Reuse single-ticket execution. |
| Slice 7: Stop Conditions and Thresholds | Planned | `feat/slice-7-stop-conditions` | `feat(slice-7): stop conditions` | `.prj/reports/slices/filtered-ticket-loop/slice-7-stop-conditions.md` | Terminate safely. |
| Slice 8: Loop Status Reporting | Planned | `feat/slice-8-loop-status-reporting` | `feat(slice-8): loop status reporting` | `.prj/reports/slices/filtered-ticket-loop/slice-8-loop-status-reporting.md` | Summarize loop outcome. |

## Slice 1: Eligible Ticket Query

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-eligible-ticket-query`
- Expected PR title: `feat(slice-1): eligible ticket query`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-1-eligible-ticket-query.md`

Candidate tasks:

- [x] Add loop eligibility option types.
- [x] Select tickets by project and eligible statuses.
- [x] Keep selection read-only.
- [x] Return deterministic results.
- [x] Add focused query tests.
- [x] Document the eligibility contract.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `feat(slice-1): eligible ticket query`
- T1.2: `feat(slice-1): eligible ticket query`
- T1.3: `feat(slice-1): eligible ticket query`
- T1.4: `feat(slice-1): eligible ticket query`
- T1.5: `feat(slice-1): eligible ticket query`
- T1.6: `feat(slice-1): eligible ticket query`
- T1.7: `feat(slice-1): eligible ticket query`
- T1.8: `feat(slice-1): eligible ticket query`

## Slice 2: Loop Filters

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-loop-filters`
- Expected PR title: `feat(slice-2): loop filters`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-2-loop-filters.md`

Candidate tasks:

- [x] Add priority, severity, area, tag, type, use-case, and status filters.
- [x] Validate filter values where existing ticket validation rules apply.
- [x] Keep filtering independent from CLI parsing.
- [x] Add focused filter tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `feat(slice-2): loop filters`
- T2.2: `feat(slice-2): loop filters`
- T2.3: `feat(slice-2): loop filters`
- T2.4: `feat(slice-2): loop filters`
- T2.5: `feat(slice-2): loop filters`
- T2.6: `feat(slice-2): loop filters`

## Slice 3: Loop Ordering

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-loop-ordering`
- Expected PR title: `feat(slice-3): loop ordering`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-3-loop-ordering.md`

Candidate tasks:

- [x] Define default ordering.
- [x] Support priority and severity ordering.
- [x] Tie-break by ticket key or created time.
- [x] Add focused ordering tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T3.1: `feat(slice-3): loop ordering`
- T3.2: `feat(slice-3): loop ordering`
- T3.3: `feat(slice-3): loop ordering`
- T3.4: `feat(slice-3): loop ordering`
- T3.5: `feat(slice-3): loop ordering`
- T3.6: `feat(slice-3): loop ordering`

## Slice 4: Loop Command Skeleton

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-loop-command-skeleton`
- Expected PR title: `feat(slice-4): loop command skeleton`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-4-loop-command-skeleton.md`

Candidate tasks:

- [ ] Add CLI command and flags.
- [ ] Resolve project and refresh index like ticket reads.
- [ ] Call the eligible ticket query.
- [ ] Support dry-run output.
- [ ] Add focused CLI tests.
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

## Slice 5: Loop Run Records

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-loop-run-records`
- Expected PR title: `feat(slice-5): loop run records`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-5-loop-run-records.md`

Candidate tasks:

- [ ] Add loop run model.
- [ ] Add SQLite migration and repository.
- [ ] Track lifecycle status and timestamps.
- [ ] Link loop runs to project IDs.
- [ ] Add repository tests.
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

## Slice 6: Repeated Ticket Execution

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-repeated-ticket-execution`
- Expected PR title: `feat(slice-6): repeated ticket execution`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-6-repeated-ticket-execution.md`

Candidate tasks:

- [ ] Run one execution per selected ticket.
- [ ] Refresh ticket state between iterations.
- [ ] Apply successful or blocked results through existing result application.
- [ ] Record iteration summaries.
- [ ] Add focused runner tests.
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

## Slice 7: Stop Conditions and Thresholds

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-stop-conditions`
- Expected PR title: `feat(slice-7): stop conditions`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-7-stop-conditions.md`

Candidate tasks:

- [ ] Add max-iteration threshold.
- [ ] Add no-eligible-ticket completion.
- [ ] Stop on blocker or execution failure.
- [ ] Avoid repeated attempts on the same ticket in one run.
- [ ] Add focused stop-condition tests.
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

## Slice 8: Loop Status Reporting

Pre-flight:

- Plan: `.prj/plans/slicing-filtered-ticket-loop-plan.md`
- Tracker: `.prj/trackers/slicing-filtered-ticket-loop-tracker.md`
- Slicing: `Filtered Ticket Loop`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-loop-status-reporting`
- Expected PR title: `feat(slice-8): loop status reporting`
- Expected report: `.prj/reports/slices/filtered-ticket-loop/slice-8-loop-status-reporting.md`

Candidate tasks:

- [ ] Print loop run summary.
- [ ] Print processed, skipped, blocked, and failed counts.
- [ ] Expose execution IDs per iteration.
- [ ] Add loop report artifact or lookup boundary if needed.
- [ ] Document foreground loop behavior.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T8.1: pending
- T8.2: pending
- T8.3: pending
- T8.4: pending
- T8.5: pending
- T8.6: pending
- T8.7: pending

## Validation Record

- Planning validation: documentation diff reviewed.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- This slicing keeps the loop as a foreground runner. Daemon ownership remains
  post-cutoff.
