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

T9.1 implementation and development validation are complete; canonical PR and
verified integration remain pending.

## Files of Interest

- `internal/core/workflow_native_dispatch.go`: host publication entry point.
- `internal/core/workflow_native_dispatch_guard.go` and `workflow_native_dispatch_store.go`: fresh authorization and retained dispatcher identity.
- `internal/core/workflow_native_dispatch*_test.go`: real receive-pack, stale owner, policy race and overlapping live dispatch fixtures.
- `internal/core/workflow_review_native.go`: shared capability-input binding with the original review fence retained.

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

Required ordered exact-head slice checks remain pending until all four tasks
and canonical PR references are committed and pushed.

T9.1 implementation: [e4531667cbbeb82922d79500cb7bd79c72e2efb9](https://forge.adrianpk.com/adrianpk/chr/commit/e4531667cbbeb82922d79500cb7bd79c72e2efb9).
Acta commit activity: `01M4VAYVXVK7AMTY6J2RA7TYC1`.

## Risks and Follow-ups

T9.1 is implemented and development-checked. T9.2 through T9.4 remain recorded
for outcome classification, canonical convergence and crash/stop boundaries.
