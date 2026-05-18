# Slice 8 SQLite Migration Foundation Tracker

Status: planned

Plan: `.prj/plans/slice-8-sqlite-migration-foundation-plan.md`
Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create slice branch from `dev`.
- [x] T8.1 SQLite driver and migration asset.
- [x] T8.2 Migrator core primitive.
- [x] T8.3 Validation and documentation.
- [x] Run validation.
- [ ] Push slice branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T8.1: `feat(db): add sqlite migration asset`
- T8.2: `feat(db): add sqlite migrator`
- T8.3: `docs(db): document sqlite migrations`

## Validation

- `make verify`: pass

## Notes

- This slice creates database schema infrastructure only.
- It must not add project registration commands, daemon behavior, watcher behavior, or Telegram behavior.
