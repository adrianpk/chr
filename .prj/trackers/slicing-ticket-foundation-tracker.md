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
| Slice 1: Ticket List | Planned | `feat/slice-1-ticket-list` | `feat(slice-1): ticket list` | List tickets for the resolved project. |
| Slice 2: Ticket Show | Planned | `feat/slice-2-ticket-show` | `feat(slice-2): ticket show` | Show one ticket by key. |
| Slice 3: Ticket Comments | Planned | `feat/slice-3-ticket-comments` | `feat(slice-3): ticket comments` | Append durable comments to ticket Markdown. |
| Slice 4: Ticket Status Transitions | Planned | `feat/slice-4-ticket-status-transitions` | `feat(slice-4): ticket status transitions` | Update ticket status through controlled commands. |
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

- [ ] Resolve project context for `ticket list`.
- [ ] Read indexed tickets for the resolved project.
- [ ] Print stable ticket rows.
- [ ] Add focused CLI tests.
- [ ] Run `make verify`.

## Validation Record

- Planning validation: documentation diff reviewed.
- Planning validation: `make verify` pass.

## Notes

- Slice numbers are local to this slicing and start at `Slice 1`.
- Repository PR numbers are review records only and do not define slice numbers.
- `ticket list` is the first implementation slice for this slicing.
- Automatic reconciliation starts after this slicing unless ticket visibility exposes a blocking index issue.
