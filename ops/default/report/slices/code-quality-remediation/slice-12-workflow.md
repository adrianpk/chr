# Slice 12: Workflow state, store and coordinator package

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-12-quality-workflow`
PR: pending

## Purpose

Establish the workflow state and persistence boundary with one store/locking
and recovery authority, direct consumers and cohesive stage contracts.

## Delivered Behavior

Slice 11 is verified integrated through PR #448. Slice 12 implementation is active;
workflow method/transaction/recovery mapping precedes package relocation.

## Implementation Notes

The approved T12.1–T12.6 tasks own this slice. All methods of the store move with
their owning Go type. Shared stage data belongs in workflow/state when needed
without transferring persistence, coordination or transport behavior.

## Contracts Added or Changed

None implemented yet.

## Files of Interest

- `internal/core/workflow_store.go` and the workflow family.
- `internal/core/workflow/state/`.
- The approved source inventory and core owner ledger.

## Validation

Activation documentation links, ticket lifecycle metadata and whitespace were
checked. Implementation checks and final controller validation have not run.

## Risks and Follow-ups

The initial store method map must identify transaction/recovery boundaries before
relocation. Candidate validation and canonical integration remain outstanding.
