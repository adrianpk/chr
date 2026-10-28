# Slice 13: Candidate validation and merge stages

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-13-quality-integration`
PR: pending

## Purpose

Give candidate preparation/validation and merge/repair policy their approved
stage owners while preserving the store's guarded transactions and recovery.

## Delivered Behavior

The first candidate-stage extraction runs validation, builds the candidate and
requests atomic acceptance through the existing bound port. Exact gate and source
observation policy now have the integrate owner. Claim admission, stop/lease
contexts, capacity release and the guarded transaction remain in workflow.Store.
Nine preparation declarations now use named selection, claim, observation and
Git results or bound completion inputs. Partial inputs on failure, cache presence,
reservation release and the existing transaction order are preserved. A private
failure mode replaces the actionable/target-moved pair with the same precedence.
These contracts resolve 11 of 87 assigned signature rows; 76 remain pending.
The first merge batch relocates 28 broker, JSON, path/scope and contained-tool
declarations into workflow/merge. Twenty-seven retain declaration tokens after
owner naming; the response formatter inlines the same 4096-byte failure bound.
The host still opens/closes the checkout and supervises the exact claimed driver.
Existing tool escape/alias assertions move beside their owner; the child fixture
retains independent original request/response fields.

The remaining candidate and merge extraction tasks are in progress.

## Implementation Notes

The predecessor PR #449 is verified merged into dev at
`a49c32c9f5e81e1bde9f585a25c6d5cc669d7a9e`. Its exact-head controller evidence
is `EVFKMQZNCCYQDPLZLRXOYO2S7M`; all four recorded commands passed.

## Contracts Added or Changed

`ValidationInput` carries immutable candidate identity, preparation, lane and
initial task result. The stage consumes `Validation`; it receives neither a
mutable store nor claim/process resources. Execution and stop contexts remain
explicit parameters. The stage import bound remains shared-state-only.

The [move ledger](../../../plan/code-quality-remediation-slice-13-moves.csv)
retains the frozen source identities and hashes for the extracted policy.

## Files of Interest

- `internal/core/workflow/integrate/`.
- `internal/core/workflow/merge/`.
- The existing workflow authority and method maps.

## Validation

- `go test ./internal/core/workflow ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|GateFailure|Stop|Guard|Journal)|NativePreparation(Methods|Failure|Stop)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; ten top-level checks, including real Git, stop, failure, journal/restart and frozen wire evidence. The existing runner owns and cleans its fixtures.

- `go test ./internal/core/workflow/integrate -count=1`: pass; deadline versus stop, detached acceptance, error preservation and accepted candidate retention.
- `make check-go-style`: pass after the named preparation contracts; ceiling six, 106 exact debts, one external exception.
- `scripts/check-package-readmes.sh`: pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./scripts/goquality`: zero issues.

- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/app ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|GateFailure|Stop|Guard|Journal|Unsupported|Conflict)|NativePreparation(Methods|Failure|Stop|Admission|Fence|Generation)|Native(DriftDuringGate|SourceDuringGate|CandidateRepair|CandidateClean|TargetDrift)|QualityRecordWire|LeafDependencies|ValidationCompletion)$' -count=1 -timeout=15m -v`: pass; 22 top-level checks through the existing owned-fixture runner. Real target/source drift, claim generations/fences, repair, normal/failed/cancelled preparation, journal recovery, app wire and stage/import contracts pass.

- `go test ./internal/core/workflow ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(NativeMerger(Execution|Input|Results|Malformed|Tools|Descendants|Stop)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; ten top-level checks through the existing owned-fixture runner. Actual child execution, protocol/input/result rejection, stop/descendant cleanup, filesystem aliases and frozen wire/import bounds pass.

## Risks and Follow-ups

The slice remains active until its approved tasks and exact-head checks pass.
