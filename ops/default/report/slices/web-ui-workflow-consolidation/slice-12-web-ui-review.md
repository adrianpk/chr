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
T12.2 and T12.3 remain.

## Implementation Notes

Use deterministic shared-workspace journeys and the delivered child fixtures.

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
Final exact-HEAD checks remain pending.

## Risks and Follow-ups

None recorded.
