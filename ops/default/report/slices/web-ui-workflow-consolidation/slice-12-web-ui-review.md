# Slice 12: Consolidated Web UI Journeys

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `test/slice-12-web-ui-review`
PR: `#409`

## Purpose

Complete the final slice of the approved [plan](../../../plan/web-ui-workflow-consolidation.md)
and [tracker](../../../tracker/web-ui-workflow-consolidation.md). PRs #398 through
#408 are verified merged into dev at their recorded exact heads; all eleven
child source tickets and reports are solved/delivered. The parent ticket remains
in progress pending this slice's controller integration.

## Delivered Behavior

One deterministic persisted-workspace journey crosses project/scoped capture,
full-body Save, retained Fast quiz answer/retry/Done, manual and generated spec
edits, contextual planning, approved order binding, Board metadata and Dash
observation. Reopening the transport reads the same records without generation,
file changes or work admission.

The browser navigates six views and verifies fixed card/preview anatomy,
desktop/narrow geometry, labeled controls, modal focus/return and pending input
through refresh and independent comments. A reproduced integration gap is fixed:
a plan-bound task title now explains its existing order ownership in the editor,
and a tampered rename receives the field guard rather than a stale-data claim.
Unbound titles and permitted metadata retain their existing writes.

## Implementation Notes

Fixtures substitute only deterministic model/readiness boundaries and use real
files, authoring journals, native materialization, HTTP forms and Chromium.
Approved materialization is an explicit fixture operation; no worker or external
forge is started. The title guard uses the set already present in the coherent
native snapshot. Plan/map images remain byte-identical through permitted edits,
comments and observations.

## Contracts Added or Changed

No scheduling, approval or persistence contract is weakened. Board now exposes
the title ownership already enforced by a retained plan revision. The existing
scheduled browser step selects both the native workflow and consolidated
journeys, and rejects missing or skipped evidence for either. Its scope remains
separate from development validation; the campaign was not dispatched here.

## Files of Interest

- `internal/app/web_journey*_test.go`: persisted journey, guard reproduction and browser boundaries.
- `internal/core/ticket_document.go` and ticket editor template: visible, shared plan-bound title guard.
- `scripts/qa/nightly.py` and `test_nightly.py`: explicit scheduled browser selection and completeness checks.
- User Guide and Web action matrix: current complete Web flow and action boundaries.
- [Parent source acceptance](../../../ticket/in_progress/20261017105432-review-draft-web-ui-usability.md#source-acceptance-matrix): all twelve sources and their canonical delivery state.

## Validation

Task checks passed:

- `go test -race ./internal/app -run '^TestWebConsolidatedJourney$' -count=1`
- `go test -race ./internal/core ./internal/app ./internal/webui -run 'TicketNative|TicketDocument|WorkflowPlanOrder|WorkflowTrackerOrder|WebConsolidated|WebTicket|HandlerTicket' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui`
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_JOURNEY_SCREENSHOTS=../../ops/default/report/slices/web-ui-workflow-consolidation/files/slice-12 go test ./internal/app -run '^TestWebConsolidatedBrowser$' -count=1 -v`
- `python3 -m unittest discover -s scripts/qa -p test_nightly.py`
- `git diff --check`

The bound-title regression failed before correction and passed afterward. The
browser uses 1280×1000 and 390×844 CSS pixels, DPR 1, browser zoom 100% and page
scale 1. Cards retain their type's exact fixed dimensions and clipped preview;
all views avoid document-width overflow. Tab remains in each modal and Escape
returns focus to its trigger. Owned claims, databases, servers and browsers
finish with their tests; live credentials and production work are unnecessary.

[Desktop Dumps](files/slice-12/dumps-desktop.png),
[narrow Quizzes](files/slice-12/quizzes-narrow.png),
[desktop Specs](files/slice-12/specs-desktop.png),
[narrow Plans](files/slice-12/plans-narrow.png),
[desktop Board](files/slice-12/board-desktop.png),
[narrow Dash](files/slice-12/dash-narrow.png),
[desktop guarded ticket](files/slice-12/ticket-editor-desktop.png) and
[narrow spec editor](files/slice-12/spec-editor-narrow.png) have adjacent JSON
measurements. Other paired catalog/editor captures are in `files/slice-12/`.
Earlier child captures remain evidence for unchanged contracts; the following
matrix makes that reuse explicit.

| Group | Delivered child evidence | Consolidated evidence |
| --- | --- | --- |
| UI-01: catalogs | [Slice 4](slice-4-dump-grid.md), [6](slice-6-quiz-catalog.md), [7](slice-7-spec-authoring.md), [10](slice-10-plan-catalog.md): short/long/unbroken geometry, Dumps 100%/90% and quiz menu states | Same workspace's Dumps, Quizzes, Specs and Plans at desktop/narrow widths; fixed white cards and clipped pale-green previews. |
| UI-02: dump integrity | [Slice 2](slice-2-project-dump.md) and [3](slice-3-dump-editor.md): concurrent capture/retry, reconciliation, namespace/restart and stale Save | One general identity, independent case context, confirmed capture retry and full-body Save/reopen. |
| UI-03: quiz state | [Slice 5](slice-5-quiz-answering.md) and [6](slice-6-quiz-catalog.md): all actions, modes, membership and retained drafts/history | One Fast batch, idempotent answer, later question binding, Done with remaining work and answered catalog. |
| UI-04: spec authoring | [Slice 7](slice-7-spec-authoring.md): manual/generated identity, lifecycle and scope isolation | Direct-approved manual creation, generated draft and exact-identity edits; dirty modal survives refresh. |
| UI-05: planning | [Slice 8](slice-8-plan-generation.md): absent/insufficient/sufficient discovery, stale/model failure and distinct spec approval | Contextual plan form consumes saved discovery from the selected case and retains a revision-bound map. |
| UI-06: ordering | [Slice 9](slice-9-plan-order-tracker.md): recoverable publication, approval/amendment, stable order and unchanged map through processing | No pre-approval tickets, idempotent approved binding, preserved plan/map bytes through Board metadata, comments and Dash reads. |
| UI-07: Board | [Slice 1](slice-1-web-board.md): native/legacy/mixed/empty, paging, modal and guarded moves | Native bound task metadata edits and independent comments; title ownership visible, tampered rename rejected without a stale-data claim. |
| UI-08: Dash | [Slice 11](slice-11-work-dash.md): concurrent/quiet/blocked/stopped/failed/completed, outages and mutation-free refresh | Dash observes the edited Board identity and reopening changes neither stored files nor model calls. |
| UI-09: journey | Child functional/browser evidence above | One retained workspace, six-view browser navigation, desktop/narrow measurements, modal labels/focus/return and dirty input across refresh/comments. |

Final clean-HEAD validation is pending at this report-reference commit. The
canonical PR and controller ledger record `make verify`, affected focused race
checks, selected owned browser checks and `git diff --check` after PR-reference
follow-up. The controller separately validates the final integrated dev
candidate. No full-suite, live-provider or nightly success is claimed by these
focused development checks; scheduled certification retains its own exact head.

## Risks and Follow-ups

Scheduled exhaustive certification remains separate from development delivery.
