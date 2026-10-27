# Slice 9: Authoring, document generation and guided preparation

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-9-quality-authoring`
PR: pending

## Purpose

Give authoring sessions, generated documents and guided preparation cohesive
owners while preserving artifact-specific approval, source and retry contracts.

## Delivered Behavior

The shared operational artifact layout and envelope behavior now belongs to
core/ops. Authoring, workflow and application callers import that owner directly.
Namespace precedence, canonical paths, ID allocation, metadata and private-file
exclusion retain their original behavior. Workflow keeps allocator reclamation,
owner proofs, guarded release and durable transaction ownership.

This is the initial T9.1 prerequisite. The complete authoring move, six duplicate
comparisons, generation sharing, signature conversions and guided ports remain
in progress or pending. No Slice 9 completion or PR readiness is claimed.

## Implementation Notes

The initial declaration scan of 101 assigned root authoring files identifies 78
shared names outside the source set. Resolve each at its genuine shared owner or
consumer-owned operation port before moving session code. Root core must not
become an imported coordinator dependency of authoring.

Four ops production files and their two focused test files move with 41 original
declaration hashes retained in the move ledger. The current dependency map and
file-owner inventory include this prerequisite. Module import checks prevent
ops from importing authoring, root core or workflow policy.

## Contracts Added or Changed

Operational envelope creation, metadata construction, slug selection, allocation
lock creation and number observation have direct shared consumers. They preserve
the existing operations rather than adding forwarding facades. The numeric
path-prefix predicate retains the old regex rule and does not tighten supported
artifact IDs. The distinct ASCII artifact slug and Unicode quiz slug rules stay
separate. Legacy ticket writers resolve an explicit path-owned/default namespace
without consulting OPS_USER.

Two debt signatures qualify their unchanged layout type through its new owner;
their parameter and result meanings are unchanged. Debt remains 301 with ceiling
seven and the same one external-interface exception.

## Files of Interest

- `internal/core/ops/` owns layout, paths, shared envelopes and allocation primitives.
- [Move evidence](../../../plan/code-quality-remediation-slice-9-moves.csv).
- [Signature reassociations](../../../plan/code-quality-remediation-slice-9-signatures.csv).
- [Current owners](../../../plan/code-quality-remediation-core-owners.csv) and
  [dependency ledger](../../../plan/code-quality-remediation-boundaries.md).

## Validation

- `go test ./internal/core/ops ./internal/core ./internal/app ./scripts/goquality -run '^Test(ResolveOps|OpsLayout|CreateOps|RefreshOps|LeafDependencies|PlanAuthoringArtifact|DumpMutation|PlanSources|SpecMutation|WorkflowAllocation|WorkflowMaterialization|WorkflowLegacy)' -count=1 -timeout=3m`
  — passed for the envelope prerequisite; the app package compiled with no matching tests.
- `make check-go-style` — passed, 301 exact debts and one external exception.
- `make check-package-readmes` — passed.

The first compile identified shared allocator calls and a legacy layout literal;
those consumers now use the owned ops contracts. The first style scan identified
two unchanged layout-type signatures; their exact reassociations are recorded.
The first strict lint identified import formatting, the legacy layout branch
complexity and old-name stutters at the mechanical move. Formatting is corrected
and the namespace resolution has one small owner. The next separate naming
commit removes stutters without a lint suppression. Final clean-head controller
validation has not run.

## Risks and Follow-ups

Complete the remaining approved Slice 9 tasks before preparing its PR. No
external fixture or live model account is required by the recorded selection.
