---
id: TKT-20261026084833
title: "Quality slice 13: Candidate validation and merge stages"
status: in_progress
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-13
source: review
reported_at: 2026-10-26T08:48:33Z
commits: [290ebfc8f299cf5ab2a6ef28174b44d2b34b18bb, 05a7a106029d64ee5aa1ba8b62703afc40437d72, c7a15133e0a07f8eb3fec735e36c967730aff038, 962e03aef96bf7aefd2b075fe218041b9cd00dc6, 8e3630b9b074ec7b15d7b0de252429c3844dd283, af02c72f04ec8fb1c8554f7fa777a2e18958dacc, a4fa20fe7344c49dbc164e726dcac9883b574c32, 7f2d18b1a04821084af93f0c825d3f3376c791cf, d9bfaa918823edcdeb698e7fff13139663ce9d63, 77b24972b7361457978889a8f7f69476fae8d860, 40ab73aaaed653ee7606801c09f9ad43b2e933c1, 3b9f43242fb4998a4c98bb1444f45334de43594d, cdf69b8822f8e498cbff8914d7d93e68cd78ff15, 1e2974b5fc5e72dc50df6f2f71583aebf2ebb857, 53304c24865b33288ac0d99d423c496db6486a55, 5ceca5a3389da5ba87c0f87d22f05fd8146a5309, 8d11ee754ff39e18d032ad6de084efd1c51509ff]
started_at: 2026-10-28T02:15:11Z
branch: refactor/slice-13-quality-integration
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

The report records the exact successful task commands and final controller validation contract.

Slice 12 PR #449 is verified integrated at
`a49c32c9f5e81e1bde9f585a25c6d5cc669d7a9e`. Candidate proof/capability and merge/repair policy now use their independent stage owners. All 87 assigned signature rows and every stored-state/mode/suppression entry are reconciled. The 146 extraction images, 908 Store methods, 100 authority boundaries and 79 dependency references resolve with frozen preimages. Each stage imports only state/contenthash and receives no mutable Store or claim resource. Real proof, repair, state, rejection, cancellation, recovery/lost-ack and wire checks pass; exact commands and prior focused races are in the report. The report is reviewing. Canonical PR preparation and final controller evidence complete T13.6 before integration readiness; the controller owns the merge.
