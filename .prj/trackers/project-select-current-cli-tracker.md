# Project Select and Current CLI Tracker

Status: planned

Plan: `.prj/plans/project-select-current-cli-plan.md`
Slicing: `.prj/trackers/slicing-user-project-context-tracker.md`

## Checklist

- [x] Create local plan and tracker.
- [ ] Commit plan and tracker directly to `dev`.
- [ ] Create implementation branch from `dev`.
- [ ] T1.1 Project registry lookup support.
- [ ] T1.2 Project select command.
- [ ] T1.3 Project current command.
- [ ] T1.4 Tests and tracker.
- [ ] Run validation.
- [ ] Push implementation branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T1.1: pending
- T1.2: pending
- T1.3: pending
- T1.4: pending

## Validation

- `make verify`: pending

## Notes

- This slice depends on `ActiveProjectStore`.
- Project context resolution for other commands belongs to a later slice in the same slicing.
