# Slice 16: Final enforcement, documentation and integrated closure

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-16-quality-closure`
PR: [#453](https://forge.adrianpk.com/adrianpk/chr/pulls/453)

## Purpose

Enforce permanent production contracts, reconcile every approved baseline finding
and publish the final capability map/glossary. Prepare the exact PR candidate for
controller-owned integration and the unchanged full delivery-set gate.

## Delivered Behavior

Production signatures have a permanent five-input maximum and zero/one result
or `(T, error)` / `(T, bool)`. The exact installed Bubble Tea Model.Update
interface is the sole result exception. Phase configuration, debt manifest,
allowance code and ratchet tests are removed. Pinned revive enforces five inputs
and two results; production dupl uses threshold 150. Complete-source AST checks
retain interface/function-type/literal, alias, build-variant, generated-source,
semantic second-result and exact exception coverage on every branch.

All 771 baseline occurrences and 454 secondary candidates have code-backed
resolutions. The final evidence covers fourteen large files, six wide values,
eleven duplicate pairs, seven dead symbols, all eight error names and all 23
baseline bare cleanup directives (eleven production, twelve fixtures). The last
three error types and their constructors/callers are corrected without aliases;
raw literals, failure kinds, sentinels and unwrapping remain unchanged. Cleanup
comments explain resource-specific handling without changing executable tokens.

The capability map and glossary locate real package/stage APIs and distinguish
ticket-native authority from forge publication. The cycle report is prepared,
with final integrated candidate and gate evidence explicitly pending.

## Implementation Notes

Predecessor PR #452 is verified merged into dev at
`a41e62be3b2a14daf04b470e8d9344d350480f34`; Slice 15 report/tracker/ticket are
closed as delivered. Slice 16 uses the approved dedicated branch/worktree. The
configured report did not exist in Git history; its exact approved path was
introduced at `0573533a59a8f1a76f1d9feb54798cf9f3df7151`.

Implementation: permanent policy `3b1d6ee0844c`, minor source corrections
`436f6742f7f224ac8c25bfde7cc61bb6ab328332`, current provenance `50d562611bfb`,
full residual reconciliation `942814b08615` and durable documentation
`33c680a17a0f`. The final signature map records 570 conforming named contracts,
200 cohesive five-input decisions and one external exception. All 230 B rows
have individual prior owning-slice decisions. Current production has 9,411
signature declarations, maximum five and exactly one external result exception.

Source reconciliation verifies all 4,445 workflow backbone hashes, 217 earlier
compatibility images and 771 final signatures after the minor source changes.
Frozen source fields/introduction commits and existing CSV line endings remain.
The 908 source Store methods have individual decisions, with 777 guarded
persistence methods; all 22 finite ports and 65 remaining root files have actual
owners. No root forwarding facade, workflow type alias, bare errcheck or unused
suppression remains. The complete secondary map retains precise supported-reader
retirement conditions and distinguishes input switches from returned observations.

D08 shares contained receipt read/absence/decode. Its mutation/control writers
retain distinct schemas, checksums, identity/revision validation and locked atomic
transactions. Existing grammar, dispatch, constructor and schema complexity
exceptions have specific in-source rationales and are not migration allowances.
Original review/detection bytes remain unchanged pending real integration.

## Contracts Added or Changed

The lint/AST enforcement becomes permanent and branch-independent. No dependency,
wire schema, enum spelling, budget, lock, approval, hash or durable authority
contract changes. Existing lint test scope remains intact; only the new
production argument/result and dupl rules exclude test fixtures. All source
findings are fixed or explicitly justified by their owning contract/measurement.

## Files of Interest

- [Final signatures](../../../plan/code-quality-remediation-slice-16-signatures.csv) and [secondary candidates](../../../plan/code-quality-remediation-slice-16-secondary.csv).
- [Final source reconciliation](../../../plan/code-quality-remediation-slice-16-findings.md), including all remaining categories and merged implementation/test evidence.
- [Cleanup dispositions](../../../plan/code-quality-remediation-slice-16-cleanup.csv) and [benchmark samples](../../../plan/code-quality-remediation-slice-16-benchmarks.csv).
- [Capability map](../../../../../docs/explanation/capability-boundaries.md) and [workflow glossary](../../../../../docs/reference/workflow-glossary.md).
- [Prepared cycle report](../../cycles/code-quality-remediation.md).
- `scripts/goquality/`, `.golangci.yml` and the actual error/cleanup owners.

## Validation

Completed task checks, before final clean-head controller acceptance:

- `go test ./scripts/goquality -run '^TestBranchIndependent$' -count=1` initially failed as expected: phase-six enforcement accepted a new six-input declaration on dev self-comparison. The existing fixture now tests the permanent boundary.
- `go test ./scripts/goquality -count=1` passed after the correction, including exact/stale/duplicate/unnecessary exceptions, alias/build-variant and nested contracts.
- `go run ./scripts/goquality` passed with 9,411 declarations, maximum five and one external exception, without a migration manifest.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./...` passed with zero issues under permanent enforcement and again after the minor source corrections.
- `go test -p=1 ./internal/core/authoring ./internal/core/execution ./internal/harness/pi -run '^Test(InvokeAssistedModel.*|RecordAssistedModelInvocation.*|RenderAssistedModelArtifacts.*|ExecutionFailure.*|ExecutionRunner.*|PiTransport(Failures|Worker|WorkerOutput))$' -count=1 -timeout=3m -json` passed all 19 selected roots, without failures/skips or owned fixture leaks.
- `go test ./internal/tui -run '^$' -bench '^BenchmarkModel(Update|View)$' -benchmem -benchtime=200ms -count=3` passed in 3.689 seconds on clean `50d562611bfb5d8e725ff62da389fd9c2e372818`: fifteen samples. Idle/resize allocate zero bytes; focus remains 1,096 bytes/nine allocations. Model size is 49,832 bytes, and View allocation counts remain 703/1,404. The linked comparison retains baseline/final samples and bounds the general heuristic-warning exclusion to actual measured paths.
- Baseline/current identity, source-line, declaration/hash, field/move/port and cleanup maps reconcile; all new Markdown links resolve; the original review/detection SHA-256 values remain unchanged.
- `git diff --check` passed for the task changes.

Initial controller receipt `KUKRMY5XTCSSLX2ILXIGDFK3ZP` passed ordered
`make verify` and `git diff --check` on clean
`178a5d91760c1efab2ffcad986542f5f4ba6c917` in 50.014 seconds before opening
canonical PR #453. Its title, target dev and immutable report-introduction link
match the approved identity. This follow-up records the actual PR and moves the
ticket to reviewing.

Final acceptance is recorded by `looper validate` after this reference commit,
keyed to the exact clean HEAD and ordered commands below. The canonical PR lists
only successful checks. Its final receipt remains outside Git so adding the
receipt cannot invalidate its own HEAD. The combined selection has 34 actual
roots, and the focused race selection has four; source bindings were checked
before execution.

- `make verify`
- `go test -p=1 ./internal/core ./internal/core/workflow ./internal/core/execution ./internal/core/telegram ./internal/harness/pi ./internal/tui ./internal/webui ./internal/app -run '^Test(QualityRecordWire|QualityStorage[A-Za-z]+|WorkflowRoundTrip|WorkflowRecoveryProcess|WorkflowClosingRecovery|OperationReceiptGuards|ExecutionRepository[A-Za-z]+|TelegramReceiver(HTTP|HTTPFailures|LockProcess|Batch|ReadJourney|Revocation|Shutdown)|TelegramPresentation(MigrationDocuments|Document(Snapshot|Rollback|Quota)|DeliveryRecovery|DocumentReadRecovery|DeliveryHTTP)|PiTransport(Failures|Worker|WorkerOutput)|ModelOwnership|DocumentCatalogFrames|GuidedViewWire)$' -count=1 -timeout=5m -json`
- `go test -race -p=1 ./internal/core/workflow ./internal/core/telegram ./internal/tui -run '^Test(WorkflowRoundTrip|OperationReceiptGuards|TelegramReceiverBatch|ModelOwnership)$' -count=1 -timeout=5m -json`
- `git diff --check`

The unchanged full gate and final integrated acceptance remain controller-owned.

## Risks and Follow-ups

T16.1 through T16.4 are implemented and pending PR-head acceptance. The benchmark
and cycle report are prepared. T16.5/T16.6 integration/full-gate closure and
T16.7 verified integration remain open until the controller supplies their exact
evidence. This report does not mark Slice 16 or the delivery set delivered.
