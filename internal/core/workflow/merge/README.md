# `workflow/merge`

## Purpose

This package owns the consumer contract for merger evidence checks.

## Contract

`Evidence` is used by `Store.runWorkflowMerger`. Evidence checks the exact retained merger input, contained evidence and hashes through Store.CheckWorkflowMergerEvidence. Prepared-claim checks and editable-checkout supervision retain their current owner.

The contract imports shared `workflow/state` protocol values. The workflow owner
implements it with the existing operations and retains claims, reservations,
source/revision guards, repository locks, journals, atomic writes and recovery.
The port exposes the required operation without exposing a mutable store.

## Validation

Existing workflow integration tests exercise this binding through real Git,
filesystem and process fixtures. The production dependency check permits shared
state and rejects coordinator, filesystem mutation and configuration imports.
