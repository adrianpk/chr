# Slice 12: Workflow state, store and coordinator package

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-12-quality-workflow`
PR: pending

## Purpose

Establish the workflow state and persistence boundary with one store/locking
and recovery authority, direct consumers and cohesive stage contracts.

## Delivered Behavior

Slice 11 is verified integrated through PR #448. Slice 12 implementation is active;
the complete method/transaction/recovery map is frozen before package relocation.
Shared cycle/review values and pure value-validation contracts now have the
existing workflow/state owner with direct native, app and adapter consumers.

## Implementation Notes

The approved T12.1–T12.6 tasks own this slice. The shared review prerequisite is
implemented in `f4be8551581dbec43746e5b676b714911af24e2c`. All methods of the store move with
their owning Go type. Shared stage data belongs in workflow/state when needed
without transferring persistence, coordination or transport behavior.

## Contracts Added or Changed

The [authority map](../../../plan/code-quality-remediation-slice-12-boundaries.md)
records all 908 store methods, 100 concrete IO/authority/recovery entries and 79
direct root dependencies at the verified predecessor revision. All store methods
retain one receiver owner; future policy consumers are recorded separately.
All 71 assigned baseline signature rows map to 68 current declarations; 48 require
refactoring, 22 are legal cohesion watches and one has a verified prior shared
owner fix. Twelve review/cycle value/validator declarations have moved to workflow/state.
Token comparison verifies their original enum strings, JSON field/tag order,
nil/empty handling, branches and error messages after qualification/naming.
The move resolves ten direct dependencies; 69 root couplings remain. No type
bridge or coordinator/transport/persistence authority moved into the value owner.
Full family relocation and assigned signature remediation remain open.

## Files of Interest

- `internal/core/workflow_store.go` and the workflow family.
- `internal/core/workflow/state/`.
- The approved source inventory and core owner ledger.

## Validation

Activation documentation links, ticket lifecycle metadata and whitespace passed.
The compiler-resolved source graph, complete method/dependency/signature identity
coverage and exact frozen declaration hashes passed. These are mapping checks;
the map is separate from runtime acceptance. The shared-value prerequisite passed:

- `go test ./internal/core/workflow/state ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(IntegrationCycle|ParseAgenticReview|ApplyAgenticReview|QualityRecordWire|QualityStorage|NativeDecisionReplay|NativeReplayFix|NativeMergerFix|WorkflowReviewModes|WorkflowReviewDecisionChecks|ServiceStartAllReadyAgentWork|Leaf)' -count=1 -timeout=3m`: focused value, frozen wire/storage, real Git replay/repair and native review/app integration checks pass; TUI compiles.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/state ./internal/core ./internal/app ./internal/tui`: zero issues.
- `make check-go-style`: pass; 9306 production declarations, ceiling six, 162 exact debts and one existing external exception.

The twelve declaration token sequences match after namespace/validator naming.
Final exact-head controller validation has not run.

## Risks and Follow-ups

Configuration/transport/projection dependencies must have real boundaries before
relocation. Assigned signatures, stage ports, candidate validation and canonical
integration remain outstanding.
