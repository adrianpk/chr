# Slice 5 Ticket New Tracker

Status: delivered

Plan: `.prj/plans/slice-5-ticket-new-plan.md`
Spec: `.prj/specs/specs/ticket-template.md`

## Checklist

- [x] Create local plan and tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Create slice branch from `dev`.
- [x] T5.1 Ticket model and renderer.
- [x] T5.2 Ticket repository and key generation.
- [x] T5.3 CLI command.
- [x] Run validation.
- [x] Push slice branch.
- [x] Open PR against `dev`.
- [x] Await user approval before merge.
- [x] Merge PR into `dev`.

## Task Commits

- T5.1: `feat(ticket): render markdown tickets`
- T5.2: `feat(ticket): create ticket files`
- T5.3: `feat(cli): add ticket new command`

## Validation

- `make verify`: pass
- Manual `chc init` and `chc ticket new`: pass

## Notes

- Status stays in ticket metadata only.
- Project-wise tickets use an empty or absent `use_case` and live directly under `.chc/tickets/`.
- Use-case tickets live under `.chc/tickets/<use_case>/`.
- This slice intentionally avoids SQLite and status transitions.
- PR #5 was merged into `dev`.
