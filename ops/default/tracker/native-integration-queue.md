# Native Integration Corrective Tracker

Status: active

Date: 2026-10-11

Approved: 2026-10-11; maintainer approved both specifications and requested Looper execution of this exact corrective set.

Delivery set: native-integration-queue

Spec: [Native Integration Queue and Merger](../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../spec/stable/native-integration-queue-model.md)

Specification approval: accepted on 2026-10-11

Plan approval: approved on 2026-10-11 under the explicit instruction to process the approved specification through Looper.

Implementation authorization: granted for the complete ordered corrective set and its defined task/PR map.

Active slice: Slice 6 — Bounded merger role

Next slice: Slice 7 — Combined validation and review

Slice strategy: behavior-first

Reason: prove a positive remote transaction and real forge observation before building dependent queue and merger behavior. Deliver types, mappings, implementation and meaningful failure checks together.

Plan: [Native Integration Corrective Delivery](../plan/native-integration-queue.md)

Progress: 5 of 13 slices integrated; 20 of 51 execution tasks delivered in this set.

Task accounting: 47 new tickets and 4 reused original assurance tickets. Retained assurance commits are implemented evidence, not integrated task completion.

## Current Checkpoint

Slice 1 is canonically integrated through [PR #375](https://forge.adrianpk.com/adrianpk/chr/pulls/375) at `79dfe32380dae9058bde15ba50b8f30c30cbaf3d`. All four ordered Looper checks passed for that exact clean HEAD; its five task tickets are solved. Slice 2 is canonically integrated through [PR #376](https://forge.adrianpk.com/adrianpk/chr/pulls/376) at `3129b5ce312df22287a4e7b78e08bf41f5958377`. All five ordered Looper checks passed for that exact clean HEAD; its three task tickets are solved. Slice 3 is canonically integrated through [PR #377](https://forge.adrianpk.com/adrianpk/chr/pulls/377) at `ef72c8cbabf06816a93c293161c12d03211979ba`. All four ordered Looper checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 3 report](../report/slices/native-integration-queue/slice-3-integration-records.md) retains task and final validation evidence. Slice 4 is canonically integrated through [PR #378](https://forge.adrianpk.com/adrianpk/chr/pulls/378) at `bb3a88888a47040f4ded39807f79623bb86d8ab7`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 4 report](../report/slices/native-integration-queue/slice-4-queue-scheduling.md) retains task and final validation evidence. Slice 5 is canonically integrated through [PR #379](https://forge.adrianpk.com/adrianpk/chr/pulls/379) at `771ae3d603f34e6c12d16243e258a125e5db4f9b`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 5 report](../report/slices/native-integration-queue/slice-5-method-preparation.md) retains task and final validation evidence. Slice 6 is active for the bounded merger role. T6.1 protocol/evidence guards pass focused race, parser fuzz and required pre-commit checks; T6.2 owned broker execution passes focused race and required pre-commit checks; T6.3 semantic/replay repair is next. No canonical Slice 6 PR is open.
The original controller remains inactive and blocked; its checkpoint and consumed budgets are preserved. Its assurance execution is transferred to Slice 13 under the approved plan.
Both live repositories and their policy/access profiles are designated in [Certification Fixtures](../report/native-integration-fixtures.md). Real default rebase through both existing adapters is certified; the Slice 2 report retains the exact-head validation boundary.

States: `ready`, `in_progress`, `validated`, `blocked`, `done`. Validated work becomes done only after verified slice integration and corresponding ticket finalization.
Only Looper may merge this set's canonical PRs. Read-only status must not mutate work.

## Slice Summary

| Slice | Name | Exact branch | Expected PR title | Report | State | PR |
| --- | --- | --- | --- | --- | --- | --- |
| Slice 1 | Guarded Git transport | `feat/slice-1-native-git-transport` | `feat(slice-1): publish exact results through atomic Git transport` | `ops/default/report/slices/native-integration-queue/slice-1-git-transport.md` | delivered | [#375](https://forge.adrianpk.com/adrianpk/chr/pulls/375) |
| Slice 2 | Forge publication profiles | `feat/slice-2-native-forge-profiles` | `feat(slice-2): observe native publication through existing forge adapters` | `ops/default/report/slices/native-integration-queue/slice-2-forge-profiles.md` | delivered | [#376](https://forge.adrianpk.com/adrianpk/chr/pulls/376) |
| Slice 3 | Durable integration records | `feat/slice-3-native-integration-records` | `feat(slice-3): persist native integration queue records` | `ops/default/report/slices/native-integration-queue/slice-3-integration-records.md` | delivered | [#377](https://forge.adrianpk.com/adrianpk/chr/pulls/377) |
| Slice 4 | Canonical queue scheduling | `feat/slice-4-native-queue-scheduling` | `feat(slice-4): schedule integration across canonical target queues` | `ops/default/report/slices/native-integration-queue/slice-4-queue-scheduling.md` | delivered | [#378](https://forge.adrianpk.com/adrianpk/chr/pulls/378) |
| Slice 5 | Method preparation and provenance | `feat/slice-5-native-method-preparation` | `feat(slice-5): preserve exact integration history and provenance` | `ops/default/report/slices/native-integration-queue/slice-5-method-preparation.md` | delivered | [#379](https://forge.adrianpk.com/adrianpk/chr/pulls/379) |
| Slice 6 | Bounded merger role | `feat/slice-6-native-merger-role` | `feat(slice-6): resolve integration conflicts through a bounded merger` | `ops/default/report/slices/native-integration-queue/slice-6-merger-role.md` | active | pending |
| Slice 7 | Combined validation and review | `feat/slice-7-native-candidate-review` | `feat(slice-7): revalidate repaired and rebased integration candidates` | `ops/default/report/slices/native-integration-queue/slice-7-candidate-review.md` | not_started | pending |
| Slice 8 | Persistent limits and decisions | `feat/slice-8-native-integration-limits` | `feat(slice-8): enforce persistent merger and invocation budgets` | `ops/default/report/slices/native-integration-queue/slice-8-integration-limits.md` | not_started | pending |
| Slice 9 | Publication and uncertainty recovery | `feat/slice-9-native-publication-recovery` | `feat(slice-9): reconcile guarded publication without duplicate submission` | `ops/default/report/slices/native-integration-queue/slice-9-publication-recovery.md` | not_started | pending |
| Slice 10 | Receipts fanout and closure | `feat/slice-10-native-receipt-closure` | `feat(slice-10): finalize exact integration receipts and task delivery` | `ops/default/report/slices/native-integration-queue/slice-10-receipt-closure.md` | not_started | pending |
| Slice 11 | Shared operation surfaces | `feat/slice-11-native-integration-surfaces` | `feat(slice-11): expose native integration progress and controls` | `ops/default/report/slices/native-integration-queue/slice-11-integration-surfaces.md` | not_started | pending |
| Slice 12 | Live provider certification | `test/slice-12-native-provider-certification` | `test(slice-12): certify native integration across supported forge profiles` | `ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md` | not_started | pending |
| Slice 13 | Assurance reuse and final closure | `test/slice-13-native-integration-assurance` | `test(slice-13): complete native integration and original workflow assurance` | `ops/default/report/slices/native-integration-queue/slice-13-integration-assurance.md` | not_started | pending |

## Slice 1: Guarded Git transport

Status: delivered

Expected branch: `feat/slice-1-native-git-transport`

Expected PR title: `feat(slice-1): publish exact results through atomic Git transport`

Expected report: `ops/default/report/slices/native-integration-queue/slice-1-git-transport.md`

Prerequisite: committed approved bootstrap on canonical `dev`.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T1.1 | [TKT-20261011133051](../ticket/solved/20261011133051-native-integration-t1-1.md) | done | `7d3ced5eac982d86c86d5368b576663a16292cde` | Execute an explicit atomic receive-pack transaction |
| T1.2 | [TKT-20261011133052](../ticket/solved/20261011133052-native-integration-t1-2.md) | done | `c1d02f6cadd5c84fd7f92feaeff11bac3ece1614` | Support host-authenticated SSH publication |
| T1.3 | [TKT-20261011133053](../ticket/solved/20261011133053-native-integration-t1-3.md) | done | `85719c14d17e956965f1f71aadb548ba12a645ca` | Support host-authenticated HTTPS publication |
| T1.4 | [TKT-20261011133054](../ticket/solved/20261011133054-native-integration-t1-4.md) | done | `cbd4fb8a6e9ab0a1471cb09db7f4a78d9b3d304d` | Seal exact source/base/result publication proof |
| T1.5 | [TKT-20261011133055](../ticket/solved/20261011133055-native-integration-t1-5.md) | done | `b7790de38a5473f3f0f3c6e5187ececb7f11089d` | Verify protocol races and review the transport boundary |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#375](https://forge.adrianpk.com/adrianpk/chr/pulls/375)

Verified merge: [#375](https://forge.adrianpk.com/adrianpk/chr/pulls/375), `79dfe32380dae9058bde15ba50b8f30c30cbaf3d` on `dev`

Report/evidence: [Slice 1 report](../report/slices/native-integration-queue/slice-1-git-transport.md)

## Slice 2: Forge publication profiles

Status: delivered

Expected branch: `feat/slice-2-native-forge-profiles`

Expected PR title: `feat(slice-2): observe native publication through existing forge adapters`

Expected report: `ops/default/report/slices/native-integration-queue/slice-2-forge-profiles.md`

Prerequisite: verified Slice 1 integration and closed task fanout.

External prerequisite: explicit disposable repository and policy/access profile designation for both existing forges, supplied in [Certification Fixtures](../report/native-integration-fixtures.md); recheck current access before live probes.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T2.1 | [TKT-20261011133056](../ticket/solved/20261011133056-native-integration-t2-1.md) | done | `131dd1733a087585add492b68ba5fa3177ef3ec7` | Admit an explicit owned-source publication profile |
| T2.2 | [TKT-20261011133057](../ticket/solved/20261011133057-native-integration-t2-2.md) | done | `17a308198809683a0bd8adff965379bbd6dd90a6` | Certify Forgejo default rebase observation |
| T2.3 | [TKT-20261011133058](../ticket/solved/20261011133058-native-integration-t2-3.md) | done | `a7e90f2ee18ba5d09e46677c101b38f981916b1b` | Certify GitHub default rebase observation |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#376](https://forge.adrianpk.com/adrianpk/chr/pulls/376)

Verified merge: [#376](https://forge.adrianpk.com/adrianpk/chr/pulls/376), `3129b5ce312df22287a4e7b78e08bf41f5958377` on `dev`

Report/evidence: [Slice 2 report](../report/slices/native-integration-queue/slice-2-forge-profiles.md)

## Slice 3: Durable integration records

Status: delivered

Expected branch: `feat/slice-3-native-integration-records`

Expected PR title: `feat(slice-3): persist native integration queue records`

Expected report: `ops/default/report/slices/native-integration-queue/slice-3-integration-records.md`

Prerequisite: verified Slice 2 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T3.1 | [TKT-20261011133059](../ticket/solved/20261011133059-native-integration-t3-1.md) | done | `64369b81579c8b4136609383568731067bc7bdfd` | Extend candidate and receipt schemas |
| T3.2 | [TKT-20261011133100](../ticket/solved/20261011133100-native-integration-t3-2.md) | done | `9a004ba55c6b72ac69aa0322fba531eebcc257d7` | Persist integration entries through WorkflowStore |
| T3.3 | [TKT-20261011133101](../ticket/solved/20261011133101-native-integration-t3-3.md) | done | `dce7077f632e9d1d93dcdde5d35ba958f4695e0f` | Extend attempt and budget journals |
| T3.4 | [TKT-20261011133102](../ticket/solved/20261011133102-native-integration-t3-4.md) | done | `c7ee88bb92a27c520489effb717cecde34f2dbdf` | Rebuild indexes and validate migrations |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#377](https://forge.adrianpk.com/adrianpk/chr/pulls/377)

Verified merge: [#377](https://forge.adrianpk.com/adrianpk/chr/pulls/377), `ef72c8cbabf06816a93c293161c12d03211979ba` on `dev`

Report/evidence: [Slice 3 report](../report/slices/native-integration-queue/slice-3-integration-records.md)

## Slice 4: Canonical queue scheduling

Status: delivered

Expected branch: `feat/slice-4-native-queue-scheduling`

Expected PR title: `feat(slice-4): schedule integration across canonical target queues`

Expected report: `ops/default/report/slices/native-integration-queue/slice-4-queue-scheduling.md`

Prerequisite: verified Slice 3 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T4.1 | [TKT-20261011133103](../ticket/solved/20261011133103-native-integration-t4-1.md) | done | `2051bb53de7b8e452f455809915d47c48352ba61` | Schedule eligible units deterministically |
| T4.2 | [TKT-20261011133104](../ticket/solved/20261011133104-native-integration-t4-2.md) | done | `05d438047230b65833949c84e7c8b1ac99190f33` | Share role capacity and prevent starvation |
| T4.3 | [TKT-20261011133105](../ticket/solved/20261011133105-native-integration-t4-3.md) | done | `35d39f6b6df235d79848fb59054fd7d8b2ef0da3` | Normalize target and coordinator identity |
| T4.4 | [TKT-20261011133106](../ticket/solved/20261011133106-native-integration-t4-4.md) | done | `c08b6d6590342a45dcb21fce12d05b740f98ef6b` | Verify real concurrent scheduling |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#378](https://forge.adrianpk.com/adrianpk/chr/pulls/378)

Verified merge: [#378](https://forge.adrianpk.com/adrianpk/chr/pulls/378), `bb3a88888a47040f4ded39807f79623bb86d8ab7` on `dev`

Report/evidence: [Slice 4 report](../report/slices/native-integration-queue/slice-4-queue-scheduling.md)

## Slice 5: Method preparation and provenance

Status: delivered

Expected branch: `feat/slice-5-native-method-preparation`

Expected PR title: `feat(slice-5): preserve exact integration history and provenance`

Expected report: `ops/default/report/slices/native-integration-queue/slice-5-method-preparation.md`

Prerequisite: verified Slice 4 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T5.1 | [TKT-20261011133107](../ticket/solved/20261011133107-native-integration-t5-1.md) | delivered | `d6532b26691da0df0002f3d29753a9ae0a52be11` | Prepare exact merge and squash candidates |
| T5.2 | [TKT-20261011133108](../ticket/solved/20261011133108-native-integration-t5-2.md) | delivered | `ed92e270edd489332191667c1c80201536363ba2` | Replay attributed rebase histories |
| T5.3 | [TKT-20261011133109](../ticket/solved/20261011133109-native-integration-t5-3.md) | delivered | `134ebaf9c8041ca92bdd184ee48f6e521a293e07` | Enforce non-no-op and signing policy |
| T5.4 | [TKT-20261011133110](../ticket/solved/20261011133110-native-integration-t5-4.md) | delivered | `a6a4d6ec83d8dc15f4bdf05712417567dc6145fc` | Verify preparation isolation and method proofs |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#379](https://forge.adrianpk.com/adrianpk/chr/pulls/379)

Verified merge: `771ae3d603f34e6c12d16243e258a125e5db4f9b` at 2026-10-12T12:53:28Z; controller Acta `01M4R2SQDTK2WNACPWQRJ7RAQR`.

Report/evidence: [Slice 5 report](../report/slices/native-integration-queue/slice-5-method-preparation.md); task checks are retained there, and the controller owns exact-HEAD candidate evidence.

## Slice 6: Bounded merger role

Status: active

Expected branch: `feat/slice-6-native-merger-role`

Expected PR title: `feat(slice-6): resolve integration conflicts through a bounded merger`

Expected report: `ops/default/report/slices/native-integration-queue/slice-6-merger-role.md`

Prerequisite: verified Slice 5 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T6.1 | [TKT-20261011133111](../ticket/reviewing/20261011133111-native-integration-t6-1.md) | validated | `dd376916016bc85c9f20677e03997e0b7f6a2a58` | Define versioned merger input and result |
| T6.2 | [TKT-20261011133112](../ticket/reviewing/20261011133112-native-integration-t6-2.md) | validated | `78793f6ce58fc5fabedcece6e9338698713d8739` | Execute the merger in an owned checkout |
| T6.3 | [TKT-20261011133113](../ticket/ready/20261011133113-native-integration-t6-3.md) | ready | pending | Preserve semantic intent during repair |
| T6.4 | [TKT-20261011133114](../ticket/ready/20261011133114-native-integration-t6-4.md) | ready | pending | Verify host acceptance of merger changes |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: [Slice 6 report](../report/slices/native-integration-queue/slice-6-merger-role.md); T6.1 and T6.2 task evidence is retained there. T6.3–T6.4 and exact-HEAD candidate checks remain pending.

## Slice 7: Combined validation and review

Status: not_started

Expected branch: `feat/slice-7-native-candidate-review`

Expected PR title: `feat(slice-7): revalidate repaired and rebased integration candidates`

Expected report: `ops/default/report/slices/native-integration-queue/slice-7-candidate-review.md`

Prerequisite: verified Slice 6 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T7.1 | [TKT-20261011133115](../ticket/ready/20261011133115-native-integration-t7-1.md) | ready | pending | Run required gates on the exact prepared result |
| T7.2 | [TKT-20261011133116](../ticket/ready/20261011133116-native-integration-t7-2.md) | ready | pending | Bind review to the current candidate |
| T7.3 | [TKT-20261011133117](../ticket/ready/20261011133117-native-integration-t7-3.md) | ready | pending | Repair actionable integration findings |
| T7.4 | [TKT-20261011133118](../ticket/ready/20261011133118-native-integration-t7-4.md) | ready | pending | Handle external target and source movement |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: pending

## Slice 8: Persistent limits and decisions

Status: not_started

Expected branch: `feat/slice-8-native-integration-limits`

Expected PR title: `feat(slice-8): enforce persistent merger and invocation budgets`

Expected report: `ops/default/report/slices/native-integration-queue/slice-8-integration-limits.md`

Prerequisite: verified Slice 7 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T8.1 | [TKT-20261011133119](../ticket/ready/20261011133119-native-integration-t8-1.md) | ready | pending | Enforce merger drift and submission counters |
| T8.2 | [TKT-20261011133120](../ticket/ready/20261011133120-native-integration-t8-2.md) | ready | pending | Reserve and settle optional token ceilings |
| T8.3 | [TKT-20261011133121](../ticket/ready/20261011133121-native-integration-t8-3.md) | ready | pending | Persist operator decisions and bounded grants |
| T8.4 | [TKT-20261011133122](../ticket/ready/20261011133122-native-integration-t8-4.md) | ready | pending | Verify process and budget recovery |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: pending

## Slice 9: Publication and uncertainty recovery

Status: not_started

Expected branch: `feat/slice-9-native-publication-recovery`

Expected PR title: `feat(slice-9): reconcile guarded publication without duplicate submission`

Expected report: `ops/default/report/slices/native-integration-queue/slice-9-publication-recovery.md`

Prerequisite: verified Slice 8 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T9.1 | [TKT-20261011133123](../ticket/ready/20261011133123-native-integration-t9-1.md) | ready | pending | Fence and journal authorized dispatch |
| T9.2 | [TKT-20261011133124](../ticket/ready/20261011133124-native-integration-t9-2.md) | ready | pending | Reconcile atomic rejection and lost responses |
| T9.3 | [TKT-20261011133125](../ticket/ready/20261011133125-native-integration-t9-3.md) | ready | pending | Observe canonical PR convergence |
| T9.4 | [TKT-20261011133126](../ticket/ready/20261011133126-native-integration-t9-4.md) | ready | pending | Verify crash and stop boundaries |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: pending

## Slice 10: Receipts fanout and closure

Status: not_started

Expected branch: `feat/slice-10-native-receipt-closure`

Expected PR title: `feat(slice-10): finalize exact integration receipts and task delivery`

Expected report: `ops/default/report/slices/native-integration-queue/slice-10-receipt-closure.md`

Prerequisite: verified Slice 9 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T10.1 | [TKT-20261011133127](../ticket/ready/20261011133127-native-integration-t10-1.md) | ready | pending | Verify immutable dual-ref receipts |
| T10.2 | [TKT-20261011133128](../ticket/ready/20261011133128-native-integration-t10-2.md) | ready | pending | Replay task finalization exactly once |
| T10.3 | [TKT-20261011133129](../ticket/ready/20261011133129-native-integration-t10-3.md) | ready | pending | Preserve descendant proof and report divergence |
| T10.4 | [TKT-20261011133130](../ticket/ready/20261011133130-native-integration-t10-4.md) | ready | pending | Enforce set closure barriers |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: pending

## Slice 11: Shared operation surfaces

Status: not_started

Expected branch: `feat/slice-11-native-integration-surfaces`

Expected PR title: `feat(slice-11): expose native integration progress and controls`

Expected report: `ops/default/report/slices/native-integration-queue/slice-11-integration-surfaces.md`

Prerequisite: verified Slice 10 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T11.1 | [TKT-20261011133131](../ticket/ready/20261011133131-native-integration-t11-1.md) | ready | pending | Expose shared integration service operations |
| T11.2 | [TKT-20261011133132](../ticket/ready/20261011133132-native-integration-t11-2.md) | ready | pending | Present CLI and TUI integration status |
| T11.3 | [TKT-20261011133133](../ticket/ready/20261011133133-native-integration-t11-3.md) | ready | pending | Present Web daemon and manual controls |
| T11.4 | [TKT-20261011133134](../ticket/ready/20261011133134-native-integration-t11-4.md) | ready | pending | Document and verify profile setup |

Required checks:

- `make verify`
- `go test ./internal/app ./internal/tui ./internal/webui -count=1`
- `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: pending

## Slice 12: Live provider certification

Status: not_started

Expected branch: `test/slice-12-native-provider-certification`

Expected PR title: `test(slice-12): certify native integration across supported forge profiles`

Expected report: `ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md`

Prerequisite: verified Slice 11 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T12.1 | [TKT-20261011133135](../ticket/ready/20261011133135-native-integration-t12-1.md) | ready | pending | Certify Forgejo merge rebase and squash |
| T12.2 | [TKT-20261011133136](../ticket/ready/20261011133136-native-integration-t12-2.md) | ready | pending | Certify GitHub merge rebase and squash |
| T12.3 | [TKT-20261011133137](../ticket/ready/20261011133137-native-integration-t12-3.md) | ready | pending | Publish bounded certification and denial matrix |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: pending

## Slice 13: Assurance reuse and final closure

Status: not_started

Expected branch: `test/slice-13-native-integration-assurance`

Expected PR title: `test(slice-13): complete native integration and original workflow assurance`

Expected report: `ops/default/report/slices/native-integration-queue/slice-13-integration-assurance.md`

Prerequisite: verified Slice 12 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T13.1 | [TKT-20261008141429](../ticket/blocked/20261008141429-ticket-native-t12-1.md) | blocked | retained: 4f10f1c60de1f164f727d37b4df5c02343ab6e11 | Adapt retained composed workflow assurance |
| T13.2 | [TKT-20261008141430](../ticket/blocked/20261008141430-ticket-native-t12-2.md) | blocked | retained: 44e8145aa618edfc51d5efb08330c641b02f7e58 | Adapt retained fault QA entrypoint |
| T13.3 | [TKT-20261008141431](../ticket/blocked/20261008141431-ticket-native-t12-3.md) | blocked | retained: pending | Reconcile real certification with final behavior |
| T13.4 | [TKT-20261008141432](../ticket/blocked/20261008141432-ticket-native-t12-4.md) | blocked | retained: pending | Run final gates and close both deliveries |

Required checks:

- `make verify`
- `go test -race ./...`
- `make qa-ticket-native-workflow`
- `make qa-native-integration`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: pending

## Final Delivery Gate

- [ ] Every slice has its verified canonical PR merge and delivered report.
- [ ] All task tickets are finalized with preserved retained assurance evidence.
- [ ] All 100 acceptance IDs and 24 invariants have evidence with honest provider limits.
- [ ] The exact integrated full gate passes.
- [ ] Original assurance obligations close with explicit replacement-unit traceability.
- [ ] No unknown submission, unfinished fanout or live owner prevents set closure.
- [ ] Plan/tracker and cycle report close with verified receipts.
