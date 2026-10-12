# Slice 6: Bounded merger role

Status: drafting
Delivery set: native-integration-queue
Plan: [Native Integration Corrective Delivery](../../../plan/native-integration-queue.md)
Tracker: [Native Integration Tracker](../../../tracker/native-integration-queue.md)
Branch: `feat/slice-6-native-merger-role`
PR: pending

## Purpose

Resolve integration conflicts through an isolated, finite merger role whose
proposals require independent host verification before candidate acceptance.

## Delivered Behavior

T6.1 defines protocol version 1 input and untrusted result contracts. The input
seals entry/unit identity, candidate revision/hash, claim generation, canonical
source/target refs, exact S/B, configured method, scope and opaque owned-workspace
identity. It includes bounded source/target intent, original history, conflict
and finding references, permitted paths/validation, replay checkpoint and finite
attempt/time/token allowances.

The parser rejects stale input or host fences, malformed/oversized output,
ambiguous JSON, escaped evidence paths, workflow/contract scope edits, incomplete
preserved requirements and invented approval/publication fields. Results retain
per-conflict rationale, bounded decision questions and supplementary local
observations without granting candidate or publication authority. Unknown usage
remains unknown.

T6.2 executes a certified model driver through an owned JSON-line tool broker.
Admission requires the current merging queue entry, host-assembled actionable
evidence, exact source/target observations, frozen scope/method and unchanged unit
gate. A private Git checkout preserves the worker and user checkout. The model
can read, write and delete permitted regular files; control files, filesystem
aliases, shell/local validation processes and publication/approval operations are
denied. The host retains proposals without committing or accepting their edits.

The existing shared agent pool, birth-identity supervisor and persistent budget
ledger own each launch. Timeout, pause and cancellation preserve consumed calls
and failure evidence. Driver exit cannot release capacity while descendants live.
Certified whole-call accounting can settle a finite allowance; unknown or partial
usage stays charged, and model-reported usage does not settle the ledger.

T6.3 retains actual Git conflicts at the correct original replay step. Host repair
records bind frozen input, per-conflict rationale, exact parents and tree diffs.
Continuations preserve already-applied and source-empty commits, byte-exact
original attribution and subsequent source changes. Merge and squash retain their
configured parents. Clean preparation produces a proof without a merger call.

Semantic planning uses original Git diffs to identify a unique source owner.
Repair can rewind that step while preserving the verified prefix. Ambiguous
ownership requires a concrete decision or an explicitly declared actionable
integration fix with host attribution and retained finding. Declared incompatible
requirements and unsupported generated/lockfile/binary/submodule/LFS procedures
cannot appear as resolved proposals. Procedures require exact regeneration and
frozen validation declarations plus host-observed tree-bound evidence.

## Implementation Notes

Protocol ownership stays in core beside existing preparation and integration
records. Actual evidence is read through the existing contained workflow store;
changed hashes, symlink traversal, foreign source manifests and incorrect replay
positions fail before acceptance. Existing task/review/receipt schemas are
unchanged. No dependency is added.

The driver certificate binds the actual executable checksum, exact arguments,
broker-only tool contract and accounting capability revision. The driver is part
of the host trust boundary, not an operating-system sandboxed agent. Built-in
filesystem, shell, network and MCP tools must be disabled. Unrestricted worker
CLIs are rejected. Exact validation commands remain frozen for later host
acceptance; they are not executable model tools.

Launch input, certificate reference, capacity/attempt identity and budget charge
are retained before the supervisor grants execution. Every model file request
rechecks the live queue and source/target fences. Proposal completion waits for
the owned process group to stop and retains stopped process/output hashes. Paused
invocations resume in integration. Both Git object formats are supported without
remotes, hooks, credential helpers or shared writable object storage.

Repair method verification recomputes actual baseline trees and compares actual
changed paths with the proposal and frozen permitted scope. It verifies original
step/predecessor identity, original attribution and actual Git file types. Removing
a repair record cannot hide a nondeterministic tree by recomputing the outer hash.
The trusted Git repair ports consume host-accepted changes; wiring retained runtime
state, independently accepting proposals and executing declared procedures belong
to T6.4. The model cannot call these ports or supply host procedure observations.

## Contracts Added or Changed

`WorkflowMergerInput`, `WorkflowMergerFence` and `WorkflowMergerResult` define the
role boundary. Input/output images use the existing 1 MiB file bound, lists use
128-entry limits and JSON nesting is limited to 32 levels. Attempts are positive
and capped at 10; runtime uses the existing worker-policy ceiling. Optional token
allowances require whole-call accounting and the exact admitted deadline.

`SealWorkflowMergerInput`, `CheckWorkflowMergerEvidence` and
`ParseWorkflowMergerResult` validate immutable inputs, actual retained evidence
and proposals. They neither launch an agent nor acquire capacity, reserve budget,
accept Git changes, run a gate or create review/publication approval.

`ExecuteWorkflowMerger`, `WorkflowMergerHarnessCertificate` and
`WorkflowMergerAttempt` add the owned runtime boundary. JSON-line exchanges are
limited to 128 messages and 8 MiB. Optional unmetered merger reservations retain
a runtime deadline and consumed attempt while reporting no token enforcement;
a finite invocation token budget requires a compatible whole-call allowance.

`WorkflowMergerRepairState`, `WorkflowIntegrationTreeRepair` and
`WorkflowMergerRepairPlan` retain actual preparation checkpoints, method effects
and bounded ownership decisions. `WorkflowMergerRequiredDecisions` rejects
declared incompatible intent and unsupported procedures. Original proof images
keep their hash representation when optional repair fields are absent.

## Files of Interest

- [Merger contracts](../../../../../internal/core/workflow_merger.go).
- [Input validation](../../../../../internal/core/workflow_merger_input.go).
- [Contained evidence](../../../../../internal/core/workflow_merger_evidence.go).
- [Result parsing](../../../../../internal/core/workflow_merger_result.go).
- [Owned execution](../../../../../internal/core/workflow_merger_execution.go).
- [Tool boundary](../../../../../internal/core/workflow_merger_tools.go).
- [Shared lifecycle](../../../../../internal/core/workflow_merger_driver.go).
- [Budget completion](../../../../../internal/core/workflow_merger_finish.go).
- [Replay repair](../../../../../internal/core/workflow_merger_repair_continue.go).
- [Semantic planning](../../../../../internal/core/workflow_merger_repair_plan.go).
- [Repair method verification](../../../../../internal/core/workflow_merger_repair_verify.go).
- [Protocol reference](../../../../../docs/reference/workflow-execution.md#merger-proposal-protocol).

## Validation

T6.1 focused checks:

- `go test ./internal/core -run '^TestNativeMerger' -count=1` — pass, 0.132 seconds.
- `go test -race ./internal/core -run '^TestNativeMerger' -count=1` — pass, 1.263 seconds.
- `go test ./internal/core -run '^$' -fuzz '^FuzzNativeMergerJSON$' -fuzztime=10s -parallel=2` — pass, 37022 executions, 11.044 seconds.

The initial strict lint check found two complexity and thirteen spacing findings.
Cohesive validators and formatting corrected those findings; strict lint then
reported zero issues. Required T6.1 pre-commit checks passed:

- `make verify` — zero lint issues; core 1282.104 seconds, app 94.013 seconds and all other packages passed.
- `git diff --check` — pass.

Owned fixtures completed naturally and their disposable data was cleaned by the
test lifecycle.

T6.1 task commit: [dd376916016bc85c9f20677e03997e0b7f6a2a58](https://forge.adrianpk.com/adrianpk/chr/commit/dd376916016bc85c9f20677e03997e0b7f6a2a58).
Acta task record: `01M4R5KBYTW86RNNWE19EJ9ZJJ`. This commit introduced the report; its
[immutable report image](https://forge.adrianpk.com/adrianpk/chr/src/commit/dd376916016bc85c9f20677e03997e0b7f6a2a58/ops/default/report/slices/native-integration-queue/slice-6-merger-role.md)
remains the canonical PR report link.

T6.2 focused checks on the completed runtime implementation:

- `go test ./internal/core -run '^TestNativeMerger' -count=1` — pass, 24.264 seconds.
- `go test -race ./internal/core -run '^TestNativeMerger' -count=1` — pass, 39.962 seconds.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./...` — pass, zero issues.

Fixtures use actual Git conflicts and a certified subprocess driver, not a live
model call. They verify stale admission before mutation, protected/aliased paths,
worker-checkout preservation, SHA-1/SHA-256 workspaces, timeout, pause/resume,
orphan descendants, retained budget charges and independent token settlement.
Required T6.2 pre-commit checks passed:

- `make verify` — zero lint issues; core 1252.301 seconds, app 95.194 seconds and all other packages passed.
- `git diff --check` — pass.

Owned fixtures completed through the test lifecycle.

T6.2 task commit: [78793f6ce58fc5fabedcece6e9338698713d8739](https://forge.adrianpk.com/adrianpk/chr/commit/78793f6ce58fc5fabedcece6e9338698713d8739).
Acta task record: `01M4RBPR0K3FB8PTX766B0K7ZH`.

T6.3 focused checks:

- `go test ./internal/core -run '^TestNative(Merger|Combined|Replay)' -count=1` — pass, 40.378 seconds.
- `go test -race ./internal/core -run '^TestNative(Merger|Combined|Replay)' -count=1` — pass, 55.628 seconds.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --fix ./...` — pass, zero issues.

Actual Git fixtures cover configured merge/squash parents, empty/already-applied
replays, successive conflicts, exact author/date/message/encoding, semantic
ownership, attributed fixes, ambiguity, unsupported procedures, disguised binary
content and recomputed hashes without repair records. Broader checks exposed an
empty replay-owner selection and fresh-timestamp recovery error. Corrected
ownership handling and explicit different host timestamps reproduce those cases
deterministically. The initial binary fixture incorrectly put a NUL in a Git argv;
the fixture now writes binary content separately from its conventional message.

Required T6.3 pre-commit checks passed:

- `make verify` — zero lint issues; core 1244.924 seconds, app 95.231 seconds and all other packages passed.
- `git diff --check` — pass.

Owned fixtures completed through their test lifecycle; the recorded gate runner
and core test process exited.

T6.3 task commit: [97518aabe9b214d18c86440c372511c17676e958](https://forge.adrianpk.com/adrianpk/chr/commit/97518aabe9b214d18c86440c372511c17676e958).
Acta task record: `01M4RFEYENF7PMNPSNESJB2DY3`.

## Risks and Follow-ups

T6.3 semantic/replay repair passes required pre-commit checks. T6.4 host acceptance remains pending. The runtime does not yet deliver automatic
conflict repair. No canonical Slice 6 PR is open yet. Required ordered final
candidate checks use the controller ledger after implementation and PR-reference
commits are pushed. No full delivery-set gate has run.
