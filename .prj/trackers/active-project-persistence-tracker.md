# Active Project Persistence Tracker

Status: planned

Plan: `.prj/plans/active-project-persistence-plan.md`
Slicing: `.prj/trackers/slicing-user-project-context-tracker.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create implementation branch from `dev`.
- [x] T1 Active project model.
- [x] T2 Active project store.
- [x] T3 Tests.
- [x] Run validation.
- [x] Push implementation branch.
- [x] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T1: `feat(chc): add active project model`
- T2: `feat(chc): persist active project state`
- T3: `test(chc): cover active project store`

## Validation

- `go test ./internal/chc`: pass
- `make verify`: pass

## Notes

- This slice intentionally does not add CLI commands.
- CLI selection and inspection belong to a later slice in the same slicing.
