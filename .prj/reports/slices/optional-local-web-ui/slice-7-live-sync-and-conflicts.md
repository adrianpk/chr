# Slice 7: Live Sync And Conflict Handling

Status: reviewing

Plan: `ops/plans/2026-06-29-optional-local-web-ui-plan.md`

Tracker: `ops/trackers/2026-06-29-optional-local-web-ui-tracker.md`

Spec: `ops/specs/2026-06-29-optional-local-web-ui.md`

Branch: `feat/slice-7-live-sync-and-conflicts`

Expected PR Title: `feat(slice-7): live sync and conflicts`

PR: pending

## Sync Transport Decision

Initial transport: htmx polling.

Rationale:

- The Web UI is local-first and server-rendered.
- Existing screens already expose fragment routes for Project, Board, Dash,
  catalogs, and detail views.
- Polling can reuse those fragments without introducing a separate event feed
  protocol in this slice.
- The later access-control slice can harden request boundaries without needing
  to secure a long-lived event stream at the same time.

Polling policy:

- Full-page shells carry a stable state revision marker.
- Major read surfaces request their matching fragment on a fixed interval.
- Hidden tabs should not poll aggressively.
- Write success paths continue to redirect to refreshed read models.
- Stale writes continue to return a conflict envelope instead of overwriting
  changed artifacts.

Deferred transport:

- Server-sent events are deferred until there is a stronger need for low-latency
  operation telemetry.
- A dedicated `/api/events` endpoint is not added in this slice because the
  selected polling transport does not require it.

## Summary

- Added htmx polling to the Web UI shell for Project, Board, Dash, artifact
  catalogs, and artifact detail views.
- Added stable revision markers to full-page shells and htmx fragments.
- Kept `/api/events` out of scope because the selected transport is polling.
- Added stale-write conflict actions so rejected writes offer reload and
  discard paths.
- Covered external workspace mutation, stale rejection, and reload retry.

## Changes

- `internal/webui/views.go`
  - Adds page-level `Revision` and `RefreshURL` fields.
  - Adds revision fields to Project, Board, Dash, and catalog view models.
  - Extends write responses with `ReloadURL` and `DiscardURL`.
- `internal/app/web_revision.go`
  - Adds stable JSON-backed revision hashing for aggregate read models.
- `internal/app/web_views.go`
  - Populates stable revisions for Project, Board, and Dash read views.
- `internal/app/web_artifacts.go`
  - Populates catalog and fallback detail revisions.
  - Keeps content-hash revisions for editable artifact details.
- `internal/webui/webui.go`
  - Maps every major page to its matching refresh fragment URL.
- `internal/webui/assets/templates/shell.html`
  - Adds polling attributes to the workspace shell.
  - Adds hidden revision markers to each fragment.
  - Adds reload/discard actions for stale writes.
- `internal/webui/assets/static/chc.css`
  - Adds compact action-row styling for conflict actions.
- `internal/webui/webui_test.go`
  - Covers polling attributes, fragment revision markers, and stale conflict
    action rendering.
- `internal/app/cli_test.go`
  - Covers external workspace mutations changing revisions.
  - Covers stale write rejection and successful retry after reload.

## Validation

- `go test ./internal/webui` (pass)
- `go test ./internal/app -run 'WebServe.*(Artifact|WriteAction|Revision|Handler|Board|Catalog)'` (pass)
- Temporary-project smoke with `go run ./cmd/chc web serve --dir /tmp/... --port 7337` plus:
  - `curl -fsS http://127.0.0.1:7337/healthz` (pass)
  - Full dump detail page exposed a `data-revision` marker (pass)
  - Direct workspace file mutation changed the dump detail fragment revision
    returned with `HX-Request: true` (pass)
  - `curl -X POST http://127.0.0.1:7337/actions` for `dump.save` with the
    refreshed revision returned `303` to the dump detail page (pass)
  - The saved dump file contained the Web UI update (pass)
- `make verify` (pass)

## Notes

- This report records the transport decision before runtime changes as required
  by the slice tracker.
