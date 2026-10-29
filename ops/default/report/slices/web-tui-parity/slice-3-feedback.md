# Slice 3: Shared operation feedback and draft retention

Status: delivered
Delivery set: web-tui-parity
Plan: [Delivery contract](../../../plan/web-tui-parity.md)
Tracker: [Tasks and evidence](../../../tracker/web-tui-parity.md)
Branch: `feat/slice-3-tui-feedback`
PR: [#457](https://forge.adrianpk.com/adrianpk/chr/pulls/457)
Base: `5e28668e2d3bdad91791231292f7e52e62af1c77` on dev, after verified PR #456 integration
Integration: PR #457 merged into dev at `c91ba4feac5b4296548a854e26b8d9cb59170ed7` using Rebase + Fast-forward.

## Purpose

Resolve G-038 with actionable manual-operation feedback, retained drafts and
results bound to their original interaction.

## Delivered Behavior

- Manual adapter results distinguish ok, validation_failed, stale and blocked.
  The existing Action panel shows summaries, canonical validation findings,
  blocking reasons and changed paths returned by supported operations.
- Failed operations open the panel; successful summaries remain in the status
  line without taking editor focus. Alt+o inspects the latest result; arrows and
  Home/End expose long findings. Only applicable reload or original retry actions
  appear, and recovery cannot act on a different current interaction.
- Project and Case fields, Dump text and Quiz answers require confirmation
  before entered text is discarded. Reload reads the captured source and only
  replaces its draft after confirmation and a successful read. Failed reloads
  retain the original input. Quiz Skip and Done confirm loss of unsent answers.
- Answer submission retains its editor through acknowledgement. Failed writes,
  delayed replies and newer editor generations preserve the complete draft.
  Quiz, Spec, Plan and Work commands capture their executor, request and reply
  owner before IO. Older replies cannot install another result or obsolete
  workflow confirmation.

## Implementation Notes

T3.1 is `eb484dd34676fa4fa8f518642d797e83b7ef21a1`. T3.2-T3.4 share
`8046584d910969a0be349bb85f9a7307e3f86d8a`: the Action panel, recovery callbacks,
draft confirmation and reply acceptance depend on one interaction owner.
`bc8e124dadb420312d6d19bff536e588741941f3` keeps unrelated Dash observations
outside document reply bindings and verifies failed reload retention.

Two initial regressions reproduced discarded Quiz input before acknowledgement
and an older Spec reply replacing a newer result. Both now pass. Reply bindings
include view, case, source/selection, draft generation, input digest and retained
workflow identity; native observations additionally reject older revisions.
Confirmation and recovery capture their original source and operation. Exact
retry is offered only with a stable captured operation ID; a reconciliation
error does not by itself prove that a write committed. A failed
retry never rereads the selected document or reconstructs a new request ID.

The production application test writes a dump, edits it through TUI keys,
commits a concurrent source update, verifies stale feedback and unchanged source
hashes after failed save/cancelled reload, then confirms reload and saves the
new image. Metadata fixtures verify real hash conflicts and reads. The owned PTY
uses the root production loader and actual input, cancels and confirms draft
loss, joins its child and compares canonical ops hashes. Provider responses in
existing generation fixtures remain deterministic; writes and revisions are real.

## Contracts Added or Changed

`internal/app/operation.Result` is an application-owned transport-neutral value.
Its small subpackage avoids an import cycle between the application assembler
and TUI. Web keeps its existing response envelope and routes while sharing
status constants, validation mapping and Dump failure classification. Terminal
adapters retain their typed error API and attach the neutral result; domain
operations, receipt identities and persistence formats are unchanged.

Manual editors own destructive-change confirmation. Guided and Harness screens
keep their existing interaction owners; the result shortcut does not intercept
their input. Pending results transfer back to the event loop, where obsolete
owners release their own busy state without replacing current input or controls.

## Files of Interest

- `internal/app/operation/result.go`, `operation_result.go`,
  `tui_operation_feedback.go`: neutral outcomes and adapter mappings.
- `internal/tui/operation_feedback.go`, `draft_recovery.go`,
  `metadata_recovery.go`: Action presentation and confirmed captured recovery.
- `internal/tui/operation_binding.go`, Quiz/Spec/Plan/Work action handlers:
  request capture and reply ownership.
- `internal/tui/operation_feedback_test.go`,
  `internal/app/tui_operation_feedback_test.go`,
  `tui_operation_terminal_linux_test.go`: sentinel, ordering and production IO
  regressions.
- User Guide, TUI reference, action matrix and changelog: delivered controls.

## Validation

Implementation checks passed locally:

- `go test ./internal/tui -count=1` — passed, including existing catalog frames.
- `go test ./internal/app -run '^Test(TUIOperation.*|OperationOutcomes|TUIDumpAdapterUsesDumpServices|TUIQuizAdapterDispatchesRuntimeQuizActions|TUISpecAdapterGeneratesAcceptanceFeature|WebDumpSaveDocument|WebQuizActions|TUIWorkflow(Preview|Stale|Controls|Start|Approval|Refresh))$' -count=1 -timeout=3m` — passed without short mode.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix --timeout=5m ./internal/tui ./internal/app ./internal/app/operation ./internal/webui` — passed after formatting and complexity corrections.
- `git diff --check` — passed.

The first controller batch `4ARNEJ3OLAZ4T4PB2SWPMQGYXB` stopped at the
repository signature-format check. Renamed methods now use single-line
parameter lists, and the neutral package includes its required README.

Controller validation `7IDEK3RU26UI7CUQSTNJ7T3NAK` passed all four ordered
checks on clean HEAD `c91ba4feac5b4296548a854e26b8d9cb59170ed7`:

- `make verify`
- `go test -race ./internal/tui -run '^Test(Operation.*|BusyQuit|Model(Ownership|AsyncRefresh|WorkflowBinding))$' -count=1`
- `go test -race ./internal/app -run '^Test(OperationOutcomes|TUIOperation.*|TUIDumpAdapterUsesDumpServices|TUIQuizAdapterDispatchesRuntimeQuizActions|TUISpecAdapterGeneratesAcceptanceFeature|TUIPlanGenerateReturns.*|TUIWorkflow(Preview|Stale|Controls|Start|Approval|Refresh)|TUIDashTarget|TUIBusyTerminal|GuidedUseCaseTUIJourney|WebDumpSaveDocument|WebQuizActions|ServiceDump(Capture|ProjectionFailure|WriteGuards|Reconciliation))$' -count=1 -timeout=3m`
- `git diff --check`

Canonical Forgejo and Acta integration event `01M63RT08W8774AQNFKE3AS3XT`
confirm PR #457 merged at the validated commit. No checks were repeated for closure.

## Risks and Follow-ups

None recorded.
