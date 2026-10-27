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
the complete method/transaction/recovery map is frozen before package relocation.

## Implementation Notes

The approved T12.1–T12.6 tasks own this slice. All methods of the store move with
their owning Go type. Shared stage data belongs in workflow/state when needed
without transferring persistence, coordination or transport behavior.

## Contracts Added or Changed

The [authority map](../../../plan/code-quality-remediation-slice-12-boundaries.md)
records all 908 store methods, 100 concrete IO/authority/recovery entries and 79
direct root dependencies at the verified predecessor revision. All store methods
retain one receiver owner; future policy consumers are recorded separately.
All 71 assigned baseline signature rows map to 68 current declarations; 48 require
refactoring, 22 are legal cohesion watches and one has a verified prior shared
owner fix. Relocation and signature changes are not yet implemented.

## Files of Interest

- `internal/core/workflow_store.go` and the workflow family.
- `internal/core/workflow/state/`.
- The approved source inventory and core owner ledger.

## Validation

Activation documentation links, ticket lifecycle metadata and whitespace passed.
The compiler-resolved source graph, complete method/dependency/signature identity
coverage and exact frozen declaration hashes passed. These are mapping checks;
runtime implementation and final controller validation have not run.

## Risks and Follow-ups

Configuration/transport/projection dependencies must have real boundaries before
relocation. Assigned signatures, stage ports, candidate validation and canonical
integration remain outstanding.
