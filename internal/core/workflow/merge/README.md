# `workflow/merge`

## Purpose

This package owns the bounded merger broker, protocol validation, repair scope
and contained model file operations. Its evidence port verifies the retained
input through the workflow owner.

## Contract

`ServeWorkflowMergerBroker` accepts bounded UTF-8 JSON messages and one terminal
proposal. Duplicate/unknown keys, excessive depth, trailing output, extra terminal
messages and byte/message budget violations remain errors. The protocol and tool
request/response fields retain their original wire encodings.

`NewWorkflowMergerTools` binds the exact input, an already opened owned checkout
and its claim guard. Every operation rechecks that guard. Path policy excludes
control records, protected paths, aliases, symlinks and multiply linked files.
Model tools can read/write/delete within the granted scope; host validation runs
through its existing supervised path after diff acceptance.

The workflow owner retains root closure, process supervision, admission, claims,
capacity/budget reservations, immutable evidence reads and guarded acceptance.
`Evidence` verifies those retained inputs without exposing a mutable store. This
package imports shared state only and cannot import its coordinator or mutation
and configuration authorities.

## Validation

Colocated tools tests retain the actual filesystem escape/alias assertions.
Workflow integration tests exercise the certified child driver, retained evidence,
proposal parsing, stop and descendant cleanup through real Git/process fixtures.
The source checker enforces the shared-state-only module dependency bound.
