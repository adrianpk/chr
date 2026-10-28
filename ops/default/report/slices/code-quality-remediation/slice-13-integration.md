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
Together with the broker/driver contracts, 18 of 87 assigned signature rows are
implemented; 69 remain pending. The exact debt manifest now has 101 entries at
ceiling six, with one external exception.

The first merge batch relocates 28 broker, JSON, path/scope and contained-tool
declarations into workflow/merge. The mechanical commit preserves declaration
tokens in 27 of 28 entries after owner naming; the response formatter inlines the same 4096-byte failure bound.
The host still opens/closes the checkout and supervises the exact claimed driver.
Existing tool escape/alias assertions move beside their owner; the child fixture
retains independent original request/response fields.

Stage APIs now use package-qualified names: integrate.ValidateCandidate,
integrate.CheckGate, merge.ServeBroker, merge.Tools and the owned scope/JSON
checks. BrokerResult retains the raw proposal and usage beside its error; driver
inputs separate the bound claim/prepared call from owned pipe/output resources.
The remaining candidate and merge extraction tasks are in progress.

## Implementation Notes

The predecessor PR #449 is verified merged into dev at
`a49c32c9f5e81e1bde9f585a25c6d5cc669d7a9e`. Its exact-head controller evidence
is `EVFKMQZNCCYQDPLZLRXOYO2S7M`; all four recorded commands passed.

Completed task commits:

- Candidate stage: `05a7a106029d64ee5aa1ba8b62703afc40437d72`.
- Preparation contracts: `c7a15133e0a07f8eb3fec735e36c967730aff038`.
- Mechanical merger broker: `962e03aef96bf7aefd2b075fe218041b9cd00dc6`.
- Stage naming and broker/driver contracts: `8e3630b9b074ec7b15d7b0de252429c3844dd283`.

All 908 Store methods retain one local receiver owner. Current method and move
hashes follow these changes while their original source identities/hashes remain
frozen. The current-file owner map records the actual stage files and private
preparation resource bindings.

## Contracts Added or Changed

`ValidationInput` carries immutable candidate identity, preparation, lane and
initial task result. The stage consumes `Validation`; it receives neither a
mutable store nor claim/process resources. Execution and stop contexts remain
explicit parameters. The stage import bound remains shared-state-only. No alias
or forwarding compatibility layer retains the moved names.

The [move ledger](../../../plan/code-quality-remediation-slice-13-moves.csv)
retains the frozen source identities and hashes for the extracted policy.

## Files of Interest

- `internal/core/workflow/integrate/`.
- `internal/core/workflow/merge/`.
- The existing workflow authority and method maps.

## Validation

- `go test ./internal/core/workflow ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|GateFailure|Stop|Guard|Journal)|NativePreparation(Methods|Failure|Stop)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; ten top-level checks, including real Git, stop, failure, journal/restart and frozen wire evidence. The existing runner owns and cleans its fixtures.

- `go test ./internal/core/workflow/integrate -count=1`: pass; deadline versus stop, detached acceptance, error preservation and accepted candidate retention.
- `make check-go-style`: pass after the named preparation contracts; ceiling six, 101 exact debts, one external exception.
- `scripts/check-package-readmes.sh`: pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./scripts/goquality`: zero issues.

- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/app ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|GateFailure|Stop|Guard|Journal|Unsupported|Conflict)|NativePreparation(Methods|Failure|Stop|Admission|Fence|Generation)|Native(DriftDuringGate|SourceDuringGate|CandidateRepair|CandidateClean|TargetDrift)|QualityRecordWire|LeafDependencies|ValidationCompletion)$' -count=1 -timeout=15m -v`: pass; 22 top-level checks through the existing owned-fixture runner. Real target/source drift, claim generations/fences, repair, normal/failed/cancelled preparation, journal recovery, app wire and stage/import contracts pass.

- `go test ./internal/core/workflow ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(NativeMerger(Execution|Input|Results|Malformed|Tools|Descendants|Stop)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; ten top-level checks through the existing owned-fixture runner. Actual child execution, protocol/input/result rejection, stop/descendant cleanup, filesystem aliases and frozen wire/import bounds pass.

- `go test -race ./internal/core/workflow ./internal/core/workflow/merge ./internal/core/workflow/integrate ./scripts/goquality -run '^Test(WorkflowPreparationStop|NativePreparationFence|NativeMerger(Execution|Tools|Malformed|Results|Descendants|Stop)|ValidationCompletion|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; 11 top-level checks, including real preparation stop/fence, child execution/cleanup, broker rejection, tools and stage/dependency contracts. No owned descendants remain.

## Risks and Follow-ups

The slice remains active until its approved tasks and exact-head checks pass.
