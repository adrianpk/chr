# Slice 3: Single Dump Editor

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `fix/slice-3-dump-editor`
PR: pending

## Purpose

Deliver the single dump editor and guarded Save defined by the
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[task map](../../../tracker/web-ui-workflow-consolidation.md).
Slice 2 is integrated. T3.1 verifies the existing shared Save owner before T3.2
removes duplicate editing surfaces and preserves pending input.

## Delivered Behavior

Implementation in progress.

## Implementation Notes

Slice 2 already routes Web Save through the canonical serialized mutation with
operation identity and observed document revision. Reuse that owner rather
than introducing another writer or changing capture append behavior.

## Contracts Added or Changed

Pending.

## Files of Interest

- `internal/app/web_actions.go` and `service_dump.go`: shared Save binding.
- `internal/webui/assets/templates/shell.html`, `chr.js` and `webui.go`: editor and write-result retention.

## Validation

Pending focused HTTP, persistence and browser regression evidence, followed by
Looper's exact clean-HEAD `make verify`, focused checks and `git diff --check`.

## Risks and Follow-ups

The controller owns integration into dev. No live project data is changed.
