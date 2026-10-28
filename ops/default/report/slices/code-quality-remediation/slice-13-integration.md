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
- `make check-go-style`: pass; ceiling six, 115 exact debts, one external exception.
- `scripts/check-package-readmes.sh`: pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./scripts/goquality`: zero issues.

## Risks and Follow-ups

The slice remains active until its approved tasks and exact-head checks pass.
