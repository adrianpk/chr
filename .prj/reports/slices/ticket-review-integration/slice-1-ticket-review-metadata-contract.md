# Slice 1: Ticket Review Metadata Contract

Status: reviewing
Slicing: Ticket Review Integration
Plan: `.prj/plans/slicing-ticket-review-integration-plan.md`
Tracker: `.prj/trackers/slicing-ticket-review-integration-tracker.md`
Branch: `feat/slice-1-ticket-review-metadata-contract`
PR: #273

## Purpose

Make ticket review workflow facts durable in ticket metadata before adding PR
attachment, provider diagnostics, or review execution behavior.

## Delivered Behavior

- Tickets can carry branch, review references, review state, review round,
  review result, review comments URL, and blocking finding references.
- Ticket parsing, rendering, validation, indexing, CLI output, and TUI detail
  rendering preserve the new metadata.
- Board columns remain unchanged. Review states that require attention are
  surfaced through the existing attention path.
- Legacy tickets without review metadata continue to parse and index.

## Implementation Notes

- `review_refs` is parsed as the supported structured frontmatter block for
  pull request references.
- Indexed `review_refs` and `blocking_findings` are stored as JSON text in
  SQLite so the filesystem ticket remains the source of truth while the read
  model can be rebuilt deterministically.
- Review attention is metadata-driven and does not introduce a new board column.

## Contracts Added or Changed

- Ticket frontmatter now includes:
  `branch`, `review_refs`, `review_state`, `review_round`, `review_result`,
  `review_comments_url`, and `blocking_findings`.
- SQLite migration `20260601000009-add_ticket_review_metadata` adds indexed
  review metadata columns and supporting indexes.
- Valid review states are `pass`, `changes_requested`,
  `insufficient_evidence`, and `needs_human_decision`.
- Valid review results are `pass`, `needs_changes`, `insufficient_evidence`,
  and `needs_human_decision`.

## Files of Interest

- `internal/chc/ticket.go`
- `internal/chc/ticket_parser.go`
- `internal/chc/ticket_validation.go`
- `internal/chc/ticket_index.go`
- `internal/chc/board_query.go`
- `internal/chcapp/ticket.go`
- `internal/chctui/model.go`
- `cmd/chc/assets/migration/sqlite/20260601000009-add_ticket_review_metadata.sql`
- `docs/reference/ticket-review-metadata.md`

## Validation

- `go test ./internal/chc -run 'Ticket|Board|Migrat'` - pass
- `go test ./internal/chcapp -run 'Ticket(List|Show|Board)|Migration'` - pass
- `go test ./internal/chctui -run 'Ticket|Board'` - pass
- `go test ./internal/chc ./internal/chcapp ./internal/chctui` - pass
- `make verify` - pass

## Risks and Follow-ups

- Slice 2 must make branch and pull request metadata attachment workflow-owned
  instead of relying on direct frontmatter edits.
- Later provider slices must define how remote review comments map back to
  local review metadata without duplicating full remote discussion content.
