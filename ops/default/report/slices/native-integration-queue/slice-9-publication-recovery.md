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

T9.1 and T9.2 implementation and development validation are complete;
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

## Files of Interest

- `internal/core/workflow_native_dispatch.go`: host publication entry point.
- `internal/core/workflow_native_dispatch_guard.go` and `workflow_native_dispatch_store.go`: fresh authorization and retained dispatcher identity.
- `internal/core/workflow_native_dispatch*_test.go`: real receive-pack, stale owner, policy race and overlapping live dispatch fixtures.
- `internal/core/workflow_review_native.go`: shared capability-input binding with the original review fence retained.
- `internal/core/workflow_native_dispatch_evidence.go` and `workflow_native_dispatch_proof.go`: durable send/response/completion facts and their exact evidence checks.
- `internal/core/workflow_native_reconcile_git.go` and its tests: independent Git classification, atomic rejection, real SSH report loss and stale observation callbacks.

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

Required ordered exact-head slice checks remain pending until all four tasks
and canonical PR references are committed and pushed.

T9.1 implementation: [e4531667cbbeb82922d79500cb7bd79c72e2efb9](https://forge.adrianpk.com/adrianpk/chr/commit/e4531667cbbeb82922d79500cb7bd79c72e2efb9).
Acta commit activity: `01M4VAYVXVK7AMTY6J2RA7TYC1`.

## Risks and Follow-ups

T9.1 and T9.2 are implemented and development-checked. T9.3 and T9.4 remain
ready for canonical convergence and crash/stop boundaries. Canonical PR and
verified integration remain pending.
