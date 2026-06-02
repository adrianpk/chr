# Slice 1: Merger Handoff Contract

Status: reviewing
Slicing: Merger Agent Workflow
Plan: `.prj/plans/slicing-merger-agent-workflow-plan.md`
Tracker: `.prj/trackers/slicing-merger-agent-workflow-tracker.md`
Branch: `feat/slice-1-merger-handoff-contract`
PR: `#283`

## Purpose

Make the reviewer-to-merger handoff explicit in ticket metadata and local
workflow validation before adding provider merge preflight or merge execution.

## Delivered Behavior

- Tickets now carry merge handoff and merge outcome metadata:
  `merge_state`, `merge_result`, `merge_actor`, `merge_commit`, `merged_at`,
  and `merge_blocker`.
- Ticket rendering, parsing, validation, and indexing preserve the new merge
  metadata.
- SQLite migration `20260602000010-add_ticket_merge_metadata.sql` adds indexed
  merge metadata fields and a merge-state index.
- `ValidateTicketMergeHandoff` validates that a ticket is in
  `ready_for_merge`, has passed review metadata, has branch metadata, and has a
  pull request review reference.
- Moving a ticket to `ready_for_merge` through the status updater now requires
  the same passed-review handoff metadata.

## Implementation Notes

The handoff check is intentionally objective and narrow. It does not inspect
remote provider state or mergeability; those concerns belong to the later
merger preflight slice.

Merge metadata is represented with scalar frontmatter fields to match the
existing ticket metadata style and keep parser/index changes small.

## Contracts Added or Changed

- Added merge metadata fields to `Ticket` and `IndexedTicket`.
- Added ticket merge state/result constants and merger actor constant.
- Added `ValidateTicketMergeHandoff`.
- Added SQLite migration for indexed merge metadata.
- Extended `ready_for_merge` status transition validation.
- Updated ticket review metadata reference with merge handoff fields.

## Files of Interest

- `internal/chc/ticket.go`
- `internal/chc/indexed_ticket.go`
- `internal/chc/ticket_parser.go`
- `internal/chc/ticket_index.go`
- `internal/chc/ticket_status.go`
- `internal/chc/ticket_merge_handoff.go`
- `cmd/chc/assets/migration/sqlite/20260602000010-add_ticket_merge_metadata.sql`
- `docs/reference/ticket-review-metadata.md`

## Validation

- `go test ./internal/chc`
- `go test ./internal/chcapp`
- `git diff --check`
- `make verify`

## Risks and Follow-ups

- Provider mergeability and remote merge execution are intentionally deferred to
  later slices.
- Blocker classification is still simple; Slice 2 will define operational
  versus repair blockers.
