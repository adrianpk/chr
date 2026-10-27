---
id: TKT-20261026084829
title: "Quality slice 9: Authoring, document generation and guided preparation"
status: in_progress
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-9
source: review
reported_at: 2026-10-26T08:48:29Z
commits: [e63688b43071]
branch: refactor/slice-9-quality-authoring
started_at: 2026-10-27T13:06:46Z
ready_at: 2026-10-26T08:57:09Z
---

# Quality Slice 9: Authoring, document generation and guided preparation

Delivery set: code-quality-remediation
Plan: [Single-cycle plan](../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../tracker/code-quality-remediation.md)
Spec: [Requirements](../../spec/code-quality-remediation.md)
Model: [Compatibility](../../spec/code-quality-remediation-model.md)
Source: [Review findings](../../report/code-quality-review-2026-10-26/report.md)

## Problem and evidence

This review unit owns R1, R2, R3, R4, R6, R10 as scoped in the plan, including every assigned declaration in [the baseline inventory](../../plan/code-quality-remediation-inventory.csv). Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It is part of one 16-slice delivery set, not a separate runtime work invocation.

## Work

1. **T9.1:** Move authoring session and dump/spec/plan artifact behavior into owned internal/core/authoring boundaries; move plan/catalog behavior into its own capability only where the dependency ledger demonstrates a distinct responsibility. Keep guided inputs and committed-source provenance intact.
2. **T9.2:** Read the governing spec/plan for both sides of D02, D04, D05, D07, D09 and D10. Extract catalog, generation, guided-preflight, authoring-step and dump-update logic only where both copies implement the same rule; record intentional semantic differences with precise evidence.
3. **T9.3:** Share the actually common context/model/validation/recording steps of spec, acceptance and plan generation. Keep artifact-specific schemas, approval, source observation, publication and idempotency decisions explicit; do not introduce a generic generation framework.
4. **T9.4:** Split service_plan.go and authoring helpers by responsibility, finish assigned context/dump/spec/plan/guided result and parameter shapes, and handle the acceptance_feature.go large-file disposition.
5. **T9.5:** Use ports to the existing workflow owner for guided preparation instead of a second coordinator. Preserve manual authoring, legacy specification formats, Git checkpoints, exact hashes and retry receipts; update existing docs in their native format.
6. **T9.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

## Acceptance criteria

1. Duplicate authoring rules have one owner while genuinely distinct generation contracts remain separate; manual and guided paths produce compatible artifacts and receipts.
2. Every assigned baseline entry has a precise fixed or retained-contract disposition, linked to task commit and actual validation.
3. Existing authoring/execution/approval/recovery and durable/wire/hash contracts are preserved as required by the specification and model.
4. The slice compiles and passes its focused selection, make verify and git diff --check; its report and tracker identify the actual evidence boundary.

## Dependencies and delivery

Predecessor: verified integration of Slice 8 into dev.
Planned branch: `refactor/slice-9-quality-authoring`.
Expected PR title: `refactor(slice-9): share authoring rules without merging distinct contracts`.
Expected report: `ops/default/report/slices/code-quality-remediation/slice-9-authoring.md`.

The tracker owns task execution; this ticket mirrors verified slice state. The maintainer approved this delivery set on 2026-10-26; this ticket is ready subject to its recorded predecessor. Use the normal in_progress/reviewing/solved transitions and record real commits and PR only when they exist. No independent ticket branch or extra PR is added beside the slice PR.

## Validation selection

Existing generation/source-revision/authoring publication recovery and guided preparation tests with real filesystem/Git; make verify.

T9.1 is active. The shared ops envelope prerequisite and focused validation are recorded in the Slice 9 report. Remaining tasks are pending; no PR is open.
