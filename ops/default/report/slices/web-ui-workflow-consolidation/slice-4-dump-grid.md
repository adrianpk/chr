# Slice 4: Dumps Grid Width

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `fix/slice-4-dump-grid`
PR: pending

## Purpose

Deliver the measured Dumps catalog sizing defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md).
Slice 3 is integrated. T4.1 establishes the width defect and fixed geometry;
T4.2 retains desktop zoom and narrow browser evidence.

## Delivered Behavior

Implementation in progress.

## Implementation Notes

Measure the existing auto-fit maximum track before changing layout. Use theme
size tokens for fixed cards and clipped inset previews. Capture and Save
semantics remain under their existing owners.

## Contracts Added or Changed

Pending measured layout evidence.

## Files of Interest

- `internal/webui/assets/static/chr.css`: Dumps grid and card geometry.
- `internal/app/web_dump_grid_browser_test.go`: owned browser measurements and captures.

## Validation

Pending focused browser checks, then Looper's exact clean-HEAD development gate.

## Risks and Follow-ups

The controller owns integration into dev. No live project data is changed.
