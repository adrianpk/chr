# `workflow/closeout`

## Purpose

This package owns closing lease, binding and result policy, plus ordered gate
execution against one frozen closing input.

## Contract

`Execute` builds the unconfirmed receipt, preserves the explicit waiver, retains
each command's partial result before checking failure and checks the final
checkout. It receives the host's watched deadline context through narrow
`Workspace` and `Commands` ports. A waived gate still prepares and checks the
exact checkout. `LeaseDuration`, `CheckBinding` and `CheckResults` preserve the
existing lease bound, retained-reference checks and ordered result validation.

The host keeps the private claim, process identity registration, bounded output,
capacity reservation, stop supervision and process journals. It rechecks current
source, candidate and authority before atomically confirming the receipt, set
and invocation. Execution alone cannot establish closure. The ports expose no
mutable store or context retained as state.

The package imports only shared `workflow/state` protocol values within the
module. Remaining closeout policy extraction and final Store responsibility
dispositions are active Slice 15 work.

## Validation

Stage boundary tests cover partial failures, nonzero exit, changed checkout,
explicit waiver and late cancellation. Existing real Git, filesystem, process,
stop and journal-recovery tests exercise the host binding. The production
dependency check rejects coordinator, mutation and configuration imports.
