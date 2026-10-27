# Slice 10: Harness contracts and runtime lifecycle

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-10-quality-harness`
PR: pending

## Purpose

Separate adapter-neutral execution from project admission, binding and host
lifecycle policy while preserving frozen identities and bounded child resources.

## Delivered Behavior

Slice 9 integration is verified. Slice 10 implementation is in progress.

## Implementation Notes

The approved inventory assigns 38 signature occurrences and 30 current root
files to this slice. Each current owner and consumer is checked before moving.

## Contracts Added or Changed

None yet.

## Files of Interest

- `internal/core/harness/`: existing binding values and planned project policy.
- `internal/harness/`: existing adapter-neutral execution and process owners.

## Validation

No Slice 10 implementation checks have run. Slice 9 closure uses its existing
three-command controller evidence at `0ce066b1599b55f6efc3f151bb960a74b75d5e91`.

## Risks and Follow-ups

T10.1-T10.6 remain in progress or pending; PR readiness is not established.
