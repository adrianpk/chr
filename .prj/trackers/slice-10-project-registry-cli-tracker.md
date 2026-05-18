# Slice 10 Project Registry CLI Tracker

Status: planned

Plan: `.prj/plans/slice-10-project-registry-cli-plan.md`
Spec: `.prj/specs/specs/user-daemon-and-global-index.md`
Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create slice branch from `dev`.
- [x] T10.1 CLI user environment wiring.
- [x] T10.2 Project add command.
- [x] T10.3 Project list command.
- [x] Run validation.
- [ ] Push slice branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T10.1: `feat(cli): wire project user environment`
- T10.2: `feat(cli): add project add command`
- T10.3: `feat(cli): add project list command`

## Validation

- `make verify`: pass

## Notes

- This slice adds CLI access to the project registry only.
- It must not add project reindex, ticket indexing, daemon behavior, watcher behavior, or Telegram behavior.
