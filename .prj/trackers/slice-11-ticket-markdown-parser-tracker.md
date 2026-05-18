# Slice 11 Ticket Markdown Parser Tracker

Status: planned

Plan: `.prj/plans/slice-11-ticket-markdown-parser-plan.md`
Spec: `.prj/specs/specs/ticket-template.md`
Index Model: `.prj/specs/specs/user-daemon-and-global-index-model.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create slice branch from `dev`.
- [x] T11.1 Indexed ticket model.
- [ ] T11.2 Ticket Markdown parser.
- [ ] T11.3 Validation and documentation.
- [ ] Run validation.
- [ ] Push slice branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T11.1: `feat(ticket): add indexed ticket model`
- T11.2: pending
- T11.3: pending

## Validation

- `make verify`: pending

## Notes

- This slice parses Markdown into an indexable model only.
- It must not add SQLite ticket index persistence, index rebuild commands, daemon watcher behavior, or Telegram behavior.
