# `workflow/integrate`

## Purpose

This package owns candidate validation execution, exact gate evidence and the
consumer contract for guarded acceptance. Pure Git attribution, replay
correspondence, bounded identity-collision retries and preparation signing
metadata rules live with that policy. Complete history/provenance, prepared-result
and observed atomic/signing capability checks use the same stage owner. Host callers still read/write Git objects
and invoke the certified signer.

## Contract

`ValidateCandidate` runs the bound validation, builds the exact candidate
and requests acceptance through `Validation`. It keeps validation deadlines
separate from operator stops and preserves bounded detached cleanup. Gate checks
compare the exact ordered commands, exit/truncation outcomes or explicit waiver.
Source observations omit prepared-only evidence.

`Store.PrepareWorkflowSliceIntegration` retains claim admission, execution ownership,
lease/stop contexts and capacity cleanup. The bound acceptance operation rechecks
the claim, prepared input and current source before the existing atomic transition.

`RepairProof` receives immutable integration proof data. A stateless native
adapter composes merge tree/regeneration and replay-history checks in their
original order; integrate imports no sibling stage.

The contract imports shared `workflow/state` protocol values. The workflow owner
implements it with the existing operations and retains claims, reservations,
source/revision guards, repository locks, journals, atomic writes and recovery.
The port exposes the required operation without exposing a mutable store.

## Validation

Existing workflow integration tests exercise this binding through real Git,
filesystem and process fixtures. The production dependency check permits shared
state and contenthash for existing evidence checksums. It rejects coordinator,
filesystem mutation and configuration imports.
