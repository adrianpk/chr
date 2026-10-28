# Slice 14: Forge publication and review stages

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-14-quality-publication`
PR: pending

## Purpose

Give forge publication and review policy independent stage owners while retaining
native execution authority, guarded host publication, journals and receipt-backed
recovery.

## Delivered Behavior

No Slice 14 runtime change is claimed yet. The source inventory freezes 61
assigned signature rows: seven category A, 19 category B and 35 category C.
Original baseline identities remain linked to their current declarations.

## Implementation Notes

Predecessor PR #450 is verified merged into dev at
`53b2adb23b9b1d20529d00bbe3b32b3be5990b4e`. Controller evidence
`5CL5UBXU2HGBPKW5NNUL6O3PMP` passed its four exact-head commands. The primary
checkout remains on dev; implementation uses the recorded slice branch.

## Contracts Added or Changed

The [assigned signature ledger](../../../plan/code-quality-remediation-slice-14-signatures.csv)
records current signatures before conversion. Publication must keep exact
source/target/candidate identity, permissions, policy revision, human approval,
idempotent dispatch and uncertain-outcome reconciliation. Ticket-native execution
and forge-native publication remain distinct terms.

## Files of Interest

- `internal/core/workflow/forge/` and `internal/core/workflow/review/`.
- Existing native provider, review and submission owners and their direct callers.
- The signature, Store method, authority and stage-port ledgers.

## Validation

- Live Git, canonical Forgejo, controller evidence and Acta confirm the predecessor merge.
- Slice 14 runtime validation has not run yet.

## Risks and Follow-ups

Stage extraction, assigned signature/state/error/suppression work, focused real
publication/review/recovery checks and final controller validation remain active.
