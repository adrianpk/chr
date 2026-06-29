# Slice 6 Core Package Consolidation Tracker

Status: delivered

Plan: `.prj/plans/slice-6-core-package-consolidation-plan.md`
Spec: `.prj/specs/specs/user-daemon-and-global-index.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit spec, plan, and tracker directly to `dev`.
- [x] Create slice branch from `dev`.
- [x] T6.1 Move workspace initialization into core package.
- [x] T6.2 Introduce workspace-oriented API names.
- [x] T6.3 Documentation and tracker alignment.
- [x] Run validation.
- [x] Push slice branch.
- [x] Open PR against `dev`.
- [x] Await user approval before merge.

## Task Commits

- T6.1: `refactor(chr): move workspace init into core package`
- T6.2: `refactor(chr): expose workspace API`
- T6.3: `docs(workspace): align core package boundary`

## Validation

- `make verify`: pass
- Manual `chr init`: pass

## Notes

- This slice is a naming and package-boundary consolidation before daemon/index work.
- It should not add user environment, registry, SQLite, daemon, watcher, or Telegram behavior.

## Reconciliation

Delivered by earlier implementation work and superseded by the later consolidated slicing trackers. This file is kept as historical planning context, not active backlog.
