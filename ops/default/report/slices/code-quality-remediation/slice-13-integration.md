# Slice 13: Candidate validation and merge stages

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-13-quality-integration`
PR: pending

## Purpose

Give candidate preparation/validation and merge/repair policy their approved
stage owners while preserving the store's guarded transactions and recovery.

## Delivered Behavior

Implementation in progress.

## Implementation Notes

The predecessor PR #449 is verified merged into dev at
`a49c32c9f5e81e1bde9f585a25c6d5cc669d7a9e`. Its exact-head controller evidence
is `EVFKMQZNCCYQDPLZLRXOYO2S7M`; all four recorded commands passed.

## Contracts Added or Changed

Pending implementation. The existing consumer-owned stage ports and shared
state values are the extraction boundary; locks and mutation resources retain
one store owner.

## Files of Interest

- `internal/core/workflow/integrate/`.
- `internal/core/workflow/merge/`.
- The existing workflow authority and method maps.

## Validation

No Slice 13 implementation checks have run.

## Risks and Follow-ups

The slice remains active until its approved tasks and exact-head checks pass.
