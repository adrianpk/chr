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

T10.2 projects a confirmed receipt as `finalizing` and reuses the existing
finalizer for all unit members. Each slice task shares that exact receipt and
canonical PR while keeping its membership identity; an independent ticket owns
its own receipt. The existing write-ahead journal retains member and set images
before the integrated queue and invocation cursor. Restart and concurrent
retries complete the same fanout without repeating task work or publication.

A stop request remains pending while fanout is incomplete. Finalization retains
the irreversible delivery before applying terminal stop, preserving consumed
budgets. Corrupted receipt, candidate or delivery proof cannot change members
or the confirmed cursor; an optimistic queue transition cannot skip fanout.

T10.1 and T10.2 are implemented and reviewing. T10.3 is active for descendant
proof and divergence reconciliation; T10.4 remains ready in the approved tracker.

## Implementation Notes

Receipt IDs retain the existing invocation/candidate/submission identity.
Physical receipt introduction precedes common-journal confirmation; restart
adopts the first verified bytes after either checkpoint. Canonical PR head
metadata may remain at S or advance after P; immutable transaction and
application evidence establish publication independently of that mutable head.

The receipt validator also requires the candidate ID and the final retained
observation to match the verified native binding. Historical receipts retain
their original bytes and verification rules.

Native finalization stores a typed receipt reference in the existing delivery
proof, derived from the physically verified receipt and its retained canonical
observation. It does not rewrite the receipt or create a parallel receipt or
member ledger. The queue receipt reader now uses the storage-envelope codec
and verifies its physical hash before checking typed native ownership.

## Contracts Added or Changed

- `ConfirmWorkflowNativeReceipt` accepts only a fenced invocation guard and
  retained host evidence; it performs no external publication or task fanout.
- Native receipts record the publication mechanism and an optional independent
  Git application reference. Unknown exchanges require that exact application
  proof before confirmation. Existing receipt fields remain readable.
- `FinalizeWorkflowSlice` finalizes a confirmed native unit from retained host
  evidence. Queue `integrated` and the confirmed cursor follow completed member
  fanout; legacy finalization retains its original verification path.

## Files of Interest

- `internal/core/workflow_native_receipt_delivery.go`: exact confirmation,
  immutable receipt construction and restart adoption.
- `internal/core/workflow_native_receipt_application.go`: independent
  application proof for lost transport responses.
- `internal/core/workflow_native_record_evidence.go` and
  `workflow_submission_observe.go`: candidate, canonical PR and source identity
  verification.
- `internal/core/workflow_native_finalization.go` and
  `workflow_native_finalization_projection.go`: native receipt projection and
  existing journal fanout.
- `internal/core/workflow_native_finalization_*test.go`: real host death at every
  changed member/set/queue/cursor image, independent tickets, concurrency, stop
  and invalid-proof barriers.
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

- `make lint-strict` — passed with zero issues on the T10.1 nine-file source.
- `git diff --check` — passed.

The T10.1 development Go source was frozen at digest
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

T10.1 implementation: [b0a64a02dbc13ad61220cde8502fca66f7a332f9](https://forge.adrianpk.com/adrianpk/chr/commit/b0a64a02dbc13ad61220cde8502fca66f7a332f9).
Acta commit activity: `01M4WWAFZQ5NQSPN1ZRKS10XTK`.

T10.2 initially reproduced the confirmed receipt remaining `observing` and the
existing finalizer rejecting native delivery without a legacy proof. The focused
fanout, receipt, real-crash, concurrent-owner, independent-ticket and stop tests
passed in 124.779 s after connecting native finalization. Strict lint passed with
zero issues and `git diff --check` passed. Acta activity:
`01M4WY0R8B4GV62EYTAHCT79JJ`.

The 17-file T10.2 development source is frozen at digest
`d4e17bdb213e312940bf2259465493988cd43927c4644241bb01a9f9005149ba`.
The owned `chr-native-slice10-t102-validation-d4e17bdb.service` executes
these pre-commit checks, with evidence under `/tmp/chr-native-slice10-t10-2-*`:

- `go test -race ./internal/core -run '^(TestNativeFinalization.*|TestNativeFanout.*|TestNativeReceipt.*|TestNativeSubmission.*|TestNativeCanonical.*|TestNativeGit.*|TestNativeStop.*|TestWorkflowFinalization.*|TestWorkflowPublication.*|TestWorkflowTicket.*|TestWorkflowFrontier.*)$' -count=1` — passed in 843.435 s.
- `make verify` — passed with zero lint issues; core tests passed in 2443.067 s
  and application tests passed in 107.273 s. All other packages passed.

The broader race check and its owned fixtures finished with exit status zero.
Acta activity: `01M4WYVYGEYJNGS5TAF486KKTJ`. Both frozen-source development
checks passed, and the owned service finished with exit status zero and no live
check process. All 17 Go source files remained unchanged. Final ordered
clean-HEAD slice validation remains pending until the remaining tasks and PR
reference commits finish.

T10.2 implementation: [c96ecba537e1bfedf6925c0324f0d49bd8332d89](https://forge.adrianpk.com/adrianpk/chr/commit/c96ecba537e1bfedf6925c0324f0d49bd8332d89).
Acta commit activity: `01M4X1988DMQYREF39TJTH8BHV`.

## Risks and Follow-ups

None recorded.
