# Slice 11: Ticket, readiness and execution boundaries

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-11-quality-tickets`
PR: pending

## Purpose

Give ticket models/indexing/rules and execution/readiness their actual owners,
finish assigned signatures and typed states, and reconcile the six-parameter phase.

## Delivered Behavior

Implementation is in progress. Slice 10 integration was verified through PR #447
at `add303a00a38df26dd23b25dbf5dc5dc59cf0e9a` before activation.

## Implementation Notes

The approved T11.1–T11.6 task list owns scope. Moves and naming remain separate
commits with direct consumers and preserved durable/wire contracts.

## Contracts Added or Changed

Pending implementation.

## Files of Interest

- `internal/core/ticket/` and native ticket/indexing code.
- `internal/core/work_readiness_service.go` and execution/readiness consumers.
- The approved baseline inventory and core owner ledger.

## Validation

No Slice 11 implementation checks have run. Required candidate validation is
`make verify`, focused real SQLite/Git regressions and `git diff --check`.

## Risks and Follow-ups

T11.1–T11.6 remain open; no Slice 11 PR is prepared yet.
