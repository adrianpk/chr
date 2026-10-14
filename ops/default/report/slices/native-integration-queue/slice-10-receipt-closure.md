# Slice 10: Receipts fanout and closure

Status: reviewing
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

T10.3 re-observes confirmed receipts through the existing canonical adapter and
submission journal. Later descendant targets retain the verified method and
original receipt. A fetched target excluding the known result records its actual
commit and a typed history-divergence blocker. Transport uncertainty or lost
merged metadata remains reconciliation without claiming removed history.

Native frontier release verifies retained prepared history and preserves global
preparation and review counts in each unit snapshot. It returns the persisted
blocker on failure. Completed task images, receipt bytes, consumed submission
budgets and cancellation intent survive re-observation. Fresh merged evidence
restores an integrated unit, or leaves unfinished fanout finalizing.

T10.4 connects native receipts to the existing set-closing path. Closure reads
the retained fanout proof and current integrated queue/journal binding rather
than decoding the native verification observation as a legacy proof. All actual
members and intended slice receipts must be finalized before a gate can start.

Unresolved common-directory publications block admission, command registration
and confirmation across namespaces. The configured host gate still runs on the
exact canonical target. Fresh native PR/Git observation and frozen native proof
hashes prevent confirmation after metadata or evidence changes during that gate.
A failed gate retains its own retry history without replaying publication or
member fanout. Original and corrective sets keep distinct closure receipts.

All four tasks are implemented and development-checked. They are reviewing;
final ordered clean-HEAD slice validation remains pending.

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

Post-receipt observations append to the common journal while confirmed PR/Git
facts continue to refer to the receipt introduction. Known result objects are
read before negative ancestry classification; fetch failures cannot prove
history removal. Recovery repairs a retained observation projection and fences
the prior owner revision before new observation IO. Known inconsistent target
fields and callbacks that skip required reconciliation are rejected.

Native set closure reuses the existing unit-delivery archive, immutable receipt,
common publication journal and closing-attempt ledger. Its original verification
reference remains unchanged. An independently reconciled lost transport response
is distinct from an unresolved publication. Closing evidence now freezes the
native capability, source/history, transaction/application and observation refs.

## Contracts Added or Changed

- `ConfirmWorkflowNativeReceipt` accepts only a fenced invocation guard and
  retained host evidence; it performs no external publication or task fanout.
- Native receipts record the publication mechanism and an optional independent
  Git application reference. Unknown exchanges require that exact application
  proof before confirmation. Existing receipt fields remain readable.
- `FinalizeWorkflowSlice` finalizes a confirmed native unit from retained host
  evidence. Queue `integrated` and the confirmed cursor follow completed member
  fanout; legacy finalization retains its original verification path.
- `ObserveWorkflowNativeSubmission` also revalidates confirmed receipts without
  another submission. Its queue projection preserves completed members and
  exposes current canonical reconciliation.
- Native frontier observation verifies the retained prepared method and keeps
  cumulative role counts bound to their delivered-unit history. Legacy
  preparation, review and receipt schemas remain readable.
- `CloseWorkflowSet` verifies settled native fanout, current confirmed journal
  ownership and all intended receipts before running its existing closing gate.
  Confirmation also rechecks the current native canonical PR and method proof.
- `QueryWorkflowIntegrationSetClosure` verifies native receipts through their
  existing delivery proof and reports unresolved common publication heads
  without mutating execution state.

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
- `internal/core/workflow_native_receipt_observation.go` and
  `workflow_native_receipt_projection.go`: immutable receipt re-observation and
  canonical queue blockers.
- `internal/core/workflow_frontier*.go`: native method proof and retained role
  counter ownership at continuation.
- `internal/core/workflow_native_receipt_*test.go`: actual atomic publication,
  lost SSH response, rejection, fencing and restart cases.
- `internal/core/workflow_closing_native.go` and `workflow_closing_publication.go`:
  native receipt/archive ownership, immutable evidence and shared publication
  barriers in the existing closing path.
- `internal/core/workflow_native_closing*test.go`: actual native publication,
  exact gates, distinct set outcomes, lost transport, changing evidence/PRs and
  cross-namespace intent before or during a real gate.

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

T10.3 first reproduced the existing observation API refusing a confirmed
receipt. Native continuation then exposed preparation/review counts without
active pointers after frontier release; retained unit snapshots now own those
counts without resetting the invocation budget. Focused descendant/divergence,
frontier, uncertainty, recovery, transition, pre-fanout, target-evidence and
cancellation checks passed in 115.599 s. Merge, rebase and squash use actual
owned Git repositories. Strict lint passed with zero issues; `git diff --check`
passed. Acta activity: `01M4X3TDFAQV2XJ9F10CDB0E6X`.

The first broader race check failed in the unchanged
`TestNativeCanonicalOutcomes` after 1197.699 s: a pre-receipt target rewrite must
retain its existing `canonical_target_unverified` condition. History-divergence
classification is now restricted to confirmed receipts. Focused canonical,
receipt and frontier checks passed in 51.338 s; strict lint again passed with
zero issues. Failed-source evidence remains under
`/tmp/chr-native-slice10-t10-3-cb2aeff9-failed`.

The 21-file T10.3 development source is frozen at digest
`976a94dd3792872b7adf022924aa276b2706d79cad96494ae59c2c4eaa8c2b01`.
Owned unit `chr-native-slice10-t103-validation-976a94dd.service` executes these
pre-commit checks, with evidence under `/tmp/chr-native-slice10-t10-3-*`:

- `go test -race ./internal/core -run '^(TestNativeReceipt.*|TestNativeFinalization.*|TestNativeFanout.*|TestNativeSubmission.*|TestNativeCanonical.*|TestNativeGit.*|TestNativeStop.*|TestWorkflowFinalization.*|TestWorkflowFrontier.*|TestWorkflowPreparation.*|TestWorkflowReview.*)$' -count=1` — passed in 1218.447 s.
- `make verify` — passed with zero lint issues; core tests passed in 2838.955 s,
  application tests passed in 122.665 s and all other packages passed.

The broader race check and its owned fixtures finished with exit status zero.
Acta activity: `01M4X6DC4KS44V9TGMAP8Z4C03`. The pre-receipt compatibility
correction and focused checks are recorded as `01M4X567V9PQ0GA2E5HWEBSVWP`.
Both frozen-source development checks passed. The owned service is inactive
with exit status zero and no live check process; all 21 Go files remained
unchanged and the owned fixtures finished. Final ordered clean-HEAD slice
validation remains pending until T10.4 and PR-reference commits finish.

T10.3 implementation: [dd7dc5b536c4a7ccd9e620824ef31b976441d87c](https://forge.adrianpk.com/adrianpk/chr/commit/dd7dc5b536c4a7ccd9e620824ef31b976441d87c).
Acta commit activity: `01M4X96ZTZ648RBXGM1EJYREJB`.

T10.4 initially failed all three native method probes because the closing path
expected a legacy verification schema. The corrected positive probe passed in
19.900 s. An unfinished standalone owner is rejected by the existing ownership
guard before fanout; its test now checks that concrete boundary. The independent
publication fixture uses the ticket executor and a real fast-forward canonical
base without relaxing the frozen-target ancestry guard. Its admission and
in-gate intent checks passed in 21.790 s.

Strict lint passed with zero issues after preserving the configured complexity
and formatting bounds. The 11-file development source is frozen at digest
`c8933a047c6f9dfe4e0b4b887cf6e2da8baf77bd7c9e8528a072ec338dd442f0`.

- `go test ./internal/core -run '^(TestNativeClosing.*|TestWorkflowClosing.*|TestNativeQueueClosure|TestNativeReceiptFrontier.*|TestNativeFinalization.*|TestNativeFanout.*)$' -count=1` — passed in 315.057 s, with output in `/tmp/chr-native-slice10-t10-4-focused.log`.

Acta focused validation activity: `01M4XBKBVZHZP2PBAKKB00NSMP`. The focused
check and its owned fixtures finished. Owned unit
`chr-native-slice10-t104-validation-c8933a04.service` executes these pre-commit
checks on the unchanged 11-file source, retaining source, command, process and
result evidence under `/tmp/chr-native-slice10-t10-4-*`:

- `go test -race ./internal/core -run '^(TestNativeClosing.*|TestWorkflowClosing.*|TestNativeQueueClosure|TestNativeReceiptFrontier.*|TestNativeFinalization.*|TestNativeFanout.*)$' -count=1` — passed in 541.991 s.
- `make verify` — passed with zero lint issues; core tests passed in 3078.743 s,
  application tests passed in 121.224 s and all other packages passed.

Both frozen-source development checks passed. The owned service is inactive
with exit status zero and no live check process; all 11 Go source files remained
unchanged and owned fixtures finished. Acta race activity:
`01M4XC51WKBG0ACGS0Q39BVECA`; pre-commit gate activity:
`01M4XF2XHX2N5XQ3KZJW7GW6BZ`.

The T10.4 task commit reference follows commit creation. Canonical Slice 10 PR
creation, its reference follow-up and final ordered clean-HEAD controller evidence
remain pending.

## Risks and Follow-ups

None recorded.
