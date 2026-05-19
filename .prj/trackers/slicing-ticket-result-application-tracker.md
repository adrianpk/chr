# Ticket Result Application Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`
Priority draft: `.prj/drafts/chc-command-surfaces-and-priority.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm `Single Ticket Execution` is completed.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Result Application Model | Delivered | `feat/slice-1-result-application-model` | `feat(slice-1): result application model` | `.prj/reports/slices/ticket-result-application/slice-1-result-application-model.md` | Repository PR #41. |
| Slice 2: Ticket State Writer | Reviewing | `feat/slice-2-ticket-state-writer` | `feat(slice-2): ticket state writer` | `.prj/reports/slices/ticket-result-application/slice-2-ticket-state-writer.md` | Update ticket status from application outcome. |
| Slice 3: Execution Comment Writer | Planned | `feat/slice-3-execution-comment-writer` | `feat(slice-3): execution comment writer` | `.prj/reports/slices/ticket-result-application/slice-3-execution-comment-writer.md` | Append execution summary comments. |
| Slice 4: Validation Evidence Writer | Planned | `feat/slice-4-validation-evidence-writer` | `feat(slice-4): validation evidence writer` | `.prj/reports/slices/ticket-result-application/slice-4-validation-evidence-writer.md` | Record validation evidence. |
| Slice 5: Changed Files Summary | Planned | `feat/slice-5-changed-files-summary` | `feat(slice-5): changed files summary` | `.prj/reports/slices/ticket-result-application/slice-5-changed-files-summary.md` | Record changed-file summaries. |
| Slice 6: Blocked Reason Writer | Planned | `feat/slice-6-blocked-reason-writer` | `feat(slice-6): blocked reason writer` | `.prj/reports/slices/ticket-result-application/slice-6-blocked-reason-writer.md` | Record blocked/failure reasons. |
| Slice 7: Markdown Index Consistency | Planned | `feat/slice-7-markdown-index-consistency` | `feat(slice-7): markdown index consistency` | `.prj/reports/slices/ticket-result-application/slice-7-markdown-index-consistency.md` | Reconcile index after ticket writes. |
| Slice 8: Result Apply CLI | Planned | `feat/slice-8-result-apply-cli` | `feat(slice-8): result apply cli` | `.prj/reports/slices/ticket-result-application/slice-8-result-apply-cli.md` | Expose result application command. |

## Slice 1: Result Application Model

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-result-application-model`
- Expected PR title: `feat(slice-1): result application model`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-1-result-application-model.md`

Candidate tasks:

- [x] Add result application domain types.
- [x] Map execution statuses and result outcomes to ticket-facing outcomes.
- [x] Define comment, evidence, changed-file, and blocked-reason fields.
- [x] Validate required references.
- [x] Add focused model tests.
- [x] Document result application contract.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T1.1: `b5d1017` `feat(result): add result application model`
- T1.2: `b5d1017` `feat(result): add result application model`
- T1.3: `b5d1017` `feat(result): add result application model`
- T1.4: `b5d1017` `feat(result): add result application model`
- T1.5: `689d99e` `test(result): cover result application model`
- T1.6: `b802623` `docs(result): document result application model`
- T1.7: `b802623` `docs(result): document result application model`
- T1.8: `docs(result): record result application validation`

## Slice 2: Ticket State Writer

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-ticket-state-writer`
- Expected PR title: `feat(slice-2): ticket state writer`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-2-ticket-state-writer.md`

Candidate tasks:

- [x] Add a ticket result state writer.
- [x] Map success to review-ready ticket state.
- [x] Map execution failure to blocked or failed ticket state.
- [x] Preserve existing ticket metadata and comments.
- [x] Add focused writer tests.
- [x] Add slice implementation report.
- [x] Run `make verify`.

Task commits:

- T2.1: `507c562` `feat(result): add ticket state writer`
- T2.2: `507c562` `feat(result): add ticket state writer`
- T2.3: `507c562` `feat(result): add ticket state writer`
- T2.4: `507c562` `feat(result): add ticket state writer`
- T2.5: `a349b26` `test(result): cover ticket state writer`
- T2.6: `3222e76` `docs(result): document ticket state writer`
- T2.7: `docs(result): record ticket state writer validation`

## Slice 3: Execution Comment Writer

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-execution-comment-writer`
- Expected PR title: `feat(slice-3): execution comment writer`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-3-execution-comment-writer.md`

Candidate tasks:

- [ ] Write execution ID, status, summary, and failure reason comments.
- [ ] Include artifact paths in the comment.
- [ ] Preserve existing comments.
- [ ] Add focused comment writer tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T3.1: pending
- T3.2: pending
- T3.3: pending
- T3.4: pending
- T3.5: pending
- T3.6: pending

## Slice 4: Validation Evidence Writer

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-validation-evidence-writer`
- Expected PR title: `feat(slice-4): validation evidence writer`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-4-validation-evidence-writer.md`

Candidate tasks:

- [ ] Define validation evidence fields.
- [ ] Write validation command and artifact path.
- [ ] Preserve validation history.
- [ ] Add focused evidence writer tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T4.1: pending
- T4.2: pending
- T4.3: pending
- T4.4: pending
- T4.5: pending
- T4.6: pending

## Slice 5: Changed Files Summary

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-changed-files-summary`
- Expected PR title: `feat(slice-5): changed files summary`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-5-changed-files-summary.md`

Candidate tasks:

- [ ] Extend result application data with changed files.
- [ ] Write changed-file summary comments or section content.
- [ ] Handle empty changed-file lists.
- [ ] Add focused summary tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending
- T5.2: pending
- T5.3: pending
- T5.4: pending
- T5.5: pending
- T5.6: pending

## Slice 6: Blocked Reason Writer

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-blocked-reason-writer`
- Expected PR title: `feat(slice-6): blocked reason writer`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-6-blocked-reason-writer.md`

Candidate tasks:

- [ ] Write blocked reason from execution failure classification.
- [ ] Include diagnostic artifact paths.
- [ ] Keep failure details concise.
- [ ] Add focused blocked reason tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending
- T6.2: pending
- T6.3: pending
- T6.4: pending
- T6.5: pending
- T6.6: pending

## Slice 7: Markdown Index Consistency

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-markdown-index-consistency`
- Expected PR title: `feat(slice-7): markdown index consistency`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-7-markdown-index-consistency.md`

Candidate tasks:

- [ ] Refresh the affected ticket after writes.
- [ ] Ensure SQLite index reflects updated ticket status and metadata.
- [ ] Surface reconciliation blockers.
- [ ] Add focused consistency tests.
- [ ] Add slice implementation report.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending
- T7.6: pending

## Slice 8: Result Apply CLI

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-result-application-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-result-application-tracker.md`
- Slicing: `Ticket Result Application`
- Slice: `Slice 8`
- Expected branch: `feat/slice-8-result-apply-cli`
- Expected PR title: `feat(slice-8): result apply cli`
- Expected report: `.prj/reports/slices/ticket-result-application/slice-8-result-apply-cli.md`

Candidate tasks:

- [ ] Add `chc ticket apply-result <execution-id>`.
- [ ] Resolve project context.
- [ ] Load one execution record.
- [ ] Apply ticket status, comments, evidence, and diagnostics.
- [ ] Write an application report artifact.
- [ ] Support dry-run output.
- [ ] Add focused CLI tests.
- [ ] Document command behavior.
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
- T8.8: pending
- T8.9: pending
- T8.10: pending

## Validation Record

- Planning validation: documentation diff reviewed.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- This slicing applies completed execution attempts to durable ticket state.
- Each implementation slice must include a Markdown implementation report
  before PR creation.
