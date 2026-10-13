# Slice 7: Combined validation and review

Status: reviewing
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-7-native-candidate-review`
PR: [#381](https://forge.adrianpk.com/adrianpk/chr/pulls/381)

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

T7.2 extends the retained review input with the native candidate, queue claim
generation and full capability. Original S and published P remain explicit;
source and history manifests, validation references, ownership, required policy,
server capability and publisher identity participate in the approval hash.
Human, trusted and agentic modes share this binding through the central service.

Every review, explicit approval, grant and publication check re-observes the
current source, target, profile, method and trusted native capability. It also
rechecks local policy after capability IO. Source-only native review and missing
trusted observers fail before a wait, decision or role claim is retained. A new
queue generation cannot inherit an older grant even with identical S/B/P.
Restart and the transition to publication readiness preserve a current grant
without repeating its host gate or decision.

T7.3 retains actionable failures of the unchanged combined gate and native
review as immutable, result-bound findings. The queue returns to merging without
rewinding accepted source tasks or removing their retained evidence. Clean
passing candidates reserve no merger attempt.

Merger admission and every owned action require the current queue generation,
source, target, method, policy, scope and finding references. The existing broker
and persistent budget bound the repair. Stopped, mutated or stale host-gate
results cannot request semantic repair, and the legacy task-repair path cannot
bypass the native route.

Host acceptance leaves the repaired result pending validation. After restart,
the existing preparation entrypoint independently verifies its Git history and
attaches exactly that accepted P to a fresh isolated checkout. It reruns the
unchanged gate, retains merger provenance in the new candidate and requires a
fresh candidate-bound review. A valid deterministic P cannot replace the
accepted semantic result. Rebase repair preserves original task attribution,
replay order and raw source commit messages.

T7.4 replaces native preparation after an external target or policy change.
Queue generation, stale preparation/review and retained drift consumption
advance together. The new isolated result receives the unchanged host gate and
a fresh review; source task proofs and prior role consumption remain retained.
A previous generation's accepted merger result cannot substitute for new-base
validation. If the unchanged gate exposes a new regression, it retains fresh
bounded merger findings.

Source observation verifies the original canonical PR identity, repository and
advertised branch as well as the PR head. New source content has no host task
proof for the frozen scope, so the host retains both observations and blocks for
an explicit ownership/scope decision. It does not admit appended content or
rewind source tasks. The same decision applies after a sealed atomic rejection
or source movement during host validation.

Target movement during a clean host gate retains a stale checkpoint and permits
replacement preparation after restart. Stops, checkout mutation and uncertain
publication retain their existing barriers. Final target drift fails before a
new submission intent; definitive non-application releases target exclusion
before replacement preparation while retaining the consumed submission.

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
capacity tied to actual process-group quiescence. The same ownership rule now
applies to reviewers, so queue movement retains a stale review checkpoint
before capacity is released.

Native findings and the queue transition are retained together through the
existing journal. Merger completion clears its finished task claim only after
process quiescence, allowing the next host gate to acquire ownership. Private
repair checkpoints import original source and base objects because rewritten
rebase and squash results need not reach those histories.

Native input replacement uses the existing queue transition and budget owners.
Queued, preparing and validating phases are checked before their combined
journal update, and each accepted replacement base is charged once within the
existing invocation-wide limit. Old candidate, review and merger observations
remain immutable history. Publication uncertainty must be sealed with existing
non-application and dispatch-quiescence proof before replacement can proceed.

## Contracts Added or Changed

`WorkflowNativePreparationProvider` observes the full capability for the exact
host-prepared Git proof. `WorkflowNativePreparationInput` freezes that capability
and the queue binding before validation. It extends the existing preparation
input with an optional native field. Retained native validation must match the
complete preparation image, including method/history and policy evidence.
`WorkflowNativeReviewInput` adds the complete native candidate, claim generation
and capability to the existing optional review input. Publication observes the
result-bound capability instead of the preliminary source-only guard.

`WorkflowCandidateRepair` and its invocation binding retain the exact failed
candidate and host findings. `ReadWorkflowCandidateRepair` exposes them for
trusted request assembly without reserving a role. `WorkflowRepairedPreparationGit`
attaches the retained host-accepted result without substituting another proof.
`BeginWorkflowMergerCombinedRepair` provides a clean merge/squash semantic
baseline; rebase repair continues through the original replay-step contract.
Preparation input and native candidate records now bind accepted merger attempt
references alongside the existing history and validation evidence.

Native preparation now handles external target changes from validation, review
and publication readiness. Source changes retain the typed
`source_revision_requires_reconciliation` blocker with
`verify_source_ownership_and_frozen_scope` as the required action. No new source
revision gains host proof or publication authority through this decision.

## Files of Interest

- `internal/core/workflow_preparation_native.go`: admission, capability and queue fences.
- `internal/core/workflow_preparation_native_evidence.go`: immutable evidence and atomic candidate attachment.
- `internal/core/workflow_preparation_result.go`: host completion checks and failure retention.
- `internal/core/workflow_native_record_evidence.go`: complete retained gate-input binding.
- `internal/core/workflow_preparation_native_test.go`: real method, restart and semantic failure fixtures.
- `internal/core/workflow_preparation_native_fence_test.go`: queue and policy movement during a host gate.
- `internal/core/workflow_preparation_native_generation_test.go`: prior-generation evidence rejection after restart.
- `internal/core/workflow_preparation_native_stop_test.go`: queue stops retain host observations and release stopped capacity.

- `internal/core/workflow_review_native.go`: native review binding and queue fences.
- `internal/core/workflow_review_native_capability_test.go`: fingerprint and source-rewrite revocation.
- `internal/core/workflow_review_native_guard_test.go`: current policy, history and generation checks.
- `internal/core/workflow_review_native_fence_test.go`: running reviewer checkpoints after queue or policy movement.
- `internal/app/service_workflow_review_native_test.go`: shared service approval and guard behavior.

- `internal/core/workflow_candidate_repair.go`: native review routing and retained findings.
- `internal/core/workflow_candidate_repair_admission.go`: current finding, source and capability guards.
- `internal/core/workflow_candidate_repair_gate.go`: queue resumption and exact accepted-result preparation.
- `internal/core/workflow_candidate_repair_git.go`: independent Git verification and gate checkout attachment.
- `internal/core/workflow_candidate_repair_combined.go`: merge/squash semantic repair baseline.
- `internal/core/workflow_candidate_repair_gate_test.go`: real merger, new host gate and review across all methods.
- `internal/core/workflow_candidate_repair_guard_test.go`: clean-result and changed finding/policy/scope rejection.
- `internal/core/workflow_candidate_repair_substitution_test.go`: accepted-result substitution rejection.

- `internal/core/workflow_preparation_drift_replace.go`: atomic queue-generation replacement and retained drift budget.
- `internal/core/workflow_preparation_source_revision.go`: canonical source observation and ownership/scope decision.
- `internal/core/workflow_preparation_drift_checkpoint.go`: clean target-drift checkpoint recovery.
- `internal/core/workflow_preparation_drift_parallel_test.go`: real shared-target queues and renewed combined evidence.
- `internal/core/workflow_preparation_drift_exclusion_test.go`: publication barriers and sealed source/target rejection recovery.

## Validation

T7.1 development checks passed before its implementation commit, using Go
1.27.1 and the existing isolated Git/agent fixture environment:

- `make verify` — zero lint issues; all packages passed, including core in 1417.044 s and app in 103.436 s.
- `go test ./internal/core -run '^(TestNativePreparation|TestNativeRecord|TestNativeSubmission|TestWorkflowPreparation)' -count=1` — passed in 183.144 s.
- `go test ./internal/core -run '^TestNativePreparationStop$' -count=1` — passed in 6.367 s.
- `go test -race ./internal/core -run '^(TestNativePreparation|TestNativeRecord|TestNativeSubmission|TestWorkflowPreparation)' -count=1` — passed in 316.053 s.
- `git diff --check` — passed.

T7.1 implementation commit: [3f1aac6910b7e8885146a544a735db69f49c1a94](https://forge.adrianpk.com/adrianpk/chr/commit/3f1aac6910b7e8885146a544a735db69f49c1a94). Acta commit activity: `01M4RW111YN23ZPXZ87AHKG9CB`.

The queue-generation regression was reproduced before correction. These are
task development checks; exact-HEAD slice evidence is retained separately by
the controller and canonical PR. All owned check processes exited naturally;
no external fixture was started.

T7.2 development checks passed before its implementation commit, using the
same Go 1.27.1 fixture environment:

- `make verify` — zero lint issues; all packages passed, including core in 1934.610 s and app in 157.802 s.
- `go test ./internal/core ./internal/app -run '^(TestNativeReview|TestNativeSubmission|TestServiceNativeReview)' -count=1` — core passed in 288.133 s and app in 11.412 s.
- `go test ./internal/core -run '^TestNativeReviewFence$' -count=1` — passed in 14.718 s.
- `go test ./internal/core -run '^TestNativeReviewPolicyIO$' -count=1` — passed in 13.381 s.
- `go test -race ./internal/core ./internal/app -run '^(TestNativeReview|TestNativeSubmission|TestNativePreparation|TestWorkflowReview|TestWorkflowPreparation|TestServiceNativeReview|TestServiceWorkflowReview|TestServiceWorkflowPreparation)' -count=1` — core passed in 1012.273 s and app in 47.812 s.
- `git diff --check` — passed.

T7.2 implementation commit: [c1c3d9dc954e9d3888c55a9a63dfac4fc04f6b59](https://forge.adrianpk.com/adrianpk/chr/commit/c1c3d9dc954e9d3888c55a9a63dfac4fc04f6b59). Acta commit activity: `01M4RZ9E4YV2JFRSJFAAG9TPTH`.

Source-only native approval and stale policy reuse were reproduced before
correction. Queue movement also reproduced loss of the running reviewer
credential. The initial repository and race checks failed only because a capacity
fixture passed a label instead of its marker path. The corrected-source checks
passed; all owned check processes exited naturally. These are task development checks;
exact-HEAD slice evidence is retained separately by the controller and canonical PR.

T7.3 implementation commit: [91676739a6706575df5ee1e644acf29189cfc62b](https://forge.adrianpk.com/adrianpk/chr/commit/91676739a6706575df5ee1e644acf29189cfc62b). Acta commit activity: `01M4S33CF0TC96K7KTAP1TNASA`.

T7.3 development checks passed before its implementation commit, using Go
1.27.1 and the same isolated fixture environment:

- `make verify` — zero lint issues; all packages passed, including core in 1981.944 s and app in 138.499 s.
- `go test ./internal/core -run '^(TestNativeCandidate|TestNativePreparationFailure)' -count=1` — passed in 204.054 s.
- `go test -race ./internal/core ./internal/app -run '^(TestNativeCandidate|TestNativeMerger|TestNativePreparation|TestNativeReview|TestWorkflowPreparation|TestWorkflowReview|TestServiceNativeReview|TestServiceWorkflowPreparation|TestServiceWorkflowReview)' -count=1` — core passed in 1324.484 s and app in 40.741 s.
- `git diff --check` — passed.

The initial native review test reproduced source-task rewind. Composed repair
checks then exposed a retained finished merger claim, missing original Git
objects for rewritten results and a stale invocation revision after the queue
transition. Each was corrected; the final focused normal and race checks passed.
The repository and race checks exited naturally. All 27 changed Go files
matched the frozen source throughout validation. No external fixture was started.
These are task development checks; exact-HEAD slice evidence is retained
separately by the controller and canonical PR.

T7.4 implementation commit: [19a339e92300df7141f57458c7892de2876cecac](https://forge.adrianpk.com/adrianpk/chr/commit/19a339e92300df7141f57458c7892de2876cecac). Acta commit activity: `01M4S6N0V0VZAQ494X453TY0SM`.

T7.4 development checks passed with Go 1.27.1:

- `make verify` — zero lint issues; all packages passed, including core in 2091.189 s and app in 147.754 s.
- `go test ./internal/core ./internal/app -run '^(TestNative.*Drift|TestNativeCandidate|TestNativePreparation|TestNativeReview|TestNativeSubmission|TestWorkflowPreparation|TestWorkflowReview|TestServiceNativeReview|TestServiceWorkflowPreparation|TestServiceWorkflowReview)' -count=1` — core passed in 918.729 s and app in 28.719 s.
- `go test ./internal/core -run '^TestNativeSourceDuringGate$' -count=1` — passed in 5.180 s.
- `go test -race ./internal/core ./internal/app -run '^(TestNative.*Drift|TestNativePreparationFence|TestNativePreparationStop|TestWorkflowParallelPublication|TestServiceNativeReview)' -count=1` — core passed in 370.330 s and app in 20.110 s.
- `go test -race ./internal/core -run '^TestNativeSourceDuringGate$' -count=1` — passed in 8.562 s.

- `git diff --check` — passed.

The baseline preparation path reproduced target-drift rejection. Initial new
fixtures required correction to avoid a target-only gate during source work,
import prepared objects into the shared remote and replace legacy fixture
bindings before native validation. The mid-gate regression exposed a paused
checkpoint that required clean target-drift recovery. Initial strict linting
rejected whitespace in a new test; the corrected lint passed. Repository
verification and all focused checks exited naturally. No external fixture was
started. Exact-HEAD slice evidence is retained separately by the controller
and canonical PR.

The first final-candidate attempt at
`b7dae01db495bc06d6a1398fe24e75eee447ae3b` passed `make verify` with zero
lint issues (core 1901.937 s, app 105.267 s), then passed
`go test ./internal/core ./internal/app -count=1` (core 1916.874 s,
app 111.218 s). The core race suite reached the inherited 45-minute outer
test deadline while `TestWorkflowReviewRepairJournal` had run for six seconds.
The app race suite passed in 297.814 s. Retained diagnostics contain a deadline
failure, with no assertion failure or data race; the final whitespace command
was not reached. No fixture process from that build remained after completion.
Controller diagnostics are retained in
`checks/b7dae01db495bc06d6a1398fe24e75eee447ae3b.log` beside the checkpoint.
Acta diagnosis: `01M4SD6C7VG06FWC2A3KRR64VB`.

The replacement final runner retains the same Go version, isolated fixture
settings, assertions and ordered commands, with a 90-minute outer Go test
deadline. The controller records the complete sequence against the clean head
that includes this report update.

## Risks and Follow-ups

T7.1–T7.4 are implemented and checked. Canonical PR [#381](https://forge.adrianpk.com/adrianpk/chr/pulls/381) retains the
slice candidate. PR-open activity: `01M4S6V1EXT2MKX1F0DWW9YCGA`.
The controller records `make verify`,
`go test ./internal/core ./internal/app -count=1`,
`go test -race ./internal/core ./internal/app -count=1` and `git diff --check`
in that order for the clean final PR head. Integration remains pending.
