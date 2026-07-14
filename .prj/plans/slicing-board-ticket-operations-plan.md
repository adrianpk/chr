# Board Ticket Operations Slicing Plan

Status: in progress

## Purpose

Implement the Board MVP as an operational ticket-flow surface backed by
Markdown tickets and SQLite metadata.

This slicing covers the Board view, manual ticket creation, generated ticket
synopsis, Board status flow, conservative attention signals, shared CLI/TUI
service boundaries, and the first active userspace shape under `.chr/default/`.

## Source

Primary sources:

- `.agents/SLICES.md`
- `.agents/playbooks/slice-delivery.md`
- `.agents/playbooks/slice-reporting.md`
- `.prj/quizzes/curated/dashboard-board-ticket-operations-quiz.md`
- `.prj/specs/specs/board-ticket-operations.md`
- `.prj/specs/specs/board-ticket-operations-model.md`
- `.prj/specs/specs/ticket-template.md`
- `.prj/specs/specs/ticket-template-model.md`
- `docs/reference/cli-tui-action-matrix.md`
- `internal/chr/ticket.go`
- `internal/chr/ticket_repository.go`
- `internal/chr/ticket_parser.go`
- `internal/chr/ticket_index.go`
- `internal/chr/indexed_ticket.go`
- `internal/chr/ticket_reconciliation.go`
- `cmd/chr/assets/migration/sqlite/20260518000002-create_indexed_tickets.sql`
- `internal/chrtui/model.go`
- `internal/chrapp/service_artifact_ticket.go`
- `internal/chrapp/tui_adapters_spec_plan_ticket.go`

Repository history reviewed:

- `f09828a feat(db): add ticket index migration`
- `479c1d7 feat(ticket): add sqlite ticket index`
- `fec9414 feat(ticket): open user ticket index`
- `7edc2d8 feat(tui): add ticket board detail slice`
- `442cbb3 feat(plan): materialize tracking artifacts`
- `5dd9c40 refactor(metadata): remove project case sqlite indexes`

## Current State

The repository already has a useful ticket foundation:

- Markdown ticket rendering and creation;
- `CHR-<number>` key generation by scanning existing ticket files;
- YAML-like frontmatter parsing;
- user-level SQLite `chr.db`;
- `indexed_tickets` and `indexed_ticket_tags`;
- index rebuild and status commands;
- TUI Board and ticket detail views grouped by status;
- ticket comment and status actions through shared services;
- plan tracking materialization that creates tracking plus tickets.

The current foundation still uses older assumptions:

- default ticket status is `open`;
- generated/manual tickets do not require `synopsis`;
- Board grouping uses sorted raw statuses instead of intentional columns;
- attention does not exist as a virtual view;
- key allocation is filesystem-scan based;
- SQLite index lacks Board-specific fields and source metadata;
- `.chr/default/` active userspace is not modeled;
- compact vs maximized Board behavior is not explicit.

## Target Model

The target model keeps Markdown as the durable ticket body and uses SQLite as
the operational metadata index for Board selection, filtering, source
traceability, synopsis display, and attention.

New Chryso-created tickets use Board statuses:

- `backlog`
- `ready_for_dev`
- `in_progress`
- `ready_for_review`
- `ready_for_merge`
- `done`

Legacy statuses are mapped or surfaced explicitly. Unknown statuses are visible
through attention.

The first userspace implementation prepares `.chr/default/` for the active
operator. Full `.chr/operators/<operator-id>/...` remains outside this slicing
but is documented in the spec/model so the insight is not lost.

## Slice Count Decision

This work is larger than a two-slice change because it crosses durable model,
SQLite migrations, generation contracts, CLI services, TUI layout, and
documentation. It does not require sixteen slices because several changes are
coherent when delivered together behind the same service boundary.

The planned count is twelve slices.

## Slice Map

Slicing name: Board Ticket Operations

Active slice number: Slice 12

Active slice short name: board docs validation

Expected branch name: `feat/slice-12-board-docs-validation`

Expected PR title: `feat(slice-12): board docs validation`

Expected report path:
`.prj/reports/slices/board-ticket-operations/slice-12-board-docs-validation.md`

Prerequisite PRs or delivered work:

- CLI/TUI Surface Parity slicing completed.
- Architecture Boundary Cleanup slicing completed.
- Plan generation safety and plan tracking materialization completed.
- Current ticket SQLite index foundation exists.

| Slice | Short Name | Expected Branch | Expected PR Title | Expected Report |
| --- | --- | --- | --- | --- |
| Slice 1 | ticket metadata contract | `feat/slice-1-ticket-metadata-contract` | `feat(slice-1): ticket metadata contract` | `.prj/reports/slices/board-ticket-operations/slice-1-ticket-metadata-contract.md` |
| Slice 2 | userspace resolver | `feat/slice-2-userspace-resolver` | `feat(slice-2): userspace resolver` | `.prj/reports/slices/board-ticket-operations/slice-2-userspace-resolver.md` |
| Slice 3 | board sqlite schema | `feat/slice-3-board-sqlite-schema` | `feat(slice-3): board sqlite schema` | `.prj/reports/slices/board-ticket-operations/slice-3-board-sqlite-schema.md` |
| Slice 4 | indexed ticket enrichment | `feat/slice-4-indexed-ticket-enrichment` | `feat(slice-4): indexed ticket enrichment` | `.prj/reports/slices/board-ticket-operations/slice-4-indexed-ticket-enrichment.md` |
| Slice 5 | ticket creation service | `feat/slice-5-ticket-creation-service` | `feat(slice-5): ticket creation service` | `.prj/reports/slices/board-ticket-operations/slice-5-ticket-creation-service.md` |
| Slice 6 | generated ticket synopsis | `feat/slice-6-generated-ticket-synopsis` | `feat(slice-6): generated ticket synopsis` | `.prj/reports/slices/board-ticket-operations/slice-6-generated-ticket-synopsis.md` |
| Slice 7 | board query service | `feat/slice-7-board-query-service` | `feat(slice-7): board query service` | `.prj/reports/slices/board-ticket-operations/slice-7-board-query-service.md` |
| Slice 8 | compact board columns | `feat/slice-8-compact-board-columns` | `feat(slice-8): compact board columns` | `.prj/reports/slices/board-ticket-operations/slice-8-compact-board-columns.md` |
| Slice 9 | attention view | `feat/slice-9-attention-view` | `feat(slice-9): attention view` | `.prj/reports/slices/board-ticket-operations/slice-9-attention-view.md` |
| Slice 10 | ticket detail surface | `feat/slice-10-ticket-detail-surface` | `feat(slice-10): ticket detail surface` | `.prj/reports/slices/board-ticket-operations/slice-10-ticket-detail-surface.md` |
| Slice 11 | maximized board | `feat/slice-11-maximized-board` | `feat(slice-11): maximized board` | `.prj/reports/slices/board-ticket-operations/slice-11-maximized-board.md` |
| Slice 12 | board docs validation | `feat/slice-12-board-docs-validation` | `feat(slice-12): board docs validation` | `.prj/reports/slices/board-ticket-operations/slice-12-board-docs-validation.md` |

## Slice 1: Ticket Metadata Contract

Define and test the durable ticket metadata contract before changing Board UI.

Expected behavior:

- `Ticket` and `IndexedTicket` include `Synopsis`, source metadata, and slice
  metadata needed by generated tickets.
- New ticket rendering writes `synopsis`.
- Ticket parsing reads `synopsis`, `source_kind`, `source_ref`, and `slice`.
- New ticket status constants define the Board vocabulary.
- Legacy status aliases are represented in core code with tests.

Out of scope:

- SQLite migrations.
- TUI Board navigation.
- Plan tracking prompt changes.

## Slice 2: Userspace Resolver

Introduce the active `.chr/default/` userspace resolver without migrating all
artifact directories.

Expected behavior:

- Core code can resolve the project-local active userspace path.
- The resolver reads a required active-operator environment value for workflows
  that need userspace ownership.
- `.chr/default/` can be created by workspace initialization or a targeted
  workspace repair path.
- Documentation explains `.envrc` as a convenience, not as a hard dependency.
- Full `.chr/operators/<operator-id>/...` remains documented as future work.

Out of scope:

- Full multi-operator artifact relocation.
- Rewriting dump, quiz, spec, plan, or runtime artifact paths.

## Slice 3: Board SQLite Schema

Extend the SQLite persistence shape for Board metadata.

Expected behavior:

- A migration adds Board metadata fields or a dedicated Board read model.
- The schema supports synopsis, normalized board column, source metadata, and
  slice metadata.
- Query indexes support project/column/status/use-case/project-area/type
  filtering.
- Ticket key allocation has a transaction-safe SQLite foundation or an explicit
  migration path toward one.
- Existing migrations remain ordered and reversible.

Out of scope:

- TUI rendering.
- LLM prompt changes.
- Manual sync for external Markdown.

## Slice 4: Indexed Ticket Enrichment

Wire the parser, reconciler, and index repository through the enriched schema.

Expected behavior:

- Rebuild and reconcile persist enriched metadata.
- Existing tickets without synopsis remain indexable.
- Legacy statuses populate normalized Board columns when possible.
- Unknown statuses remain queryable and attention-ready.
- Existing `chr index rebuild` and `chr index status` keep working.

Out of scope:

- Creating new tickets.
- Board-specific TUI navigation.

## Slice 5: Ticket Creation Service

Upgrade manual ticket creation through the shared service boundary.

Expected behavior:

- CLI ticket creation can accept synopsis.
- Board/TUI-created tickets use `backlog` by default.
- Service-level ticket creation validates required fields for
  Chryso-created tickets.
- Ticket creation updates or refreshes the SQLite index.
- Key allocation uses the new Board-safe allocation path when available.

Out of scope:

- Interactive multi-field TUI form polish beyond the minimal creation path.
- Generated tickets from plan tracking.

## Slice 6: Generated Ticket Synopsis

Update plan tracking materialization so generated tickets are self-describing.

Expected behavior:

- Plan tracking generation contract includes synopsis per ticket.
- Materialized tickets persist synopsis and source metadata.
- Generated tickets retain slice identifiers and source plan/tracking
  references.
- Existing plan tracking behavior remains available from CLI and TUI.

Out of scope:

- Changing plan generation itself unless required to produce ticket metadata.
- Board rendering changes.

## Slice 7: Board Query Service

Add a service-level Board read model used by both CLI and TUI.

Expected behavior:

- Board queries return normal columns, counts, tickets, and selected detail
  metadata.
- CLI can inspect Board state in a scriptable form.
- TUI loads Board data through the service model instead of regrouping raw
  tickets ad hoc.
- The service preserves CLI/TUI parity by sharing query logic.

Out of scope:

- TUI column navigation.
- Attention toggling.

## Slice 8: Compact Board Columns

Implement compact left-pane column navigation.

Expected behavior:

- The Board compact view shows one normal column at a time.
- Left/right or approved keybindings move through normal columns.
- Up/down selects tickets within the active column.
- Column labels are human-readable.
- The right pane continues to show selected ticket detail.

Out of scope:

- Attention virtual view.
- Maximized multi-column Board.

## Slice 9: Attention View

Add the special attention view outside normal column navigation.

Expected behavior:

- A dedicated keybinding toggles attention mode.
- Repeating the keybinding returns to the previous normal column.
- Attention includes invalid/unknown statuses and conservative available
  signals.
- Rows show restrained text hints.
- The detail pane shows plain-text attention reasons.

Out of scope:

- Rich reviewer/PR-state attention.
- LLM-assisted sync.
- Old-done archival.

## Slice 10: Ticket Detail Surface

Improve selected ticket detail for Board operation.

Expected behavior:

- Detail shows synopsis prominently.
- Detail shows source metadata when present.
- Detail shows attention reasons when present.
- Detail loads enough Markdown body/comment/checklist context to understand the
  selected ticket.
- Layout stays readable in normal right-pane size.

Out of scope:

- Full external editor workflow.
- Complex Markdown rendering.

## Slice 11: Maximized Board

Implement the broader maximized Board view.

Expected behavior:

- Maximized Board shows multiple normal columns at once.
- The view remains dense, keyboard-first, and readable in terminal constraints.
- Selection stays coherent when toggling between compact and maximized mode.
- Attention remains a special view, not part of normal column cycling.

Out of scope:

- F8 metrics dashboard.
- Visual analytics.

## Slice 12: Board Docs Validation

Close the slicing with tests, references, and user-facing documentation.

Expected behavior:

- Command reference documents Board/ticket changes.
- TUI reference documents compact Board, attention, manual ticket creation, and
  maximized Board behavior.
- Project workspace/runtime docs describe `.chr/default/` MVP and future
  multi-operator direction.
- SQLite migration reference documents Board metadata changes.
- CLI/TUI action matrix reflects new shared operations.
- `make verify` passes.

Out of scope:

- New behavior beyond fixes required by validation.

## Delivery Order

Implement slices sequentially:

1. Slice 1 stabilizes the file-level metadata contract.
2. Slice 2 introduces userspace path resolution before new Board artifacts need
   it.
3. Slice 3 changes the database schema.
4. Slice 4 wires parsing/reconciliation into the new schema.
5. Slice 5 updates manual ticket creation.
6. Slice 6 updates generated ticket creation.
7. Slice 7 creates the shared Board read model.
8. Slice 8 implements compact column navigation.
9. Slice 9 adds attention.
10. Slice 10 improves ticket detail.
11. Slice 11 adds maximized Board rendering.
12. Slice 12 finalizes documentation and validation.

No implementation slice should start before the previous slice is reviewed,
merged to `dev`, and validated.
