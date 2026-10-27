# Slice 10: Harness contracts and runtime lifecycle

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-10-quality-harness`
PR: pending

## Purpose

Separate adapter-neutral execution from project admission, binding and host
lifecycle policy while preserving frozen identities and bounded child resources.

## Delivered Behavior

Binding sealing and validation now live beside the retained binding values in
core/harness. Capability and binding hashes use the original JSON representation;
profile-name and selection-digest checks have one owner. Root and app consumers
call the owner directly. Native command fencing retains its original workflow
errors and execution checks.

Runtime planning consumes its existing named artifact path set directly; the
seven-result facade is removed. Canonical layout selection has a named outcome.
Pi preparation returns distinct credential, call-resource, model-mount, container
and publication records. Expected absence and publication stability remain
separate from failure. Output limits, digest guards, credential refresh exclusion,
cleanup ordering and finite deadlines are unchanged.

The bridge derives and defers its own cancellation before input reading. The
input watcher receives that exact cancellation handle and returns prompt/error;
there is no stored operation context or context hidden inside a result record.
The real malformed-input test now verifies cancellation on every error branch.
The process exchange groups its two owned spools and retains explicit context,
cancellation and protocol handler arguments.

## Implementation Notes

The approved inventory assigns 38 signature occurrences and 30 current root
files to this slice. Each current owner and consumer is checked before moving.
The seven declaration preimages in slice-10-moves.csv match after namespace
qualification and the equivalent standard-library membership predicate.
Dependency guards allow only the existing neutral capability and exact digest
owner; workflow/coordinator imports remain forbidden.

## Contracts Added or Changed

Pure binding validation and lexical identity predicates moved without a root
forwarding API. A separate naming commit shortens the public contracts to
harness.SealBinding, ValidateBinding and ValidProfileName; the focused binding,
selection and wire regressions passed again after that rename.

## Files of Interest

- `internal/core/harness/`: existing binding values and planned project policy.
- `internal/harness/`: existing adapter-neutral execution and process owners.

## Validation

Slice 9 closure uses its existing
three-command controller evidence at `0ce066b1599b55f6efc3f151bb960a74b75d5e91`.

Focused implementation checks passed:

- `go test ./internal/core/harness ./internal/core ./internal/app -run '^Test(HarnessBinding|HarnessSelection|QualityRecordWire)' -count=1 -timeout=3m`
- `go test ./scripts/goquality -run '^Test' -count=1`
- `make check-go-style`: 9304 declarations, ceiling seven, 237 debts, one external exception.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/harness/... ./internal/core ./internal/app`

The signature disposition ledger records ten fixed occurrences and six cohesive
five-input contracts. Twenty-two assigned occurrences remain pending. Ten exact
debt allowances are removed: 227 remain, with ceiling seven and one external
exception unchanged. No migration debt or broader allowance is introduced.

Focused runtime/resource and concurrency checks passed:

- `go test ./internal/harness/pi ./internal/harness/process ./internal/core ./internal/app -run '^Test(Pi|AgentRunProcess|RuntimePlan|QualityRecordWire)' -count=1 -timeout=3m`
- `go test ./internal/core ./internal/app ./internal/harness/pi ./internal/harness/codex -run '^Test(PlanRuntime|ValidateRuntime|RuntimeArtifactCleanup|HarnessBinding|QualityRecordWire|PiWorkerInput|PiWorkerPreparation|PiGenerationResources|PiGenerationResourceIdentity|Codex)' -count=1 -timeout=3m`: Codex compiled with no matching tests; runtime paths, resources and bridge preparation exercised directly.
- `go test -race ./internal/harness/pi ./internal/harness/process -run '^Test(PiWorkerConnection(Bounds)?|PiCredentialRefresh(Failure|Reject|Deadline|Resources|Canceled|Home)?|PiTransportDriver|AgentRunProcess(Streams|Cancellation))$' -count=1 -timeout=3m`
- `make check-go-style`: 9303 production declarations, 227 debts, unchanged ceiling and external exception.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/harness/... ./internal/core ./internal/app`: zero issues.

## Risks and Follow-ups

T10.1-T10.6 remain in progress or pending; PR readiness is not established.
