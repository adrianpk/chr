# Slice 15: Workflow closeout and coordinator cleanup

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-15-quality-closeout`
PR: pending

## Purpose

Give closing-gate policy and execution a cohesive closeout stage, complete bounded
lifecycle values and remove temporary migration bridges while preserving durable
transaction, lock, recovery and explicit human authority.

## Delivered Behavior

No Slice 15 runtime change is claimed yet. The assigned signature ledger freezes
40 baseline rows: four category A, 16 category B and 20 category C. Existing
current IDs and signatures remain linked to their original inventory identities.

## Implementation Notes

Predecessor PR #451 is verified merged into dev at
`8b5659a71af5d393919593c5e0d8c9a6ec0f497c`. Controller evidence `DWWEUV4RMALE7SB4LMQKN2PGSB`
passed its four exact-head commands. The primary checkout stays on dev;
implementation uses the recorded dedicated slice worktree and branch.

The 908-method Store map and current declaration preimages remain the starting
point for responsibility extraction. Moving files or relabeling method counts is
not a completed policy/persistence separation.

## Contracts Added or Changed

The [assigned signature ledger](../../../plan/code-quality-remediation-slice-15-signatures.csv)
records current lifecycle, closing, frontier, queue, continuation, amendment,
process and capacity signatures before conversion. Exact candidate confirmation,
explicit closing waiver, stopped versus completed outcomes, finite budgets,
partial results and guarded replay/receipt ownership must remain distinct.

## Files of Interest

- `internal/core/workflow/closeout/` and closing admission/execution/confirmation owners.
- Workflow lifecycle, frontier, continuation, amendment, process and capacity owners.
- Existing stage contracts, Store method and authority/dependency/ownership ledgers.

## Validation

- Live Git, canonical Forgejo, exact controller evidence and Acta confirm PR #451's merge.
- Slice 15 runtime checks have not run yet.

## Risks and Follow-ups

Closeout extraction, all 40 signature dispositions, method-level policy ownership,
bridge removal and naming/compatibility reconciliation remain active Slice 15
work. Required focused real lifecycle/recovery checks, shared guided parity,
focused races and final exact-head controller validation remain to be completed.
