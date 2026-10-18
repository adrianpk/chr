# Slice 12: Consolidated Web UI Journeys

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `test/slice-12-web-ui-review`
PR: pending

## Purpose

Complete the final slice of the approved [plan](../../../plan/web-ui-workflow-consolidation.md)
and [tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 11 is
verified integrated through PR #408 at `f6a01cf8ec95012109fa78bdf06629233c5088db`.

## Delivered Behavior

T12.1 verifies one persisted workspace across general/scoped capture and Save,
Fast quiz answer/retry/Done, manual and generated spec editing, contextual
planning, approved revision-bound ticket allocation, Board metadata and Dash
observation. Reopening the views changes neither stored files nor model calls.
T12.2 adds one browser navigation across the same workspace, fixed card and
preview measurements, labeled controls, modal focus/return, narrow geometry and
dirty form retention through refresh/comments. A reproduced gap is corrected:
plan-bound task titles now expose their existing ownership guard in the Board
instead of offering a rename rejected as stale. T12.3 remains.

## Implementation Notes

Use deterministic shared-workspace journeys and the delivered child fixtures.
The title guard uses the already inspected set revision; ordinary unbound titles
remain editable. Server-side field validation and the existing order owner both
retain authority.

## Contracts Added or Changed

The journey reuses the delivered owners and their existing contracts.

## Files of Interest

The source ticket owns task progress; this report owns consolidated evidence.

## Validation

T12.1 checks passed:

- `go test -race ./internal/app -run '^TestWebConsolidatedJourney$' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/app`
- `git diff --check`

The fixture checks one general destination, independent named context, one Fast
batch, exact manual/generated spec identities, no ticket allocation before
approval, idempotent allocation, and unchanged plan/map bytes through Board and
Dash operations. No worker, model service or external forge is invoked.
T12.2 checks passed:

- `go test -race ./internal/core ./internal/app ./internal/webui -run 'TicketNative|TicketDocument|WorkflowPlanOrder|WorkflowTrackerOrder|WebConsolidated|WebTicket|HandlerTicket' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui`
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_JOURNEY_SCREENSHOTS=../../ops/default/report/slices/web-ui-workflow-consolidation/files/slice-12 go test ./internal/app -run '^TestWebConsolidatedBrowser$' -count=1 -v`
- `git diff --check`

The title guard regression failed before the correction and passed afterward.
The browser covers six views at 1280×1000 and 390×844 CSS pixels, DPR 1, 100%
browser zoom and page scale 1. Adjacent JSON records card/preview and modal
geometry. No view has document-width overflow. Keyboard Tab stays in each
modal, Escape restores its trigger, and labeled fields retain pending input
through workspace refresh. Comments preserve dirty ownership fields, and their
subsequent Save leaves the plan and map unchanged.

[Desktop Board](files/slice-12/board-desktop.png),
[narrow Dash](files/slice-12/dash-narrow.png),
[desktop guarded ticket](files/slice-12/ticket-editor-desktop.png),
[narrow spec editor](files/slice-12/spec-editor-narrow.png) and the other catalog/editor
captures reside under `files/slice-12/` with JSON measurements.
Final exact-HEAD checks remain pending.

## Risks and Follow-ups

None recorded.
