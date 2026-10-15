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

Active slice: Slice 13 — Assurance reuse and final closure

Next slice: none; final integrated gate follows Slice 13

Slice strategy: behavior-first

Reason: prove a positive remote transaction and real forge observation before building dependent queue and merger behavior. Deliver types, mappings, implementation and meaningful failure checks together.

Plan: [Native Integration Corrective Delivery](../plan/native-integration-queue.md)

Progress: 12 of 13 slices integrated; 47 of 51 execution tasks delivered in this set.

Task accounting: 47 new tickets and 4 reused original assurance tickets. Retained assurance commits are implemented evidence, not integrated task completion.

## Current Checkpoint

Slice 1 is canonically integrated through [PR #375](https://forge.adrianpk.com/adrianpk/chr/pulls/375) at `79dfe32380dae9058bde15ba50b8f30c30cbaf3d`. All four ordered Looper checks passed for that exact clean HEAD; its five task tickets are solved. Slice 2 is canonically integrated through [PR #376](https://forge.adrianpk.com/adrianpk/chr/pulls/376) at `3129b5ce312df22287a4e7b78e08bf41f5958377`. All five ordered Looper checks passed for that exact clean HEAD; its three task tickets are solved. Slice 3 is canonically integrated through [PR #377](https://forge.adrianpk.com/adrianpk/chr/pulls/377) at `ef72c8cbabf06816a93c293161c12d03211979ba`. All four ordered Looper checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 3 report](../report/slices/native-integration-queue/slice-3-integration-records.md) retains task and final validation evidence. Slice 4 is canonically integrated through [PR #378](https://forge.adrianpk.com/adrianpk/chr/pulls/378) at `bb3a88888a47040f4ded39807f79623bb86d8ab7`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 4 report](../report/slices/native-integration-queue/slice-4-queue-scheduling.md) retains task and final validation evidence. Slice 5 is canonically integrated through [PR #379](https://forge.adrianpk.com/adrianpk/chr/pulls/379) at `771ae3d603f34e6c12d16243e258a125e5db4f9b`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 5 report](../report/slices/native-integration-queue/slice-5-method-preparation.md) retains task and final validation evidence. Slice 6 is canonically integrated through [PR #380](https://forge.adrianpk.com/adrianpk/chr/pulls/380) at `725128da941a5c149a2b0a5d9a6a7065f842823d`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 6 report](../report/slices/native-integration-queue/slice-6-merger-role.md) retains task and final validation evidence. Slice 7 is canonically integrated through [PR #381](https://forge.adrianpk.com/adrianpk/chr/pulls/381) at `79af9e38426e681e5b0d778ae4df2b0a84952d73`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 7 report](../report/slices/native-integration-queue/slice-7-candidate-review.md) retains task and final validation evidence. Slice 8 is canonically integrated through [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382) at `0096f6e960ecc25e7aae72a8a97fdc53c689418c`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 8 report](../report/slices/native-integration-queue/slice-8-integration-limits.md) retains task and final validation evidence. Slice 9 is canonically integrated through [PR #383](https://forge.adrianpk.com/adrianpk/chr/pulls/383) at `8853fdb14638c853f641b82c7d19d805b41fa8fe`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 9 report](../report/slices/native-integration-queue/slice-9-publication-recovery.md) retains task and final validation evidence. Slice 10 is canonically integrated through [#384](https://forge.adrianpk.com/adrianpk/chr/pulls/384) at `9726200044d15b4e66bacfe1334de4a9410f4bdc`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 10 report](../report/slices/native-integration-queue/slice-10-receipt-closure.md) retains task and final validation evidence. Slice 11 is canonically integrated through [PR #385](https://forge.adrianpk.com/adrianpk/chr/pulls/385) at `12094bddb2d298f6c60fbfcb49f513b6178ff595`. All four ordered controller checks passed for that exact clean HEAD; its four task tickets are solved. The [Slice 11 report](../report/slices/native-integration-queue/slice-11-integration-surfaces.md) retains task and final validation evidence. Slice 12 is canonically integrated through [PR #386](https://forge.adrianpk.com/adrianpk/chr/pulls/386) at `acfb3e445c8056bcc198fef82bbec5a18f132304`. All five final controller checks passed on that exact clean HEAD; its three task tickets are solved. Slice 13 is active for assurance reuse and final closure.
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
| Slice 6 | Bounded merger role | `feat/slice-6-native-merger-role` | `feat(slice-6): resolve integration conflicts through a bounded merger` | `ops/default/report/slices/native-integration-queue/slice-6-merger-role.md` | delivered | [#380](https://forge.adrianpk.com/adrianpk/chr/pulls/380) |
| Slice 7 | Combined validation and review | `feat/slice-7-native-candidate-review` | `feat(slice-7): revalidate repaired and rebased integration candidates` | `ops/default/report/slices/native-integration-queue/slice-7-candidate-review.md` | delivered | [#381](https://forge.adrianpk.com/adrianpk/chr/pulls/381) |
| Slice 8 | Persistent limits and decisions | `feat/slice-8-native-integration-limits` | `feat(slice-8): enforce persistent merger and invocation budgets` | `ops/default/report/slices/native-integration-queue/slice-8-integration-limits.md` | delivered | [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382) |
| Slice 9 | Publication and uncertainty recovery | `feat/slice-9-native-publication-recovery` | `feat(slice-9): reconcile guarded publication without duplicate submission` | `ops/default/report/slices/native-integration-queue/slice-9-publication-recovery.md` | delivered | [#383](https://forge.adrianpk.com/adrianpk/chr/pulls/383) |
| Slice 10 | Receipts fanout and closure | `feat/slice-10-native-receipt-closure` | `feat(slice-10): finalize exact integration receipts and task delivery` | `ops/default/report/slices/native-integration-queue/slice-10-receipt-closure.md` | delivered | [#384](https://forge.adrianpk.com/adrianpk/chr/pulls/384) |
| Slice 11 | Shared operation surfaces | `feat/slice-11-native-integration-surfaces` | `feat(slice-11): expose native integration progress and controls` | `ops/default/report/slices/native-integration-queue/slice-11-integration-surfaces.md` | delivered | [#385](https://forge.adrianpk.com/adrianpk/chr/pulls/385) |
| Slice 12 | Live provider certification | `test/slice-12-native-provider-certification` | `test(slice-12): certify native integration across supported forge profiles` | `ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md` | delivered | [#386](https://forge.adrianpk.com/adrianpk/chr/pulls/386) |
| Slice 13 | Assurance reuse and final closure | `test/slice-13-native-integration-assurance` | `test(slice-13): complete native integration and original workflow assurance` | `ops/default/report/slices/native-integration-queue/slice-13-integration-assurance.md` | in_progress | pending |

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

Status: delivered

Expected branch: `feat/slice-6-native-merger-role`

Expected PR title: `feat(slice-6): resolve integration conflicts through a bounded merger`

Expected report: `ops/default/report/slices/native-integration-queue/slice-6-merger-role.md`

Prerequisite: verified Slice 5 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T6.1 | [TKT-20261011133111](../ticket/solved/20261011133111-native-integration-t6-1.md) | done | `dd376916016bc85c9f20677e03997e0b7f6a2a58` | Define versioned merger input and result |
| T6.2 | [TKT-20261011133112](../ticket/solved/20261011133112-native-integration-t6-2.md) | done | `78793f6ce58fc5fabedcece6e9338698713d8739` | Execute the merger in an owned checkout |
| T6.3 | [TKT-20261011133113](../ticket/solved/20261011133113-native-integration-t6-3.md) | done | `97518aabe9b214d18c86440c372511c17676e958` | Preserve semantic intent during repair |
| T6.4 | [TKT-20261011133114](../ticket/solved/20261011133114-native-integration-t6-4.md) | done | `679063f2fca4b95d5dd9dbefe686ecb7332be6c3` | Verify host acceptance of merger changes |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#380](https://forge.adrianpk.com/adrianpk/chr/pulls/380)

Verified merge: `725128da941a5c149a2b0a5d9a6a7065f842823d` at 2026-10-12T19:14:21Z; controller Acta `01M4RRK4KT3HRPZ57T5VAW4FM0`.

Report/evidence: [Slice 6 report](../report/slices/native-integration-queue/slice-6-merger-role.md); all four task commits, Acta records and successful pre-commit checks are retained there. All four ordered exact-HEAD candidate checks and canonical integration are verified.

## Slice 7: Combined validation and review

Status: delivered

Expected branch: `feat/slice-7-native-candidate-review`

Expected PR title: `feat(slice-7): revalidate repaired and rebased integration candidates`

Expected report: `ops/default/report/slices/native-integration-queue/slice-7-candidate-review.md`

Prerequisite: verified Slice 6 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T7.1 | [TKT-20261011133115](../ticket/solved/20261011133115-native-integration-t7-1.md) | done | [3f1aac6910b7e8885146a544a735db69f49c1a94](https://forge.adrianpk.com/adrianpk/chr/commit/3f1aac6910b7e8885146a544a735db69f49c1a94) | Run required gates on the exact prepared result |
| T7.2 | [TKT-20261011133116](../ticket/solved/20261011133116-native-integration-t7-2.md) | done | [c1c3d9dc954e9d3888c55a9a63dfac4fc04f6b59](https://forge.adrianpk.com/adrianpk/chr/commit/c1c3d9dc954e9d3888c55a9a63dfac4fc04f6b59) | Bind review to the current candidate |
| T7.3 | [TKT-20261011133117](../ticket/solved/20261011133117-native-integration-t7-3.md) | done | [91676739a6706575df5ee1e644acf29189cfc62b](https://forge.adrianpk.com/adrianpk/chr/commit/91676739a6706575df5ee1e644acf29189cfc62b) | Repair actionable integration findings |
| T7.4 | [TKT-20261011133118](../ticket/solved/20261011133118-native-integration-t7-4.md) | done | [19a339e92300df7141f57458c7892de2876cecac](https://forge.adrianpk.com/adrianpk/chr/commit/19a339e92300df7141f57458c7892de2876cecac) | Handle external target and source movement |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#381](https://forge.adrianpk.com/adrianpk/chr/pulls/381)

Verified merge: `79af9e38426e681e5b0d778ae4df2b0a84952d73` at 2026-10-13T05:58:15Z; controller Acta `01M4SXE559V2QJJB40S9H1S9QB`.

Report/evidence: [Slice 7 report](../report/slices/native-integration-queue/slice-7-candidate-review.md); all four task commits and development checks are retained there. The required race check exposed aggregate stop fixture readiness timing; [4f4a3cc3dc1ffe0a791d484cdf1b3e2982935125](https://forge.adrianpk.com/adrianpk/chr/commit/4f4a3cc3dc1ffe0a791d484cdf1b3e2982935125) isolates that readiness wait after source-task setup, with focused normal/race and lint checks passing. Final exact-HEAD evidence is owned by the controller and canonical PR.

## Slice 8: Persistent limits and decisions

Status: delivered

Expected branch: `feat/slice-8-native-integration-limits`

Expected PR title: `feat(slice-8): enforce persistent merger and invocation budgets`

Expected report: `ops/default/report/slices/native-integration-queue/slice-8-integration-limits.md`

Prerequisite: verified Slice 7 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T8.1 | [TKT-20261011133119](../ticket/solved/20261011133119-native-integration-t8-1.md) | done | [94ea7e86745e9ab30d55bc9fa931fd23a91826bc](https://forge.adrianpk.com/adrianpk/chr/commit/94ea7e86745e9ab30d55bc9fa931fd23a91826bc) | Enforce merger drift and submission counters |
| T8.2 | [TKT-20261011133120](../ticket/solved/20261011133120-native-integration-t8-2.md) | done | [e1e574d4c031a25452de66da886be0e1df7271a9](https://forge.adrianpk.com/adrianpk/chr/commit/e1e574d4c031a25452de66da886be0e1df7271a9) | Reserve and settle optional token ceilings |
| T8.3 | [TKT-20261011133121](../ticket/solved/20261011133121-native-integration-t8-3.md) | done | `cb24ddffc886af7db8fe2f746fdd0d9982bbcdd8` | Persist operator decisions and bounded grants |
| T8.4 | [TKT-20261011133122](../ticket/solved/20261011133122-native-integration-t8-4.md) | done | [b0574cca109c65d46a00bbbfb0b426f35fe751d1](https://forge.adrianpk.com/adrianpk/chr/commit/b0574cca109c65d46a00bbbfb0b426f35fe751d1) | Verify process and budget recovery |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382)

Verified merge: [PR #382](https://forge.adrianpk.com/adrianpk/chr/pulls/382) at `0096f6e960ecc25e7aae72a8a97fdc53c689418c` on 2026-10-13T18:09:40Z; Acta `01M4V79DVNVSXZG5D9TMZ0NMSX`.

Report/evidence: [Slice 8 report](../report/slices/native-integration-queue/slice-8-integration-limits.md); all four ordered controller checks passed for the exact merged head and all four task tickets are solved.

## Slice 9: Publication and uncertainty recovery

Status: delivered

Expected branch: `feat/slice-9-native-publication-recovery`

Expected PR title: `feat(slice-9): reconcile guarded publication without duplicate submission`

Expected report: `ops/default/report/slices/native-integration-queue/slice-9-publication-recovery.md`

Prerequisite: verified Slice 8 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T9.1 | [TKT-20261011133123](../ticket/solved/20261011133123-native-integration-t9-1.md) | done | [e4531667cbbeb82922d79500cb7bd79c72e2efb9](https://forge.adrianpk.com/adrianpk/chr/commit/e4531667cbbeb82922d79500cb7bd79c72e2efb9) | Fence and journal authorized dispatch |
| T9.2 | [TKT-20261011133124](../ticket/solved/20261011133124-native-integration-t9-2.md) | done | [ae54f89aff1248e7cee90c5e033ac923ec7c77c7](https://forge.adrianpk.com/adrianpk/chr/commit/ae54f89aff1248e7cee90c5e033ac923ec7c77c7) | Reconcile atomic rejection and lost responses |
| T9.3 | [TKT-20261011133125](../ticket/solved/20261011133125-native-integration-t9-3.md) | done | [48a9539d023e63c96f04aa61ecb6baa46d77f9e2](https://forge.adrianpk.com/adrianpk/chr/commit/48a9539d023e63c96f04aa61ecb6baa46d77f9e2) | Observe canonical PR convergence |
| T9.4 | [TKT-20261011133126](../ticket/solved/20261011133126-native-integration-t9-4.md) | done | [a06c392fc7a6b9169694989e436ea6afdec4010e](https://forge.adrianpk.com/adrianpk/chr/commit/a06c392fc7a6b9169694989e436ea6afdec4010e) | Verify crash and stop boundaries |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#383](https://forge.adrianpk.com/adrianpk/chr/pulls/383)

Verified merge: `8853fdb14638c853f641b82c7d19d805b41fa8fe` at 2026-10-14T08:13:24Z; controller Acta `01M4WQJAN50D6ZWVGHEVQ52SSW`.

Report/evidence: [Slice 9 report](../report/slices/native-integration-queue/slice-9-publication-recovery.md); T9.1 focused normal/race checks and strict lint passed. The required pre-commit repository check passed with seven unchanged Go files. T9.2 broader normal/race checks and the required pre-commit repository check passed with fourteen unchanged Go files. T9.2 implementation is recorded. T9.3 normal/race checks and the required pre-commit `make verify` passed on nine unchanged Go files. T9.3 implementation is recorded. T9.4 normal/race checks and required pre-commit `make verify` passed on seven unchanged Go files. All four tasks are implemented and development-checked. All four configured ordered exact-HEAD checks passed at the integrated commit. Canonical merge and all four task tickets are closed.

## Slice 10: Receipts fanout and closure

Status: delivered

Expected branch: `feat/slice-10-native-receipt-closure`

Expected PR title: `feat(slice-10): finalize exact integration receipts and task delivery`

Expected report: `ops/default/report/slices/native-integration-queue/slice-10-receipt-closure.md`

Prerequisite: verified Slice 9 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T10.1 | [TKT-20261011133127](../ticket/solved/20261011133127-native-integration-t10-1.md) | done | [b0a64a02dbc13ad61220cde8502fca66f7a332f9](https://forge.adrianpk.com/adrianpk/chr/commit/b0a64a02dbc13ad61220cde8502fca66f7a332f9) | Verify immutable dual-ref receipts |
| T10.2 | [TKT-20261011133128](../ticket/solved/20261011133128-native-integration-t10-2.md) | done | [c96ecba537e1bfedf6925c0324f0d49bd8332d89](https://forge.adrianpk.com/adrianpk/chr/commit/c96ecba537e1bfedf6925c0324f0d49bd8332d89) | Replay task finalization exactly once |
| T10.3 | [TKT-20261011133129](../ticket/solved/20261011133129-native-integration-t10-3.md) | done | [dd7dc5b536c4a7ccd9e620824ef31b976441d87c](https://forge.adrianpk.com/adrianpk/chr/commit/dd7dc5b536c4a7ccd9e620824ef31b976441d87c) | Preserve descendant proof and report divergence |
| T10.4 | [TKT-20261011133130](../ticket/solved/20261011133130-native-integration-t10-4.md) | done | [a9b33ea1f09e16fa5167c0fe75b2a4218b7a8d47](https://forge.adrianpk.com/adrianpk/chr/commit/a9b33ea1f09e16fa5167c0fe75b2a4218b7a8d47) | Enforce set closure barriers |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#384](https://forge.adrianpk.com/adrianpk/chr/pulls/384)

Verified merge: `9726200044d15b4e66bacfe1334de4a9410f4bdc` at 2026-10-14T18:36:18Z; controller Acta `01M4XV6YGZZBS8CGG3AJ2EHKT7`.

Report/evidence: [Slice 10 report](../report/slices/native-integration-queue/slice-10-receipt-closure.md); all four task implementations, development checks and Acta references are recorded. Exact receipt confirmation, recoverable task fanout, descendant-history reconciliation and verified native closure are integrated. All four configured ordered exact-HEAD checks passed at the merged commit. The canonical merge is verified and all four task tickets are solved.

## Slice 11: Shared operation surfaces

Status: delivered

Expected branch: `feat/slice-11-native-integration-surfaces`

Expected PR title: `feat(slice-11): expose native integration progress and controls`

Expected report: `ops/default/report/slices/native-integration-queue/slice-11-integration-surfaces.md`

Prerequisite: verified Slice 10 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T11.1 | [TKT-20261011133131](../ticket/solved/20261011133131-native-integration-t11-1.md) | done | [98915f97987104ae0cb3bc67ddcf3f8f91dbab0d](https://forge.adrianpk.com/adrianpk/chr/commit/98915f97987104ae0cb3bc67ddcf3f8f91dbab0d) | Expose shared integration service operations |
| T11.2 | [TKT-20261011133132](../ticket/solved/20261011133132-native-integration-t11-2.md) | done | [18b307141f84534b78d38429db1f79a380417c1a](https://forge.adrianpk.com/adrianpk/chr/commit/18b307141f84534b78d38429db1f79a380417c1a) | Present CLI and TUI integration status |
| T11.3 | [TKT-20261011133133](../ticket/solved/20261011133133-native-integration-t11-3.md) | done | [e37e275b49c8a16f1349c916be9f5045b4be811f](https://forge.adrianpk.com/adrianpk/chr/commit/e37e275b49c8a16f1349c916be9f5045b4be811f) | Present Web daemon and manual controls |
| T11.4 | [TKT-20261011133134](../ticket/solved/20261011133134-native-integration-t11-4.md) | done | [266f652b0f3367a37f0ea88b7e5825c69a2a9f16](https://forge.adrianpk.com/adrianpk/chr/commit/266f652b0f3367a37f0ea88b7e5825c69a2a9f16) | Document and verify profile setup |

Required checks:

- `make verify`
- `go test ./internal/app ./internal/tui ./internal/webui -count=1`
- `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`
- `git diff --check`

PR: [#385](https://forge.adrianpk.com/adrianpk/chr/pulls/385)

Verified merge: `12094bddb2d298f6c60fbfcb49f513b6178ff595` at 2026-10-14T23:48:48Z; controller Acta `01M4YD33D6YKR6AASWSDNZEEV0`.

Report/evidence: [Slice 11 report](../report/slices/native-integration-queue/slice-11-integration-surfaces.md); T11.1 shared native inspection and displayed revision-bound actions are implemented. Focused core/app checks passed in 69.682 s and 33.102 s with zero strict lint issues, recorded as Acta `01M4XXSWZ91P9JV7JBWDQ20S78`. Broader race checks passed in 115.109 s (core) and 53.292 s (app), recorded as Acta `01M4XY9J5W70A20HWQ4D5VXTXM`. Required pre-commit `make verify` passed with zero lint issues, core 2754.086 s, app 126.105 s and all TUI/Web packages passed, recorded as Acta `01M4Y0NX6YKJTA466DP5TVQ8DA`. The owned service exited successfully on frozen 11-file source `8ddc53dce797218763bebd341c36bd8578efd8b5e4e46bfa598a2caf727428ad`; T11.1 is reviewing at [98915f97987104ae0cb3bc67ddcf3f8f91dbab0d](https://forge.adrianpk.com/adrianpk/chr/commit/98915f97987104ae0cb3bc67ddcf3f8f91dbab0d), recorded as Acta `01M4Y0R3R8REK566SDYW7AMC0M`. T11.2 retained CLI/TUI views are implemented; focused app regressions passed in 33.812 s with zero strict lint issues, recorded as Acta `01M4Y19DFC1ADK6N1T1ZWWYEH6`. Focused race passed in app 64.400 s and TUI 1.023 s, recorded as Acta `01M4Y1EB7NM6TFQ16NCK6CEYN3`. Required pre-commit `make verify` passed with zero lint issues, core 2809.193 s, app 128.702 s and all TUI/Web packages passed; Acta `01M4Y44X498KK6Z58E062WY9MX`. The owned unit `chr-native-slice11-t112-validation-3d98c772.service` exited successfully on frozen five-file source `3d98c77273f1916314042ea00886b32695cb45b5068c69dbf01cff119069f44a`. T11.2 is reviewing at [18b307141f84534b78d38429db1f79a380417c1a](https://forge.adrianpk.com/adrianpk/chr/commit/18b307141f84534b78d38429db1f79a380417c1a), recorded as Acta `01M4Y4616JTEX8TRQ2J10X1GKA`. T11.3 guarded surface controls are implemented; focused app/TUI/Web regressions passed in 107.195 s, 0.007 s and 0.112 s with zero strict lint issues, recorded as Acta `01M4Y68YY6QKNVF97KGNP8BEW5`. Focused race passed in app 216.665 s, TUI 1.032 s and Web handlers 1.218 s, recorded as Acta `01M4Y6HYY5E4VEBPX4MXV2CRE7`. Required pre-commit `make verify` passed with zero lint issues, core 2855.905 s, app 198.637 s, TUI 0.073 s and Web 0.094 s; all remaining Web packages passed. Acta `01M4Y9980W7F4RHA53HB2PC1WF`. The owned service exited successfully on unchanged 15-file source `ff3a974203742015969c6f47c4e35dfd1809c55d2dd87a211ad532b2238461e7`. T11.3 is reviewing at [e37e275b49c8a16f1349c916be9f5045b4be811f](https://forge.adrianpk.com/adrianpk/chr/commit/e37e275b49c8a16f1349c916be9f5045b4be811f), recorded as Acta `01M4Y9ACQZTX6W5SSREHCWNX35`. T11.4 setup/reference guidance and executable policy verification are complete. Focused core/app checks passed in 14.878 s and 0.919 s; documentation links, anchors and whitespace passed. Acta `01M4Y9EP3EKTFGEZ30S6PTBGXA`. The owned validation unit exited successfully on unchanged runtime/documentation source `b51a7dea41e2566816729d985c704a99f8699fe7aa527523dfabbd5baa9846f3`. T11.4 is reviewing at [266f652b0f3367a37f0ea88b7e5825c69a2a9f16](https://forge.adrianpk.com/adrianpk/chr/commit/266f652b0f3367a37f0ea88b7e5825c69a2a9f16), recorded as Acta `01M4Y9GT1HTNDJ4EYQX8BH6ABC`. Canonical PR [#385](https://forge.adrianpk.com/adrianpk/chr/pulls/385) retains all four task implementations and their evidence. Final ordered validation is recorded through the controller after PR-reference commits; verified integration is complete. All four configured clean-HEAD checks passed; Acta `01M4YD0NT7KDTBNP3ARA2RNMD8`.

## Slice 12: Live provider certification

Status: delivered

Expected branch: `test/slice-12-native-provider-certification`

Expected PR title: `test(slice-12): certify native integration across supported forge profiles`

Expected report: `ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md`

Prerequisite: verified Slice 11 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T12.1 | [TKT-20261011133135](../ticket/solved/20261011133135-native-integration-t12-1.md) | done | [ae806985740281a4a77ff7e06c1d448f4a6a1a86](https://forge.adrianpk.com/adrianpk/chr/commit/ae806985740281a4a77ff7e06c1d448f4a6a1a86) | Certify Forgejo merge rebase and squash |
| T12.2 | [TKT-20261011133136](../ticket/solved/20261011133136-native-integration-t12-2.md) | done | [a9b780729b8d7d1fab45e0cde03ba02240e8a625](https://forge.adrianpk.com/adrianpk/chr/commit/a9b780729b8d7d1fab45e0cde03ba02240e8a625) | Certify GitHub merge rebase and squash |
| T12.3 | [TKT-20261011133137](../ticket/solved/20261011133137-native-integration-t12-3.md) | done | [d44cdb16a6588446cba41d6a46d2ebfebad4db91](https://forge.adrianpk.com/adrianpk/chr/commit/d44cdb16a6588446cba41d6a46d2ebfebad4db91) | Publish bounded certification and denial matrix |

Required checks:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

PR: [#386](https://forge.adrianpk.com/adrianpk/chr/pulls/386)

Verified merge: `acfb3e445c8056bcc198fef82bbec5a18f132304` at 2026-10-15T12:29:44Z; controller Acta `01M4ZRME308JWKS62JG7DADA4X`.

Report/evidence: [Slice 12 report](../report/slices/native-integration-queue/slice-12-provider-certification.md). T12.1 passed all nine real Forgejo method/publication/contention cases in 51.915 s. The provider-specific merge-policy defect is corrected with denial-preserving regression coverage. Required pre-commit repository verification passed with zero lint issues, core 2747.252 s and app 197.636 s; focused race checks passed in 10.642 s. Acta `01M4YGRAHG25NY28CDGH3JQKKN`. The owned unit exited successfully on unchanged seven-file source. T12.1 is committed and pushed; T12.2 passed all nine GitHub cases under the race detector (221.147 s package; 222.095 s command), with zero strict-lint issues and unchanged source. Acta `01M4YHQ5TQDSP5NXAZ10HHQVTZ`. T12.2 is committed and pushed; a subsequent repeatability run blocked T12.3 on GitHub rebase result mismatch. The later correction and final validation below supersede that development blocker.

The remaining focused race check passed independently on the unchanged retained source: 12.113 s package time and 34.826 s command time. Acta `01M4YKF8DNFZTDZ6YFY5VJ7D0G`. This local development evidence does not resolve live certification or final clean-HEAD validation.

Resolved v1 incompatibility: explicit GitHub source-attribution v2 preserves raw S separately from exact delivered P with complete transaction, history/tree, source and original merged PR proof. The corrective matrix passed all 46 cases; historical v1 evidence remains strict. Pre-commit `make verify` passed with zero lint issues, core 2919.809 s, app 192.959 s and all remaining packages passing. Validation Acta `01M4ZDFXDP7YC3Z1NXWAYS3YHK`; recovery Acta `01M4ZARSAGBGADWYHS9D0N5M9V`. All five final clean-HEAD checks passed; verified integration is complete.

Verified integration: [PR #386](https://forge.adrianpk.com/adrianpk/chr/pulls/386) merged through Rebase + Fast-forward into `dev` at `acfb3e445c8056bcc198fef82bbec5a18f132304` on 2026-10-15T12:29:44Z. All five final controller checks passed on that exact clean HEAD in the recorded order. Validation `4IZI6UNUBQ2KL7PMUYI64E5T2Q`; merge Acta `01M4ZRME308JWKS62JG7DADA4X`.

## Slice 13: Assurance reuse and final closure

Status: in_progress

Expected branch: `test/slice-13-native-integration-assurance`

Expected PR title: `test(slice-13): complete native integration and original workflow assurance`

Expected report: `ops/default/report/slices/native-integration-queue/slice-13-integration-assurance.md`

Prerequisite: verified Slice 12 integration and closed task fanout.

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T13.1 | [TKT-20261008141429](../ticket/reviewing/20261008141429-ticket-native-t12-1.md) | validated | retained: 4f10f1c60de1f164f727d37b4df5c02343ab6e11 | Adapt retained composed workflow assurance |
| T13.2 | [TKT-20261008141430](../ticket/in_progress/20261008141430-ticket-native-t12-2.md) | in_progress | retained: 44e8145aa618edfc51d5efb08330c641b02f7e58 | Adapt retained fault QA entrypoint |
| T13.3 | [TKT-20261008141431](../ticket/ready/20261008141431-ticket-native-t12-3.md) | ready | retained: pending | Reconcile real certification with final behavior |
| T13.4 | [TKT-20261008141432](../ticket/blocked/20261008141432-ticket-native-t12-4.md) | blocked | retained: pending | Run final gates and close both deliveries |

Required checks:

- `make verify`
- `go test -race ./...`
- `make qa-ticket-native-workflow`
- `make qa-native-integration`
- `git diff --check`

PR: pending

Verified merge: pending

Report/evidence: [Slice 13 report](../report/slices/native-integration-queue/slice-13-integration-assurance.md). T13.1 adapts the retained scenarios to production atomic publication and native receipt continuation. Ordered, loose and parallel scenarios passed in 51.739 s; strict lint passed with zero issues. Expanded focused race checks passed in 452.844 s; Acta `01M4ZT1TSNNZ5CB01K2H1VM7S8`. Required pre-commit `make verify` passed in 2939.843 s on unchanged source `b28e8b30e9b4fb93c0d21079d4c21f1af2918fa80d82a5639b775a4ae2c29e67`; core 2907.443 s, app 207.917 s and all remaining packages passed with zero strict-lint issues. Acta `01M4ZWVHWVZ43BM362QWXH8QPH`. T13.1 is validated for its task commit; canonical PR and integration remain pending.

Slice 12 prerequisite verified. T13.1 validation is complete; T13.2 now adapts retained fault QA, followed by T13.3 certification reconciliation. T13.4 retains its final integrated-gate dependency and will close only with controller evidence after integration.

## Final Delivery Gate

- [ ] Every slice has its verified canonical PR merge and delivered report.
- [ ] All task tickets are finalized with preserved retained assurance evidence.
- [ ] All 100 acceptance IDs and 24 invariants have evidence with honest provider limits.
- [ ] The exact integrated full gate passes.
- [ ] Original assurance obligations close with explicit replacement-unit traceability.
- [ ] No unknown submission, unfinished fanout or live owner prevents set closure.
- [ ] Plan/tracker and cycle report close with verified receipts.
