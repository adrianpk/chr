# Slice 7: Live Sync And Conflict Handling

Status: implementing

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

Pending implementation.

## Changes

Pending implementation.

## Validation

Pending implementation.

## Notes

- This report records the transport decision before runtime changes as required
  by the slice tracker.
