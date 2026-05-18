# Slice 9 Project Registry Repository Tracker

Status: planned

Plan: `.prj/plans/slice-9-project-registry-repository-plan.md`
Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [ ] Create slice branch from `dev`.
- [ ] T9.1 Project model.
- [ ] T9.2 Project registry repository.
- [ ] T9.3 Validation and documentation.
- [ ] Run validation.
- [ ] Push slice branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T9.1: pending
- T9.2: pending
- T9.3: pending

## Validation

- `make verify`: pending

## Notes

- This slice adds core registry persistence only.
- It must not add project CLI commands, daemon behavior, watcher behavior, or Telegram behavior.
