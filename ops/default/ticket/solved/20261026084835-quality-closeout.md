---
id: TKT-20261026084835
title: "Quality slice 15: Workflow closeout and coordinator cleanup"
status: solved
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-15
source: review
reported_at: 2026-10-26T08:48:35Z
commits: [30743d17b45afc3656cc08aac01e1b8950422984, 6f7184ab351c8548d43b0ee25c30d3a23e1cd8b5, f3fc48d21dae09efb4c329d4e985546b6bb3739a, 178e8fcdcc0d5091d5e51b92813a7fa8f576f58b, 3e1896c7f42a0a802d78a23ff036adbe3298d8e4, e5fbc60981a76a7b25cf76e69b63ce6260140252, 6afffc97a34bc31483d9f7f8acd4ff045e827bc3, 85256e808df3824ddaf4b1eb230ff4a5ad8ae757, 4d4f74804e857046b099ea3db91e87ec5d1258f6, 6e281666ac251b7f6cbdf997122dee845fff87bf, 12eb09162d1a2dd2256bd5197b28e9be70c66312, 72253efe04a5f0ca41a2393b442f89c98aac8637, 0ce09547cb796604fb01ee41851a1e9a709cf02e, b5b75ade5cbf21eb68ab3521b26545a20c5ee970, 3ed65239b85570183865b74e169bcbf5a62c1819, 5f506a41dee1e172f8915f5712bbc9d6292a7bea, e7255f915be24ad6e3fecda8d9ed5a76ecc35f8c, 93708ffd89d4eee64d111ad71424a5fc3fcfed19, 66222b13904fd86cba76f67e69eafa922fe48a28, 72d968844ee6ee3d4915cb7f2e200a3f049487d1, 38b7f97593a2df3d2af6c465c809dca4bb4c1345, 663c49213963ec459a32015b8feedbda32271f2e, d4844af16cce52e3a44ad9bba5e2a59e80adb520, c111c9df77898601331cd83c6df4932a9c2add8b, c3d57b9927bb8d739808fea8d5db2809c5915398, 5f58bc1cf3f22c0553035b9935d66530cb4762dd, 05d68ed95bc5b154df6be54f0279bc18856eab96, c737a56d85e5fb0532d20b18a3c56c9d86a21fb1]
started_at: 2026-10-28T08:07:25Z
branch: refactor/slice-15-quality-closeout
pr: https://forge.adrianpk.com/adrianpk/chr/pulls/452
closed_at: 2026-10-28T16:07:20Z
resolution: fixed
reviewed_at: 2026-10-28T15:59:24Z
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

Five unused Store receivers are removed with exact body comparison and six real
callers updated. Four coordinator/image functions and the contained Markdown
reader retain their existing guards, partial outcomes and complete atomic
commits. Thirty-three real fixtures, four races, affected short tests and strict
lint pass. Final source inspection now records 104 decisions: 75 guarded
persistence methods, 23 closing policies, one host evidence adapter, four
coordinator functions and one persistence reader. The remaining 804 source rows
stay pending. The compatibility ledger adds the default-only omission contract
for older immutable upgrade snapshots; publication or capacity authority cannot
be hidden by an absent historical key. Final API/bridge/compatibility work and
slice acceptance remain active.

Closing orchestration now uses actual CloseSet/CloseInvocationSet coordinator
operations and a private four-method lifecycle host. Store retains the same
claim, process-journal, confirmation and recovery authority and critical
sections. Eleven normalized declarations match; 31 workflow fixtures, one
application fixture, six races, affected short tests and strict lint pass.
Source inspection covers all closing methods and tracker replacement guards:
156/908 dispositions are recorded, with 752 Store rows still pending. C15.13
records the supported v1/v2 tracker schema and projection distinction without
implicit upgrade or new execution authority. Final repository-wide API,
compatibility, guided parity and exact-head slice/PR acceptance remain active.

Task/slice/manual execution and stop watching now use actual coordinator and
private lifecycle owners, with 13 Store methods removed and the original guarded
claim/input/replay/candidate/result transactions retained. Twenty-seven production
and 85 test declaration streams match after explicit owner/call substitutions.
All 58 selected real fixtures and 11 races pass without skips/leaks; affected
short tests, signature checks and strict lint pass. Source inspection now records
242/908 specific decisions, with 666 Store rows pending. C15.14 preserves manual
empty-worker handoff, frozen execution and explicit verified reconciliation.
Final API/compatibility, guided parity and exact-head slice/PR work remain open.

Preparation/review/merger/submission execution now uses actual coordinator and
private role-host owners; 17 Store methods are removed without changing original
claim/approval/budget/process/candidate/result/receipt transactions. All 25
production and 135 test streams match after owner/call and exact repair-port
pairing. All 107 selected fixtures pass across the original batch and ten unfinished
submission retries after its retained 15-minute package timeout. Eleven races
and short/signature/strict lint checks pass without test skips or owned leaks. Method assessment covers 331/908 rows with
577 pending. C15.15 preserves proposal-only versus trusted repair acceptance.
At that revision, review-policy/preparation-Git separation, remaining
API/compatibility work, guided parity, exact-head slice validation and PR
preparation remained open.

Review admission/restriction and private preparation Git ownership are separated
in `12eb09162d1a2dd2256bd5197b28e9be70c66312` under the unchanged admission lock and lease. Store retains exact
configuration/evidence reads and complete claim/input/approval/result transactions.
Fifteen comparable production and three test streams match; semantic admission
preserves original trusted decision, errors and trusted-only binding read. All
16 selected real fixtures, four focused races, short/signature/style and strict
lint checks pass without skips or owned leaks. The initial lint whitespace
failure and successful unchanged-command rerun are retained in the report.
The method map records 378/908 body-specific decisions with 530 pending; 841
methods remain on Store. C15.8 names the actual human-only stage restriction.
Final API/compatibility assessment, guided parity, exact-head slice validation
and PR preparation remain open.

Publication advancement, explicit native reconciliation and continuation have
actual coordinator/private-loop owners in `6e281666ac251b7f6cbdf997122dee845fff87bf`. Nineteen Store methods
are removed, including the sole poll-only forwarding method; original guard
rereads, acceptance locks and complete journal/budget/decision/receipt/checkpoint
transactions remain. All 22 production and 38 test declaration comparisons pass
after explicit normalization and exact forwarding consolidation. All 27 selected
real fixtures, six races, short/signature/style and strict lint checks pass with
no test skips or owned leaks. The method map now records 449/908 decisions, with
459 pending and 822 methods remaining on Store. C15.16 preserves observation
versus bounded polling without resending or refilling budget. At that revision,
mixed lifecycle control, remaining API/compatibility work, guided parity,
exact-head slice validation and PR preparation remained open.

Lifecycle control coordination and guarded Store acceptance are separated in
`4d4f74804e857046b099ea3db91e87ec5d1258f6`. The original minute-bounded timeout/lock/cancel block is exact;
three production streams match after explicit recomposition and five changed
test declarations match. Frozen request/digest, partial acceptance, receipt-only
replay and continuation after exclusion retain their original behavior. All 20
selected real fixtures, four races, short/signature/style and strict lint checks
pass without skips or owned leaks. Sixty-five new persistence decisions bring
the method map to 514/908 assessed rows with 394 pending; Store still has 822
methods. C15.17 records lifecycle receipt/replay and post-acceptance execution.
Remaining API/compatibility assessment, guided parity, exact-head slice
validation and PR preparation remain open.

Canonical PR ownership separation is in `85256e808df3824ddaf4b1eb230ff4a5ad8ae757`. The private provider
host retains the unchanged repository lock, Store owns exact candidate/creating
intent and complete observation/binding/task/set transactions, and forge owns
pure canonical PR identity/URL policy with direct consumers. All 16 normalized
or explicitly recomposed production streams match; no tests are edited. All 19
selected real fixtures, four races, short/signature/style and strict lint checks
pass without skips or owned leaks; six initial import-group lint findings were
corrected before the successful lint run. Seventy source-specific PR/order/
adoption/materialization decisions bring the method map to 584/908 assessed rows
with 324 pending; Store has 821 methods. C15.18 records uncertain creation and
canonical identity without replacement. Remaining API/compatibility assessment,
guided parity, exact-head slice validation and PR preparation remain open.

Native observation coordination is separated in `6afffc97a34bc31483d9f7f8acd4ff045e827bc3`. Actual once/
poll/Git coordinators and a private host retain exact deadline/backoff/partial/
detached-cleanup sequencing; Store retains all reconciliation/binding/callback
locks and complete journal/projection/finalization transactions. Eight production
and 23 changed test streams match after owner/call normalization. All 33 selected
real fixtures, five races and the isolated closing race pass; the initial
three-second gate-start wait failure is retained with no test change or
reproduced cause. Short/signature/style and strict lint pass without owned leaks
or skips. One hundred five source-specific decisions bring
the map to 689/908 assessed rows with 219 pending; Store has 816 methods. C15.16
records actual observation owners and C15.19 resolves registration/read control.
Remaining API/compatibility assessment, guided parity, exact-head slice validation
and PR preparation remain open.

Frontier/capacity observation and role selection are separated in `e5fbc60981a76a7b25cf76e69b63ce6260140252`.
Private observers and role selector retain original canonical proof ordering,
partial results, shared process/slot exclusion and stable eligibility. Store
retains all query/admission/frontier locks, exact evidence fences and complete
receipt/entry/invocation transactions. All 16 normalized production streams match, including two verified private-owner
conversions required by strict lint; no tests are edited. All 34 real fixtures, six races, short/signature/style and
strict lint pass without skips or owned leaks. Ninety-nine source-specific
decisions bring the map to 788/908 assessed rows with 120 pending; Store has 807
methods. Merger checkpoint Git seeding remains pending separately. C15.20 resolves
shared capacity/role priority without granting execution. Remaining API/
compatibility assessment, guided parity, exact-head slice validation and PR
preparation remain open.

Merger workspace/checkpoint/broker IO is separated in `3e1896c7f42a0a802d78a23ff036adbe3298d8e4`. The existing
private host owns exact isolated repository setup, four-object checkpoint import
and materialization, and bounded ordered UTF-8 evidence assembly. Store retains
original admission locks, claims, contained evidence validation and complete
atomic writes. All six affected declaration streams match; the sole test edit
changes its direct owner call. All 17 real fixtures, five races, affected short
tests, signature/style and strict lint pass without skips or owned leaks.
Sixty-two individual method decisions bring the map to 850/908 assessed rows
with 58 pending and 804 actual Store methods. C15.15 records these isolation and
checkpoint boundaries. Final API/compatibility assessment, guided parity,
exact-head slice validation and canonical PR preparation remain open.

Construction/lane/preview/scheduler/capacity sequencing is separated in
`178e8fcdcc0d5091d5e51b92813a7fa8f576f58b`. Thirteen procedures use actual private owners under unchanged
Store locks, allocation/current-evidence guards and complete atomic writes. All
30 production streams match after owner/call normalization and a verified sole-
Store-field conversion; no tests change. All 38 real fixtures and seven races
pass without skips or owned leaks. Corrected short/style/strict lint passes;
initial omitted target arguments were restored before fixtures, and later Godoc/
conversion lint corrections are recorded separately. Forty-three new individual
decisions bring the map to 893/908 assessed rows with 15 pending and 791 actual
Store methods. Final API/compatibility assessment, guided parity, exact-head
slice validation and canonical PR preparation remain open.

Final read/provider procedures are separated in `6f7184ab351c8548d43b0ee25c30d3a23e1cd8b5`; harness subprocess
fixtures are correctly colocated in `f3fc48d21dae09efb4c329d4e985546b6bb3739a`. Fourteen methods use actual
private owners and source publication delegates only its query/provider/reread
facet while Store retains original locks and exact evidence/atomic authority.
All 25 production and two helper streams/raw literals match with explicit
normalization/recomposition. The initial 38-case command passed 37 and failed one;
that binding case and old-reader failure reproduce, then both pass after moving
the exact helper bodies into their invoking binary. Six races and corrected
short/style/strict lint pass without skips or owned leaks. All 34 exact workflow
subprocess selectors resolve. All 908 source-method decisions are recorded;
Store has 777 actual methods. Final bridge/compatibility acceptance, guided
parity, exact-head slice validation and canonical PR preparation remain open.

Final source/compatibility acceptance records all 908 original method decisions
and 217 exact owned secondary references with individual contract/control bases,
current hashes, named regressions and retirement conditions. Six terms and 32
cases preserve supported retained schemas, paths, statuses, authority and public
flags. All 65 root core production files match the shared ownership map; zero
workflow aliases/root facades and 22 actual finite ports are verified. Two stale
observer operation names are corrected. These are metadata-only changes; runtime
source is unchanged. Required development and shared guided parity checks,
ordered exact-HEAD controller acceptance, canonical PR/reference follow-up and
controller-owned dev integration remain the delivery boundary.

Canonical PR [#452](https://forge.adrianpk.com/adrianpk/chr/pulls/452) is open
against dev. The controller recorded successful `make verify` and
`git diff --check` at `30743d17b45afc3656cc08aac01e1b8950422984` without owned
fixture leaks. This ticket and the report are reviewing. Final clean PR-head
acceptance uses the report's ordered commands and the canonical PR's successful
Validation list. The controller retains the exact HEAD/command evidence and owns
dev integration; this transition does not mark the slice delivered.

PR #452 is verified merged into dev at `a41e62be3b2a14daf04b470e8d9344d350480f34`. All five final
ordered controller checks passed on that exact clean head. Slice 15 is delivered;
final set enforcement and integrated gate closure remain Slice 16 work.
