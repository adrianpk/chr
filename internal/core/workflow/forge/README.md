# `workflow/forge`

## Purpose

This package owns the consumer contract for submission response retention.

## Contract

`Responses` is used by `Store.SubmitWorkflowSliceIntegration`. Retain binds the provider response and original IO failure to one invocation and journaled submission ID. Observation and confirmed delivery remain separate from an untrusted response.

The contract imports shared `workflow/state` protocol values. The workflow owner
implements it with the existing operations and retains claims, reservations,
source/revision guards, repository locks, journals, atomic writes and recovery.
The port exposes the required operation without exposing a mutable store.

## Validation

Existing workflow integration tests exercise this binding through real Git,
filesystem and process fixtures. The production dependency check permits shared
state and rejects coordinator, filesystem mutation and configuration imports.
