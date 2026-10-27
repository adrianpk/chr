---
id: TKT-20261026084832
title: "Quality slice 12: Workflow state, store and coordinator package"
status: in_progress
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-12
source: review
reported_at: 2026-10-26T08:48:32Z
commits: [6dd94097895b, 5002e793644e, f4be8551581dbec43746e5b676b714911af24e2c, a96c05cee2000654f23e2ce0b0347b1155ec17ca]
started_at: 2026-10-27T23:03:10Z
branch: refactor/slice-12-quality-workflow
ready_at: 2026-10-26T08:57:09Z
---

# Quality Slice 12: Workflow state, store and coordinator package

Delivery set: code-quality-remediation
Plan: [Single-cycle plan](../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../tracker/code-quality-remediation.md)
Spec: [Requirements](../../spec/code-quality-remediation.md)
Model: [Compatibility](../../spec/code-quality-remediation-model.md)
Source: [Review findings](../../report/code-quality-review-2026-10-26/report.md)

## Problem and evidence

This review unit owns R1, R2, R3, R5, R6, R10 as scoped in the plan, including every assigned declaration in [the baseline inventory](../../plan/code-quality-remediation-inventory.csv). Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It is part of one 16-slice delivery set, not a separate runtime work invocation.

## Work

1. **T12.1:** Freeze an explicit map of all current WorkflowStore methods: durable read/mutation/transaction, coordinator, integration, merge, forge/review or closeout. Retain one persistence/locking authority and identify every transaction and recovery boundary.
2. **T12.2:** Move the workflow family and its actual callers to internal/core/workflow as a compiling mechanical relocation. Put shared workflow values under internal/core/workflow/state only where stages must share them without importing their coordinator. Never split methods of a nonlocal Go type across packages.
3. **T12.3:** Separate storage and mutation from workflow policy. Define consumer-owned, cohesive stage ports implemented by the store/coordinator; move logic to stage-owned methods/functions in subsequent slices. Keep lock acquisition, journaling and atomic multi-record writes at their current owner.
4. **T12.4:** Replace the assigned store/coordinator tuple results and repeated inputs with named results and typed modes/states. Deduplicate receipt writing D08 only after confirming operation receipt semantics match.
5. **T12.5:** Switch app, authoring, ticket and adapters to the new workflow boundary. Any temporary compile-time type bridge must be enumerated with an owning removal task by Slice 15; no new exported internals solely to bypass package privacy.
6. **T12.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

## Acceptance criteria

1. Workflow files have a real package owner and compatible state encodings; stage interfaces do not depend on the coordinator; one store owns transactions and recovery.
2. Every assigned baseline entry has a precise fixed or retained-contract disposition, linked to task commit and actual validation.
3. Existing authoring/execution/approval/recovery and durable/wire/hash contracts are preserved as required by the specification and model.
4. The slice compiles and passes its focused selection, make verify and git diff --check; its report and tracker identify the actual evidence boundary.

## Dependencies and delivery

Predecessor: verified integration of Slice 11 into dev.
Planned branch: `refactor/slice-12-quality-workflow`.
Expected PR title: `refactor(slice-12): establish workflow state and persistence boundaries`.
Expected report: `ops/default/report/slices/code-quality-remediation/slice-12-workflow.md`.

The tracker owns task execution; this ticket mirrors verified slice state. The maintainer approved this delivery set on 2026-10-26; this ticket is ready subject to its recorded predecessor. Use the normal in_progress/reviewing/solved transitions and record real commits and PR only when they exist. No independent ticket branch or extra PR is added beside the slice PR.

## Validation selection

Existing WorkflowRoundTrip, mutation/receipt/idempotency/materialization and process recovery regressions on owned filesystem/Git; make verify.

All commands above are planned. None ran as implementation validation during this planning task.

Slice 11 PR #448 is verified integrated at `d0d65c114cf8ddffcd0492f7fc0b746a832adc0d`.
Slice 12 is active; T12.1 method/transaction/recovery mapping precedes relocation.


T12.1 freezes all 908 store methods across 234 files with exact identities/hashes,
100 concrete IO/authority/recovery entries and 79 direct root dependencies. The
[authority map](../../plan/code-quality-remediation-slice-12-boundaries.md)
keeps one receiver/transaction owner and records future policy consumers.
All 71 assigned baseline rows map to 68 current declarations. Compiler-resolved
mapping and exact coverage/hash checks pass; T12.2 boundary resolution is active.
No runtime move or final candidate validation is represented as completed.


The first T12.2 prerequisite moves twelve shared review/cycle value and pure
validator declarations to workflow/state, with direct actual consumers and no
root aliases. Exact token comparison and focused value/wire/storage/native Git/
review/app regressions pass; strict lint is clean and the signature policy stays
at 9306 declarations, ceiling six, 162 exact debts and one external exception.
Ten frozen direct dependencies resolve; 69 root couplings remain before family
relocation. Stage ports, assigned signatures and final validation stay open.


The next T12.2 prerequisite moves ticket queries and Board/document projections
to their existing ticket owner with actual native, app and adapter consumers.
The projection implementation is `a96c05cee2000654f23e2ce0b0347b1155ec17ca`.
Forty-two moved declaration token sequences match after qualification/naming and
cmp.Compare substitution; one local generic integer comparator is removed.
Public query tests colocate with their owner and share its existing SQLite
fixtures. Twelve more direct dependencies resolve, leaving 57 root couplings;
all 908 store methods remain together at their original receiver. Native writes,
locks and journals retain their authority. Focused ticket, native filesystem/Git,
wire/storage, app and boundary checks, strict lint and the ceiling-six policy pass.


Configuration has one shared workspace/config owner for the complete parsed
image, raw revision/explicit presence, defaults, upgrade fences and guarded
harness-selection journal/image/receipt recovery. All 114 moved declarations
retain exact tokens after qualification/API naming and the original literal
rebase default. Config.Explicit exposes presence without a mutable-map escape.
Fifteen more direct dependencies resolve, leaving 42; all 908 store methods and
native workflow upgrade transactions remain at their original receiver. The
existing Slice 15 upgrade-journal tuple now qualifies the same Config value;
no new signature debt or relaxed check is introduced.
