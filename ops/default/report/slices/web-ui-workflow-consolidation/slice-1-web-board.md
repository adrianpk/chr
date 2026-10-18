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
This report currently records T1.1; the slice is not ready for integration.

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

## Implementation Notes

T1.1 commit: `02c24fcfdc174a41755d89259810300e369d901d`.

The reproduced defect was a required `.chr/default/tickets` directory in the
legacy index read. Canonical-only and empty stores failed; mixed stores omitted
canonical tickets. The new shared read observes documents directly instead of
rebuilding that index. Native reads use existing workflow inspection exclusion
and reject pending transitions without recovering them.

Observation remains bounded by the existing 4096-record and 1 MiB file limits.
Symlink paths, ambiguous canonical keys and invalid page bounds are rejected.
Malformed loose frontmatter becomes an attention card; read errors are not
converted to empty data. Tests use isolated local stores and native Git fixtures.

## Contracts Added or Changed

`ReadTicketBoard` is a central application operation over the core read owner.
`BoardRequest` supplies the page offset to the Web provider. Board views expose
loaded counts and continuation separately from total column counts. The former
standalone native-ticket preview is removed.

## Files of Interest

- `internal/core/ticket_board_read.go`: authority selection, bounded observation and paging.
- `internal/core/ticket_board_document.go`: loose-document reads and status vocabulary mapping.
- `internal/app/service_ticket_board.go`: shared application entry point.
- `internal/app/web_views.go` and `internal/webui/assets/templates/shell.html`: one Board projection and continuation.

## Validation

Task development checks passed on the implementation tree committed as
`02c24fcfdc174a41755d89259810300e369d901d`:

- `go test -race ./internal/core ./internal/app ./internal/webui -run '^(TestTicketBoard.*|TestWorkflowInspectionLinked|TestBoardQuery.*|TestWebBoard.*|TestWebServeBoardUsesProjectTickets|TestWebWorkflowViews|TestHandlerProjectBoardDashRoutes|TestHandlerFullPagesConfigurePolling|TestHandlerBoardAndDashFragments)$' -count=1` — passed. Covers 130 native tickets across pages, deduplication, canonical document preservation, linked pending-transition rejection, status mappings, attention, invalid stores, all four source layouts, HTTP continuation and stable polling revision.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m --new-from-rev=HEAD ./internal/core ./internal/app ./internal/webui/...` — passed with zero issues; HEAD was the planning baseline during this task check.
- `scripts/check-go-func-signatures.sh` — passed.
- `git diff --check` — passed.

Before the correction, `go test ./internal/app -run '^TestWebBoardSources$' -count=1`
failed for canonical-only, mixed and empty stores, reproducing the read defect.

Final clean-HEAD slice validation through Looper, including `make verify`, has
not run. These task checks do not establish complete slice or nightly evidence.

## Risks and Follow-ups

T1.2 remains: canonical detail reads, shared creation and full-body/metadata
editing, delivery references, comments and revision/lifecycle guards.
T1.3 remains: the shared normal/maximized editor, New, drag moves, dirty-form
preservation and desktop/narrow browser evidence. The source ticket stays in
progress until all slice requirements are implemented.
