# Slice 5 Ticket New Tracker

Status: planned

Plan: `.prj/plans/slice-5-ticket-new-plan.md`
Spec: `.prj/specs/specs/ticket-template.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create slice branch from `dev`.
- [x] T5.1 Ticket model and renderer.
- [x] T5.2 Ticket repository and key generation.
- [ ] T5.3 CLI command.
- [ ] Run validation.
- [ ] Push slice branch.
- [ ] Open PR against `dev`.
- [ ] Await user approval before merge.

## Task Commits

- T5.1: `feat(ticket): render markdown tickets`
- T5.2: `feat(ticket): create ticket files`
- T5.3: pending

## Validation

- `make verify`: pending
- Manual `chc init` and `chc ticket new`: pending

## Notes

- Status stays in ticket metadata only.
- Project-wise tickets use an empty or absent `use_case` and live directly under `.chc/tickets/`.
- Use-case tickets live under `.chc/tickets/<use_case>/`.
- This slice intentionally avoids SQLite and status transitions.
