# Board Ticket Operations Slicing Tracker

Status: in progress

Plan: `.prj/plans/slicing-board-ticket-operations-plan.md`

## Active Slice

Slicing name: Board Ticket Operations

Active slice number: Slice 1

Active slice short name: ticket metadata contract

Expected branch name: `feat/slice-1-ticket-metadata-contract`

Expected PR title: `feat(slice-1): ticket metadata contract`

Expected report path:
`.prj/reports/slices/board-ticket-operations/slice-1-ticket-metadata-contract.md`

Prerequisite PRs or delivered work:

- CLI/TUI Surface Parity slicing completed.
- Architecture Boundary Cleanup slicing completed.
- Plan generation safety and plan tracking materialization completed.
- Current ticket SQLite index foundation exists.

## Slicing Checklist

- [x] Capture Board/Dashboard/ticket operations quiz.
- [x] Define Board ticket operations spec.
- [x] Define Board ticket operations model.
- [x] Review existing ticket, SQLite, CLI, and TUI foundation.
- [x] Define complete ordered slice map.
- [x] Create slicing plan.
- [x] Create slicing tracker.
- [x] Commit plan and tracker directly to `dev`.
- [x] Select the first pending slice for implementation.
- [ ] Complete all planned slices.
- [ ] Validate completed slicing with `make verify`.
- [ ] Decide whether post-slice tuning is needed.

## Slice Status

| Slice | Status | Expected Branch | Expected PR Title | Expected Report | Notes |
| --- | --- | --- | --- | --- | --- |
| Slice 1: Ticket Metadata Contract | In progress | `feat/slice-1-ticket-metadata-contract` | `feat(slice-1): ticket metadata contract` | `.prj/reports/slices/board-ticket-operations/slice-1-ticket-metadata-contract.md` | Metadata, synopsis, source fields, Board statuses, legacy aliases. |
| Slice 2: Userspace Resolver | Pending | `feat/slice-2-userspace-resolver` | `feat(slice-2): userspace resolver` | `.prj/reports/slices/board-ticket-operations/slice-2-userspace-resolver.md` | Active `.chc/user/` resolver and env contract. |
| Slice 3: Board SQLite Schema | Pending | `feat/slice-3-board-sqlite-schema` | `feat(slice-3): board sqlite schema` | `.prj/reports/slices/board-ticket-operations/slice-3-board-sqlite-schema.md` | Migration/index shape for Board metadata and key allocation foundation. |
| Slice 4: Indexed Ticket Enrichment | Pending | `feat/slice-4-indexed-ticket-enrichment` | `feat(slice-4): indexed ticket enrichment` | `.prj/reports/slices/board-ticket-operations/slice-4-indexed-ticket-enrichment.md` | Parser/reconciler/repository through enriched metadata. |
| Slice 5: Ticket Creation Service | Pending | `feat/slice-5-ticket-creation-service` | `feat(slice-5): ticket creation service` | `.prj/reports/slices/board-ticket-operations/slice-5-ticket-creation-service.md` | Manual creation with synopsis, backlog default, index refresh. |
| Slice 6: Generated Ticket Synopsis | Pending | `feat/slice-6-generated-ticket-synopsis` | `feat(slice-6): generated ticket synopsis` | `.prj/reports/slices/board-ticket-operations/slice-6-generated-ticket-synopsis.md` | Plan tracking materialization ticket synopsis/source metadata. |
| Slice 7: Board Query Service | Pending | `feat/slice-7-board-query-service` | `feat(slice-7): board query service` | `.prj/reports/slices/board-ticket-operations/slice-7-board-query-service.md` | Shared CLI/TUI Board read model. |
| Slice 8: Compact Board Columns | Pending | `feat/slice-8-compact-board-columns` | `feat(slice-8): compact board columns` | `.prj/reports/slices/board-ticket-operations/slice-8-compact-board-columns.md` | One-column compact Board navigation. |
| Slice 9: Attention View | Pending | `feat/slice-9-attention-view` | `feat(slice-9): attention view` | `.prj/reports/slices/board-ticket-operations/slice-9-attention-view.md` | Virtual attention view and plain-text reasons. |
| Slice 10: Ticket Detail Surface | Pending | `feat/slice-10-ticket-detail-surface` | `feat(slice-10): ticket detail surface` | `.prj/reports/slices/board-ticket-operations/slice-10-ticket-detail-surface.md` | Synopsis, source refs, attention, body/comment context. |
| Slice 11: Maximized Board | Pending | `feat/slice-11-maximized-board` | `feat(slice-11): maximized board` | `.prj/reports/slices/board-ticket-operations/slice-11-maximized-board.md` | Multi-column maximized Board. |
| Slice 12: Board Docs Validation | Pending | `feat/slice-12-board-docs-validation` | `feat(slice-12): board docs validation` | `.prj/reports/slices/board-ticket-operations/slice-12-board-docs-validation.md` | Documentation, parity matrix, migration reference, final validation. |

## Slice 1: Ticket Metadata Contract

Candidate tasks:

- [x] T1.1 Add core constants/types for Board statuses and legacy status aliases.
- [x] T1.2 Extend `Ticket` with synopsis, source metadata, and slice metadata.
- [x] T1.3 Render new metadata fields in ticket frontmatter.
- [x] T1.4 Parse new metadata fields into `IndexedTicket`.
- [x] T1.5 Add metadata validation tests for Board-created tickets and legacy
  tickets.
- [x] T1.6 Create slice implementation report.
- [x] T1.7 Run validation.

Task commits:

- `5cf1805 feat(ticket): define board status aliases`
- `7bd604a feat(ticket): add board metadata fields`
- `6a67fba feat(ticket): render board metadata`
- `b7adc7b feat(ticket): parse board metadata`
- `14e4b6d test(ticket): cover board metadata validation`
- `2af922b docs(board): report ticket metadata slice`
- `b87e8dd refactor(ticket): simplify metadata rendering`

Validation:

- `go test ./internal/chc -run 'TestTicketStatus|TestTicketRenderer|TestTicketParser'`: pass.
- `go test ./internal/chc -run 'TestTicketRenderer'`: pass.
- `go test ./internal/chc -run 'TestTicketParser'`: pass.
- `go test ./internal/chc -run 'TestValidateIndexedTicket|TestTicketStatus'`: pass.
- `git diff --check`: pass.
- `go test ./...`: pass.
- `make verify`: pass.

## Slice 2: Userspace Resolver

Candidate tasks:

- [ ] T2.1 Add project-local userspace resolver for `.chc/user/`.
- [ ] T2.2 Add active-operator environment resolution and validation.
- [ ] T2.3 Ensure workspace initialization or repair can create `.chc/user/`.
- [ ] T2.4 Document `.envrc` usage and non-direnv setup.
- [ ] T2.5 Add resolver tests.
- [ ] T2.6 Create slice implementation report.
- [ ] T2.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 3: Board SQLite Schema

Candidate tasks:

- [ ] T3.1 Decide direct `indexed_tickets` migration vs dedicated Board read
  model and record the decision in the report.
- [ ] T3.2 Add migration for synopsis, normalized Board column, source metadata,
  and slice metadata.
- [ ] T3.3 Add indexes for Board queries.
- [ ] T3.4 Add transaction-safe ticket key allocation persistence.
- [ ] T3.5 Add migration tests.
- [ ] T3.6 Update SQLite migration reference.
- [ ] T3.7 Create slice implementation report.
- [ ] T3.8 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 4: Indexed Ticket Enrichment

Candidate tasks:

- [ ] T4.1 Update `IndexedTicket` and scan/upsert logic for enriched fields.
- [ ] T4.2 Normalize Board columns during parse/reconcile.
- [ ] T4.3 Preserve indexability for legacy tickets without synopsis.
- [ ] T4.4 Surface unknown statuses as attention-ready metadata.
- [ ] T4.5 Update index rebuild/status tests.
- [ ] T4.6 Create slice implementation report.
- [ ] T4.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 5: Ticket Creation Service

Candidate tasks:

- [ ] T5.1 Add synopsis to `CreateTicketRequest` and CLI flags.
- [ ] T5.2 Default Board/TUI-created tickets to `backlog`.
- [ ] T5.3 Validate required metadata for ChrysoCoord-created tickets.
- [ ] T5.4 Use SQLite-backed key allocation or the Slice 3 allocation boundary.
- [ ] T5.5 Refresh affected ticket index after creation.
- [ ] T5.6 Add CLI/service tests.
- [ ] T5.7 Create slice implementation report.
- [ ] T5.8 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 6: Generated Ticket Synopsis

Candidate tasks:

- [ ] T6.1 Update plan tracking request/response contracts for ticket synopsis.
- [ ] T6.2 Include source metadata and slice identifiers in generated tickets.
- [ ] T6.3 Materialize generated tickets with synopsis in frontmatter.
- [ ] T6.4 Update TUI/CLI plan tracking paths to reload enriched tickets.
- [ ] T6.5 Add tests for generated ticket synopsis and source refs.
- [ ] T6.6 Create slice implementation report.
- [ ] T6.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 7: Board Query Service

Candidate tasks:

- [ ] T7.1 Add Board query structs for columns, counts, rows, and detail
  metadata.
- [ ] T7.2 Implement shared Board query service over the ticket index.
- [ ] T7.3 Add CLI inspection command for Board state where meaningful.
- [ ] T7.4 Route TUI Board loading through the shared query service.
- [ ] T7.5 Add service and CLI tests.
- [ ] T7.6 Create slice implementation report.
- [ ] T7.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 8: Compact Board Columns

Candidate tasks:

- [ ] T8.1 Add TUI state for active normal Board column.
- [ ] T8.2 Implement normal column navigation.
- [ ] T8.3 Filter the compact left pane to the active column.
- [ ] T8.4 Render human-readable column labels and counts.
- [ ] T8.5 Preserve selection when changing columns where possible.
- [ ] T8.6 Add TUI model tests.
- [ ] T8.7 Create slice implementation report.
- [ ] T8.8 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 9: Attention View

Candidate tasks:

- [ ] T9.1 Add attention reason model and conservative detectors.
- [ ] T9.2 Add TUI attention mode toggle that preserves previous normal column.
- [ ] T9.3 Render attention row hints.
- [ ] T9.4 Show plain-text attention reasons in detail.
- [ ] T9.5 Add tests for invalid statuses and toggle behavior.
- [ ] T9.6 Create slice implementation report.
- [ ] T9.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 10: Ticket Detail Surface

Candidate tasks:

- [ ] T10.1 Add synopsis to Board detail.
- [ ] T10.2 Add source metadata and slice metadata to Board detail.
- [ ] T10.3 Load selected Markdown body/comment/checklist context safely.
- [ ] T10.4 Fit detail rendering in normal right-pane constraints.
- [ ] T10.5 Add TUI detail rendering tests.
- [ ] T10.6 Create slice implementation report.
- [ ] T10.7 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 11: Maximized Board

Candidate tasks:

- [ ] T11.1 Add maximized Board renderer for multiple normal columns.
- [ ] T11.2 Preserve selected ticket and active column across maximize toggles.
- [ ] T11.3 Keep attention separate from normal maximized columns.
- [ ] T11.4 Add terminal-size-aware rendering tests.
- [ ] T11.5 Create slice implementation report.
- [ ] T11.6 Run validation.

Task commits:

- Pending.

Validation:

- Pending.

## Slice 12: Board Docs Validation

Candidate tasks:

- [ ] T12.1 Update TUI reference.
- [ ] T12.2 Update command reference.
- [ ] T12.3 Update project workspace and runtime/config docs for userspace.
- [ ] T12.4 Update SQLite migration reference.
- [ ] T12.5 Update CLI/TUI parity matrix.
- [ ] T12.6 Run `gofmt` and targeted tests.
- [ ] T12.7 Run `make verify`.
- [ ] T12.8 Create slice implementation report.

Task commits:

- Pending.

Validation:

- Pending.

## Completion

Planned slices are pending.

Final validation:

- Pending.

Post-slice tuning needed:

- Pending.

Known deferred work:

- Full `.chc/operators/<operator-id>/...` multi-operator layout.
- Manual Markdown sync with LLM-assisted empty synopsis generation.
- Daemon-based watching, retained only as an evaluated non-preferred option.
- Done-ticket archival heuristics.
- Rich reviewer/PR-state attention.
- F8 metrics dashboard.
