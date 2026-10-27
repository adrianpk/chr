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

## Implementation Notes

The approved inventory assigns 38 signature occurrences and 30 current root
files to this slice. Each current owner and consumer is checked before moving.
The seven declaration preimages in slice-10-moves.csv match after namespace
qualification and the equivalent standard-library membership predicate.
Dependency guards allow only the existing neutral capability and exact digest
owner; workflow/coordinator imports remain forbidden.

## Contracts Added or Changed

Pure binding validation and lexical identity predicates moved without a root
forwarding API. Package naming cleanup remains a separate task commit.

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

## Risks and Follow-ups

T10.1-T10.6 remain in progress or pending; PR readiness is not established.
