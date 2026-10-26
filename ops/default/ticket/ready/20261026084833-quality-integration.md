---
id: TKT-20261026084833
title: "Quality slice 13: Candidate validation and merge stages"
status: ready
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-13
source: review
reported_at: 2026-10-26T08:48:33Z
commits: []
ready_at: 2026-10-26T08:57:09Z
---

# Quality Slice 13: Candidate validation and merge stages

Delivery set: code-quality-remediation
Plan: [Single-cycle plan](../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../tracker/code-quality-remediation.md)
Spec: [Requirements](../../spec/code-quality-remediation.md)
Model: [Compatibility](../../spec/code-quality-remediation-model.md)
Source: [Review findings](../../report/code-quality-review-2026-10-26/report.md)

## Problem and evidence

This review unit owns R1, R2, R3, R10 as scoped in the plan, including every assigned declaration in [the baseline inventory](../../plan/code-quality-remediation-inventory.csv). Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It is part of one 16-slice delivery set, not a separate runtime work invocation.

## Work

1. **T13.1:** Extract candidate preparation/validation into internal/core/workflow/integrate through the ports established in Slice 12. Preserve exact source/target observation, guard checks, validation claims and atomic acceptance.
2. **T13.2:** Extract merge and repair policy into internal/core/workflow/merge. Convert executeWorkflowMergerDriver, preparation claim/result tuples and repeated claim/prepared/tool/process inputs into cohesive execution values; contexts stay explicit.
3. **T13.3:** Retain store-owned transactions, claim generations, capacity/budget reservations, process identity and exact validation evidence. Stage code must not import core/workflow or obtain a mutable store through an untyped escape hatch.
4. **T13.4:** Preserve stale input rejection, already-applied versus absent outcomes, supervised child cleanup, interrupted preparation and lost acknowledgement reconciliation. Finish all assigned signatures, typed modes, flags and error/suppression entries.
5. **T13.5:** Move each package in its own mechanical task commit, then rename to package-qualified verbs/nouns with tests unchanged. Record concrete import and responsibility reductions rather than only file-count changes.
6. **T13.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

## Acceptance criteria

1. Validation and merge logic are independently owned stages with narrow ports; the same receipts, budgets and recovery invariants still decide integration.
2. Every assigned baseline entry has a precise fixed or retained-contract disposition, linked to task commit and actual validation.
3. Existing authoring/execution/approval/recovery and durable/wire/hash contracts are preserved as required by the specification and model.
4. The slice compiles and passes its focused selection, make verify and git diff --check; its report and tracker identify the actual evidence boundary.

## Dependencies and delivery

Predecessor: verified integration of Slice 12 into dev.
Planned branch: `refactor/slice-13-quality-integration`.
Expected PR title: `refactor(slice-13): extract candidate validation and merge policy`.
Expected report: `ops/default/report/slices/code-quality-remediation/slice-13-integration.md`.

The tracker owns task execution; this ticket mirrors verified slice state. The maintainer approved this delivery set on 2026-10-26; this ticket is ready subject to its recorded predecessor. Use the normal in_progress/reviewing/solved transitions and record real commits and PR only when they exist. No independent ticket branch or extra PR is added beside the slice PR.

## Validation selection

Existing preparation/admission/merger repair/cancellation and lost-ack fixtures with real Git and child processes; focused race for changed claim ownership; make verify.

All commands above are planned. None ran as implementation validation during this planning task.
