# Automatic Reconciliation Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`
Priority draft: `.prj/drafts/chc-command-surfaces-and-priority.md`
Specs:

- `.prj/specs/specs/user-daemon-and-global-index.md`
- `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Confirm `Ticket Foundation` is completed.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Notes |
| --- | --- | --- | --- | --- |
| Slice 1: Ticket File Discovery | Delivered | `feat/slice-1-ticket-file-discovery` | `feat(slice-1): ticket file discovery` | Repository PR #22. |
| Slice 2: Ticket Index Rebuild | Delivered | `feat/slice-2-ticket-index-rebuild` | `feat(slice-2): ticket index rebuild` | Repository PR #23. |
| Slice 3: Index Status | Delivered | `feat/slice-3-index-status` | `feat(slice-3): index status` | Repository PR #24. |
| Slice 4: Index Rebuild CLI | Delivered | `feat/slice-4-index-rebuild-cli` | `feat(slice-4): index rebuild cli` | Repository PR #25. |
| Slice 5: Index Status CLI | Planned | `feat/slice-5-index-status-cli` | `feat(slice-5): index status cli` | Expose index health inspection. |
| Slice 6: Refresh Before Ticket Reads | Planned | `feat/slice-6-refresh-before-ticket-reads` | `feat(slice-6): refresh before ticket reads` | Make ticket read commands refresh automatically. |
| Slice 7: Reconciliation Boundary for Execution | Planned | `feat/slice-7-reconciliation-execution-boundary` | `feat(slice-7): reconciliation execution boundary` | Prepare future execution and loop use. |

## Slice 1: Ticket File Discovery

Pre-flight:

- Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
- Tracker: `.prj/trackers/slicing-automatic-reconciliation-tracker.md`
- Slicing: `Automatic Reconciliation`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-ticket-file-discovery`
- Expected PR title: `feat(slice-1): ticket file discovery`

Candidate tasks:

- [x] Add project ticket discovery core primitive.
- [x] Return stable relative and absolute ticket paths.
- [x] Add focused discovery tests.
- [x] Document ticket discovery behavior.
- [x] Run `make verify`.

Task commits:

- T1.1: `feat(ticket): add file discovery primitive`
- T1.2: `feat(ticket): return stable discovered paths`
- T1.3: `test(ticket): cover file discovery`
- T1.4: `docs(ticket): document file discovery`
- T1.5: `docs(planning): record ticket discovery validation`

## Slice 2: Ticket Index Rebuild

Pre-flight:

- Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
- Tracker: `.prj/trackers/slicing-automatic-reconciliation-tracker.md`
- Slicing: `Automatic Reconciliation`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-ticket-index-rebuild`
- Expected PR title: `feat(slice-2): ticket index rebuild`

Candidate tasks:

- [x] Add project ticket index rebuild service.
- [x] Replace one project's indexed ticket set from parsed Markdown.
- [x] Preserve parse errors in rebuild results.
- [x] Add focused rebuild tests.
- [x] Run `make verify`.

Task commits:

- T2.1: `feat(ticket): add index rebuild service`
- T2.2: `test(ticket): cover index rebuild`
- T2.3: `docs(ticket): document index rebuild`
- T2.4: `refactor(ticket): satisfy index rebuild lint`
- T2.5: `docs(planning): record ticket index rebuild validation`

## Slice 3: Index Status

Pre-flight:

- Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
- Tracker: `.prj/trackers/slicing-automatic-reconciliation-tracker.md`
- Slicing: `Automatic Reconciliation`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-index-status`
- Expected PR title: `feat(slice-3): index status`

Candidate tasks:

- [x] Add index status core model.
- [x] Detect missing index rows.
- [x] Detect stale index rows.
- [x] Detect orphaned index rows.
- [x] Report parse errors separately.
- [x] Add focused status tests.
- [x] Run `make verify`.

Task commits:

- T3.1: `feat(ticket): add index status checker`
- T3.2: `test(ticket): cover index status`
- T3.3: `docs(ticket): document index status`
- T3.4: `refactor(ticket): satisfy index status lint`
- T3.5: `docs(planning): record index status validation`

## Slice 4: Index Rebuild CLI

Pre-flight:

- Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
- Tracker: `.prj/trackers/slicing-automatic-reconciliation-tracker.md`
- Slicing: `Automatic Reconciliation`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-index-rebuild-cli`
- Expected PR title: `feat(slice-4): index rebuild cli`

Candidate tasks:

- [x] Add index command group.
- [x] Add `chc index rebuild`.
- [x] Resolve project context through the existing resolver.
- [x] Print rebuild results.
- [x] Add focused CLI tests.
- [x] Document command behavior.
- [x] Run `make verify`.

Task commits:

- T4.1: `feat(cli): add index rebuild command`
- T4.2: `test(cli): cover index rebuild command`
- T4.3: `docs(index): document rebuild command`
- T4.4: `docs(planning): record index rebuild cli validation`

## Slice 5: Index Status CLI

Pre-flight:

- Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
- Tracker: `.prj/trackers/slicing-automatic-reconciliation-tracker.md`
- Slicing: `Automatic Reconciliation`
- Slice: `Slice 5`
- Expected branch: `feat/slice-5-index-status-cli`
- Expected PR title: `feat(slice-5): index status cli`

Candidate tasks:

- [ ] Add `chc index status`.
- [ ] Resolve project context through the existing resolver.
- [ ] Print index health summary.
- [ ] Return non-zero for stale or invalid state.
- [ ] Add focused CLI tests.
- [ ] Document command behavior.
- [ ] Run `make verify`.

Task commits:

- T5.1: pending
- T5.2: pending
- T5.3: pending
- T5.4: pending
- T5.5: pending
- T5.6: pending
- T5.7: pending

## Slice 6: Refresh Before Ticket Reads

Pre-flight:

- Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
- Tracker: `.prj/trackers/slicing-automatic-reconciliation-tracker.md`
- Slicing: `Automatic Reconciliation`
- Slice: `Slice 6`
- Expected branch: `feat/slice-6-refresh-before-ticket-reads`
- Expected PR title: `feat(slice-6): refresh before ticket reads`

Candidate tasks:

- [ ] Refresh before `ticket list`.
- [ ] Refresh before `ticket show`.
- [ ] Refresh before `ticket comment`.
- [ ] Refresh before `ticket status`.
- [ ] Add focused CLI tests for direct Markdown edits.
- [ ] Document automatic refresh behavior.
- [ ] Run `make verify`.

Task commits:

- T6.1: pending
- T6.2: pending
- T6.3: pending
- T6.4: pending
- T6.5: pending
- T6.6: pending
- T6.7: pending

## Slice 7: Reconciliation Boundary for Execution

Pre-flight:

- Plan: `.prj/plans/slicing-automatic-reconciliation-plan.md`
- Tracker: `.prj/trackers/slicing-automatic-reconciliation-tracker.md`
- Slicing: `Automatic Reconciliation`
- Slice: `Slice 7`
- Expected branch: `feat/slice-7-reconciliation-execution-boundary`
- Expected PR title: `feat(slice-7): reconciliation execution boundary`

Candidate tasks:

- [ ] Add reusable reconciliation service boundary.
- [ ] Define result semantics for future execution blockers.
- [ ] Add tests for clean refresh and parse-error outcomes.
- [ ] Document execution precondition.
- [ ] Run `make verify`.

Task commits:

- T7.1: pending
- T7.2: pending
- T7.3: pending
- T7.4: pending
- T7.5: pending

## Validation Record

- Planning validation: documentation diff reviewed.
- Slice 1 validation: `go test ./internal/chc` pass.
- Slice 1 validation: `make verify` pass.
- Slice 1 merged into `dev` as PR #22.
- Slice 2 validation: initial `make verify` failed on lint whitespace.
- Slice 2 validation: `make verify` pass after lint-focused refactor.
- Slice 2 merged into `dev` as PR #23.
- Slice 3 validation: initial `make verify` failed on formatting and staticcheck.
- Slice 3 validation: `make verify` pass after lint-focused refactor.
- Slice 3 merged into `dev` as PR #24.
- Slice 4 validation: `go test ./cmd/chc` pass.
- Slice 4 validation: `make verify` pass.
- Slice 4 merged into `dev` as PR #25.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- Manual rebuild and status commands are recovery tools, not normal workflow requirements.
- Automatic refresh before ticket reads is required before ticket execution or loops rely on the index.
