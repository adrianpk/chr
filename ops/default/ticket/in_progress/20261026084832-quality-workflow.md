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
commits: [6dd94097895b, 5002e793644e, f4be8551581dbec43746e5b676b714911af24e2c, a96c05cee2000654f23e2ce0b0347b1155ec17ca, bedb07a4cca4d39c5f7099d594a83c6650002398, 3fb0f46aff5421452d6b602fe91e4e483e93dfbc, 20ca6c3e508b7c0964baaef4f345c87199e3d81b, 9c74db45a790d03f7d710b2ffac2e7dcdc4d5e88, fe0df936e6152c0ad346e1e680ad2dd246825e60]
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
T12.1 and T12.2 now have the complete frozen map and compiling workflow owner.
Relocation implementation: `20ca6c3e508b7c0964baaef4f345c87199e3d81b`.
All 908 methods move with `workflow.Store`; all 79 original direct dependencies
and 100 authority/recovery entries resolve to actual current declarations.

The relocation covers 2,710 production declarations and 1,666 test declarations.
All 4,342 declaration groups retain exact tokens after compiler-resolved API and
implicit embedded-field names, owner qualification and the two real guarded legacy
operations. Wire fixture bytes, lock/journal order and recovery guards remain
unchanged. The existing review/state, ticket/metadata and complete configuration
prerequisites retain their direct capability owners.

Actual core/app/adapter consumers and acceptance/fault source bindings follow the
new package. The 78 selected tests and frozen controller gate are unchanged.
Focused wire/storage, native Git/filesystem/recovery, legacy ticket/app and import
boundary checks pass; strict lint is clean. Runner/admission/evidence regressions
and package README checks pass. All seven focused journal/crash-recovery checks
pass through the existing QA fixture-owning runner. The signature policy remains 9,305 declarations,
ceiling six, 162 exact debts and one external exception.

T12.3 implementation: `9c74db45a790d03f7d710b2ffac2e7dcdc4d5e88`.
It connects integrate.Validation, merge.Evidence, review.Decisions,
forge.Responses and closeout.Gate to their actual production operations. Private
bindings retain claim/prepared/provider or journal/closing identities; all 908
Store methods and existing mutation/recovery authority remain local. Sixty shared
protocol types and two evidence enumerators retain exact tokens, wire order/tags
and nil/empty values. Twenty AST cases enforce state-only stage imports. Focused
real stage, stop/lost-response, wire/storage, policy/admission and dependency tests
pass; strict lint is clean with 9,316 production declarations, unchanged ceiling
six, 162 exact debts and one external exception.

T12.4's receipt/materialization batch (`fe0df936e6152c0ad346e1e680ad2dd246825e60`) implements ten assigned declarations (11
baseline rows): named mutation lookup/normalization/preparation and materialization
replay/task results, plus cohesive set/allocation inputs. Actual authoring and
amendment/adoption callers consume those values. D08 shares only contained receipt
read/absence/decode; operation-specific schemas, checksum/order and request guards,
receipt writers and lock/journal authority remain. Typed operation states retain
existing bytes. Thirty focused real IO/recovery/replay/adoption/amendment/authoring
checks and ten receipt corruption/cross-schema cases pass; strict lint is clean.
The current signature policy has 9,317 declarations, unchanged ceiling six,
152 debts and one external exception.

Thirty-seven required declarations, typed mode and cohesion-watch reconciliation,
final caller reconciliation, exact-head controller validation and canonical PR
preparation remain open. The ticket stays in_progress.
