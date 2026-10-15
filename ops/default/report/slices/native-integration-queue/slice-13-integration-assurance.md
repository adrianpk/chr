# Slice 13: Native Integration and Original Workflow Assurance

Status: delivered
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `test/slice-13-native-integration-assurance`
PR: [#387](https://forge.adrianpk.com/adrianpk/chr/pulls/387)

## Purpose

Complete the original workflow assurance through the native atomic publisher,
reusing the four original task identities and retained implementation. Final
delivery closure requires the controller's exact integrated gate evidence.

## Delivered Behavior

T13.1 implementation and pre-commit validation are complete. It adapts the original ordered,
loose and parallel scenarios from
`4f10f1c60de1f164f727d37b4df5c02343ab6e11`. The existing cumulative preparation
and review frontier corrections are already present and are not replayed.

T13.1 implementation: [91a00cf358255332c9e3f9365771efbfbe5e640a](https://forge.adrianpk.com/adrianpk/chr/commit/91a00cf358255332c9e3f9365771efbfbe5e640a). Acta `01M4ZWXQQ90GGQHNRX72TNKCK7`.

T13.2 reuses the retained fault runner from
`44e8145aa618edfc51d5efb08330c641b02f7e58`. Its acceptance/fault manifests
resolve 78 distinct tests, including six existing native fault tests.
The adapted three-publication lost-response queue passed once in 18.863 s;
Final `go test -race ./internal/core -run '^TestWorkflowAssurance' -count=1 -timeout=5m` passed in 170.989 s, including actual native SSH report loss across three publications. Strict lint passed with zero issues. Acta `01M4ZXGH4NM29T2Q1ENX0PN03C`. Bounded `make qa-ticket-native-workflow` passed all 78 tests in 410.319 s on unchanged source `6fa423e7c5ffa4c93282d1e8f16f2f15e3759586c889749dc92f22303e52e2c7`, with no leftover owned children. Acta `01M4ZXX216AR8NEA7XBGAD4AD9`. The owned unit exited successfully.

T13.2 implementation: [8e041477fa4401412316da7b1b91af4ca900c09e](https://forge.adrianpk.com/adrianpk/chr/commit/8e041477fa4401412316da7b1b91af4ca900c09e). Acta `01M4ZXZSH26BFZXTS6PRD01264`.

T13.3 implementation: [b2f479d536aba6df4c151303c30d822744252697](https://forge.adrianpk.com/adrianpk/chr/commit/b2f479d536aba6df4c151303c30d822744252697). Acta `01M4ZY5FCRR8Q188427ZAH8DDN`.

T13.4 closure preparation: [4cb701e668d8fbf3b263d08254112ccd62b542f7](https://forge.adrianpk.com/adrianpk/chr/commit/4cb701e668d8fbf3b263d08254112ccd62b542f7). Acta `01M4ZYAEA0WPNBZ07ETNXTD7G6`. Its final integrated-gate outcome remains blocked.

## Implementation Notes

The composed fixture uses real worker processes, linked checkouts and the
production atomic receive-pack publisher against an isolated bare repository.
Controlled canonical metadata is interpreted by the existing forge adapter.
This is deterministic local assurance; live-provider certification remains the
separate Slice 12 evidence and native-integration QA target.

The adapted ordered scenario reproduced missing native routing in shared
continuation: at cursor 3, before any confirmed delivery, the old path returned
`publication_guard_unsupported`. Native continuation now uses existing queue,
preparation, review, dispatch, observation and receipt owners, with fresh revision
guards between phases. Shared service entrypoints supply the native ports.

Receipt confirmation also exposed a premature frontier advance. Continuation
now completes retained member fanout before attempting the next unit barrier.
No task is considered delivered from a transport response alone.

## Contracts Added or Changed

- Native publication ports are separate from the legacy provider submission
  interface. The configured profile selects native routing explicitly.
- Each continuation phase uses the existing guarded operation and checkpoint
  bound. Restart observes an existing intent without dispatching it again.
- Original ticket IDs, task labels, retained commits and source branch remain
  provenance under the approved T13.1–T13.4 execution transfer.

## Files of Interest

- `internal/core/workflow_assurance*_test.go`: retained composed assurance,
  adapted to actual native publication and receipt barriers.
- `internal/core/workflow_publication_native.go`: shared native phase routing.
- `internal/core/workflow_continuation_steps.go`: fanout before frontier advance.
- `internal/app/service_workflow_publication.go` and
  `service_workflow_continuation.go`: shared service native port binding.

## Certification Reconciliation

T13.3 fulfills the transferred original live-provider obligation using the
[integrated Slice 12 certification report](https://forge.adrianpk.com/adrianpk/chr/src/commit/1c625ecce15ab3250498b4a0e24e0bc87b6ec842/ops/default/report/slices/native-integration-queue/slice-12-provider-certification.md).
All 46 retained live-case files still match their immutable hashes. The host
index `chr/qa/native-integration-unblock-20261015/live-evidence.json` has SHA-256
`27ab099281cdc6c2726b525467000fd522475af21fdfc42957d2432391716c19`.

All twelve positive provider/method/transport bindings preserve distinct original
S and exact delivered P, compatible receipt proof and the original merged PR.
Fresh canonical reads confirmed all twelve PRs and all 24 actual source/target
refs at P. Historical PR base metadata matches its retained observation; actual
Git refs independently prove the current target. GitHub raw source attribution
remains admitted only by its pinned v2 contract and complete independent P proof.

| Provider | Method | SSH original PR | HTTPS original PR |
| --- | --- | --- | --- |
| forgejo | merge | [#61](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/61) | [#52](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/52) |
| forgejo | rebase | [#64](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/64) | [#55](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/55) |
| forgejo | squash | [#67](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/67) | [#58](https://forge.adrianpk.com/adrianpk/chr-integration-qa/pulls/58) |
| github | merge | [#47](https://github.com/adrianpk/chr-integration-qa/pull/47) | [#56](https://github.com/adrianpk/chr-integration-qa/pull/56) |
| github | rebase | [#50](https://github.com/adrianpk/chr-integration-qa/pull/50) | [#59](https://github.com/adrianpk/chr-integration-qa/pull/59) |
| github | squash | [#53](https://github.com/adrianpk/chr-integration-qa/pull/53) | [#62](https://github.com/adrianpk/chr-integration-qa/pull/62) |

The other 34 cases retain source/target contention, permission/review denial,
external close/retarget and lost HTTPS report outcomes. The strict-v1
counterexample remains preserved in the Slice 12 report.

This slice changes continuation routing and deterministic QA wiring. Production
transport, history preparation, capability and native receipt/observer contracts
remain those certified by Slice 12; no individual profile case was invalidated
by these changes. The configured final candidate QA rechecked live policy
and all profiles at the exact clean HEAD and passed. No new live access or policy change
was made for this reconciliation.

All twelve predecessor reports are delivered. Their canonical PRs #375–#386
are merged into `dev`, their merge commits are ancestral to the current base,
and all 47 predecessor task tickets are solved with closure metadata.
Reconciliation Acta `01M4ZXY81C2RHH091D0M6QPP6T`.

## Acceptance and Closure Handoff

The [approved acceptance coverage](../../../plan/native-integration-queue.md#acceptance-coverage)
retains one primary owner for each AC-001–AC-100. Earlier slices remain primary
implementation evidence; this replacement assurance unit composes and reconciles
that evidence without introducing duplicate task tickets.

| Obligation | Current evidence and remaining boundary |
| --- | --- |
| Original A01–A24 and AC-091 | Ordered, loose and parallel native continuation plus real SSH report-loss recovery pass; the 78-test acceptance/fault target passed within its bounds. |
| AC-001–AC-090 and AC-096–AC-100 | The approved primary ownership map, twelve delivered reports and reconciled certification retain implementation evidence; final candidate evidence belongs to the exact clean PR head. |
| AC-092 | Required candidate checks passed in configured order on the final clean PR head; the controller's unchanged full gate must then pass on the exact integrated development revision. |
| AC-093 | All thirteen corrective merges and 50 solved tasks are verified. T13.4 final closure remains pending on the integrated gate. |
| AC-094 | Original T12.1–T12.4 execute as T13.1–T13.4 through this replacement PR. Original cycle closure requires the replacement merge/report and exact integrated gate evidence. |
| AC-095 | Final closure must retain reconciliation, absence of unknown publication and verified original/corrective closure evidence. It remains pending. |

T13.4 prepares these bindings and the existing configured closure paths. The
controller owns integration and the post-integration full gate:

```sh
make verify && go test -race ./... && make qa-ticket-native-workflow && make qa-native-integration && git diff --check
```

After that gate succeeds, only approved closure documentation may record the
exact development SHA, ordered results, replacement PR/report and original versus
corrective task outcomes. The outstanding cycle reports are created during that
closure phase, using the controller's retained evidence. Runtime changes require
the authorized correction route and fresh validation.

The original reserved branch remains at
`6100f7d166cc3c61466dcacae05c26baabb2460b`. Its historical blocked execution and
consumed budgets are preserved. No original assurance PR or historical controller
completion is invented; this Slice 13 report is the canonical replacement unit.
Replacement integration is verified. T13.4 remains blocked on the exact integrated
full gate and subsequent cycle closure. Neither delivery set is complete yet.

## Validation

- `go test ./internal/core -run '^TestWorkflowAssuranceOrdered$' -count=1 -timeout=5m`
  reproduced the missing native routing in 3.253 s before the correction.
- `go test ./internal/core -run '^TestWorkflowAssurance' -count=1 -timeout=5m`
  passed ordered, loose and parallel scenarios in 51.739 s after native routing
  and the fanout barrier correction.
- `make lint-strict` passed with zero issues after extracting bounded phase
  helpers and correcting statement spacing.
- `go test -race ./internal/core -run '^(TestWorkflowAssurance|TestWorkflowContinue|TestWorkflowContinuation|TestWorkflowPublication|TestNativeFinalization|TestNativeClosingGate)' -count=1 -timeout=10m` passed in 452.844 s. Acta `01M4ZT1TSNNZ5CB01K2H1VM7S8`.
- Required pre-commit `make verify` passed in 2939.843 s on unchanged source
  `b28e8b30e9b4fb93c0d21079d4c21f1af2918fa80d82a5639b775a4ae2c29e67`:
  strict lint zero issues, core 2907.443 s, app 207.917 s and all TUI/Web
  packages passed. Acta `01M4ZWVHWVZ43BM362QWXH8QPH`.
  The owned unit exited successfully; host manifest, log and result are retained
  under `chr/qa/native-integration-slice-13/`.
  The delivery controller retains exact clean-HEAD candidate evidence after the
  PR-reference commit. Development checks do not replace that evidence or the
  post-integration gate.

Verified merge: `88a9c6344f5e80d5a9a03357a1ce71d106e1231f` at 2026-10-15T16:48:49Z; controller Acta `01M507ET2KEAG9ATT5QAEGKMJ6`.

All five configured candidate checks passed in order on that exact clean HEAD: `make verify`, `go test -race ./...`, `make qa-ticket-native-workflow`, `make qa-native-integration`, and `git diff --check`. Validation `3GRRMCM246GSGUCZVTSUTZM6XY`; Acta `01M507ENSTTAR7FZY5KXB66DE0`. The separate integrated full gate remains controller-owned and pending.

Successful candidate commands, in order: `make verify`,
`go test -race ./...`, `make qa-ticket-native-workflow`,
`make qa-native-integration`, and `git diff --check`.

## Risks and Follow-ups

T13.4 exact integrated-gate closure remains pending. Retained implementation
alone does not establish original delivery completion.
