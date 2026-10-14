# Slice 9: Publication and uncertainty recovery

Status: reviewing
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

Recovery retains actual transport facts, independent ref/history proof and
bounded canonical PR observations for the same submission. Stop intent preserves
unresolved publication and receipt finalization; independently sealed changed-input
non-application may complete the retained stop without refunding consumption.

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

All four tasks are implemented and development-checked. The controller records
ordered exact-HEAD validation after all task and PR-reference commits.

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

T9.4 retains stop across crash and recovery boundaries. Native stop
records intent in both invocation and queue entry while retaining reconciliation
instead of reporting terminal pause/cancellation. Receipt-confirmed work remains
available for required finalization. Independently sealed changed-input
non-application can finalize the retained stop. The trusted send boundary reads
stop intent under the mutation lock before any commands; completion still
retains actual transport facts.

Repository fixtures now use actual host exits around intent, dispatch, send,
server application and response retention. Restart and expired task leases
cannot replace an earlier dispatcher or refill submission counts. Pending stop
can observe canonical merged metadata while receipt/fanout remains unfinished.

## Files of Interest

- `internal/core/workflow_native_dispatch.go`: host publication entry point.
- `internal/core/workflow_native_dispatch_guard.go` and `workflow_native_dispatch_store.go`: fresh authorization and retained dispatcher identity.
- `internal/core/workflow_native_dispatch*_test.go`: real receive-pack, stale owner, policy race and overlapping live dispatch fixtures.
- `internal/core/workflow_review_native.go`: shared capability-input binding with the original review fence retained.
- `internal/core/workflow_native_dispatch_evidence.go` and `workflow_native_dispatch_proof.go`: durable send/response/completion facts and their exact evidence checks.
- `internal/core/workflow_native_reconcile_git.go` and its tests: independent Git classification, atomic rejection, real SSH report loss and stale observation callbacks.

- `internal/core/workflow_native_observation*.go`: bounded canonical observation, retained proof, projection and real Git/adapter fixtures.

- `internal/core/workflow_native_stop.go`, `workflow_native_stop_test.go` and `workflow_native_dispatch_crash_test.go`: publication-aware stop, crash/restart and lease fixtures.

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
All four tasks are implemented and development-checked. Canonical PR
preparation follows the recorded task commits.

Required ordered exact-HEAD slice checks are recorded separately by the
controller for the clean PR head after all task and reference commits.

T9.1 implementation: [e4531667cbbeb82922d79500cb7bd79c72e2efb9](https://forge.adrianpk.com/adrianpk/chr/commit/e4531667cbbeb82922d79500cb7bd79c72e2efb9).
Acta commit activity: `01M4VAYVXVK7AMTY6J2RA7TYC1`.

T9.2 implementation: [ae54f89aff1248e7cee90c5e033ac923ec7c77c7](https://forge.adrianpk.com/adrianpk/chr/commit/ae54f89aff1248e7cee90c5e033ac923ec7c77c7).
Acta commit activity: `01M4VFXT6NZ0S5TT0V9BJ8W6PF`.

T9.3 implementation: [48a9539d023e63c96f04aa61ecb6baa46d77f9e2](https://forge.adrianpk.com/adrianpk/chr/commit/48a9539d023e63c96f04aa61ecb6baa46d77f9e2).
Acta commit activity: `01M4VKZWK3H5X4ZDRMEDMHBP6J`.

A temporary overlay probe for the pending T9.4 stop boundary reproduced a
terminal `cancelled` status for a Git-applied native submission still awaiting
canonical observation with no receipt. The probe failed in 5.949 s and did not
change any T9.3 repository source file. Acta investigation activity:
`01M4VHYTVR3WFMJQRF7Z5H30GX`. The correction is implemented in T9.4.

Preliminary T9.4 temporary-overlay probes used actual host exits at intent,
dispatch, send, server application and response retention. All five cases passed
in 35.373 s without replacement dispatch or budget refill; actual applied Git
results were independently recovered. A live-host lease-expiry probe passed in
6.937 s without transferring publication authority. Owned probe subprocesses
finished. Acta activity: `01M4VJR8909TNKYHFRDP9PCE3P`,
`01M4VK4XH0AKPKE937H3NK7WJV`. The final repository-source checks below supersede these preliminary probes.

T9.4's initial focused repository tests passed in 57.018 s before adding the
physical-receipt/fanout stop case and splitting functions for strict linting.
The first broader check failed in `TestNativeStopReceipt`: its receipt fixture
referenced a second transaction rather than the retained submission's exact
real atomic Git response. Correcting that fixture reference preserved the
production identity check; `go test ./internal/core -run '^TestNativeStopReceipt$' -count=1`
passed in 5.464 s. Acta activity: `01M4VN3ZEC0YMJ5CDNNKP1B27P`.
Failed source/command/process evidence is retained under
`/tmp/chr-native-slice9-t9-4-failed-3eb53332/`; its owned service finished.

The corrected source passed `make lint-strict` with zero issues and
`git diff --check`. The prior six-file source remained unchanged at digest
`ec41aea2918b199642bb38a26fea98958d8704031ae7c6aade663dafdde30d6b`.
The owned `chr-native-slice9-t94-validation-ec41aea2.service` executed these
development checks on that unchanged source:

- `go test ./internal/core -run '^(TestNativeStop.*|TestNativeFault.*|TestNativeDispatch.*|TestNativeGit.*|TestNativeCanonical.*|TestNativeSubmission.*|TestWorkflowStop.*|TestWorkflowClosingStop|TestNativePreparationStop)$' -count=1` — passed in 251.101 s.
- `go test -race ./internal/core -run '^(TestNativeStop.*|TestNativeFault.*|TestNativeDispatch.*|TestNativeGit.*|TestNativeCanonical.*|TestNativeSubmission.*|TestWorkflowStop.*|TestWorkflowClosingStop|TestNativePreparationStop)$' -count=1` — passed in 423.165 s.
- `make verify` — failed in `TestNativeCanonicalDeadline`; strict lint, app (110.899 s) and all other packages passed, while core reported failure at 2438.219 s.

Normal-check Acta activity: `01M4VND29SGWM7T1H9WYXPNEE4`.
Race-check Acta activity: `01M4VNTTG6B32V9XGNCVG0SMMT`.
That owned service finished with no live check process. Its complete source,
commands, logs and process evidence are retained under
`/tmp/chr-native-slice9-t9-4-failed-ec41aea2/`.

The full-suite failure exposed a timing assumption: the last real Git read can
reach the persisted polling deadline and return `context.DeadlineExceeded`.
The assertion now accepts that documented typed error while still requiring the
same retained deadline blocker, unchanged deadline, zero external IO after
expiry, and explicit recovery of the same original PR. Other errors still fail.
Deadline/stalled-observer tests passed in 14.256 s before extracting the explicit
recovery helper for the unchanged strict complexity bound. Acta activity:
`01M4VR8F4YW7EYTJE90EMFFM68`.

The seven-file corrected source passed `make lint-strict` with zero issues and
is frozen at digest `e305eefe0d88cebe669d08a4f8d2a7388251eec1b61a2793dd767627bcff3758`.
The owned `chr-native-slice9-t94-validation-e305eefe.service` runs the same
ordered normal/race development commands and required pre-commit `make verify`
listed above. The normal check passed in 247.969 s on the unchanged seven-file
source; its owned process finished. Acta activity: `01M4VRMT3JFMWMM91G9Q3HKBVG`.
The race check passed in 433.723 s on that same unchanged source; its owned
process finished. Race-check Acta activity: `01M4VS2X5ZKDF6DWJBP6CS463R`.
Pre-commit `make verify` passed with zero lint issues; core 2473.043 s, app
113.530 s and all other packages passed on the unchanged seven-file source.
The owned service finished with exit status zero and no live check process.
All owned check fixtures finished. Evidence is retained under
`/tmp/chr-native-slice9-t9-4-*`. These are development checks; ordered clean-HEAD
slice evidence is recorded by the controller after task and PR references.

T9.4 implementation: [a06c392fc7a6b9169694989e436ea6afdec4010e](https://forge.adrianpk.com/adrianpk/chr/commit/a06c392fc7a6b9169694989e436ea6afdec4010e).
Acta commit activity: `01M4VVFZY92DBXEJYPP80BP2W8`.

## Risks and Follow-ups

None recorded.
