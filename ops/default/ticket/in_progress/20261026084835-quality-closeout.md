---
id: TKT-20261026084835
title: "Quality slice 15: Workflow closeout and coordinator cleanup"
status: in_progress
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-15
source: review
reported_at: 2026-10-26T08:48:35Z
commits: []
started_at: 2026-10-28T08:07:25Z
branch: refactor/slice-15-quality-closeout
ready_at: 2026-10-26T08:57:09Z
---

# Quality Slice 15: Workflow closeout and coordinator cleanup

Delivery set: code-quality-remediation
Plan: [Single-cycle plan](../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../tracker/code-quality-remediation.md)
Spec: [Requirements](../../spec/code-quality-remediation.md)
Model: [Compatibility](../../spec/code-quality-remediation-model.md)
Source: [Review findings](../../report/code-quality-review-2026-10-26/report.md)

## Problem and evidence

This review unit owns R1, R2, R3, R7, R10 as scoped in the plan, including every assigned declaration in [the baseline inventory](../../plan/code-quality-remediation-inventory.csv). Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It is part of one 16-slice delivery set, not a separate runtime work invocation.

## Work

1. **T15.1:** Extract closing-gate and closure behavior into internal/core/workflow/closeout; leave core/workflow responsible for sequencing, dispatch and lifecycle. Preserve confirmation, closing-gate waiver, exact-candidate and terminal outcome distinctions.
2. **T15.2:** Complete result/input conversions for remaining workflow lifecycle, loose queue, frontier, continuation, amendment, process and capacity paths. Remove workflowClosingInvocationFiles unused context and prepareWorkflowTask/withoutWorkflowPublicationProfile production-only wrappers where still applicable.
3. **T15.3:** Drain policy methods from the persistence owner into their proper stages without moving atomic state transitions or locks to callers. Inspect the final store API method by method for cohesive persistence ownership; do not close merely because 909 methods moved to another directory.
4. **T15.4:** Remove all temporary core/workflow type aliases, pass-through facades and migration-only interfaces introduced by the set. Check that core contains only the shared responsibilities recorded in the dependency ledger.
5. **T15.5:** Finish the naming/compatibility ledger: define frontier, adoption, loose selection, continuation, amendment and closing from governing specs, then make only internal unambiguous renames. Resolve each legacy path and control flag as a supported contract, safe removal or evidence-backed retained case within this cycle.
6. **T15.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

## Acceptance criteria

1. Workflow policy is split by stage, persistence owns only durable operations, migration bridges are gone, and all assigned lifecycle and compatibility items have an evidence-backed disposition.
2. Every assigned baseline entry has a precise fixed or retained-contract disposition, linked to task commit and actual validation.
3. Existing authoring/execution/approval/recovery and durable/wire/hash contracts are preserved as required by the specification and model.
4. The slice compiles and passes its focused selection, make verify and git diff --check; its report and tracker identify the actual evidence boundary.

## Dependencies and delivery

Predecessor: verified integration of Slice 14 into dev.
Planned branch: `refactor/slice-15-quality-closeout`.
Expected PR title: `refactor(slice-15): complete bounded workflow stages and remove bridges`.
Expected report: `ops/default/report/slices/code-quality-remediation/slice-15-closeout.md`.

The tracker owns task execution; this ticket mirrors verified slice state. The maintainer approved this delivery set on 2026-10-26; this ticket is ready subject to its recorded predecessor. Use the normal in_progress/reviewing/solved transitions and record real commits and PR only when they exist. No independent ticket branch or extra PR is added beside the slice PR.

## Validation selection

Existing continuation/closing/process/budget/stop recovery tests, one shared guided parity journey, focused race for lifecycle changes; make verify.

All commands above are planned. None ran as implementation validation during this planning task.

Predecessor PR #451 is verified merged into dev at `8b5659a71af5d393919593c5e0d8c9a6ec0f497c`. Slice 15 is active. Its source ledger freezes 40 assigned signature rows and current declaration identities before closing/lifecycle conversion. No runtime extraction or Store policy disposition is claimed yet.
