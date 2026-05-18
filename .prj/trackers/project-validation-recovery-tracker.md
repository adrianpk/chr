# Project Validation and Recovery Tracker

Status: planned

Plan: `.prj/plans/project-validation-recovery-plan.md`
Slicing: `.prj/trackers/slicing-user-project-context-tracker.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create implementation branch from `dev`.
- [x] T1.1 Project workspace validation helpers.
- [x] T1.2 Apply validation to selection and active context.
- [x] T1.3 Tests and tracker.
- [x] Run validation.
- [x] Push implementation branch.
- [x] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T1.1: `feat(cli): validate registered project workspaces`
- T1.2: `feat(cli): validate selected project workspace`
- T1.3: `test(cli): cover project validation recovery`

## Validation

- `go test ./cmd/chc`: pass
- `make verify`: pass

## Notes

- This slice closes the planned `User Project Context` slicing.
- Project removal or repair commands are explicitly out of scope.
- PR #16 is open against `dev`.
