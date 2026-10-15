# Slice 13: Native Integration and Original Workflow Assurance

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `test/slice-13-native-integration-assurance`
PR: pending

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
  No clean-HEAD candidate validation or integrated delivery gate is claimed yet.

Required candidate commands remain, in order: `make verify`,
`go test -race ./...`, `make qa-ticket-native-workflow`,
`make qa-native-integration`, and `git diff --check`.

## Risks and Follow-ups

T13.3 certification reconciliation and T13.4 exact integrated-gate closure
remain pending. The original controller remains inactive;
retained implementation alone does not establish original delivery completion.
