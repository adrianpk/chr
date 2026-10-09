# Ticket-Native Work Tracker

Status: active

Date: 2026-10-08

Approved: 2026-10-08; maintainer requested implementation through slice-loop.

Delivery set: ticket-native-workflow

Specification: [Ticket-Native Work Execution](../spec/stable/ticket-native-workflow.md)

Model: [Ticket-Native Work Model](../spec/stable/ticket-native-workflow-model.md)

Plan: [Ticket-Native Work Delivery](../plan/ticket-native-workflow.md)

Specification approval: accepted on 2026-10-08

Plan approval: approved on 2026-10-08

Implementation authorization: granted on 2026-10-08 for the approved delivery set through slice-loop

Planning commit on dev: the bootstrap commit introducing this tracker; resolve with `git log --diff-filter=A --format=%H -- ops/default/tracker/ticket-native-workflow.md`

Active slice: Slice 7 — Ordered set and use-case continuation

Next slice: Slice 8 — Loose-ticket integration loop

Slice strategy: behavior-first

Progress: 6 of 12 slices integrated; 24 of 48 tasks done.

## Current Checkpoint

Slices 1 through 6 are verified integrated through canonical PRs #364–#369.
Slice 6 merged head `22050846d7373395ae776b37bb98df76208dbdf7` passed the controller's
ordered `make verify`, uncached core/app tests, core/app race tests and
`git diff --check` with the retained CLI curation fixture and scoped test timeout.
Its report is delivered and all four task tickets are solved. Acta records the
canonical merge in `01M4F3YG8WEZMVZYGE4EZ7XM80`.
Slice 7 is active for ordered slice/set/use-case traversal, durable continuation,
membership amendments and closing validation. T7.1 is locally validated on the dedicated Slice 7 branch. T7.2 is in progress;
T7.3 and T7.4 remain ready. The slice PR has not been opened.

Only slice-loop may merge this set's canonical slice PRs using Rebase +
Fast-forward into `dev`.

## State and Evidence Rules

- Task rows remain `pending` until their durable ticket key is bound; `ready`
  records a bound, actionable task awaiting its active slice and execution turn.
- During implementation use `in_progress`, `validated` or `blocked` with
  the ticket's observed state, task commit and evidence reference.
- `validated` means local task implementation passed its checks. `done`
  requires verified slice integration and the corresponding ticket transition.
- Slice state advances through `not_started`, `active`, `reviewing` and
  `integrated`. A blocker is recorded with its failed command, reason and
  next-action owner.
- Commit/PR/report fields reflect actual objects only. Expected report paths
  and branch/title metadata below do not imply creation.
- Confirm each merge against canonical `dev`, close its report and tasks,
  then advance the approved delivery set's frontier. A tracker checkbox alone
  never proves delivery.

## Slice Summary

| Slice | Name | Exact branch | Expected PR title | Report | State | Tasks done | PR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Slice 1 | Guarded workflow records | `feat/slice-1-workflow-records` | `feat(slice-1): persist guarded workflow records` | `ops/default/report/slices/ticket-native-workflow/slice-1-workflow-records.md` | integrated | 4/4 | [#364](https://forge.adrianpk.com/adrianpk/chr/pulls/364) |
| Slice 2 | Task-ticket materialization | `feat/slice-2-workflow-materialization` | `feat(slice-2): materialize ordered task tickets` | `ops/default/report/slices/ticket-native-workflow/slice-2-workflow-materialization.md` | integrated | 4/4 | [#365](https://forge.adrianpk.com/adrianpk/chr/pulls/365) |
| Slice 3 | Deterministic work selection | `feat/slice-3-workflow-selection` | `feat(slice-3): select ordered workflow extents` | `ops/default/report/slices/ticket-native-workflow/slice-3-workflow-selection.md` | integrated | 4/4 | [#366](https://forge.adrianpk.com/adrianpk/chr/pulls/366) |
| Slice 4 | Owned isolated task execution | `feat/slice-4-workflow-execution` | `feat(slice-4): execute owned isolated tasks` | `ops/default/report/slices/ticket-native-workflow/slice-4-workflow-execution.md` | integrated | 4/4 | [#367](https://forge.adrianpk.com/adrianpk/chr/pulls/367) |
| Slice 5 | Slice candidates and delivery receipts | `feat/slice-5-workflow-delivery` | `feat(slice-5): verify aggregate slice delivery` | `ops/default/report/slices/ticket-native-workflow/slice-5-workflow-delivery.md` | integrated | 4/4 | [#368](https://forge.adrianpk.com/adrianpk/chr/pulls/368) |
| Slice 6 | Guarded automatic publication | `feat/slice-6-workflow-publication` | `feat(slice-6): publish guarded integration candidates` | `ops/default/report/slices/ticket-native-workflow/slice-6-workflow-publication.md` | integrated | 4/4 | [#369](https://forge.adrianpk.com/adrianpk/chr/pulls/369) |
| Slice 7 | Ordered set and use-case continuation | `feat/slice-7-workflow-continuation` | `feat(slice-7): continue ordered sets and use cases` | `ops/default/report/slices/ticket-native-workflow/slice-7-workflow-continuation.md` | active | 0/4 | pending |
| Slice 8 | Loose-ticket integration loop | `feat/slice-8-workflow-loose-tickets` | `feat(slice-8): integrate filtered loose tickets` | `ops/default/report/slices/ticket-native-workflow/slice-8-workflow-loose-tickets.md` | not_started | 0/4 | pending |
| Slice 9 | Independent parallel runners | `feat/slice-9-workflow-concurrency` | `feat(slice-9): coordinate independent parallel runners` | `ops/default/report/slices/ticket-native-workflow/slice-9-workflow-concurrency.md` | not_started | 0/4 | pending |
| Slice 10 | Shared operation controls | `feat/slice-10-workflow-controls` | `feat(slice-10): expose shared workflow controls` | `ops/default/report/slices/ticket-native-workflow/slice-10-workflow-controls.md` | not_started | 0/4 | pending |
| Slice 11 | Defaults and authority transition | `feat/slice-11-workflow-defaults` | `feat(slice-11): enable versioned ticket workflow defaults` | `ops/default/report/slices/ticket-native-workflow/slice-11-workflow-defaults.md` | not_started | 0/4 | pending |
| Slice 12 | End-to-end workflow assurance | `test/slice-12-workflow-assurance` | `test(slice-12): verify ticket workflow end to end` | `ops/default/report/slices/ticket-native-workflow/slice-12-workflow-assurance.md` | not_started | 0/4 | pending |

## Slice 1: Guarded workflow records

Status: integrated

Prerequisite: satisfied by the approved bootstrap commit, 48 ticket bindings and explicit activation.

Expected branch: `feat/slice-1-workflow-records`

Expected PR title: `feat(slice-1): persist guarded workflow records`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-1-workflow-records.md`

Plan contract: [Slice 1](../plan/ticket-native-workflow.md#slice-1-guarded-workflow-records)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T1.1 | [TKT-20261008141345](../ticket/solved/20261008141345-ticket-native-t1-1.md) | done | 7fa91666a7c3b473dc1651994863413d53217f38 | Define versioned task/set/slice/receipt identities, fields and lifecycle validation in native core types. |
| T1.2 | [TKT-20261008141346](../ticket/solved/20261008141346-ticket-native-t1-2.md) | done | c28ecb7ddf5f595b6f631db9665ce524b8a5a117 | Implement revision-guarded canonical and adopted-Markdown read/write with unknown-field preservation. |
| T1.3 | [TKT-20261008141347](../ticket/solved/20261008141347-ticket-native-t1-3.md) | done | 2f51be25a0afe1231cffe74ddeccdb8af89a685e | Implement process-safe mutation locking and journaled multi-file transitions with guarded replay. |
| T1.4 | [TKT-20261008141348](../ticket/solved/20261008141348-ticket-native-t1-4.md) | done | 9db5940c692003107398db9489c9a263fb6b7144 | Extend rebuildable SQLite queries/migrations and record/file bounds. |

### Validation

Status: all required controller checks passed against the verified integrated head.

Expected focused checks: `go test ./internal/core -count=1` and
`go test -race ./internal/core -count=1`.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A02, A22; foundations for A24.

### Delivery Evidence

- PR: [#364](https://forge.adrianpk.com/adrianpk/chr/pulls/364); verified merged into `dev`.
- Merge commit: `b041521e94d46ee46aa35a9133e8746f31a27328`.
- Report status: delivered; [Slice 1 report](../report/slices/ticket-native-workflow/slice-1-workflow-records.md).
- Task commits: T1.1 `7fa91666a7c3b473dc1651994863413d53217f38`; T1.2 `c28ecb7ddf5f595b6f631db9665ce524b8a5a117`; T1.3 `2f51be25a0afe1231cffe74ddeccdb8af89a685e`; T1.4 `9db5940c692003107398db9489c9a263fb6b7144`. All four tickets are solved with resolution fixed.
- Validation evidence: controller checks for `b041521e94d46ee46aa35a9133e8746f31a27328` passed `make verify`, `go test ./internal/core -count=1`, `go test -race ./internal/core -count=1` and `git diff --check` in that order.
- Blockers: none observed.

## Slice 2: Task-ticket materialization

Status: integrated

Prerequisite: satisfied by verified PR #364 integration into dev.

Expected branch: `feat/slice-2-workflow-materialization`

Expected PR title: `feat(slice-2): materialize ordered task tickets`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-2-workflow-materialization.md`

Plan contract: [Slice 2](../plan/ticket-native-workflow.md#slice-2-task-ticket-materialization)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T2.1 | [TKT-20261008141349](../ticket/solved/20261008141349-ticket-native-t2-1.md) | done | 8add9b52a3154dd48b4b1b881b1c2a1d9c8deba1 | Extend plan-generation contracts to distinguish set, slice and task identity/order. |
| T2.2 | [TKT-20261008141350](../ticket/solved/20261008141350-ticket-native-t2-2.md) | done | 93a19da5472e6ca323d671ea1e0ffbd54cd6be59 | Persist one ticket per task and bind task-owned membership to set-owned slice contracts. |
| T2.3 | [TKT-20261008141351](../ticket/solved/20261008141351-ticket-native-t2-3.md) | done | ae26d0e520ef011c466cd0aae5e49e56b3c4c1e4 | Add materialization receipts and idempotent recovery after partial generation. |
| T2.4 | [TKT-20261008141352](../ticket/solved/20261008141352-ticket-native-t2-4.md) | done | 1585ee89b21a51c565fee54c4f87d05fddab4dac | Add explicit quiescent adoption for selected active legacy plans/tickets and update generation guidance. |

### Validation

Status: all four required controller checks passed against the verified integrated head.

Expected focused checks: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A01, A02.

### Delivery Evidence

- PR: [#365](https://forge.adrianpk.com/adrianpk/chr/pulls/365); verified merged into `dev`.
- Merge commit: `7d4008d32eccefcd3abad5e80955fb7f19110b6a`.
- Report status: delivered; [Slice 2 report](../report/slices/ticket-native-workflow/slice-2-workflow-materialization.md).
- Task commits: T2.1 `8add9b52a3154dd48b4b1b881b1c2a1d9c8deba1`; T2.2 `93a19da5472e6ca323d671ea1e0ffbd54cd6be59`; T2.3 `ae26d0e520ef011c466cd0aae5e49e56b3c4c1e4`; T2.4 `1585ee89b21a51c565fee54c4f87d05fddab4dac`. All four tickets are solved with resolution fixed.
- Validation evidence: local materialization/adoption checks passed. Initial required race validation at `7f695ded2be57b115a84b5f047621466bcee3e58` failed the external-model-dependent CLI dump fixture. Its isolated check passed with the owned deterministic curation fixture recorded in the report. Controller evidence at `7d4008d32eccefcd3abad5e80955fb7f19110b6a` passed `make verify`, `go test ./internal/core ./internal/app -count=1`, `go test -race ./internal/core ./internal/app -count=1` and `git diff --check` in that order with the documented owned curation fixture.
- Blockers: none observed.

## Slice 3: Deterministic work selection

Status: integrated

Prerequisite: satisfied by verified PR #365 integration into dev.

Expected branch: `feat/slice-3-workflow-selection`

Expected PR title: `feat(slice-3): select ordered workflow extents`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-3-workflow-selection.md`

Plan contract: [Slice 3](../plan/ticket-native-workflow.md#slice-3-deterministic-work-selection)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T3.1 | [TKT-20261008141353](../ticket/solved/20261008141353-ticket-native-t3-1.md) | done | 559000591223cdc0ce5df4c3ffd6dc0f367124f6 | Add exact ticket/slice/set/use-case/loose selectors and bounded preview results. |
| T3.2 | [TKT-20261008141354](../ticket/solved/20261008141354-ticket-native-t3-2.md) | done | 785a1d90fae0f62ea6a20f14d8efbf8a34257cfa | Resolve set-local/task-local order and homogeneous use-case set order. |
| T3.3 | [TKT-20261008141355](../ticket/solved/20261008141355-ticket-native-t3-3.md) | done | 2404d19149774e675e3c161ae2da8020d186e033 | Validate implicit and explicit prerequisites, including target delivery proof. |
| T3.4 | [TKT-20261008141356](../ticket/solved/20261008141356-ticket-native-t3-4.md) | done | 7d4a35032ad432ebf307140ec692038031665262 | Preserve native classification/rank rules and implement loose ownership exclusion and thresholds. |

### Validation

Status: all required controller checks passed against the verified integrated head.

Expected focused checks: `go test ./internal/core ./internal/app -count=1`.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A04, A05, A07; query bounds in A24.

### Delivery Evidence

- PR: [#366](https://forge.adrianpk.com/adrianpk/chr/pulls/366); verified merged into `dev`.
- Merge commit: `b39a0051818b1fa9c08b88cf2e79658b173cf16f`.
- Report status: delivered; [Slice 3 report](../report/slices/ticket-native-workflow/slice-3-workflow-selection.md).
- Task commits: T3.1 `559000591223cdc0ce5df4c3ffd6dc0f367124f6`; T3.2 `785a1d90fae0f62ea6a20f14d8efbf8a34257cfa`; T3.3 `2404d19149774e675e3c161ae2da8020d186e033`; T3.4 `7d4a35032ad432ebf307140ec692038031665262`. All four tickets are solved with resolution fixed.
- Validation evidence: `make verify` and focused real-files/Git/concurrency selection race checks passed. The retained owned CLI curation fixture is documented in the report. Controller evidence at `b39a0051818b1fa9c08b88cf2e79658b173cf16f` passed `make verify`, `go test ./internal/core ./internal/app -count=1` and `git diff --check` in that order.
- Blockers: none observed.

## Slice 4: Owned isolated task execution

Status: integrated

Prerequisite: satisfied by verified PR #366 integration into dev.

Expected branch: `feat/slice-4-workflow-execution`

Expected PR title: `feat(slice-4): execute owned isolated tasks`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-4-workflow-execution.md`

Plan contract: [Slice 4](../plan/ticket-native-workflow.md#slice-4-owned-isolated-task-execution)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T4.1 | [TKT-20261008141357](../ticket/solved/20261008141357-ticket-native-t4-1.md) | done | 33e4cbbfb32ab9dd27592ba5a7fe03276848e2b2 | Implement guarded invocation/attempt identity, exclusive claims and fencing generations under the existing serial lock. |
| T4.2 | [TKT-20261008141358](../ticket/solved/20261008141358-ticket-native-t4-2.md) | done | a4d454df0264d7bedce0452296243a5324a027c0 | Execute against a retained isolated cumulative lane branch with frozen inputs and host validation. |
| T4.3 | [TKT-20261008141359](../ticket/solved/20261008141359-ticket-native-t4-3.md) | done | 8a3a0fbf75abc200fdff27b4419aa32bcff22d67 | Persist stop intent, lease/process ownership, attempt budgets and safe replacement. |
| T4.4 | [TKT-20261008141400](../ticket/solved/20261008141400-ticket-native-t4-4.md) | done | df517f43250d1f7e52e7599984c9e33d0b920850 | Expose the shared manual claim/checkpoint/evidence-reconciliation protocol. |

### Validation

Status: all four ordered controller checks passed on the verified integrated head.

Required checks, in order: `make verify`,
`go test ./internal/core ./internal/app -count=1`,
`go test -race ./internal/core ./internal/app -count=1`, `git diff --check`.
Real worker subprocess and linked-worktree recovery fixtures are included.
The retained CLI curation fixture is documented in the Slice 2 report.

Acceptance: A06, A09, A12, A17, A21; execution budgets in A24.

### Delivery Evidence

- PR: [#367](https://forge.adrianpk.com/adrianpk/chr/pulls/367); verified merged into `dev`.
- Merge commit: `517fdbb3322b4e02b951e5301b8f6c8e04740f57`.
- Report introduction: `17733105712a59a758a33d25ff13631e31ee40e9`; canonical full immutable introduction link recorded in the PR.
- Report status: delivered; [Slice 4 report](../report/slices/ticket-native-workflow/slice-4-workflow-execution.md).
- Task lifecycle: all four tickets are solved with resolution fixed; task rows are done.
- Controller evidence: exact head `517fdbb3322b4e02b951e5301b8f6c8e04740f57` passed `make verify`, `go test ./internal/core ./internal/app -count=1`, `go test -race ./internal/core ./internal/app -count=1` and `git diff --check` in that order.
- Validation evidence: task `make verify`, focused real-files/Git/process/manual tests, focused core/app race checks and `git diff --check` passed. Exact successful commands and the corrected Linux process-observation race are recorded in the report and task tickets.
- Acta task events: T4.1 `01M4E972DPE5DM4Q7ATJEBCRAK`; T4.2 `01M4ECD42K1X3N30FQXME60JCV`; T4.3 `01M4EESF2QHEVH0KPD8GHN6SB0`; T4.4 `01M4EG734Q2TMJWEH6N6V5GX2M`.
- PR activity: Acta opened event `01M4EGDF7ME9BN2PS22735YBVH`; merged event `01M4EGRYV77Q7CJ4ZJ00TE0JS0`.
- Blockers: none observed.

## Slice 5: Slice candidates and delivery receipts

Status: integrated

Prerequisite: satisfied by verified PR #367 integration into dev.

Expected branch: `feat/slice-5-workflow-delivery`

Expected PR title: `feat(slice-5): verify aggregate slice delivery`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-5-workflow-delivery.md`

Plan contract: [Slice 5](../plan/ticket-native-workflow.md#slice-5-slice-candidates-and-delivery-receipts)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T5.1 | [TKT-20261008141401](../ticket/solved/20261008141401-ticket-native-t5-1.md) | done | 0bbbd64e691ee6832d495a2b8d4de3b920895e47 | Run remaining tasks cumulatively and host-validate aggregate slice acceptance on an immutable tree. |
| T5.2 | [TKT-20261008141402](../ticket/solved/20261008141402-ticket-native-t5-2.md) | done | 2c6c9081403013bad4916a1a2734059274ccda22 | Create/reuse one canonical PR and retain branch/candidate/PR bindings. |
| T5.3 | [TKT-20261008141403](../ticket/solved/20261008141403-ticket-native-t5-3.md) | done | fb9ed4247ca27e18b3cb365590ad6033b464d166 | Verify external integration through canonical receipts and method-specific target/tree/history proof. |
| T5.4 | [TKT-20261008141404](../ticket/solved/20261008141404-ticket-native-t5-4.md) | done | 3c09f32119bc0eb9b376663b0d03709f0f948628 | Journal shared receipt creation and task-status finalization. |

### Validation

Status: all task checks and the final four ordered controller checks passed on
`6cfaedd9c12f7a6dcfc562804917c32c5f6daaca`.

Successful T5.1 checks: `make verify` with the retained Slice 2 CLI curation
fixture; `go test -race ./internal/core ./internal/app -run '^(TestWorkflowSlice|TestServiceWorkflowSlice)' -count=1`;
`git diff --check`. Real fixtures cover aggregate retries, proof/contract/Git
mutation, pause/cancel, fenced replacement, budgets, complete membership and
aggregate/repair journal replay after host death.

Successful T5.2 checks: `make verify` with the retained CLI curation fixture;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowSlicePR|TestWorkflowPR|TestServiceWorkflowSlicePR)' -count=1`;
`git diff --check`. Real fixtures cover lost creation responses, host death after
creation and between review journal writes, concurrent callers, closed/mismatched
PRs, same-PR repair and bounded configured provider IO. Read-only Forgejo API
checks confirmed canonical routing and its branch commit field.

Successful T5.3 checks: `make verify` with the retained CLI curation fixture;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowIntegration|TestServiceWorkflowIntegration)' -count=1`;
`git diff --check`. Real bare remotes verify explicit merge/rebase/squash trees,
parents and ordered history, including moved bases and later target commits.
Wrong identities, wrong target history, worker claims, forged intermediate trees,
changed candidates and edited immutable proof are rejected.

Successful T5.4 checks: `make verify` with the retained CLI curation fixture;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowFinalization|TestServiceWorkflowFinalization)' -count=1`;
`git diff --check`. Real fixtures cover host death at receipt/task/set/run writes,
replay once, concurrent callers, reachable later target commits, unchanged
remaining set members and central service finalization.

Expected focused checks: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A01, A08, A13, A22.

### Delivery Evidence

- PR: [#368](https://forge.adrianpk.com/adrianpk/chr/pulls/368); merged through Rebase + Fast-forward into dev.
- Merge commit: `6cfaedd9c12f7a6dcfc562804917c32c5f6daaca`; canonical dev and local dev match.
- Report introduction: `c5ed935a203f041c1da30761768b8466cf93a955`; full immutable report link recorded in the PR.
- Report status: delivered; [Slice 5 report](../report/slices/ticket-native-workflow/slice-5-workflow-delivery.md).
- All four task implementation commits remain recorded in their done rows and solved tickets.
- Acta canonical merge event: `01M4ESBJ8W5EGFTJR23A3CCXN3`.
- Final clean-head controller evidence: `checks/6cfaedd9c12f7a6dcfc562804917c32c5f6daaca.json` and matching log in the controller state directory; the exact four ordered commands passed.
- Blockers: none observed.

## Slice 6: Guarded automatic publication

Status: integrated

Prerequisite: satisfied by verified PR #368 integration into dev.

Expected branch: `feat/slice-6-workflow-publication`

Expected PR title: `feat(slice-6): publish guarded integration candidates`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-6-workflow-publication.md`

Plan contract: [Slice 6](../plan/ticket-native-workflow.md#slice-6-guarded-automatic-publication)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T6.1 | [TKT-20261008141405](../ticket/solved/20261008141405-ticket-native-t6-1.md) | done | 085747eb4bc083318d6673b80ec7d9e437e45f8f | Bind review decisions, bounded repairs and exact-candidate manual approval to current policy and immutable inputs. |
| T6.2 | [TKT-20261008141406](../ticket/solved/20261008141406-ticket-native-t6-2.md) | done | 497af614315fe80ce69c0f43c216f1a81d15200d | Prepare and validate configured merge/rebase/squash results and enforce atomic source/base publication guards. |
| T6.3 | [TKT-20261008141407](../ticket/solved/20261008141407-ticket-native-t6-3.md) | done | b9e728a265668bac4f94fe92580f71eafaa3e6e1 | Journal submission before provider IO and retain target exclusion for uncertain responses. |
| T6.4 | [TKT-20261008141408](../ticket/solved/20261008141408-ticket-native-t6-4.md) | done | 84875e1775a89efd2bdfd0eeb782eda1c2bee8a6 | Reconcile target movement, external delivery and stale candidate evidence through the shared operation. |

### Validation

Status: all four task checks and all four ordered controller checks passed on clean head `22050846d7373395ae776b37bb98df76208dbdf7`.

Successful T6.1 checks: `make verify` with the retained CLI curation fixture;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowReview|TestServiceWorkflowReview)' -count=1`;
`git diff --check`. Real fixtures cover policy/candidate/target revocation, exact
manual approval, mode stops, three persistent repairs, role cancellation,
concurrent decisions and process-death recovery.

Successful T6.2 checks: `make verify` with the retained CLI curation fixture;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowPreparation|TestWorkflowReview|TestServiceWorkflowPreparation|TestServiceWorkflowReview)' -count=1`;
`git diff --check`. Real fixtures cover combined merge/rebase/squash gates,
exact approval, moved target, immutable evidence, conflicts, stop intent,
concurrent callers and journal recovery. Current source-only native provider
paths report the required guard blocker without fallback.

Successful T6.3 checks: `make verify` with the retained CLI curation fixture;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowSubmission|TestServiceWorkflowSubmission)' -count=1`;
`git diff --check`. Actual bare-Git transactions protect source and base for all
three methods. Fixtures cover lost responses, host death at intent/projection/
provider/response/confirmation boundaries, cancellation, concurrent callers,
linked namespace aliases, consumed budgets and corrupt or missing journals.

Successful T6.4 checks: `make verify` with the retained CLI curation fixture
and scoped Go test timeout of 30 minutes;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowPublication|TestServiceWorkflowPublication)' -count=1`;
`git diff --check`. Real fixtures cover policy publication, moved bases, external
merge/rebase/squash, paused/cancelled delivery, late responses, corrupt evidence,
repaired candidates, receipt fan-out death and restart reuse.

Expected focused checks: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`. Exercise real target
movement and controlled provider responses; live-provider claims require Slice 12
evidence on designated disposable repositories.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A11, A13, A14, A15, A21; publication budgets in A24.

### Delivery Evidence

- PR: [#369](https://forge.adrianpk.com/adrianpk/chr/pulls/369); merged into `dev` through Rebase + Fast-forward.
- Merge commit: `22050846d7373395ae776b37bb98df76208dbdf7`.
- Report status: delivered; [complete Slice 6 report](../report/slices/ticket-native-workflow/slice-6-workflow-publication.md).
- Task commits: T6.1 `085747eb4bc083318d6673b80ec7d9e437e45f8f`; T6.2 `497af614315fe80ce69c0f43c216f1a81d15200d`; T6.3 `b9e728a265668bac4f94fe92580f71eafaa3e6e1`; T6.4 `84875e1775a89efd2bdfd0eeb782eda1c2bee8a6`. All four tickets are solved.
- Acta task events: T6.1 `01M4EVSNQG6RK6XRXA0YAKW199`; T6.2 `01M4EXHWH60EG9YSZ35DF9TFQ2`; T6.3 `01M4EZMN5MW4KQC7ANKETSBPFV`; T6.4 `01M4F1W257AMYZHJ46X250RD7J`.
- Final clean-head controller evidence: `checks/22050846d7373395ae776b37bb98df76208dbdf7.json` and matching log in the controller state directory; the exact four ordered commands passed.
- Acta canonical merge event: `01M4F3YG8WEZMVZYGE4EZ7XM80`.
- Blockers: none observed.

## Slice 7: Ordered set and use-case continuation

Status: active

Prerequisite: satisfied by verified PR #369 integration into dev.

Expected branch: `feat/slice-7-workflow-continuation`

Expected PR title: `feat(slice-7): continue ordered sets and use cases`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-7-workflow-continuation.md`

Plan contract: [Slice 7](../plan/ticket-native-workflow.md#slice-7-ordered-set-and-use-case-continuation)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T7.1 | [TKT-20261008141409](../ticket/in_progress/20261008141409-ticket-native-t7-1.md) | validated | 96576db5caf4b9ecf115f4a9b3490e53e151b1de | Traverse slice/set/use-case extents through the verified integration barrier. |
| T7.2 | [TKT-20261008141410](../ticket/in_progress/20261008141410-ticket-native-t7-2.md) | in_progress | pending | Persist equivalent-request identity, confirmed cursor and bounded continuation checkpoints. |
| T7.3 | [TKT-20261008141411](../ticket/ready/20261008141411-ticket-native-t7-3.md) | ready | pending | Implement frozen/default and explicit later-admission rules, quiescent amendments and set reopening. |
| T7.4 | [TKT-20261008141412](../ticket/ready/20261008141412-ticket-native-t7-4.md) | ready | pending | Confirm closing validation against all actual members and exact target/plan snapshots. |

### Validation

Status: T7.1 pre-commit checks passed; whole-slice final validation is pending.

Successful T7.1 checks: `make verify`;
`go test -race ./internal/core ./internal/app -run '^(TestWorkflowFrontier|TestServiceWorkflowFrontier)' -count=1`;
`git diff --check`. The repository gate used the retained CLI curation fixture
and a scoped uncached 30-minute Go test timeout. Real two-PR Git fixtures,
receipt/set barriers, unsliced exclusions, stopped/corrupt frontiers and both
journal replacements passed. Acta: `01M4F5ND4089A1DT2QV8QKMWGF`.

Expected focused checks: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`; a multi-slice real-Git
scenario must create separate canonical PR fixture identities and integrate in order.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A03, A04, A07, A13, A19, A20; checkpoint bounds in A24.

### Delivery Evidence

- PR: pending.
- Merge commit: pending.
- Report status: not created.
- Task commits and ticket lifecycle evidence: pending.
- Validation evidence: pending.
- Blockers: none observed; implementation has not started.

## Slice 8: Loose-ticket integration loop

Status: not_started

Prerequisite: Slice 7 verified integrated into dev.

Expected branch: `feat/slice-8-workflow-loose-tickets`

Expected PR title: `feat(slice-8): integrate filtered loose tickets`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-8-workflow-loose-tickets.md`

Plan contract: [Slice 8](../plan/ticket-native-workflow.md#slice-8-loose-ticket-integration-loop)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T8.1 | [TKT-20261008141413](../ticket/ready/20261008141413-ticket-native-t8-1.md) | ready | pending | Compose exact independent-ticket and loose extents with the common candidate/integration lifecycle. |
| T8.2 | [TKT-20261008141414](../ticket/ready/20261008141414-ticket-native-t8-2.md) | ready | pending | Reselect eligible loose work after each integration without claiming slice-owned tickets. |
| T8.3 | [TKT-20261008141415](../ticket/ready/20261008141415-ticket-native-t8-3.md) | ready | pending | Preserve bounded selection, blocker-stop, duplicate-request and resume behavior. |
| T8.4 | [TKT-20261008141416](../ticket/ready/20261008141416-ticket-native-t8-4.md) | ready | pending | Map legacy `run-all` and single-ticket entrypoints according to explicit authority mode. |

### Validation

Status: not run.

Expected focused checks: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A05, A07, A13, A19.

### Delivery Evidence

- PR: pending.
- Merge commit: pending.
- Report status: not created.
- Task commits and ticket lifecycle evidence: pending.
- Validation evidence: pending.
- Blockers: none observed; implementation has not started.

## Slice 9: Independent parallel runners

Status: not_started

Prerequisite: Slice 8 verified integrated into dev.

Expected branch: `feat/slice-9-workflow-concurrency`

Expected PR title: `feat(slice-9): coordinate independent parallel runners`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-9-workflow-concurrency.md`

Plan contract: [Slice 9](../plan/ticket-native-workflow.md#slice-9-independent-parallel-runners)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T9.1 | [TKT-20261008141417](../ticket/ready/20261008141417-ticket-native-t9-1.md) | ready | pending | Introduce repository-wide process reservations and validated capacity/queue admission across worktrees/namespaces. |
| T9.2 | [TKT-20261008141418](../ticket/ready/20261008141418-ticket-native-t9-2.md) | ready | pending | Enforce exclusive task/slice/set/use-case frontier claims and fenced result application on every start path. |
| T9.3 | [TKT-20261008141419](../ticket/ready/20261008141419-ticket-native-t9-3.md) | ready | pending | Serialize publication per canonical remote target, including uncertain outcomes and external target movement. |
| T9.4 | [TKT-20261008141420](../ticket/ready/20261008141420-ticket-native-t9-4.md) | ready | pending | Replace the coarse lock only after subprocess/crash/race acceptance passes and document the ownership boundary. |

### Validation

Status: not run.

Expected focused checks: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`; subprocess barriers must
prove actual overlap/exclusion, not infer it from a race-detector pass.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A09, A10, A11, A12; capacity/admission bounds in A24.

### Delivery Evidence

- PR: pending.
- Merge commit: pending.
- Report status: not created.
- Task commits and ticket lifecycle evidence: pending.
- Validation evidence: pending.
- Blockers: none observed; implementation has not started.

## Slice 10: Shared operation controls

Status: not_started

Prerequisite: Slice 9 verified integrated into dev.

Expected branch: `feat/slice-10-workflow-controls`

Expected PR title: `feat(slice-10): expose shared workflow controls`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-10-workflow-controls.md`

Plan contract: [Slice 10](../plan/ticket-native-workflow.md#slice-10-shared-operation-controls)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T10.1 | [TKT-20261008141421](../ticket/ready/20261008141421-ticket-native-t10-1.md) | ready | pending | Expose proposed CLI extents, preview/status/continue and exact-bound approval/reconciliation controls. |
| T10.2 | [TKT-20261008141422](../ticket/ready/20261008141422-ticket-native-t10-2.md) | ready | pending | Add TUI controls and progress/wait/blocker rendering over shared operations. |
| T10.3 | [TKT-20261008141423](../ticket/ready/20261008141423-ticket-native-t10-3.md) | ready | pending | Add Web operation controls, guarded writes and resumable progress with visible failure envelopes. |
| T10.4 | [TKT-20261008141424](../ticket/ready/20261008141424-ticket-native-t10-4.md) | ready | pending | Wire daemon resume, pause/cancel and manual status parity; document Telegram's notification-only boundary. |

### Validation

Status: not run.

Expected focused checks: `go test ./internal/app ./internal/tui ./internal/webui -count=1`
and `go test -race ./internal/app ./internal/tui ./internal/webui -count=1`;
extend repository-owned browser smoke validation for the new controls.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A09, A17, A21, A23.

### Delivery Evidence

- PR: pending.
- Merge commit: pending.
- Report status: not created.
- Task commits and ticket lifecycle evidence: pending.
- Validation evidence: pending.
- Blockers: none observed; implementation has not started.

## Slice 11: Defaults and authority transition

Status: not_started

Prerequisite: Slice 10 verified integrated into dev.

Expected branch: `feat/slice-11-workflow-defaults`

Expected PR title: `feat(slice-11): enable versioned ticket workflow defaults`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-11-workflow-defaults.md`

Plan contract: [Slice 11](../plan/ticket-native-workflow.md#slice-11-defaults-and-authority-transition)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T11.1 | [TKT-20261008141425](../ticket/ready/20261008141425-ticket-native-t11-1.md) | ready | pending | Deliver version marker, approved configuration keys, bounds and effective-value/source reporting. |
| T11.2 | [TKT-20261008141426](../ticket/ready/20261008141426-ticket-native-t11-2.md) | ready | pending | Align standard/minimal profiles, bootstrap and fallback defaults for new projects. |
| T11.3 | [TKT-20261008141427](../ticket/ready/20261008141427-ticket-native-t11-3.md) | ready | pending | Upgrade selected projects with prior-policy snapshots and explicit authority transition. |
| T11.4 | [TKT-20261008141428](../ticket/ready/20261008141428-ticket-native-t11-4.md) | ready | pending | Export/repair trackers one way and publish the complete operation/configuration/manual guidance. |

### Validation

Status: not run.

Expected focused checks: `go test ./internal/core ./internal/app -count=1` and
`go test -race ./internal/core ./internal/app -count=1`; cover every generator
and effective-default path with a new/existing/reinitialization matrix.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: A16, A18; explicit adoption and configuration bounds in A24.

### Delivery Evidence

- PR: pending.
- Merge commit: pending.
- Report status: not created.
- Task commits and ticket lifecycle evidence: pending.
- Validation evidence: pending.
- Blockers: none observed; implementation has not started.

## Slice 12: End-to-end workflow assurance

Status: not_started

Prerequisite: Slice 11 verified integrated into dev.

Expected branch: `test/slice-12-workflow-assurance`

Expected PR title: `test(slice-12): verify ticket workflow end to end`

Expected report: `ops/default/report/slices/ticket-native-workflow/slice-12-workflow-assurance.md`

Plan contract: [Slice 12](../plan/ticket-native-workflow.md#slice-12-end-to-end-workflow-assurance)

| Task | Ticket | State | Commit | Work |
| --- | --- | --- | --- | --- |
| T12.1 | [TKT-20261008141429](../ticket/ready/20261008141429-ticket-native-t12-1.md) | ready | pending | Compose real-files/subprocess/Git scenarios for ordered slices, loose queues and parallel independent work. |
| T12.2 | [TKT-20261008141430](../ticket/ready/20261008141430-ticket-native-t12-2.md) | ready | pending | Add restart/crash/lost-response/manual-edit fault scenarios and a bounded repository QA entrypoint. |
| T12.3 | [TKT-20261008141431](../ticket/ready/20261008141431-ticket-native-t12-3.md) | ready | pending | Verify supported existing provider/method paths in designated disposable live repositories. |
| T12.4 | [TKT-20261008141432](../ticket/ready/20261008141432-ticket-native-t12-4.md) | ready | pending | Run final exact-candidate gates, reconcile the acceptance matrix and publish the cycle evidence/report. |

### Validation

Status: not run.

Expected focused checks: run the composed scenarios via the QA entrypoint delivered by
T12.2, then the standard `make verify`, `go test -race ./...` and
`git diff --check` on one immutable candidate. Future `make qa-ticket-native-workflow`
is not a command available at planning time.

Required candidate checks: `make verify` and `git diff --check`.

Acceptance: composed confirmation of A01-A24 and live-provider limitations.

### Delivery Evidence

- PR: pending.
- Merge commit: pending.
- Report status: not created.
- Task commits and ticket lifecycle evidence: pending.
- Validation evidence: pending.
- Blockers: none observed; implementation has not started.

## Acceptance Evidence

Slice 12 confirms the composed workflow and live-provider boundaries. Required
behavior is tested in its primary delivery slices rather than deferred here.

| Criterion | Primary slices | Evidence | State |
| --- | --- | --- | --- |
| A01 | 2, 5 | pending | pending |
| A02 | 1, 2 | pending | pending |
| A03 | 6, 7 | pending | pending |
| A04 | 3, 7 | pending | pending |
| A05 | 3, 8 | pending | pending |
| A06 | 4 | pending | pending |
| A07 | 3, 4, 7, 8 | pending | pending |
| A08 | 5 | pending | pending |
| A09 | 4, 9, 10 | pending | pending |
| A10 | 9 | pending | pending |
| A11 | 6, 9 | pending | pending |
| A12 | 4, 9 | pending | pending |
| A13 | 5, 6, 7, 8 | pending | pending |
| A14 | 6 | pending | pending |
| A15 | 6 | pending | pending |
| A16 | 11 | pending | pending |
| A17 | 4, 10 | pending | pending |
| A18 | 11 | pending | pending |
| A19 | 7, 8 | pending | pending |
| A20 | 7 | pending | pending |
| A21 | 4, 6, 10 | pending | pending |
| A22 | 1, 5 | pending | pending |
| A23 | 10 | pending | pending |
| A24 | 1, 3, 4, 6, 7, 9, 11 | pending | pending |

## External Prerequisites

- Designate disposable existing-provider repositories and authorized credentials
  before live publication validation. Absence blocks the affected live claim;
  provider doubles are not substituted as live success.
- Establish atomic source/base guarding for each claimed provider/method path.
  Unsupported paths stay blocked without policy fallback.
- Preserve unrelated checkout/worktree changes throughout delivery.

## Final Closure

Status: not started.

Expected cycle report: `ops/default/report/ticket-native-workflow-cycle.md`.

Final exact-dev candidate: pending.

Required final evidence: `make verify`, `go test -race ./...`,
the `make qa-ticket-native-workflow` entrypoint to be delivered in Slice 12,
and `git diff --check` on the recorded immutable candidate.

Closure requires 48 done task tickets, 12 verified slice integrations, delivered
canonical slice reports and A01-A24 evidence. Releases and promotion to `main`
are separate from this delivery set.

## Open Questions

No product-contract decision is open. Execution approval and 48 ticket bindings
are recorded in the bootstrap commit. Live-provider fixture designation remains
an external prerequisite before the affected publication validation.
