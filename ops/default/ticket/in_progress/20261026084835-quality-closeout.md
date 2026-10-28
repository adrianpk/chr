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
commits: [5f506a41dee1e172f8915f5712bbc9d6292a7bea, e7255f915be24ad6e3fecda8d9ed5a76ecc35f8c, 93708ffd89d4eee64d111ad71424a5fc3fcfed19, 66222b13904fd86cba76f67e69eafa922fe48a28, 72d968844ee6ee3d4915cb7f2e200a3f049487d1, 38b7f97593a2df3d2af6c465c809dca4bb4c1345, 663c49213963ec459a32015b8feedbda32271f2e, d4844af16cce52e3a44ad9bba5e2a59e80adb520, c111c9df77898601331cd83c6df4932a9c2add8b, c3d57b9927bb8d739808fea8d5db2809c5915398, 5f58bc1cf3f22c0553035b9935d66530cb4762dd, 05d68ed95bc5b154df6be54f0279bc18856eab96, c737a56d85e5fb0532d20b18a3c56c9d86a21fb1]
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

Predecessor PR #451 is verified merged into dev at `8b5659a71af5d393919593c5e0d8c9a6ec0f497c`. Slice 15 is active. Its source ledger freezes 40 assigned signature rows and current declaration identities before closing/lifecycle conversion. Closeout execution/pure policy and all 40 assigned signature dispositions are committed: 26 implemented and 14 retained cohesive five-input contracts. The Slice 15 report records real closing, capacity/process, continuation, control/receipt, amendment/frontier, upgrade and focused race checks plus strict lint. Production signatures have at most five parameters and zero exact debts. Remaining policy ownership, final bridge/compatibility work and slice acceptance remain active.

Closing history extraction now gives attempt validation, receipt binding, retry
budgets and ordered evidence checks a real closeout owner through a read-only
port. Thirteen mechanical relocations match after owner/name qualification;
shared process fields and enum values remain exact. Seven Store methods have
stage/adapter owners and 901 remain; final method ownership is still active.
The report records 30 real closing/recovery/budget/process/capacity tests and
five focused races, plus wire, short and strict checks.

The vocabulary ledger now defines the six governing workflow terms and disposes
ten current compatibility/control cases with concrete consumers, existing
regression references and retirement conditions. Links, case rows and named
regression declarations are checked. Remaining legacy schema/path/control
references stay open under T15.5; no exhaustive compatibility acceptance is
claimed.

Native record/task/receipt/snapshot/proof values now have one shared state owner,
while record/snapshot validation stays explicit in the host. Closing input
assembly and actual-member selection execute in closeout through finite
InputSource observations. Nine direct and five normalized validation moves
match; 31 real tests, five races, affected broad short tests, wire and strict
checks pass. Original 908 Store source rows retain current references to 899
Store methods and nine stage/adapter owners. Slice-proof policy, final method
ownership, compatibility and slice acceptance remain active.

Closing slice/proof/member/canonical/hash policy now has closeout owners through
ProofSource. Immutable receipt/task/target bindings, the ordered task digest and
observation-required identity are shared directly without aliases. Twelve
normalized shared/pure declarations match; 42 real regressions and six races pass
without skips or leaked owned descendants, along with short/style/strict checks.
Original 908 method rows now map to 890 Store methods and 18 stage/adapter owners;
20 actual ports preserve host native guards and atomic confirmation. A concrete
legacy receipt/proof disposition raises the compatibility ledger to eleven
cases. Remaining final method/bridge/compatibility and exact-head slice acceptance
are still active.

Public invocation checkpoints and their complete retained-data dependency graph
now share exact state definitions. Closing readiness, frontier/projection,
retained closure/evidence reads and budget enumeration have closeout owners.
RunGuard retains ordered live-process checks and final source/input fences;
projection preserves the live target alias and immutable history value. Forty
normalized declarations match; 34 workflow tests, one projection test, seven
races, affected broad short tests and strict lint pass. The original 908 method
rows resolve to 885 Store methods and 23 stage/adapter owners through 21 actual
ports. Closing status aggregation, final Store/bridge/compatibility work and
exact-head slice acceptance remain open.

Closing status aggregation now executes in closeout through direct Proofs and
finite StatusSource observations. Receipt-first checks, duplicate publication
blockers, live capacity, retained gate proof and partial errors retain their
order. Exact capacity/role records have shared state owners without aliases.
Thirty-six real workflow tests, six races, affected short tests and strict lint
pass; three declaration comparisons match. The method map records 25 specific
durable/lock retention decisions and this stage policy decision. Its remaining
882 source rows await final inspection; current Store has 884 methods and 22
actual ports. Final API/bridge/compatibility work and slice acceptance remain
active.
