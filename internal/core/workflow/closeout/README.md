# `workflow/closeout`

## Purpose

This package owns the consumer contract for closing gate execution.

## Contract

`Gate` is used by `Store.CloseWorkflowSet`. Execute returns the receipt for one reserved closing claim and exact retained input. The owner subsequently rechecks and commits the closing result; execution alone does not confirm closure.

The contract imports shared `workflow/state` protocol values. The workflow owner
implements it with the existing operations and retains claims, reservations,
source/revision guards, repository locks, journals, atomic writes and recovery.
The port exposes the required operation without exposing a mutable store.

## Validation

Existing workflow integration tests exercise this binding through real Git,
filesystem and process fixtures. The production dependency check permits shared
state and rejects coordinator, filesystem mutation and configuration imports.
