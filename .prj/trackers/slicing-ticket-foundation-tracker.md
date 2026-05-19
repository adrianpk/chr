# Ticket Foundation Slicing Tracker

Status: active

Plan: `.prj/plans/slicing-ticket-foundation-plan.md`
Feature map: `.prj/drafts/chc-feature-slicing-set-map.md`
Specs:

- `.prj/specs/specs/ticket-template.md`
- `.prj/specs/specs/ticket-template-model.md`
- `.prj/specs/specs/user-daemon-and-global-index.md`
- `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Slicing Checklist

- [x] Identify slicing from `.prj/drafts/chc-feature-slicing-set-map.md`.
- [x] Review existing ticket creation, parsing, index, and project context implementation.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Notes |
| --- | --- | --- | --- | --- |
| Slice 1: Ticket List | Delivered | `feat/slice-1-ticket-list` | `feat(slice-1): ticket list` | Repository PR #17. |
| Slice 2: Ticket Show | Delivered | `feat/slice-2-ticket-show` | `feat(slice-2): ticket show` | Repository PR #18. |
| Slice 3: Ticket Comments | Delivered | `feat/slice-3-ticket-comments` | `feat(slice-3): ticket comments` | Repository PR #19. |
| Slice 4: Ticket Status Transitions | Reviewing | `feat/slice-4-ticket-status-transitions` | `feat(slice-4): ticket status transitions` | Update ticket status through controlled commands. |
| Slice 5: Ticket Lifecycle Validation | Planned | `feat/slice-5-ticket-lifecycle-validation` | `feat(slice-5): ticket lifecycle validation` | Validate metadata and lifecycle consistency. |

## Slice 1: Ticket List

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-foundation-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-foundation-tracker.md`
- Slicing: `Ticket Foundation`
- Slice: `Slice 1`
- Expected branch: `feat/slice-1-ticket-list`
- Expected PR title: `feat(slice-1): ticket list`

Candidate tasks:

- [x] Resolve project context for `ticket list`.
- [x] Read indexed tickets for the resolved project.
- [x] Print stable ticket rows.
- [x] Add focused CLI tests.
- [x] Run `make verify`.

Task commits:

- T1.1: `feat(cli): resolve ticket list project context`
- T1.2: `feat(ticket): open user ticket index`
- T1.3: `feat(cli): add ticket list command`
- T1.4: `test(cli): cover ticket list output`
- T1.5: `docs(planning): record ticket list validation`

## Slice 2: Ticket Show

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-foundation-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-foundation-tracker.md`
- Slicing: `Ticket Foundation`
- Slice: `Slice 2`
- Expected branch: `feat/slice-2-ticket-show`
- Expected PR title: `feat(slice-2): ticket show`

Candidate tasks:

- [x] Add indexed ticket lookup by key.
- [x] Add `chc ticket show`.
- [x] Add focused CLI tests.
- [x] Run `make verify`.

Task commits:

- T2.1: `feat(ticket): find indexed ticket by key`
- T2.2: `feat(cli): add ticket show command`
- T2.3: `test(cli): cover ticket show command`
- T2.4: `docs(planning): record ticket show validation`

## Slice 3: Ticket Comments

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-foundation-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-foundation-tracker.md`
- Slicing: `Ticket Foundation`
- Slice: `Slice 3`
- Expected branch: `feat/slice-3-ticket-comments`
- Expected PR title: `feat(slice-3): ticket comments`

Candidate tasks:

- [x] Add ticket Markdown comment writer.
- [x] Add `chc ticket comment`.
- [x] Add focused CLI tests.
- [x] Document ticket comment behavior.
- [x] Run `make verify`.

Task commits:

- T3.1: `feat(ticket): append ticket comments`
- T3.2: `feat(cli): add ticket comment command`
- T3.3: `test(cli): cover ticket comments`
- T3.4: `test(cli): satisfy ticket comment lint`
- T3.5: `docs(ticket): document ticket comments`
- T3.6: `docs(planning): record ticket comments validation`

## Validation Record

- Planning validation: documentation diff reviewed.
- Planning validation: `make verify` pass.
- Slice 1 validation: `go test ./cmd/chc` pass.
- Slice 1 validation: `go test ./internal/chc` pass.
- Slice 1 validation: `make verify` pass.
- Slice 1 merged into `dev` as PR #17.
- Slice 2 validation: `go test ./cmd/chc` pass.
- Slice 2 validation: `go test ./internal/chc` pass.
- Slice 2 validation: `make verify` pass.
- Slice 2 merged into `dev` as PR #18.
- Slice 3 validation: `go test ./cmd/chc` pass.
- Slice 3 validation: `go test ./internal/chc` pass.
- Slice 3 validation: `make verify` pass.
- Slice 3 merged into `dev` as PR #19.
- Slice 4 validation: `go test ./cmd/chc` pass.
- Slice 4 validation: `go test ./internal/chc` pass.
- Slice 4 validation: `make verify` pass.

## Slice 4: Ticket Status Transitions

Pre-flight:

- Plan: `.prj/plans/slicing-ticket-foundation-plan.md`
- Tracker: `.prj/trackers/slicing-ticket-foundation-tracker.md`
- Slicing: `Ticket Foundation`
- Slice: `Slice 4`
- Expected branch: `feat/slice-4-ticket-status-transitions`
- Expected PR title: `feat(slice-4): ticket status transitions`

Candidate tasks:

- [x] Share ticket metadata update helpers.
- [x] Add ticket status writer with allowed values.
- [x] Add `chc ticket status`.
- [x] Add focused CLI tests.
- [x] Document ticket status behavior.
- [x] Run `make verify`.

Task commits:

- T4.1: `refactor(ticket): share metadata updates`
- T4.2: `feat(ticket): update ticket status`
- T4.3: `feat(cli): add ticket status command`
- T4.4: `test(cli): cover ticket status command`
- T4.5: `docs(ticket): document ticket status`
- T4.6: `docs(planning): record ticket status validation`

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- `ticket list` is the first implementation slice for this slicing.
- Automatic reconciliation starts after this slicing unless ticket visibility exposes a blocking index issue.
