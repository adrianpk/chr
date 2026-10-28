# Slice 12: Workflow state, store and coordinator package

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-12-quality-workflow`
PR: `#449`

## Purpose

Establish the workflow state and persistence boundary with one store, locking
and recovery authority, direct consumers and cohesive stage contracts.

## Delivered Behavior

`internal/core/workflow` owns the actual `Store` and all 908 receiver methods.
Core composition, application operations and adapters use that owner directly.
Repository/file locks, revision/source/claim guards, journals, atomic multi-record
writes and restart recovery remain at the same authority. Legacy read/replace
operations retain adoption and exact-source checks.

Five consumer-owned stage contracts bind actual integration validation, merger
evidence, review acceptance, submission response and closing-gate operations.
Shared protocol values live in workflow/state; stage contracts import those
values without importing their coordinator. Ticket projections and metadata edits
and the complete parsed configuration retain their existing leaf owners.

All 71 assigned signature rows are reconciled to 68 declarations: 47 implemented
contracts, 20 cohesive five-parameter contracts retained with individual reasons,
and one prior shared-owner fix. Named observations preserve error-associated
values, absence, pagination, byte budgets, nil/empty collections and replay.
Invocation admission, attempt preparation and supervised execution have cohesive
inputs. Actual app, authoring, ticket, daemon and adapter consumers use the values
directly; the app integration-guard alias is removed.

Coordinator progress/resume, supervised-process and receipt/reservation states
have named string types/constants with unchanged stored spellings. Preparation
keeps its existing typed automatic/manual mode. The D08 readers share contained
read/absence/decode only; each operation still validates its own schema, checksum,
request, operator and revision before confirming replay. Writers remain separate.

## Implementation Notes

The relocation is `20ca6c3e508b7c0964baaef4f345c87199e3d81b`; stage bindings are
`9c74db45a790d03f7d710b2ffac2e7dcdc4d5e88`. The first ten signature changes are
`fe0df936e6152c0ad346e1e680ad2dd246825e60`.
T12.4 completion and T12.5 reconciliation: `4ffe06de32d98e02ab388fa0613fd9f0b61e2933`.

The [authority map](../../../plan/code-quality-remediation-slice-12-boundaries.md)
and its method, boundary, dependency, signature, move and port CSVs retain frozen
source identities/hashes and explicit current ownership/provenance. The compiling
relocation covers 2,710 production and 1,666 test declarations; its 4,342 declaration
groups retain tokens after compiler-resolved naming/qualification. Subsequent
contract changes have separate current hashes and implementation evidence.

Sixty protocol types and two evidence enumerators have shared-state ownership.
Field/tag order, hashed encodings, optional/zero values and nil/empty semantics
remain compatible. All 79 original root dependencies and 100 authority/recovery
entries resolve to actual current owners. No compile-time bridge remains.

Child contexts are created explicitly by their existing callers. Stop watchers
receive a bound claim/cancel input; heartbeats return one finish handle that
cancels and waits. Goroutine/ticker, process, reservation, cleanup and error-join
behavior remains covered by real stop, process and daemon regressions.

## Contracts Added or Changed

Stage interfaces are `integrate.Validation`, `merge.Evidence`, `review.Decisions`,
`forge.Responses` and `closeout.Gate`; private bound operations or Store implement
them. Storage and claim resources remain private. The state-only stage import
bounds and existing source/revision/admission guards remain enforced.

Public query results include `InvocationLookup`, `InvocationPage`,
`LegacyTicketSelection` and `MutationLookup`; operation-specific private inputs
and observations replace the remaining assigned tuples. Coordinator and process
state types preserve existing JSON and validation behavior. The source policy
retains ceiling six and one external exception; resolved debts fall from 162 to
115, with 9,317 production declarations and no new allowance.

Acceptance/fault IDs, order and 78 distinct test selections remain unchanged
while their 82 source bindings follow the real workflow package. The frozen
controller gate is unchanged.

## Files of Interest

- `internal/core/workflow/workflow_store.go`, `workflow_lock.go` and `workflow_transition.go`.
- `internal/core/workflow/workflow_stage_ports.go` and the five stage contract packages.
- `internal/core/workflow/workflow_operation_receipt.go`, query/result/input owners and colocated recovery tests.
- `internal/core/workflow/state/invocation_status.go` and `workflow_process_identity.go`.
- `internal/core/workspace/config/`, ticket metadata/query owners and the authority-map CSVs.

## Validation

The completed contract batch passes:

- `go test ./internal/core/workflow ./internal/core ./internal/app ./internal/tui ./scripts/goquality -run '^Test(QualityRecordWire|QualityStorage[A-Za-z]+|LeafDependencies|WorkflowRoundTrip|WorkflowEntry[A-Za-z]*|WorkflowLegacyEntry|WorkflowInspection[A-Za-z]+|WorkflowTracker[A-Za-z]+|WorkflowClaim(Fence|Journal|Expiry|LinkedRoots|ForeignClone|ProcessRace)|WorkflowTask(Exact|Cumulative|Failures|OutputBounds|Journal|ArgvBounds|PrerequisiteEdit)|WorkflowManual[A-Za-z]+|WorkflowStop(Barrier)?|TelegramControlStopOwner|WorkflowPreparation(Methods|Guard|Stop|Journal)|NativeMergerExecution|NativeAgentAdmission|DaemonWorkflow[A-Za-z]+|WorkflowProcess(Birth|Edit|Signal)|ServiceIntegrationGrant|ServiceIntegrationApproval|GuidedUseCaseControlAdmissionAuthority)$' -count=1 -timeout=20m -v`: 75 top-level checks pass through the existing fixture-owning runner. Frozen wire/SQLite, real Git/filesystem, claims/fences/journals, tracker crash/restart, task/manual execution, process/stop, daemon, integration controls and dependency checks pass; TUI compiles. No owned descendants remain.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/core/authoring ./internal/core ./internal/app ./internal/tui ./scripts/goquality`: zero issues, including unchanged context-inheritance checks.
- `make check-go-style`: pass; ceiling six, 115 exact debts and one external exception.
- `scripts/check-package-readmes.sh`: pass.
- `git diff --check`: pass.

Earlier task evidence includes the five real stage-operation regressions,
stop/lost-response cases, seven focused journal/crash checks, ten receipt
presence/corruption/cross-schema cases and 30 receipt/materialization/adoption/
amendment/authoring checks. Frozen source identities and all 908 current local
method hashes are reconciled. Exact clean-head `make verify` and focused/race
checks are recorded through the controller after the PR-reference commit.

## Risks and Follow-ups

Stage policy extraction retains the approved owners in Slices 13–15. This report
remains reviewing until the controller verifies integration into dev.
