# Slice 1: Board and Shared Ticket Modal

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-1-web-board`
PR: `#398`

## Purpose

Deliver the Board and shared ticket editor defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[ordered task map](../../../tracker/web-ui-workflow-consolidation.md).
All three Slice 1 tasks are implemented.

## Delivered Behavior

Canonical-only, legacy, mixed, native and empty projects open one six-lane
Board. Native records win over their derived projections. Attention cards keep
identity and reasons; storage failures remain unavailable responses. Pages load
up to 128 tickets with continuation, while column counts cover all observed
records. Cards wrap long titles, show bounded excerpts and retain assignment.

One New/Edit form exposes full description, workflow, classification, named
case, descriptive ownership and delivery references. Save derives synopsis and
preserves discussion and unknown metadata. Native contract descriptions, set
membership, claims, workflow-owned statuses and verified delivery values keep
explicit read-only reasons. Repository classification maps to its own vocabulary;
status-directory transitions remain owned by the repository ticket workflow.

Normal and maximized layouts retain the same form and identity. Narrow layouts
stack fields and keep close and submit reachable. Focus stays within the dialog
and returns to the trigger. Selector saves and dragging use the same guarded
writer. Rejected or stale moves retain the actual column and display the reason.
Failed saves retain input and provide an explicit reload/discard action.

Comments append independently without saving or discarding dirty fields. The
response advances form guards only to the owner's committed comment revision,
so a later concurrent edit cannot be silently accepted by the dirty form.
Conditional polling pauses when hidden and preserves Board scroll, selection
and any open modal. New starts fresh and selecting a case creates no slice
membership or execution claim.

## Implementation Notes

- T1.1: `02c24fcfdc174a41755d89259810300e369d901d` — canonical/native reads, deduplication, attention and bounded continuation.
- T1.2: `f6907f9fa8b6644225a07ce2245c9aeb28299f97` — shared document creation and guarded body, metadata and discussion writes.
- T1.3: `2083b518f34574a001d98ac3b5e83d2286e8c581` — shared responsive editor, New, moves, independent comment fragments and browser evidence.

The original failure required `.chr/default/tickets` even for canonical-only or
empty projects. Shared reads now observe canonical documents without rebuilding
an index. Native inspection rejects pending transitions without recovery.
Bounds remain 4096 records and 1 MiB per file; ambiguous keys and unsafe paths
are rejected. Unadopted writes compare exact bytes under process-safe exclusion
and replace atomically; native writes retain workflow mutation receipts and guards.

## Contracts Added or Changed

`ReadTicketBoard`, `ReadTicketDocument`, `SaveTicketDocument` and
`CreateBoardTicket` are shared application operations. Web create/update/move
and comment actions remain thin adapters. `TicketDocument.CommittedRevision`
identifies the mutation's exact source image separately from a subsequent read.

Board views expose loaded counts and continuation separately from complete
column totals. Board fragments use semantic ETags. Ticket saves return inline
results, while successful comments replace only the discussion fragment and
retain the main form. Maximize changes CSS layout without recreating inputs.

## Files of Interest

- `internal/core/ticket_board_read.go` and `ticket_board_document.go`: authority selection, bounded observation and loose-document mapping.
- `internal/core/ticket_document.go`, `ticket_edit.go` and `ticket_edit_validation.go`: document identity, guarded edits and retained ownership.
- `internal/app/service_ticket_board.go`, `web_ticket.go` and `web_ticket_form.go`: shared operations, adapter mappings and named field choices.
- `internal/webui/ticket_editor.go`, `assets/templates/ticket_editor.html`, `chr.js` and `chr.css`: fragments, form retention, moves, focus and responsive layouts.
- `internal/app/web_board_browser_test.go` and its helpers: owned Chromium journey over real HTTP and isolated canonical/native stores.

## Validation

Task development checks passed before their implementation commits:

- `go test -race ./internal/core ./internal/app ./internal/webui -run '^(TestTicket(Document.*|Native.*|Key.*|Parser.*|Metadata.*|Commenter.*|ReviewMetadata.*|StatusUpdater.*|Repository.*|Board.*)|TestWorkflow(Legacy.*|RoundTrip|Record|InspectionLinked)|TestWeb(Ticket.*|Board.*|WorkflowViews)|TestHandler(TicketComment.*|FullPagesConfigurePolling|BoardAndDashFragments|ProjectBoardDashRoutes))$' -count=1` — passed; covers canonical/legacy/repository/native round trips, unknown metadata, stale and competing saves, comments, transitions, paging, pending inspection and HTTP projections.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app ./internal/webui/...` — passed with zero issues.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_BOARD_SCREENSHOTS=/home/adrian/Projects/worktrees/chr/20261018-1111-slice-1-web-board/ops/default/report/slices/web-ui-workflow-consolidation/files/slice-1 go test ./internal/app -run '^TestWebBoardBrowser$' -count=1` — passed; covers creation, selector and allowed/rejected/stale drags, dirty comments, stale comment/save retention, reload/discard, native read-only contracts, maximize/restore, focus and scroll preservation.

The controller records final clean-HEAD evidence after the PR-reference follow-up
commit, in this order: `make verify`, the focused race command above,
`CHR_WEB_BROWSER=/usr/sbin/chromium go test ./internal/app -run '^TestWebBoardBrowser$' -count=1`, then `git diff --check`.
These task checks are development evidence; exhaustive and live certification
remain in the existing scheduled nightly.

Browser captures use CSS viewports **1440 × 1000** (desktop) and **390 × 844**
(narrow), **DPR 1**, and **100% zoom**. The fixture owns its server, project
stores and Chromium process; all stop when the check finishes.

| View | Desktop | Narrow |
| --- | --- | --- |
| Populated Board | [Board](files/slice-1/desktop-board.png), [remaining lanes](files/slice-1/desktop-board-right.png) | [Board](files/slice-1/narrow-board.png) |
| Normal edit | [Edit](files/slice-1/desktop-edit.png) | [Edit](files/slice-1/narrow-edit.png) |
| Lower scroll area | [Delivery and comments](files/slice-1/desktop-edit-lower.png) | [Delivery and comments](files/slice-1/narrow-edit-lower.png) |
| Maximized edit | [Maximized](files/slice-1/desktop-maximized.png) | [Maximized](files/slice-1/narrow-maximized.png) |
| Fresh New form | [New](files/slice-1/desktop-new.png) | [New](files/slice-1/narrow-new.png) |

## Risks and Follow-ups

Repository ticket status-directory transitions and workflow-owned native fields
remain read-only with their owner's reason. Final integration is controller
owned; this report remains reviewing until that merge is verified.
