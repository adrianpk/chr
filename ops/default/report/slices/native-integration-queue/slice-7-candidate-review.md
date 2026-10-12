# Slice 7: Combined validation and review

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-7-native-candidate-review`
PR: pending

## Purpose

Bind combined validation and review to the exact prepared result, including
repaired candidates and changes to the source, target, method or policy.

## Delivered Behavior

T7.1 connects the existing host gate to native candidate records. Before any
combined command executes, the retained input freezes the admitted queue entry,
claim generation and full trusted capability for exact S/B/P. Explicit native
publication requires the native observer; a preliminary or source-only guard
cannot substitute for this result-bound capability.

The host verifies the clean isolated checkout and configured method history,
runs the unchanged repository-owned unit gate through existing owned processes,
and rechecks the queue fence before each command and at completion. Current
source, target, policy and capability are observed again before acceptance.

Successful validation retains content-addressed source, history, preparation,
capability and command-result evidence. The candidate, prepared attempt and
queue pointer advance together through the existing transition journal to
reviewing. Restart reads the retained evidence without repeating the host gate
or previously accepted task execution. Reuse requires the same queue claim
generation even when S/B/P and the candidate pointer are unchanged.

A textually clean semantic regression, checkout mutation or stale queue/policy
observation cannot produce candidate authority. Failure checkpoints preserve
observations and the original task evidence. A passing command exit alone does
not grant review approval or publication.

## Implementation Notes

The existing preparation and process owners execute validation. The native
provider is a trusted host port for current policy and exact-result transport
observations. The caller must have admitted and advanced the queue entry to
validating before preparation; the host rejects foreign or stale entry bindings.
Immutable evidence may survive interruption without an authoritative candidate
pointer. Successful candidate and queue pointers are committed together.

Queue phase and stop projection preserve the running host-validation credential
until the existing owner retains completion or stopped observations. This allows
a stale queue fence to fail with a retained checkpoint and keeps pause/cancel
capacity tied to actual process-group quiescence.

## Contracts Added or Changed

`WorkflowNativePreparationProvider` observes the full capability for the exact
host-prepared Git proof. `WorkflowNativePreparationInput` freezes that capability
and the queue binding before validation. It extends the existing preparation
input with an optional native field. Retained native validation must match the
complete preparation image, including method/history and policy evidence.

## Files of Interest

- `internal/core/workflow_preparation_native.go`: admission, capability and queue fences.
- `internal/core/workflow_preparation_native_evidence.go`: immutable evidence and atomic candidate attachment.
- `internal/core/workflow_preparation_result.go`: host completion checks and failure retention.
- `internal/core/workflow_native_record_evidence.go`: complete retained gate-input binding.
- `internal/core/workflow_preparation_native_test.go`: real method, restart and semantic failure fixtures.
- `internal/core/workflow_preparation_native_fence_test.go`: queue and policy movement during a host gate.
- `internal/core/workflow_preparation_native_generation_test.go`: prior-generation evidence rejection after restart.
- `internal/core/workflow_preparation_native_stop_test.go`: queue stops retain host observations and release stopped capacity.

## Validation

T7.1 development checks passed before its implementation commit, using Go
1.27.1 and the existing isolated Git/agent fixture environment:

- `make verify` — zero lint issues; all packages passed, including core in 1417.044 s and app in 103.436 s.
- `go test ./internal/core -run '^(TestNativePreparation|TestNativeRecord|TestNativeSubmission|TestWorkflowPreparation)' -count=1` — passed in 183.144 s.
- `go test ./internal/core -run '^TestNativePreparationStop$' -count=1` — passed in 6.367 s.
- `go test -race ./internal/core -run '^(TestNativePreparation|TestNativeRecord|TestNativeSubmission|TestWorkflowPreparation)' -count=1` — passed in 316.053 s.
- `git diff --check` — passed.

T7.1 implementation commit: [3f1aac6910b7e8885146a544a735db69f49c1a94](https://forge.adrianpk.com/adrianpk/chr/commit/3f1aac6910b7e8885146a544a735db69f49c1a94). Acta commit activity: `01M4RW111YN23ZPXZ87AHKG9CB`.

The queue-generation regression was reproduced before correction. The final
slice candidate has not been validated. All owned check processes exited
naturally; no external fixture was started.

## Risks and Follow-ups

T7.2 is active and T7.3–T7.4 remain ready in the approved tracker: candidate-bound grants, bounded
merger repair routing and integration drift handling. Final slice candidate
checks and canonical PR evidence remain pending.
