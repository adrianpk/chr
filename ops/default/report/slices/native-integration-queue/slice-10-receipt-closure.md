# Slice 10: Receipts fanout and closure

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-10-native-receipt-closure`
PR: pending

## Purpose

Finalize exact native delivery through immutable receipts, task fanout and
verified closure barriers.

## Delivered Behavior

T10.1 adds host-owned receipt confirmation from the retained submission and its
latest canonical observation. Pending, closed-unmerged and incompatible PR
results cannot produce receipt bytes. Exact merged proof retains original S,
published P, validated B, method history, candidate and capability binding.

Actual lost receive-pack responses remain immutable `unknown` transport facts.
A separately retained, exact Git application observation can attribute P after
later source movement. Receipt confirmation records that proof without another
publication or submission charge. Stale invocation, entry, budget and claim
generation callbacks cannot create receipts or change the common journal.

T10.1 is implemented and development-checked. T10.2 through T10.4
remain pending in the approved tracker.

## Implementation Notes

Receipt IDs retain the existing invocation/candidate/submission identity.
Physical receipt introduction precedes common-journal confirmation; restart
adopts the first verified bytes after either checkpoint. Canonical PR head
metadata may remain at S or advance after P; immutable transaction and
application evidence establish publication independently of that mutable head.

The receipt validator also requires the candidate ID and the final retained
observation to match the verified native binding. Historical receipts retain
their original bytes and verification rules.

## Contracts Added or Changed

- `ConfirmWorkflowNativeReceipt` accepts only a fenced invocation guard and
  retained host evidence; it performs no external publication or task fanout.
- Native receipts record the publication mechanism and an optional independent
  Git application reference. Unknown exchanges require that exact application
  proof before confirmation. Existing receipt fields remain readable.

## Files of Interest

- `internal/core/workflow_native_receipt_delivery.go`: exact confirmation,
  immutable receipt construction and restart adoption.
- `internal/core/workflow_native_receipt_application.go`: independent
  application proof for lost transport responses.
- `internal/core/workflow_native_record_evidence.go` and
  `workflow_submission_observe.go`: candidate, canonical PR and source identity
  verification.
- `internal/core/workflow_native_receipt_*test.go`: actual atomic publication,
  lost SSH response, rejection, fencing and restart cases.

## Validation

The first delivery probe reached the missing runtime confirmation boundary.
The implementation then exposed a submission validator that required current
PR head metadata to equal P despite certified observations allowing lag at S.
Confirmation now uses immutable publication proof and exact canonical PR/result
identity. The initial focused receipt/restart tests passed in 32.843 s before
adding substituted-application proof cases. Strict lint corrections preserve the
configured complexity and formatting bounds.

- `make lint-strict` — passed with zero issues on the current nine-file source.
- `git diff --check` — passed.

The current Go source is frozen at digest
`7301223b366b09e47bd4953885134d0d6f9c545210de73f0ee5cf3c1da22f20b`.
The owned `chr-native-slice10-t101-validation-7301223b.service` executed these
development checks in order:

- `go test ./internal/core -run '^(TestNativeReceipt.*|TestNativeSubmission.*|TestNativeCanonical.*|TestNativeGit.*|TestNativeStop.*|TestWorkflowFinalization.*|TestWorkflowPublication.*)$' -count=1` — passed in 375.131 s.
- `go test -race ./internal/core -run '^(TestNativeReceipt.*|TestNativeSubmission.*|TestNativeCanonical.*|TestNativeGit.*|TestNativeStop.*|TestWorkflowFinalization.*|TestWorkflowPublication.*)$' -count=1` — passed in 603.402 s.
- `make verify` — passed with zero lint issues; core 2401.672 s, app 107.021 s and all other packages passed.

Source, command, process and result evidence is retained under
`/tmp/chr-native-slice10-t10-1-*`. Broader normal/race checks passed on the unchanged nine-file source.
Their owned check processes and fixtures finished. Acta activity:
`01M4WSYK9EZM99G585ZDXFTXRP`. Required pre-commit `make verify` passed on that same source. The owned service
finished with exit status zero and no live check process. All owned check
fixtures finished.

Final ordered clean-HEAD slice validation belongs to the controller after all task and PR
reference commits.

## Risks and Follow-ups

None recorded.
