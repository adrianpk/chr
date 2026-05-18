# Slice 12 Ticket Index Repository Tracker

Status: planned

Plan: `.prj/plans/slice-12-ticket-index-repository-plan.md`
Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create slice branch from `dev`.
- [x] T12.1 Ticket index migration.
- [x] T12.2 Ticket index repository.
- [ ] T12.3 Validation and documentation.
- [ ] Run validation.
- [ ] Push slice branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T12.1: `feat(db): add ticket index migration`
- T12.2: `feat(ticket): add sqlite ticket index`
- T12.3: pending

## Validation

- `make verify`: pending

## Notes

- This slice persists parsed ticket index rows only.
- It must not add index rebuild commands, daemon watcher behavior, Telegram behavior, or ticket search CLI.
