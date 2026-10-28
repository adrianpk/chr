# `workflow/review`

## Purpose

This package owns the consumer contract for review decision acceptance.

## Contract

`Decisions` is used by `Store.reviewWorkflowSlice`. Accept binds the role decision, execution failure and interruption to the reserved claim and exact input. Current policy, approval, repair and receipt guards remain with the owner.

The contract imports shared `workflow/state` protocol values. The workflow owner
implements it with the existing operations and retains claims, reservations,
source/revision guards, repository locks, journals, atomic writes and recovery.
The port exposes the required operation without exposing a mutable store.

## Validation

Existing workflow integration tests exercise this binding through real Git,
filesystem and process fixtures. The production dependency check permits shared
state and rejects coordinator, filesystem mutation and configuration imports.
