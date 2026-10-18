# Slice 1: Board and Shared Ticket Modal

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-1-web-board`
PR: pending

## Purpose

Deliver the Board and shared ticket editor defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[ordered task map](../../../tracker/web-ui-workflow-consolidation.md).
T1.1 and T1.2 are implemented. T1.3 remains; the slice is not ready for integration.

## Delivered Behavior

T1.1 repairs the Board read path for canonical-only, legacy, mixed and empty
projects. Native tasks appear in the same six lanes as supported loose tickets;
their derived legacy projections do not produce duplicate cards. Repository
ticket resolution and legacy runtime status aliases retain their distinct
meanings. Attention cards retain their document identity and reasons.

Reads expose up to 128 tickets per page, with continuation and previous-page
links. Column totals cover the entire bounded observation, independently of
the loaded page. Polling retains the selected page and uses a semantic revision
derived from source images. Storage failures remain unavailable responses.

T1.2 adds canonical document detail, independent ticket creation, full-description
replacement, classification, named case selection, descriptive assignee/owner,
branch, PR and commit references, and separate comments through shared operations.
Description saves derive the synopsis without replacing discussion or unknown
frontmatter. Comment append preserves earlier entries, including when Comments
is the final section. Nested metadata no longer shadows top-level fields.

Saves reject stale observations at the persistence owner. Native writes retain
the existing workflow mutation and claim guards; unadopted writes compare source
bytes under process-safe exclusion and replace atomically. Metadata-only native
edits preserve authoritative status/case bindings even when a body projection is
outdated. Key recovery includes native envelopes and other ops namespaces.

## Implementation Notes

T1.1 commit: `02c24fcfdc174a41755d89259810300e369d901d`.

T1.2 commit: `f6907f9fa8b6644225a07ce2245c9aeb28299f97`.

The reproduced defect was a required `.chr/default/tickets` directory in the
legacy index read. Canonical-only and empty stores failed; mixed stores omitted
canonical tickets. The new shared read observes documents directly instead of
rebuilding that index. Native reads use existing workflow inspection exclusion
and reject pending transitions without recovering them.

Observation remains bounded by the existing 4096-record and 1 MiB file limits.
Symlink paths, ambiguous canonical keys and invalid page bounds are rejected.
Malformed loose frontmatter becomes an attention card; read errors are not
converted to empty data. Tests use isolated local stores and native Git fixtures.

Native descriptions bound to approved execution contracts, set-owned case
associations, claimed tickets, workflow-owned statuses and verified delivery
values retain their existing owners and explicit read-only reasons. Repository
status-directory transitions remain read-only in this editor and use the
repository ticket workflow; their classification fields map explicitly to
`kind`, `scope` and `source`. The UI must render these constraints in T1.3.

## Contracts Added or Changed

`ReadTicketBoard` is a central application operation over the core read owner.
`BoardRequest` supplies the page offset to the Web provider. Board views expose
loaded counts and continuation separately from total column counts. The former
standalone native-ticket preview is removed.

`ReadTicketDocument`, `SaveTicketDocument` and `CreateBoardTicket` are shared
application operations. Their core owner binds edits to the observed source
image, accepts only supported descriptive fields and preserves native membership,
claim and receipt data. Web update and move actions call the same guarded save;
comments use a separate request without other form fields.

## Files of Interest

- `internal/core/ticket_board_read.go`: authority selection, bounded observation and paging.
- `internal/core/ticket_board_document.go`: loose-document reads and status vocabulary mapping.
- `internal/app/service_ticket_board.go`: shared application entry point.
- `internal/app/web_views.go` and `internal/webui/assets/templates/shell.html`: one Board projection and continuation.
- `internal/core/ticket_document.go` and `ticket_edit.go`: document reads, source binding, full-body edits and native writes.
- `internal/core/ticket_edit_validation.go`: field vocabularies, read-only constraints and guarded transitions.
- `internal/app/web_ticket.go`: thin create/update/move/comment mappings to shared operations.

## Validation

Task development checks passed on the implementation tree committed as
`02c24fcfdc174a41755d89259810300e369d901d`:

- `go test -race ./internal/core ./internal/app ./internal/webui -run '^(TestTicketBoard.*|TestWorkflowInspectionLinked|TestBoardQuery.*|TestWebBoard.*|TestWebServeBoardUsesProjectTickets|TestWebWorkflowViews|TestHandlerProjectBoardDashRoutes|TestHandlerFullPagesConfigurePolling|TestHandlerBoardAndDashFragments)$' -count=1` — passed. Covers 130 native tickets across pages, deduplication, canonical document preservation, linked pending-transition rejection, status mappings, attention, invalid stores, all four source layouts, HTTP continuation and stable polling revision.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app ./internal/webui/...` — passed with zero issues; HEAD was the planning baseline during this task check.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

Before the correction, `go test ./internal/app -run '^TestWebBoardSources$' -count=1`
failed for canonical-only, mixed and empty stores, reproducing the read defect.

T1.2 task development checks passed before its commit:

- `go test -race ./internal/core ./internal/app ./internal/webui -run '^(TestTicket(Document.*|Native.*|Key.*|Parser.*|Metadata.*|Commenter.*|ReviewMetadata.*|StatusUpdater.*|Repository.*)|TestWorkflow(Legacy.*|RoundTrip|Record|InspectionLinked)|TestWeb(Ticket.*|Board.*|WorkflowViews)|TestHandlerTicketCommentRefreshesModal)$' -count=1` — passed. Covers canonical/legacy/repository document round trips, retained fields and unknown metadata, stale and competing saves, final-section discussion append, native ownership and projected-status guards, independent creation, key recovery, and shared Web creation/reload/moves.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app ./internal/webui/...` — passed with zero issues; HEAD was `323ff51fb3f13c1c266b1843549360ba460368d8` during this task check.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

Final clean-HEAD slice validation through Looper, including `make verify`, has
not run. These task checks do not establish complete slice or nightly evidence.

## Risks and Follow-ups

T1.3 remains: the shared normal/maximized editor, New, drag moves, dirty-form
preservation and desktop/narrow browser evidence. The source ticket stays in
progress until all slice requirements are implemented.
