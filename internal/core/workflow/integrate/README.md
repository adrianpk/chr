# `workflow/integrate`

## Purpose

This package owns the consumer contract for candidate validation.

## Contract

`Validation` is used by `Store.PrepareWorkflowSliceIntegration`. Validate runs the bound host validation. Accept rechecks the claim, prepared input and current source before the existing atomic acceptance transition.

The contract imports shared `workflow/state` protocol values. The workflow owner
implements it with the existing operations and retains claims, reservations,
source/revision guards, repository locks, journals, atomic writes and recovery.
The port exposes the required operation without exposing a mutable store.

## Validation

Existing workflow integration tests exercise this binding through real Git,
filesystem and process fixtures. The production dependency check permits shared
state and rejects coordinator, filesystem mutation and configuration imports.
