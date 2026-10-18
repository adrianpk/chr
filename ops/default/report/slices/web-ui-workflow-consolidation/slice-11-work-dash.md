# Slice 11: Live Work Dash

Status: reviewing
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-11-work-dash`
PR: `#408`

## Purpose

Deliver the approved [plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md) Slice 11. Slice 10
is verified integrated through PR #407 at
`ce1a19c0c10a50c27807b976fa353c81a0f6912e`.

## Delivered Behavior

Dash shows the shared pending queue at left, independent active contexts at
upper right, and recent activity and retained results below. Narrow screens
stack these panels. Quiet validation remains active without a phase event;
known outcomes survive event silence and reopening. Native host validation,
confirmed ticket delivery and legacy harness success remain distinct.

Visible tabs refresh bounded observations with conditional revision reads.
Unchanged responses preserve the DOM. Changed responses retain scroll, focused
links and selected text that still exists in its original context. Read outages
retain the last view with a stale timestamp, then recover without starting work.
Confirmed Preview Work and Stop Daemon forms remain usable outside the refresh
region, with separate results and visible monitoring during pending requests.

## Implementation Notes

Native observation uses the inspection lock and shared selection, prerequisite
and declared ordering rules without an unrelated legacy Board dependency.
Legacy observation reads authoritative ticket files and existing SQLite records
in read-only mode, without initializing a database. Event identity includes its
original execution context; foreign project, execution and ticket events cannot
supply another context's phase. Completed or stopped invocations retain their
own outcome instead of borrowing a newer attempt's task phase.

## Contracts Added or Changed

`ObserveWork` is the central adapter-neutral observation. Native inspection does
not recover pending transitions or rewrite maps. Queue pages contain at most
128 items with total and continuation; contexts, retained results and event tails
have explicit bounds in the [owning reference](../../../../../docs/reference/work-observation.md).
Semantic view revisions exclude observation time. Conditional responses return
304 with the latest observation time; failed polling preserves stale content.
Opening, polling and reopening cannot claim, execute, stop, migrate or publish.
Retained controls use their existing shared operations and confirmation guards.

## Files of Interest

- `internal/app/service_work_observation*.go`: shared authority-aware observations.
- `internal/core/workflow_work_observation.go`: coherent native inspection.
- `internal/core/legacy_work_observation.go` and `agent_event_recent.go`: read-only records and bounded event tails.
- `internal/app/web_dash.go` and `internal/webui/dash.go`: bounded pages and conditional HTTP reads.
- `internal/webui/assets/templates/dash.html` and `assets/static/dash.js`: independent panels, controls and refresh recovery.
- User Guide, work observation reference and Web action matrix: operator-facing behavior and evidence boundaries.

## Validation

Task checks passed:

- `go test -race ./internal/core ./internal/app ./internal/webui -run 'WorkflowWorkObservation|WorkflowWorkPending|LegacyWorkObservation|RecentWorkEvents|WorkObservation|WorkflowSelection|WebDash|DashControls|DashRefresh|DashUnavailable|HandlerProjectBoardDashRoutes|HandlerBoardAndDashFragments|HandlerRendersGenerationAndWorkActionForms' -count=1`
- `go test -race ./internal/app -run 'WorkObservation|WebDash' -count=1`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/core ./internal/app ./internal/webui`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/app`
- `CHR_WEB_BROWSER=/usr/sbin/chromium CHR_WEB_DASH_SCREENSHOTS=../../ops/default/report/slices/web-ui-workflow-consolidation/files/slice-11 go test ./internal/app -run '^TestWebDashBrowser$' -count=1 -v`
- `git diff --check`

Real fixtures cover two persisted native claims, concurrent mutation-free reads,
blocked recovery, corrupt unrelated legacy files, absent/read-only databases,
130-ticket paging, stable event deduplication and retained results without phase
events. Lifecycle cases include quiet validation, blocked, stopped, cancelled,
failed and completed work. Browser transitions verify unchanged DOM, changed
scroll/focus/selection, pending controls, stale recovery, reopening, hidden-tab
pause and automatic visible polling. Tests close their owned claims, databases,
servers and browser processes without launching a live worker.

[Queued desktop](files/slice-11/queued-desktop.png),
[queued narrow](files/slice-11/queued-narrow.png),
[running desktop](files/slice-11/running-desktop.png),
[running narrow](files/slice-11/running-narrow.png),
[completed desktop](files/slice-11/completed-desktop.png),
[completed narrow](files/slice-11/completed-narrow.png) and
[stale narrow](files/slice-11/stale-narrow.png) have adjacent JSON measurements.
At 1280×1000 and 390×844 CSS pixels, DPR 1, browser zoom 100% and page scale 1,
the layout uses two desktop columns and one narrow column without document-width
overflow. Running observations retain two independent contexts and one deduplicated
event; completed observations retain two results with no event requirement.

Final clean-HEAD validation is pending at this report-reference commit. The
canonical PR and controller ledger record required `make verify`, affected race
checks, the Dash browser check and whitespace validation after PR-reference
follow-up. Scheduled exhaustive verification remains a separate evidence boundary.

## Risks and Follow-ups

None recorded.
