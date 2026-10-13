# Slice 9: Publication and uncertainty recovery

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-9-native-publication-recovery`
PR: pending

## Purpose

Connect the retained native submission to exact atomic Git publication and
canonical observation without duplicate submissions or premature delivery.

## Delivered Behavior

T9.1 connects the retained native intent to the trusted atomic Git publisher.
The store checks the current owner, candidate, scope and publication policy,
retains intent and the live host identity before IO, then invokes the exact
publisher once. A restarted or overlapping caller receives observation-required
for that same submission; it cannot dispatch again or refill the consumed
submission count. A raw successful Git response retains the common uncertainty
barrier and cannot manufacture delivery or a receipt.

## Implementation Notes

The existing invocation, common publication journal, queue entry and budget
remain the owners of dispatch authority and target exclusion. The common
repository mutation lock covers authorization and dispatch retention; the
publisher runs after that lock is released while the durable target journal
retains exclusion. Other targets do not wait for this transport call.

The dispatch-phase recheck shares current review/capability inputs but requires
the exact retained submission in the publishing phase. The original review
phase fence remains unchanged. Ownership is Linux boot/PID/birth identity;
a three-minute caller deadline bounds the dispatch operation.

## Contracts Added or Changed

- `PublishWorkflowNativeSubmission` connects the existing native journal to the exact trusted Git publication port.
- Dispatch ownership is retained before the publisher receives control; unresolved retries only observe.
- Live candidate, review and capability checks run again after intent retention, under a dispatch-specific phase fence.

T9.1, T9.2 and T9.3 implementation and development validation are complete;
canonical PR and verified integration remain pending.

T9.2 retains the actual receive-pack advertisement and exact atomic command
manifest before sending. The journal records a conservative unknown outcome at
that boundary, then retains the returned response and host completion proof.
Completion establishes local quiescence; an unknown response still permits no
replacement dispatch or target release.

`ReconcileWorkflowNativeGitSubmission` reads the actual source/target refs and
independently fetches the retained method history. A lost response remains
unknown evidence while a verified actual result can become Git-applied. No
forge receipt is created. Definitive rejection requires both a completed
transport and independently observed refs. Changed inputs seal non-application
and block stale authorization; unchanged refs retain a permissions/capability
blocker. Missing responses, timeouts and changed refs alone cannot seal an
unknown attempt. Observation callbacks retain the invocation, budget, journal,
entry and generation fences.

T9.3 retains bounded canonical observations and explicit reconciliation.
`ObserveWorkflowNativeSubmission` retains a single canonical observation;
`PollWorkflowNativeSubmission` uses exponential backoff within the original
persisted deadline. Both invoke the existing forge observation port outside the
common repository mutation lock, then retain sanitized evidence only under the
unchanged invocation, budget, journal, entry and claim-generation fence.
Repeated unchanged observations are deduplicated. Deadline expiry, external PR
edits and unavailable target history project explicit reconciliation blockers.
A fresh explicit observation can reconcile that same intent after the polling
bound without resetting its deadline or consuming another submission.

Lost transport output remains unknown evidence. A retained, independently
verified source-P and target-history observation can attribute the original
transaction after later source work arrives. Neither polling nor canonical
merged metadata creates a receipt, replaces a candidate or rolls back refs.

## Files of Interest

- `internal/core/workflow_native_dispatch.go`: host publication entry point.
- `internal/core/workflow_native_dispatch_guard.go` and `workflow_native_dispatch_store.go`: fresh authorization and retained dispatcher identity.
- `internal/core/workflow_native_dispatch*_test.go`: real receive-pack, stale owner, policy race and overlapping live dispatch fixtures.
- `internal/core/workflow_review_native.go`: shared capability-input binding with the original review fence retained.
- `internal/core/workflow_native_dispatch_evidence.go` and `workflow_native_dispatch_proof.go`: durable send/response/completion facts and their exact evidence checks.
- `internal/core/workflow_native_reconcile_git.go` and its tests: independent Git classification, atomic rejection, real SSH report loss and stale observation callbacks.

- `internal/core/workflow_native_observation*.go`: bounded canonical observation, retained proof, projection and real Git/adapter fixtures.

## Validation

The initial real-dispatch test reached the absent dispatch boundary and failed
without publishing. A first implementation then exposed the existing review
phase fence after intent moved the entry into publication; a dispatch-specific
recheck preserves the original review phase rules.

T9.1 development checks passed with Go 1.27.1:

- `make verify` — zero lint issues; core 2243.062 s, app 108.440 s and all other packages passed.

- `go test ./internal/core -run '^(TestNativeDispatch.*|TestNativeSubmission.*|TestNativeReview.*|TestNativeBudget.*)$' -count=1` — passed in 254.104 s.
- `go test -race ./internal/core -run '^(TestNativeDispatch.*|TestNativeSubmission.*|TestNativeReview.*|TestNativeBudget.*)$' -count=1` — passed in 381.919 s.
- `make lint-strict` — zero issues.
- `git diff --check` — passed.

The seven changed Go files remained unchanged throughout the successful checks.
The required pre-commit repository check passed on that same frozen source.
All owned focused and repository check processes and fixtures finished.

T9.2 development checks on the current fourteen-file Go source passed:

- `make verify` — zero lint issues; core 2311.960 s, app 110.734 s and all other packages passed.
- `go test ./internal/core -run '^(TestNativeDispatchEvidence|TestNativeGit.*)$' -count=1` — passed in 61.909 s, including actual SSH report truncation after server application.
- `go test ./internal/core -run '^(TestNativeDispatch.*|TestNativeGit.*|TestNativeSubmission.*|TestNativeReview.*|TestNativePreparation.*|TestNativeBudget.*|TestWorkflowGit.*)$' -count=1` — passed in 355.662 s.
- `go test -race ./internal/core -run '^(TestNativeDispatch.*|TestNativeGit.*|TestNativeSubmission.*|TestNativeReview.*|TestNativePreparation.*|TestNativeBudget.*|TestWorkflowGit.*)$' -count=1` — passed in 560.525 s.
- `make lint-strict` — zero issues.
- `git diff --check` — passed.

The initial real-dispatch evidence test failed because the completed transport
facts were absent. The implementation now retains those facts and independently
classifies the actual refs/history. Unknown unchanged and changed-ref outcomes
remain excluded, while stale observer results leave the common journal intact.

Focused-check Acta activity: `01M4VCSGYGAEN61C0KYC9XAK7W`.
Broader normal/race activity: `01M4VDRC8Z5TQAX002QF2DW5TW`.

Broader focused normal/race checks and the required pre-commit `make verify`
passed on fourteen unchanged Go files, with source digest
`839b196f9e05fd84bab9e3e0e0f88df0fceb627442bea018de2de3e53905f53d`.
The owned `chr-native-slice9-t92-validation-839b196f.service` finished with
exit status zero and no live check process. Command/process evidence is retained
under `/tmp/chr-native-slice9-t9-2-*`. All owned fixtures finished. These are
development checks; final exact-head slice evidence remains pending.

T9.3's later-source regression initially failed because retained actual Git
application could not attribute a lost-response transaction after source
advancement. The observer now validates that prior exact proof without
rewriting the raw response. An initial focused run passed in 36.039 s before
adding the stalled-read and actual lost-SSH canonical cases. The final formatter
and strict linter reported zero issues, and `git diff --check` passed.

The nine changed Go files remained unchanged at source digest
`f3d69423789f84537de347f594c0701a76bd2a3bcb43bb5040465b8381a4bcbc`.
The owned `chr-native-slice9-t93-validation-f3d69423.service` executes these
development checks in order:

- `go test ./internal/core -run '^(TestNativeCanonical.*|TestNativeLaterSource|TestNativeObservation.*|TestNativeSubmission.*|TestNativeGit.*|TestNativeDispatch.*|TestNativeReceipt.*)$' -count=1`
- `go test -race ./internal/core -run '^(TestNativeCanonical.*|TestNativeLaterSource|TestNativeObservation.*|TestNativeSubmission.*|TestNativeGit.*|TestNativeDispatch.*|TestNativeReceipt.*)$' -count=1`
- `make verify`

Command/process evidence is retained under `/tmp/chr-native-slice9-t9-3-*`.
The focused normal check passed in 162.892 s on the unchanged nine-file source;
its owned check process finished. Acta activity: `01M4VHAK680FBXG2P887XS9Q90`.
The race check passed in 271.828 s on the same unchanged source, and its owned
check process finished. Acta activity: `01M4VHM9JAY36FRCT1QB06Z57X`.
Pre-commit `make verify` passed with zero lint issues; core 2394.437 s, app
108.657 s and all other packages passed. The owned service finished with exit
status zero and no live check process. All owned check fixtures finished.
T9.3 implementation and development validation are complete; T9.4 and canonical
PR preparation remain pending.

Required ordered exact-head slice checks remain pending until all four tasks
and canonical PR references are committed and pushed.

T9.1 implementation: [e4531667cbbeb82922d79500cb7bd79c72e2efb9](https://forge.adrianpk.com/adrianpk/chr/commit/e4531667cbbeb82922d79500cb7bd79c72e2efb9).
Acta commit activity: `01M4VAYVXVK7AMTY6J2RA7TYC1`.

T9.2 implementation: [ae54f89aff1248e7cee90c5e033ac923ec7c77c7](https://forge.adrianpk.com/adrianpk/chr/commit/ae54f89aff1248e7cee90c5e033ac923ec7c77c7).
Acta commit activity: `01M4VFXT6NZ0S5TT0V9BJ8W6PF`.

## Risks and Follow-ups

T9.1, T9.2 and T9.3 are implemented and development-checked. T9.4 remains
ready for crash/stop boundaries.
Canonical PR and verified integration remain pending.

A temporary overlay probe for the pending T9.4 stop boundary reproduced a
terminal `cancelled` status for a Git-applied native submission still awaiting
canonical observation with no receipt. The probe failed in 5.949 s and did not
change any T9.3 repository source file. Acta investigation activity:
`01M4VHYTVR3WFMJQRF7Z5H30GX`. The correction remains part of T9.4.

Preliminary T9.4 temporary-overlay probes used actual host exits at intent,
dispatch, send, server application and response retention. All five cases passed
in 35.373 s without replacement dispatch or budget refill; actual applied Git
results were independently recovered. A live-host lease-expiry probe passed in
6.937 s without transferring publication authority. Owned probe subprocesses
finished. Acta activity: `01M4VJR8909TNKYHFRDP9PCE3P`,
`01M4VK4XH0AKPKE937H3NK7WJV`. T9.4 still requires its implementation and final
development checks.
