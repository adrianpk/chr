# Slice 6 Core Package Consolidation Tracker

Status: planned

Plan: `.prj/plans/slice-6-core-package-consolidation-plan.md`
Spec: `.prj/specs/specs/user-daemon-and-global-index.md`

## Checklist

- [x] Create local plan and tracker.
- [ ] Commit spec, plan, and tracker directly to `dev`.
- [ ] Create slice branch from `dev`.
- [ ] T6.1 Move workspace initialization into core package.
- [ ] T6.2 Introduce workspace-oriented API names.
- [ ] T6.3 Documentation and tracker alignment.
- [ ] Run validation.
- [ ] Push slice branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T6.1: pending
- T6.2: pending
- T6.3: pending

## Validation

- `make verify`: pending
- Manual `chc init`: pending

## Notes

- This slice is a naming and package-boundary consolidation before daemon/index work.
- It should not add user environment, registry, SQLite, daemon, watcher, or Telegram behavior.
