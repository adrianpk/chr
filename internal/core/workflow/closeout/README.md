# `workflow/closeout`

## Purpose

This package owns closing attempt, history, retry-budget, receipt, lease and
result policy, plus ordered gate execution against one frozen closing input.

## Contract

`Execute` builds the unconfirmed receipt, preserves the explicit waiver, retains
each command's partial result before checking failure and checks the final
checkout. It receives the host's watched deadline context through narrow
`Workspace` and `Commands` ports. A waived gate still prepares and checks the
exact checkout. `LeaseDuration`, `CheckBinding` and `CheckResults` preserve the
existing lease bound, retained-reference checks and ordered result validation.

`Attempt` retains the original exact journal fields. `ReadAttempt` and
`ReadReceipt` preserve hash verification, partial decode results, schema and
confirmation binding. `CheckHistory` checks attempt order, per-budget counts and
result/input/process evidence. `CheckBudget` keeps the invocation or standalone
reopening budget. The read-only `Evidence` port delegates contained bounded reads
and original prerequisite hash verification to the host. Failed or interrupted
attempts remain distinct from confirmed completion; the stage never confirms one.

The host keeps the private claim, process identity registration, bounded output,
capacity reservation, stop supervision and process journals. It rechecks current
source, candidate and authority before atomically confirming the receipt, set
and invocation. Execution alone cannot establish closure. The ports expose no
mutable store or context retained as state.

The package imports shared `workflow/state` protocol values and `contenthash`
within the module. Remaining closeout policy extraction and final Store
responsibility dispositions are active Slice 15 work.

`Request` owns the unchanged observed-set/revision/source/target input. The host
continues to enforce admission and authority. `BuildInput` owns contract hashing,
ordered frozen input assembly and partial results. Its `InputSource` supplies
bounded configuration, target/tree, namespace, slice-proof and hash observations.
`Members` selects all actual integrated members without active attempts and
returns their deterministic key order. Empty membership remains an empty slice;
selection failure returns nil. The stage retains the same shared evidence map
through ordered slice observations and never reorders the source descriptors.
The host keeps record/snapshot validation and final current-image/member guards.

## Validation

Stage boundary tests cover partial failures, nonzero exit, changed checkout,
explicit waiver and late cancellation. Existing real Git, filesystem, process,
stop and journal-recovery tests exercise the host binding. The production
dependency check rejects coordinator, mutation and configuration imports.
