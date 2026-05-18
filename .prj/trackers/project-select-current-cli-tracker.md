# Project Select and Current CLI Tracker

Status: planned

Plan: `.prj/plans/project-select-current-cli-plan.md`
Slicing: `.prj/trackers/slicing-user-project-context-tracker.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create implementation branch from `dev`.
- [x] T1.1 Project registry lookup support.
- [x] T1.2 Project select command.
- [x] T1.3 Project current command.
- [x] T1.4 Tests and tracker.
- [x] Run validation.
- [x] Push implementation branch.
- [x] Open PR against `dev`.
- [x] Await user approval before merge.

## Task Commits

- T1.1: `feat(project): add registry lookups`
- T1.2: `feat(cli): add project select command`
- T1.3: `feat(cli): add project current command`
- T1.4: `test(cli): cover project selection commands`

## Validation

- `go test ./cmd/chc`: pass
- `make verify`: pass

## Notes

- This slice depends on `ActiveProjectStore`.
- Project context resolution for other commands belongs to a later slice in the same slicing.
- PR #14 was merged into `dev`.
