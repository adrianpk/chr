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
commits: [e63688b43071, a2e76eec801e, 2f9e9717c027, 0e1d949a3fca, f7df89655484, dbdb1ac183be, 1d108d03c5c2, d577def590fa, ae88f240639d, e2d01335ce70, 129fb1b9d256, 678d9e6abaf4, bba7d53e5f34, 712e812ac131, 571c431a25d8, d647834c037d, 701bbd85d763, 65353580359a, 34d663404459, 0adb83936123, ea96e86c861a, d8b7a9921cb5, eac5ca5dd8e1, 115897594fc5, 969eba2252d7, 051c19eac8fb, bf04bf96c5d6, 7d8acfadb0ff, 1671d3ea3af1, 89898ddd5bd8, fcd5f288a097, 2527d57f8339, 8490f7aeb303, 4e8ed9c2792b, b263991b7f52, 769e2047686c, 23b7fe8d1a9b, 5c533bfdb8cf, 64376e9f69c1, ce0f625d1ca1, 59c5e103482c, 63a843aa48a6]
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

T9.1 through T9.5 remain active. All six duplicate comparisons and shared
model-request mechanics are recorded; plan and acceptance responsibilities are
split. Nine pure native control/evidence values now have a shared state owner,
with original guided/session/request bytes frozen before relocation. Retained
admission/selection and harness identity values also have shared owners.
StartInput separates the durable request from host-only callbacks and commands;
guided/native admission and recovery preserve the frozen wire bytes. The move
ledger retains 273 declaration/extraction/transition preimages.
The session value graph, enum methods and original session/guided validation
belong to core/authoring. Shared default-user paths belong to workspace. Three
read-only native observation ports retain the original store, scope/authority
guards, quiescence, expected absence and verified result references/time.
The exact legacy frontmatter codec, native plan binding values, projection/path
and metadata primitives and bounded Git transport have shared owners. Native
admission/preview/materialization/result proof operations are separate from
authoring transitions and read sessions under the original active owner claim.
Native plan resume, preparation, credentials and signing remain with their owner.
Seven context/host dependency files remain before the complete repository move.
All 89 assigned signature occurrences have exact dispositions: 60 tuple fixes,
three parameter fixes (two already delivered by Slice 4) and 26 cohesive
five-parameter retentions. Focused real-boundary checks and pinned strict lint
pass; exact debt is 237 with ceiling seven and one external exception. The
complete repository move and remaining context/host boundaries remain unfinished. The
report is drafting; no PR is open and final controller validation is pending.
