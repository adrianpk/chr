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
Ticket projections have the existing ticket owner. The complete project
configuration image and its admission/recovery operations now have the
workspace/config owner, shared directly by workflow, app and adapters.

## Implementation Notes

The approved T12.1–T12.6 tasks own this slice. The shared review prerequisite is
implemented in `f4be8551581dbec43746e5b676b714911af24e2c`. All methods of the store move with
their owning Go type. The ticket-projection prerequisite is implemented in
`a96c05cee2000654f23e2ce0b0347b1155ec17ca`. Shared stage data belongs in workflow/state when needed
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
Shared review values resolve ten direct dependencies. Ticket eligibility/order,
Board values and document projections now have their existing ticket owner with
direct app, adapter and native read consumers. Forty-two declaration moves retain
exact tokens after qualification, API naming and cmp.Compare substitution; the
removed integer comparator has the same standard-library result. These resolve
another twelve dependencies. Configuration relocation resolves fifteen more,
leaving 42 direct root couplings. No type
bridge or coordinator/transport/persistence authority moved into the value owner.
The 114 configuration declarations retain exact tokens after qualification/API
names, insignificant parameter commas and the original rebase default's literal
value. Raw-image revision, explicit presence, defaults, validation order, locks
and journal/image/receipt replay are unchanged. Config.Explicit exposes only a
read-only key-presence query. The existing upgrade-journal tuple debt now names
workspaceconfig.Config; its shape and Slice 15 ownership are retained.
All 908 store methods still retain their original root receiver. Full family
relocation and assigned signature remediation remain open.

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
The ticket-projection prerequisite also passed:

- `go test ./internal/core/ticket/... -count=1 -timeout=3m`: complete ticket-owner suite, including SQLite query/filter/order and Board regressions.
- `go test ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(TicketBoard|TicketDocument|WorkflowSelection|WorkflowOrder|QualityRecordWire|QualityStorage|TicketLoopDryRunPrintsEligibleTickets|Leaf)' -count=1 -timeout=3m`: native filesystem/Git, selection/order, frozen wire/storage, CLI and dependency boundaries pass; TUI compiles.
- `go test ./internal/core/ticket/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(TicketLoopQuery|BoardQuery|BoardObservation|TicketDocument|LegacyWorkObservation|WorkflowSelection|WorkflowOrder|WorkflowRoundTrip|QualityRecordWire|QualityStorage|TicketBoardPrintsBoardColumns|TicketLoopDryRunPrintsEligibleTickets|LeafDependencies)$' -count=1 -timeout=3m`: the anchored exact-name selection includes WorkflowRoundTrip, LegacyWorkObservation, CLI Board output and dependency checks. Prefix cases are covered by the two commands above.
- `make check-go-style`: pass; 9305 production declarations, ceiling six, 162 exact debts and one existing external exception.

- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/ticket/... ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues after removing two exact duplicate, newly unused root fixture helpers and splitting the boundary case list.

The move ledger records 42 matching declaration token sequences and the exact
standard-library replacement. The frozen entry map is unchanged.
The configuration prerequisite passed:

- `go test ./internal/core/workspace/config -count=1 -timeout=3m`: complete colocated selection preparation, recovery, receipt, bound and capacity tests.
- `go test ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(LoadWorkspaceConfig|WorkflowConfig|WorkflowUpgrade|NativeProfile|NativeLegacySnapshot|NativePreparedSnapshot|WorkAuthority|HarnessSelection|HarnessGenerationConfigFence|WorkflowReviewPolicy|QualityRecordWire|QualityStorage|Leaf)' -count=1 -timeout=3m`: real filesystem/Git/process, frozen configuration/deadline, upgrade/recovery, wire/storage, app and boundary regressions pass; TUI compiles.
- `go test ./internal/core -run '^Test(WorkflowConfigDeadline|HarnessSelectionUpgradeFence)$' -count=1 -timeout=3m`: unchanged deadline checks and the real cross-owner receipt-failure/upgrade/replay fixture pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workspace/config ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues.
- `make check-go-style`: pass; 9306 production declarations, ceiling six, 162 exact debts and one existing external exception.

All 114 configuration declaration token comparisons pass. Earlier test/lint
overlap hit the existing one-second deadline fixture; unchanged dev and the same
candidate selection run after lint passed. Assertions and timeout bounds are
unchanged. This is focused batch evidence, not final controller validation.
Final exact-head controller validation has not run.

## Risks and Follow-ups

Remaining transport, guided admission, native document and operation-lock
dependencies must have real boundaries before family relocation. Assigned signatures, stage ports, candidate validation and canonical
integration remain outstanding.
