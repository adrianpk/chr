# Slice 11: Live Work Dash

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-11-work-dash`
PR: pending

## Purpose

Deliver the approved [plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md) Slice 11. Slice 10
is verified integrated through PR #407 at
`ce1a19c0c10a50c27807b976fa353c81a0f6912e`.

## Delivered Behavior

Pending implementation of authority-aware queue, current work, recent activity
and retained results with bounded, mutation-free refresh and stale recovery.

## Implementation Notes

T11.1 is active; T11.2 and T11.3 follow in the recorded order.

## Contracts Added or Changed

Pending implementation.

## Files of Interest

- `internal/app/web_views.go`: current Dash read projection.
- `internal/webui/assets/templates/shell.html`: existing Dash presentation.

## Validation

- `git diff --check` passed for activation and verified Slice 10 closure.

## Risks and Follow-ups

None recorded.
