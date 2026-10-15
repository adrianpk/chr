# Native Integration Corrective Delivery

Status: active

Date: 2026-10-11

Approved: 2026-10-11; maintainer approved both specifications and requested Looper execution of this exact corrective set.

Delivery set: native-integration-queue

Spec: [Native Integration Queue and Merger](../spec/stable/native-integration-queue.md)

Model: [Native Integration Queue Model](../spec/stable/native-integration-queue-model.md)

Specification approval: accepted on 2026-10-11

Plan approval: approved on 2026-10-11 under the explicit instruction to process the approved specification through Looper.

Implementation authorization: granted for the complete ordered corrective set and its defined task/PR map.

Active slice: Slice 12 — Live provider certification

Next slice: Slice 13 — Assurance reuse and final closure

Slice strategy: behavior-first

Reason: prove a positive remote transaction and real forge observation before building dependent queue and merger behavior. Deliver types, mappings, implementation and meaningful failure checks together.

Tracker: [Native Integration Tracker](../tracker/native-integration-queue.md)

Full gate: `make verify && go test -race ./... && make qa-ticket-native-workflow && make qa-native-integration && git diff --check`

## Scope and Delegation

Deliver exactly the accepted specification in one set of thirteen ordered slices.
The maintainer delegated implementation, canonical slice PR preparation and Looper-controlled Rebase + Fast-forward merges into `dev`.
Workers do not submit merge mutations. Keep development serial with one Looper worker; the product retains its approved bounded parallel mode.
Do not change `main`, release artifacts, credentials, account access or production branch protection.

All 100 acceptance criteria and 24 invariants are mapped below. A safe all-unsupported provider result cannot complete this set.
The approved profile explicitly permits rewriting the exact CHR-owned source branch. External Git metadata races and unsupported policies retain their documented blockers.

## Entry and Prerequisites

- Commit and push the accepted spec/model, this plan, tracker and task bindings to canonical `dev` before creating Slice 1's branch.
- Baseline verified on 2026-10-11: `67c6edecf7b7b0f32474f04e6e34a7d24f530201`; recheck current remote state on activation.
- Original delivery: eleven canonical slices integrated, no open slice PR, no pending merge; its inactive controller remains blocked at final assurance.
- Preserve primary checkout changes and retained worktrees. Use the clean development controller checkout and isolated implementation worktrees.
- Slices 1 through 12 use 47 new task tickets. Slice 13 reuses the four original assurance tickets, giving 51 execution tasks without duplicate tickets.
- Slice 1 uses owned temporary Git/SSH/HTTPS fixtures only. Do not require a live production repository for its transport proof.
- Slice 2's disposable Forgejo and GitHub repositories, host-owned access roles, full target/source refs and explicit branch-policy profiles are supplied in [Certification Fixtures](../report/native-integration-fixtures.md). The maintainer authorized their creation/configuration and synthetic test pushes, PRs and merges on 2026-10-11. Setup access is verified; publisher certification remains Slice 2 and Slice 12 work.
- Use only the two designated private test repositories. Do not create additional remote repositories, alter account permissions, disable production protections or use production test PRs to fill missing certification authority.
- Stop at Slice 2 if the designation/access gate remains unmet. If either forge cannot support the approved mechanism, stop dependent work and request an approved specification revision.

## Original Assurance Execution Transfer

This approved corrective plan replaces the old final-assurance execution path while preserving its work.
The four original ticket keys become T13.1 through T13.4 aliases; IDs, original task labels, original branch and retained evidence remain provenance.
The current execution binding supersedes their old start gate. Keep them blocked until Slice 13 becomes active; do not reset the historical controller checkpoint or its consumed budgets.

Original T12.1 implementation: `4f10f1c60de1f164f727d37b4df5c02343ab6e11`.
Original T12.2 implementation: `44e8145aa618edfc51d5efb08330c641b02f7e58`.
Retained branch checkpoint: `6100f7d166cc3c61466dcacae05c26baabb2460b`.

Adapt the relevant code once in the Slice 13 task commits, retaining original commit references. Do not blindly replay stale plan/tracker state from that branch.
Original T12.3 uses current certification from Slice 12 and reruns only evidence invalidated by later behavior changes. Original T12.4 retains final exact-candidate and cycle-closure obligations.
The corrective Slice 13 is the replacement assurance integration unit; do not open another PR for the old reserved assurance branch or claim that nonexistent PR merged.
The original plan/tracker may close with verified replacement receipt/PR/report references after their remaining outcomes are satisfied. Report that execution transfer explicitly, preserving the historical controller's blocked checkpoint rather than calling it completed.

## Validation and Controller Contract

Use Go 1.27.1 and repository-owned validation. For each immutable slice candidate, record `make verify`, the exact focused checks below in order, any additional successful focused checks, and `git diff --check` last.
Record them with `looper validate` for the clean PR HEAD. Every relevant head change requires fresh evidence; avoid rerunning broader checks without a changed candidate or unresolved concern.

Use real subprocess/Git/transport fixtures for correctness boundaries. Doubles do not certify a forge.
Slice reports identify exact commit, commands, outcome, material limitations and acceptance evidence.
Use report paths below, following the repository's current `ops/default/report/` location.

Canonical PR bodies use `## Summary`, `## Changes`, `## Validation` in that order.
Summary is one concise paragraph. Changes are plain outcome bullets ending with the immutable slice-report link.
Validation is successful backtick-wrapped command bullets in configured order, with `git diff --check` last.
Complete PR reference commits and push before returning a candidate to Looper.

After each verified canonical merge, update the report, task ticket lifecycle/directories and tracker and activate the next slice.
No worker may bypass a failed check, claim live evidence from mocks, substitute a merge method, invent an extra slice, or expand remote access.

## Ordered Slice Map

| Slice | Name | Exact branch | Expected PR title | Report |
| --- | --- | --- | --- | --- |
| Slice 1 | Guarded Git transport | `feat/slice-1-native-git-transport` | `feat(slice-1): publish exact results through atomic Git transport` | `ops/default/report/slices/native-integration-queue/slice-1-git-transport.md` |
| Slice 2 | Forge publication profiles | `feat/slice-2-native-forge-profiles` | `feat(slice-2): observe native publication through existing forge adapters` | `ops/default/report/slices/native-integration-queue/slice-2-forge-profiles.md` |
| Slice 3 | Durable integration records | `feat/slice-3-native-integration-records` | `feat(slice-3): persist native integration queue records` | `ops/default/report/slices/native-integration-queue/slice-3-integration-records.md` |
| Slice 4 | Canonical queue scheduling | `feat/slice-4-native-queue-scheduling` | `feat(slice-4): schedule integration across canonical target queues` | `ops/default/report/slices/native-integration-queue/slice-4-queue-scheduling.md` |
| Slice 5 | Method preparation and provenance | `feat/slice-5-native-method-preparation` | `feat(slice-5): preserve exact integration history and provenance` | `ops/default/report/slices/native-integration-queue/slice-5-method-preparation.md` |
| Slice 6 | Bounded merger role | `feat/slice-6-native-merger-role` | `feat(slice-6): resolve integration conflicts through a bounded merger` | `ops/default/report/slices/native-integration-queue/slice-6-merger-role.md` |
| Slice 7 | Combined validation and review | `feat/slice-7-native-candidate-review` | `feat(slice-7): revalidate repaired and rebased integration candidates` | `ops/default/report/slices/native-integration-queue/slice-7-candidate-review.md` |
| Slice 8 | Persistent limits and decisions | `feat/slice-8-native-integration-limits` | `feat(slice-8): enforce persistent merger and invocation budgets` | `ops/default/report/slices/native-integration-queue/slice-8-integration-limits.md` |
| Slice 9 | Publication and uncertainty recovery | `feat/slice-9-native-publication-recovery` | `feat(slice-9): reconcile guarded publication without duplicate submission` | `ops/default/report/slices/native-integration-queue/slice-9-publication-recovery.md` |
| Slice 10 | Receipts fanout and closure | `feat/slice-10-native-receipt-closure` | `feat(slice-10): finalize exact integration receipts and task delivery` | `ops/default/report/slices/native-integration-queue/slice-10-receipt-closure.md` |
| Slice 11 | Shared operation surfaces | `feat/slice-11-native-integration-surfaces` | `feat(slice-11): expose native integration progress and controls` | `ops/default/report/slices/native-integration-queue/slice-11-integration-surfaces.md` |
| Slice 12 | Live provider certification | `test/slice-12-native-provider-certification` | `test(slice-12): certify native integration across supported forge profiles` | `ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md` |
| Slice 13 | Assurance reuse and final closure | `test/slice-13-native-integration-assurance` | `test(slice-13): complete native integration and original workflow assurance` | `ops/default/report/slices/native-integration-queue/slice-13-integration-assurance.md` |

## Slice 1: Guarded Git transport

Status: delivered

PR: [#375](https://forge.adrianpk.com/adrianpk/chr/pulls/375)

Expected branch: `feat/slice-1-native-git-transport`

Expected PR title: `feat(slice-1): publish exact results through atomic Git transport`

Expected report: `ops/default/report/slices/native-integration-queue/slice-1-git-transport.md`

Prerequisite: committed approved bootstrap on canonical `dev`.

Acceptance evidence owner: Slice 1 report and its linked retained command/provider evidence.

Acceptance IDs: AC-036, AC-037, AC-038, AC-039, AC-040, AC-041, AC-042, AC-043, AC-051, AC-052, AC-053, AC-054, AC-055, AC-096, AC-097, AC-098

Invariants: INV-07, INV-08, INV-09, INV-10, INV-14, INV-15

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T1.1 | [TKT-20261011133051](../ticket/solved/20261011133051-native-integration-t1-1.md) | Implement the publication port with advertisement comparison, both explicit old/new ref commands, atomic negotiation, bounded report-status handling and a positive local receive-pack path. | A real local repository updates S->P and B->P together; stale advertised values abort before commands and uncertain IO is retained. |
| T1.2 | [TKT-20261011133052](../ticket/solved/20261011133052-native-integration-t1-2.md) | Connect the guarded exchange through SSH using host-owned authentication and argument vectors, with subprocess ownership and bounded IO. | The same positive and stale-ref cases pass over a real SSH fixture without credentials entering evidence. |
| T1.3 | [TKT-20261011133053](../ticket/solved/20261011133053-native-integration-t1-3.md) | Connect the same exchange through smart HTTPS with host credential integration, canonical identity checks and bounded packet/object processing. | The positive and stale-ref cases pass over a real HTTP receive-pack fixture; redirects and partial responses cannot become success. |
| T1.4 | [TKT-20261011133054](../ticket/solved/20261011133054-native-integration-t1-4.md) | Bind the port to immutable prepared candidate identity, fast-forward ancestry, explicit expected refs, original/published source distinction and no-op defenses. | P equal to S, an already-P advertisement, fork sources, invalid refs and non-fast-forward target updates cannot be partially published. |
| T1.5 | [TKT-20261011133055](../ticket/solved/20261011133055-native-integration-t1-5.md) | Add meaningful real-process fault fixtures for advertisement/receive races, atomic rejection, malformed packets and lost response; record protocol/security review. | Both refs participate; external source/target writers are preserved and SSH/HTTPS behavior agrees without an unguarded fallback. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 2: Forge publication profiles

Status: delivered

Expected branch: `feat/slice-2-native-forge-profiles`

Expected PR title: `feat(slice-2): observe native publication through existing forge adapters`

Expected report: `ops/default/report/slices/native-integration-queue/slice-2-forge-profiles.md`

Canonical PR: [#376](https://forge.adrianpk.com/adrianpk/chr/pulls/376)

Verified merge: `3129b5ce312df22287a4e7b78e08bf41f5958377` on `dev`

Prerequisite: verified Slice 1 integration and closed task fanout.

External prerequisite: explicit disposable repository and policy/access profile designation for both existing forges, supplied in [Certification Fixtures](../report/native-integration-fixtures.md); recheck current access before live probes.

Acceptance evidence owner: Slice 2 report and its linked retained command/provider evidence.

Acceptance IDs: AC-032, AC-044, AC-045, AC-046, AC-047, AC-048, AC-061, AC-062, AC-063, AC-080, AC-081, AC-082, AC-083, AC-084, AC-085, AC-086, AC-087, AC-089, AC-090, AC-099

Invariants: INV-08, INV-11, INV-14, INV-15, INV-23

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T2.1 | [TKT-20261011133056](../ticket/solved/20261011133056-native-integration-t2-1.md) | Add opt-in profile configuration, current capability/ownership/policy fingerprints, unsupported reasons and original/published source-aware binding. | Upgrade preserves policy and keeps the profile disabled; incompatible protections, native approval and fork ownership block before publication. |
| T2.2 | [TKT-20261011133057](../ticket/solved/20261011133057-native-integration-t2-2.md) | Implement same-canonical-PR Git application observation for Forgejo and prove one default rebase in a designated disposable repository. | The actual prepared result is reachable and the original PR is observed merged; metadata lag and close/retarget races remain explicit. |
| T2.3 | [TKT-20261011133058](../ticket/solved/20261011133058-native-integration-t2-3.md) | Implement the equivalent existing GitHub adapter contract and prove one default rebase in a designated disposable repository. | The same source/base/result and canonical merged-PR proof holds without depending on a provider merge queue. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 3: Durable integration records

Status: delivered

Expected branch: `feat/slice-3-native-integration-records`

Expected PR title: `feat(slice-3): persist native integration queue records`

Expected report: `ops/default/report/slices/native-integration-queue/slice-3-integration-records.md`

Canonical PR: [#377](https://forge.adrianpk.com/adrianpk/chr/pulls/377)

Verified merge: `ef72c8cbabf06816a93c293161c12d03211979ba` on `dev`, 2026-10-12T02:39:12Z

Prerequisite: verified Slice 2 integration and closed task fanout.

Acceptance evidence owner: Slice 3 report and its linked retained command/provider evidence.

Acceptance IDs: AC-005, AC-006, AC-056, AC-058, AC-063, AC-066, AC-067, AC-068, AC-071, AC-073, AC-074

Invariants: INV-01, INV-02, INV-03, INV-06, INV-16, INV-17, INV-19, INV-24

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T3.1 | [TKT-20261011133059](../ticket/solved/20261011133059-native-integration-t3-1.md) | Add original/published source identities, method manifests and versioned capability/transport references without rewriting historical receipts. | Old schemas remain verifiable and new readers reject missing required proof or unsupported versions. |
| T3.2 | [TKT-20261011133100](../ticket/solved/20261011133100-native-integration-t3-2.md) | Implement idempotent unit entry identity, queue sequence allocation and guarded entry/invocation transitions with the existing file transaction owner. | Duplicate admission consumes one entry; interrupted transitions preserve authoritative records and replay safely. |
| T3.3 | [TKT-20261011133101](../ticket/solved/20261011133101-native-integration-t3-3.md) | Add merger/drift/token reservations and explicit Git-applied/forge-pending submission facts, preserving existing consumed counters. | Restart and candidate replacement cannot refill counters or release unknown reservations. |
| T3.4 | [TKT-20261011133102](../ticket/solved/20261011133102-native-integration-t3-4.md) | Add rebuildable queue queries, strict cross-record validation and additive migration with malformed-byte preservation. | Index loss is recoverable; stale or inconsistent records block mutation without manufacturing execution state. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 4: Canonical queue scheduling

Status: delivered

Expected branch: `feat/slice-4-native-queue-scheduling`

Expected PR title: `feat(slice-4): schedule integration across canonical target queues`

Expected report: `ops/default/report/slices/native-integration-queue/slice-4-queue-scheduling.md`

Canonical PR: [#378](https://forge.adrianpk.com/adrianpk/chr/pulls/378)

Verified merge: `bb3a88888a47040f4ded39807f79623bb86d8ab7` on `dev`

Prerequisite: verified Slice 3 integration and closed task fanout.

Acceptance evidence owner: Slice 4 report and its linked retained command/provider evidence.

Acceptance IDs: AC-001, AC-002, AC-003, AC-004, AC-005, AC-006, AC-007, AC-008, AC-009, AC-010, AC-011, AC-012, AC-013, AC-014, AC-015, AC-058, AC-059

Invariants: INV-03, INV-04, INV-05, INV-06, INV-18, INV-20, INV-21, INV-22

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T4.1 | [TKT-20261011133103](../ticket/solved/20261011133103-native-integration-t4-1.md) | Use the admitted frontier, durable queue sequence and existing prerequisites to select independent units and skip paused/review-blocked entries. | Task and slice order remain intact; an independent eligible successor progresses when another unit waits for a human. |
| T4.2 | [TKT-20261011133104](../ticket/solved/20261011133104-native-integration-t4-2.md) | Route executor, merger, reviewer and closing launches through the same repository pool; bound prepared candidates and prefer finishable integration. | Default mode runs one agent and sustained implementation cannot starve eligible integration. |
| T4.3 | [TKT-20261011133105](../ticket/solved/20261011133105-native-integration-t4-3.md) | Resolve aliases/worktrees/namespaces to common repository/target keys and enforce registered coordination across clones/hosts. | Aliases cannot bypass capacity or locks; unsupported coordinator connectivity blocks before mutable work. |
| T4.4 | [TKT-20261011133106](../ticket/solved/20261011133106-native-integration-t4-4.md) | Exercise subprocess overlap, per-target exclusion, independent targets, dynamic capacity and unresolved publication barriers. | Two independent lanes overlap within bounds; one uncertain target does not block unrelated targets or read-only status. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 5: Method preparation and provenance

Status: delivered

Expected branch: `feat/slice-5-native-method-preparation`

Expected PR title: `feat(slice-5): preserve exact integration history and provenance`

Expected report: `ops/default/report/slices/native-integration-queue/slice-5-method-preparation.md`

Prerequisite: verified Slice 4 integration and closed task fanout.

Acceptance evidence owner: Slice 5 report and its linked retained command/provider evidence.

Acceptance IDs: AC-016, AC-017, AC-018, AC-019, AC-027, AC-028, AC-029, AC-030, AC-031, AC-032, AC-033, AC-041, AC-043, AC-052

Invariants: INV-05, INV-08, INV-09, INV-10, INV-11, INV-12, INV-15

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T5.1 | [TKT-20261011133107](../ticket/solved/20261011133107-native-integration-t5-1.md) | Extend deterministic preparation to preserve B/S parents or the one-parent squash proof with complete source attribution. | Host-verified parents, combined diff and result tree match the immutable prepared candidate. |
| T5.2 | [TKT-20261011133108](../ticket/solved/20261011133108-native-integration-t5-2.md) | Preserve ordered source authors/messages/dates, empty replay mappings and explicit attributed integration-fix commits. | No silent squash, dropped duplicate intent or merge-containing history bypass is accepted. |
| T5.3 | [TKT-20261011133109](../ticket/solved/20261011133109-native-integration-t5-3.md) | Reprepare source-equal candidates within the approved bound and sign rewritten commits through the trusted host where required. | No artificial unrelated commit is inserted; missing certified signing capability blocks before validation/publication. |
| T5.4 | [TKT-20261011133110](../ticket/solved/20261011133110-native-integration-t5-4.md) | Add real repository method/history fixtures and prove that worker lanes and user checkouts are untouched. | Merge/rebase/squash proofs reject altered trees, mappings and provenance, including no-op edge cases. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 6: Bounded merger role

Status: delivered

Canonical PR: [#380](https://forge.adrianpk.com/adrianpk/chr/pulls/380)

Verified merge: `725128da941a5c149a2b0a5d9a6a7065f842823d` at 2026-10-12T19:14:21Z; controller Acta `01M4RRK4KT3HRPZ57T5VAW4FM0`.

Expected branch: `feat/slice-6-native-merger-role`

Expected PR title: `feat(slice-6): resolve integration conflicts through a bounded merger`

Expected report: `ops/default/report/slices/native-integration-queue/slice-6-merger-role.md`

Prerequisite: verified Slice 5 integration and closed task fanout.

Acceptance evidence owner: Slice 6 report and its linked retained command/provider evidence.

Acceptance IDs: AC-020, AC-021, AC-022, AC-023, AC-024, AC-025, AC-026, AC-027, AC-028, AC-029, AC-033, AC-034, AC-035

Invariants: INV-04, INV-05, INV-13, INV-14, INV-18, INV-19, INV-20

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T6.1 | [TKT-20261011133111](../ticket/solved/20261011133111-native-integration-t6-1.md) | Bind intent, target changes, conflict evidence, scope, replay checkpoint and finite allowances to one immutable role input. | Malformed, oversized or stale results and escaped evidence paths are rejected before candidate acceptance. |
| T6.2 | [TKT-20261011133112](../ticket/solved/20261011133112-native-integration-t6-2.md) | Run the configured harness only for actionable conflicts/regressions with shared capacity, local tool permissions and host-owned lifecycle. | The role cannot push, merge, approve, change scope or forge receipts, and timeout preserves ownership of a live child. |
| T6.3 | [TKT-20261011133113](../ticket/solved/20261011133113-native-integration-t6-3.md) | Support per-conflict decisions and correct replay-step repair, declared regeneration, explicit integration fixes and ambiguity requests. | Conflicting product requirements or unsupported binary/submodule/LFS resolution stop with a concrete decision request. |
| T6.4 | [TKT-20261011133114](../ticket/solved/20261011133114-native-integration-t6-4.md) | Compare actual diffs and method proof, commit through the host port and retain attributed evidence of accepted/rejected resolutions. | Blanket side selection, weakened requirements and stale or out-of-scope changes cannot yield authorization. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 7: Combined validation and review

Status: delivered

PR: [#381](https://forge.adrianpk.com/adrianpk/chr/pulls/381)

Expected branch: `feat/slice-7-native-candidate-review`

Expected PR title: `feat(slice-7): revalidate repaired and rebased integration candidates`

Expected report: `ops/default/report/slices/native-integration-queue/slice-7-candidate-review.md`

Prerequisite: verified Slice 6 integration and closed task fanout.

Acceptance evidence owner: Slice 7 report and its linked retained command/provider evidence.

Acceptance IDs: AC-019, AC-021, AC-022, AC-024, AC-026, AC-028, AC-035, AC-048, AC-049, AC-050, AC-051, AC-052, AC-077

Invariants: INV-08, INV-11, INV-12, INV-13, INV-15, INV-19, INV-21

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T7.1 | [TKT-20261011133115](../ticket/solved/20261011133115-native-integration-t7-1.md) | Execute repository-owned combined validation on clean P and retain result-tree/history-bound evidence independently of agent reports. | A textually clean semantic regression fails the combined gate instead of being published. |
| T7.2 | [TKT-20261011133116](../ticket/solved/20261011133116-native-integration-t7-2.md) | Extend review grants with capability, source rewrite, history manifest and required policy fingerprints across every entrypoint. | Any relevant S/B/P/contract/method/policy change invalidates previous publication approval. |
| T7.3 | [TKT-20261011133117](../ticket/solved/20261011133117-native-integration-t7-3.md) | Route bounded combined-gate/review repair through the merger while preserving task evidence and original replay attribution. | Clean passing candidates avoid extra merger calls; repaired candidates receive a new host gate and review. |
| T7.4 | [TKT-20261011133118](../ticket/solved/20261011133118-native-integration-t7-4.md) | Reprepare on a changed target, verify changed source ownership/scope and release publication exclusion during repair. | Serial and parallel runs tolerate external writers without reusing stale combined evidence or silently expanding scope. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 8: Persistent limits and decisions

Status: delivered

Expected branch: `feat/slice-8-native-integration-limits`

Expected PR title: `feat(slice-8): enforce persistent merger and invocation budgets`

Expected report: `ops/default/report/slices/native-integration-queue/slice-8-integration-limits.md`

Prerequisite: verified Slice 7 integration and closed task fanout.

Acceptance evidence owner: Slice 8 report and its linked retained command/provider evidence.

Acceptance IDs: AC-023, AC-034, AC-035, AC-067, AC-068, AC-069, AC-070, AC-071, AC-100

Invariants: INV-04, INV-13, INV-18, INV-19, INV-20

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T8.1 | [TKT-20261011133119](../ticket/solved/20261011133119-native-integration-t8-1.md) | Apply approved default/bounded limits and retained worker timeout across unit and invocation lifecycle. | Restart, rebase and queue re-entry cannot reset consumed attempts or extend deadlines. |
| T8.2 | [TKT-20261011133120](../ticket/solved/20261011133120-native-integration-t8-2.md) | Implement checked invocation-wide reservations across all roles with certified whole-call harness enforcement and compatible usage accounting. | Concurrent calls cannot oversubscribe; output-only metering rejects finite-token mode and unknown usage stays charged. |
| T8.3 | [TKT-20261011133121](../ticket/solved/20261011133121-native-integration-t8-3.md) | Expose candidate-bound ambiguity resolution and explicit positive budget extensions using existing revision guards. | Silence does not approve; stale answers fail and grants preserve all prior consumption. |
| T8.4 | [TKT-20261011133122](../ticket/solved/20261011133122-native-integration-t8-4.md) | Exercise launch/timeout/crash/unknown-usage checkpoints and slot retention with real owned subprocesses. | A surviving child or uncertain usage prevents hidden replacement and refunds, while serial mode remains one active agent. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 9: Publication and uncertainty recovery

Status: delivered

PR: [#383](https://forge.adrianpk.com/adrianpk/chr/pulls/383)

Verified merge: `8853fdb14638c853f641b82c7d19d805b41fa8fe` at 2026-10-14T08:13:24Z; controller Acta `01M4WQJAN50D6ZWVGHEVQ52SSW`.

Expected branch: `feat/slice-9-native-publication-recovery`

Expected PR title: `feat(slice-9): reconcile guarded publication without duplicate submission`

Expected report: `ops/default/report/slices/native-integration-queue/slice-9-publication-recovery.md`

Prerequisite: verified Slice 8 integration and closed task fanout.

Acceptance evidence owner: Slice 9 report and its linked retained command/provider evidence.

Acceptance IDs: AC-036, AC-037, AC-038, AC-039, AC-040, AC-041, AC-042, AC-043, AC-048, AC-049, AC-050, AC-051, AC-052, AC-053, AC-054, AC-055, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-064, AC-065, AC-096, AC-097, AC-098, AC-099

Invariants: INV-06, INV-07, INV-08, INV-09, INV-10, INV-11, INV-15, INV-16, INV-17, INV-18, INV-22, INV-23

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T9.1 | [TKT-20261011133123](../ticket/solved/20261011133123-native-integration-t9-1.md) | Acquire canonical target exclusion, re-observe current candidate/policy, persist intent and dispatch identity before external IO. | Stale callbacks cannot publish and an earlier potentially live dispatcher prevents replacement publication. |
| T9.2 | [TKT-20261011133124](../ticket/solved/20261011133124-native-integration-t9-2.md) | Classify definitive not-applied, unknown and Git-applied outcomes with actual refs and method proof instead of command exit alone. | No lost response triggers another submission or partial ref update. |
| T9.3 | [TKT-20261011133125](../ticket/solved/20261011133125-native-integration-t9-3.md) | Poll bounded forge/Git observations, preserve the target barrier for pending metadata and handle external PR edits or target divergence. | Closed-unmerged and inconsistent metadata cannot create receipts or trigger automatic target rollback. |
| T9.4 | [TKT-20261011133126](../ticket/solved/20261011133126-native-integration-t9-4.md) | Add fault injection around intent, dispatch, server apply, response, restart, lease expiry and stop intent. | All uncertain outcomes retain exclusion and stopped work reconciles any already delivered result before terminal cancellation. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 10: Receipts fanout and closure

Status: delivered

Expected branch: `feat/slice-10-native-receipt-closure`

Expected PR title: `feat(slice-10): finalize exact integration receipts and task delivery`

Expected report: `ops/default/report/slices/native-integration-queue/slice-10-receipt-closure.md`

Canonical PR: [#384](https://forge.adrianpk.com/adrianpk/chr/pulls/384)

Verified merge: `9726200044d15b4e66bacfe1334de4a9410f4bdc` at 2026-10-14T18:36:18Z; controller Acta `01M4XV6YGZZBS8CGG3AJ2EHKT7`.

Prerequisite: verified Slice 9 integration and closed task fanout.

Acceptance evidence owner: Slice 10 report and its linked retained command/provider evidence.

Acceptance IDs: AC-002, AC-004, AC-062, AC-063, AC-064, AC-065, AC-066, AC-073, AC-093, AC-094, AC-095

Invariants: INV-01, INV-02, INV-16, INV-17, INV-21, INV-23, INV-24

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T10.1 | [TKT-20261011133127](../ticket/solved/20261011133127-native-integration-t10-1.md) | Bind the exact candidate/submission/profile, original S, published P, validated B, method history and same canonical merged PR. | An incompatible merge result, missing proof or closed-only PR cannot be accepted. |
| T10.2 | [TKT-20261011133128](../ticket/solved/20261011133128-native-integration-t10-2.md) | Use existing receipt/member journal keys for all slice tasks and independent tickets and expose finalizing until fanout completes. | Interrupted fanout resumes without duplicate completion or missing task receipts. |
| T10.3 | [TKT-20261011133129](../ticket/solved/20261011133129-native-integration-t10-3.md) | Verify receipt reachability under later target advancement and expose destructive remote rewrites without erasing history. | A later descendant keeps valid delivery; removed proof creates a visible reconciliation blocker. |
| T10.4 | [TKT-20261011133130](../ticket/solved/20261011133130-native-integration-t10-4.md) | Require intended receipts, full-gate evidence, settled finalization and no unknown publication before closure. | Original and corrective delivery closure remain separate verified outcomes with preserved historical evidence. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 11: Shared operation surfaces

Status: delivered

PR: [#385](https://forge.adrianpk.com/adrianpk/chr/pulls/385)

Expected branch: `feat/slice-11-native-integration-surfaces`

Expected PR title: `feat(slice-11): expose native integration progress and controls`

Expected report: `ops/default/report/slices/native-integration-queue/slice-11-integration-surfaces.md`

Prerequisite: verified Slice 10 integration and closed task fanout.

Acceptance evidence owner: Slice 11 report and its linked retained command/provider evidence.

Acceptance IDs: AC-023, AC-067, AC-068, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-080, AC-081

Invariants: INV-01, INV-11, INV-14, INV-19, INV-20, INV-22, INV-23

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T11.1 | [TKT-20261011133131](../ticket/solved/20261011133131-native-integration-t11-1.md) | Extend the existing application service with observed queue/phase/role/budget facts and revision-bound review/decision/grant actions. | Surface callers cannot introduce alternative state transitions or optimistic delivery status. |
| T11.2 | [TKT-20261011133132](../ticket/solved/20261011133132-native-integration-t11-2.md) | Show serial/parallel lanes, target queue position, active role and actionable blockers through shared service projections. | Read-only views do not retry, repair or mutate records and a serial user sees one active agent. |
| T11.3 | [TKT-20261011133133](../ticket/solved/20261011133133-native-integration-t11-3.md) | Map existing Web/daemon/manual entrypoints to the same candidate-bound actions and evidence with stale-action rejection. | All surfaces agree on phase and blocker; private tokens, raw lease fields and credentials remain hidden. |
| T11.4 | [TKT-20261011133134](../ticket/solved/20261011133134-native-integration-t11-4.md) | Add user-facing setup/reference guidance for source rewriting, policy support and optional bounded parallel/token operation. | Upgrade preserves old configuration, explicit opt-in is checked before work admission and supported boundaries match the executable. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/app ./internal/tui ./internal/webui -count=1`
- `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`
- `git diff --check`

Verified integration: [PR #385](https://forge.adrianpk.com/adrianpk/chr/pulls/385) merged into `dev` at `12094bddb2d298f6c60fbfcb49f513b6178ff595` on 2026-10-14T23:48:48Z. All four configured checks passed on that exact clean HEAD. Controller merge activity: `01M4YD33D6YKR6AASWSDNZEEV0`; final validation activity: `01M4YD0NT7KDTBNP3ARA2RNMD8`.

## Slice 12: Live provider certification

Status: in_progress

Expected branch: `test/slice-12-native-provider-certification`

Expected PR title: `test(slice-12): certify native integration across supported forge profiles`

Expected report: `ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md`

Prerequisite: verified Slice 11 integration and closed task fanout.

Acceptance evidence owner: Slice 12 report and its linked retained command/provider evidence.

Acceptance IDs: AC-036, AC-037, AC-038, AC-046, AC-047, AC-052, AC-061, AC-063, AC-082, AC-083, AC-084, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-097, AC-099

Invariants: INV-07, INV-08, INV-09, INV-10, INV-11, INV-12, INV-15, INV-16, INV-17, INV-23

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T12.1 | [TKT-20261011133135](../ticket/reviewing/20261011133135-native-integration-t12-1.md) | Run all three exact-result methods against designated disposable Forgejo repositories, including source/target contention and metadata convergence. | Each case records canonical PR, actual Git history/tree, policy/credential role and compatible receipt proof. |
| T12.2 | [TKT-20261011133136](../ticket/ready/20261011133136-native-integration-t12-2.md) | Run the equivalent three cases against designated disposable GitHub repositories using the existing adapter. | Each selected method publishes exact P with the same PR observed merged and no native-queue prerequisite. |
| T12.3 | [TKT-20261011133137](../ticket/ready/20261011133137-native-integration-t12-3.md) | Deliver a repeatable native-integration QA entrypoint and profile evidence for permission denial, incompatible native review, fork/no-atomic and transport failures. | All six positive paths and negative boundaries are explicit; missing live access remains a blocker rather than simulated success. |

### Required Candidate Checks

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

## Slice 13: Assurance reuse and final closure

Status: not_started

Expected branch: `test/slice-13-native-integration-assurance`

Expected PR title: `test(slice-13): complete native integration and original workflow assurance`

Expected report: `ops/default/report/slices/native-integration-queue/slice-13-integration-assurance.md`

Prerequisite: verified Slice 12 integration and closed task fanout.

Acceptance evidence owner: Slice 13 report and its linked retained command/provider evidence.

Acceptance IDs: AC-001, AC-002, AC-003, AC-004, AC-005, AC-006, AC-007, AC-008, AC-009, AC-010, AC-011, AC-012, AC-013, AC-014, AC-015, AC-016, AC-017, AC-018, AC-019, AC-020, AC-021, AC-022, AC-023, AC-024, AC-025, AC-026, AC-027, AC-028, AC-029, AC-030, AC-031, AC-032, AC-033, AC-034, AC-035, AC-036, AC-037, AC-038, AC-039, AC-040, AC-041, AC-042, AC-043, AC-044, AC-045, AC-046, AC-047, AC-048, AC-049, AC-050, AC-051, AC-052, AC-053, AC-054, AC-055, AC-056, AC-057, AC-058, AC-059, AC-060, AC-061, AC-062, AC-063, AC-064, AC-065, AC-066, AC-067, AC-068, AC-069, AC-070, AC-071, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-080, AC-081, AC-082, AC-083, AC-084, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-091, AC-092, AC-093, AC-094, AC-095, AC-096, AC-097, AC-098, AC-099, AC-100

Invariants: INV-01, INV-02, INV-03, INV-04, INV-05, INV-06, INV-07, INV-08, INV-09, INV-10, INV-11, INV-12, INV-13, INV-14, INV-15, INV-16, INV-17, INV-18, INV-19, INV-20, INV-21, INV-22, INV-23, INV-24

| Task | Ticket | Work | Acceptance boundary |
| --- | --- | --- | --- |
| T13.1 | [TKT-20261008141429](../ticket/blocked/20261008141429-ticket-native-t12-1.md) | Reuse original T12.1 at 4f10f1c60de1f164f727d37b4df5c02343ab6e11 and adapt ordered/loose/parallel scenarios to the real publisher. | Composed original and new acceptance passes without duplicate task implementation. |
| T13.2 | [TKT-20261008141430](../ticket/blocked/20261008141430-ticket-native-t12-2.md) | Reuse original T12.2 at 44e8145aa618edfc51d5efb08330c641b02f7e58 and adapt its bounded QA/restart/crash scenarios. | Both QA targets retain meaningful fault evidence and bounded real subprocess coverage. |
| T13.3 | [TKT-20261008141431](../ticket/blocked/20261008141431-ticket-native-t12-3.md) | Fulfill original T12.3 using current Slice 12 certification and rerun cases invalidated by later implementation or policy changes. | Every claimed provider/method has positive real proof matching delivered behavior. |
| T13.4 | [TKT-20261008141432](../ticket/blocked/20261008141432-ticket-native-t12-4.md) | Fulfill original T12.4 by running exact integrated full gates and publishing original/replacement-unit and corrective closure evidence. | All tasks and acceptance obligations close after verified integration; old controller history is preserved without false completion. |

### Required Candidate Checks

- `make verify`
- `go test -race ./...`
- `make qa-ticket-native-workflow`
- `make qa-native-integration`
- `git diff --check`

## Final Gate and Closure

Full gate: `make verify && go test -race ./... && make qa-ticket-native-workflow && make qa-native-integration && git diff --check`

`make qa-native-integration` is delivered by this set before Slice 13; its absence is a blocker, never a skipped success. The final gate runs once on the exact integrated development candidate and repeats only after relevant corrections.
Final live certification must match the delivered implementation and policy. Reuse evidence whose relevant code/profile binding remains valid; rerun invalidated cases.

Closure requires every slice's actual canonical merge, all 51 task outcomes finalized, all 100 acceptance criteria reconciled, no uncertain publication, and verified original/corrective closure evidence.
Use a bounded implementation report for the corrective set and the original cycle's outstanding report. Mark old execution transfer rather than inventing the original reserved PR or controller completion.
After the full gate, change only configured closure artifacts; runtime code or gate changes require fresh validation through authorized delivery.

## Acceptance Coverage

Every acceptance ID maps to primary evidence owners below. Slice 13 owns the composed/final gate and reconciles the full matrix; it does not replace earlier primary implementation evidence.

| Acceptance ID | Primary slices |
| --- | --- |
| AC-001 | 4 |
| AC-002 | 4, 10 |
| AC-003 | 4 |
| AC-004 | 4, 10 |
| AC-005 | 3, 4 |
| AC-006 | 3, 4 |
| AC-007 | 4 |
| AC-008 | 4 |
| AC-009 | 4 |
| AC-010 | 4 |
| AC-011 | 4 |
| AC-012 | 4 |
| AC-013 | 4 |
| AC-014 | 4 |
| AC-015 | 4 |
| AC-016 | 5 |
| AC-017 | 5 |
| AC-018 | 5 |
| AC-019 | 5, 7 |
| AC-020 | 6 |
| AC-021 | 6, 7 |
| AC-022 | 6, 7 |
| AC-023 | 6, 8, 11 |
| AC-024 | 6, 7 |
| AC-025 | 6 |
| AC-026 | 6, 7 |
| AC-027 | 5, 6 |
| AC-028 | 5, 6, 7 |
| AC-029 | 5, 6 |
| AC-030 | 5 |
| AC-031 | 5 |
| AC-032 | 2, 5 |
| AC-033 | 5, 6 |
| AC-034 | 6, 8 |
| AC-035 | 6, 7, 8 |
| AC-036 | 1, 9, 12 |
| AC-037 | 1, 9, 12 |
| AC-038 | 1, 9, 12 |
| AC-039 | 1, 9 |
| AC-040 | 1, 9 |
| AC-041 | 1, 5, 9 |
| AC-042 | 1, 9 |
| AC-043 | 1, 5, 9 |
| AC-044 | 2 |
| AC-045 | 2 |
| AC-046 | 2, 12 |
| AC-047 | 2, 12 |
| AC-048 | 2, 7, 9 |
| AC-049 | 7, 9 |
| AC-050 | 7, 9 |
| AC-051 | 1, 7, 9 |
| AC-052 | 1, 5, 7, 9, 12 |
| AC-053 | 1, 9 |
| AC-054 | 1, 9 |
| AC-055 | 1, 9 |
| AC-056 | 3, 9 |
| AC-057 | 9 |
| AC-058 | 3, 4, 9 |
| AC-059 | 4, 9 |
| AC-060 | 9 |
| AC-061 | 2, 9, 12 |
| AC-062 | 2, 10 |
| AC-063 | 2, 3, 10, 12 |
| AC-064 | 9, 10 |
| AC-065 | 9, 10 |
| AC-066 | 3, 10 |
| AC-067 | 3, 8, 11 |
| AC-068 | 3, 8, 11 |
| AC-069 | 8 |
| AC-070 | 8 |
| AC-071 | 3, 8 |
| AC-072 | 11 |
| AC-073 | 3, 10, 11 |
| AC-074 | 3, 11 |
| AC-075 | 11 |
| AC-076 | 11 |
| AC-077 | 7, 11 |
| AC-078 | 11 |
| AC-079 | 11 |
| AC-080 | 2, 11 |
| AC-081 | 2, 11 |
| AC-082 | 2, 12 |
| AC-083 | 2, 12 |
| AC-084 | 2, 12 |
| AC-085 | 2, 12 |
| AC-086 | 2, 12 |
| AC-087 | 2, 12 |
| AC-088 | 12 |
| AC-089 | 2, 12 |
| AC-090 | 2, 12 |
| AC-091 | 13 |
| AC-092 | 13 |
| AC-093 | 10 |
| AC-094 | 10 |
| AC-095 | 10 |
| AC-096 | 1, 9 |
| AC-097 | 1, 9, 12 |
| AC-098 | 1, 9 |
| AC-099 | 2, 9, 12 |
| AC-100 | 8 |

| Invariant | Primary slices |
| --- | --- |
| INV-01 | 3, 10, 11 |
| INV-02 | 3, 10 |
| INV-03 | 3, 4 |
| INV-04 | 4, 6, 8 |
| INV-05 | 4, 5, 6 |
| INV-06 | 3, 4, 9 |
| INV-07 | 1, 9, 12 |
| INV-08 | 1, 2, 5, 7, 9, 12 |
| INV-09 | 1, 5, 9, 12 |
| INV-10 | 1, 5, 9, 12 |
| INV-11 | 2, 5, 7, 9, 11, 12 |
| INV-12 | 5, 7, 12 |
| INV-13 | 6, 7, 8 |
| INV-14 | 1, 2, 6, 11 |
| INV-15 | 1, 2, 5, 7, 9, 12 |
| INV-16 | 3, 9, 10, 12 |
| INV-17 | 3, 9, 10, 12 |
| INV-18 | 4, 6, 8, 9 |
| INV-19 | 3, 6, 7, 8, 11 |
| INV-20 | 4, 6, 8, 11 |
| INV-21 | 4, 7, 10 |
| INV-22 | 4, 9, 11 |
| INV-23 | 2, 9, 10, 11, 12 |
| INV-24 | 3, 10 |
