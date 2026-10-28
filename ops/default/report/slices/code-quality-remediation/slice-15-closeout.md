# Slice 15: Workflow closeout and coordinator cleanup

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-15-quality-closeout`
PR: pending

## Purpose

Give closing policy and execution a cohesive closeout stage, complete bounded
lifecycle values and remove temporary migration bridges while preserving durable
transaction, lock, recovery and explicit human authority.

## Delivered Behavior

The closeout stage builds unconfirmed receipts, retains ordered partial command
results and checks the exact checkout through narrow Workspace and Commands
ports. Explicit waivers still prepare and check that checkout. Pure lease,
retained-binding and result checks have one closeout owner. Watched deadlines,
process supervision, capacity and guarded confirmation retain their host owners.

Closing attempt shape, ordered history, retry budgets and receipt/input/result
binding now belong to closeout through read-only Evidence. Reads keep the
original containment, file bounds and storage error semantics. Shared process
records and the eight-slot bound live in state; live process observation and
registration stay with the host.

Closing admission pairs an optional reserved claim with the existing retained
closure result. Delivery proof and its verification hash use one named result
for ticket-native and legacy evidence, including partial decoded proof on error.
A supervised command value binds its exact argv, process, permit, output pair and
output base. Unused context forwarding through closing projection and commit
helpers is removed without changing lock or journal ownership.

Capacity observations name process/liveness separately from owned worker
slot/guard reservations. Continuation names accepted steps, barriers, membership
refresh and retained role counts. Control lookups distinguish missing receipts,
verified replay and failed status reads without constructing execution ports.
The neutral application lookup also serves the Telegram adapter. Frozen control
requests retain exact request bytes/digests and accepted decision guards.

Amendment construction pairs partial materialization with its atomic file set;
frontier verification binds one retained invocation/receipt/candidate unit and
preserves partial target/proof observations. Upgrade preparation distinguishes
journal presence from validity and retains its original pending/archive order,
partial journal/image outcomes and joined reconstruction errors.

## Implementation Notes

Predecessor PR #451 is verified merged into dev at
`8b5659a71af5d393919593c5e0d8c9a6ec0f497c`. Its exact controller evidence
`DWWEUV4RMALE7SB4LMQKN2PGSB` passed. The primary checkout remains on dev.

Mechanical relocation `72d968844ee6ee3d4915cb7f2e200a3f049487d1` preserves all three normalized
policy declaration streams. Separate stage extraction and API naming are in
`38b7f97593a2df3d2af6c465c809dca4bb4c1345`; named closing exchanges are in
`663c49213963ec459a32015b8feedbda32271f2e`. The former Gate pass-through interface is removed.
Execution extraction is a policy-flow change, not a whole-method move or a
whole-method token-equivalence claim.

The [move ledger](../../../plan/code-quality-remediation-slice-15-moves.csv)
retains 31 source/extraction/consolidation images. Current identities and hashes are updated
in the original declaration, method, dependency and ownership ledgers while their
frozen source fields remain unchanged. Seven former Store methods now have actual
closeout policy or host evidence adapter owners. The original 908 method rows retain frozen source evidence and
current identities; 901 methods remain on Store. This count does not establish
final policy/persistence separation.

Capacity/continuation conversion is in `d4844af16cce52e3a44ad9bba5e2a59e80adb520`, control/receipt
conversion in `c111c9df77898601331cd83c6df4932a9c2add8b`, and amendment/frontier/upgrade conversion
in `c3d57b9927bb8d739808fea8d5db2809c5915398`. The preparation retry reuses the same continuation step
value; the neutral receipt lookup preserves both original fields after its
owner/name change. Normalized tokens match for both value relocations. The
unused production publication-profile wrapper is removed; its sole test caller
uses the live configuration-key filter with the same key. `prepareWorkflowTask`
already exists only in test fixtures.

Closing history extraction is in `5f58bc1cf3f22c0553035b9935d66530cb4762dd`. Thirteen mechanical declaration moves
match after owner/name qualification. Read/validation/budget changes use the
actual Evidence port and have explicit semantic evidence, not a whole-method
token-equivalence claim. No compatibility alias or forwarding Store facade
replaces the removed policy methods.

The [vocabulary and compatibility ledger](../../../plan/code-quality-remediation-slice-15-vocabulary.md)
defines frontier, adoption, loose selection, continuation, amendment and closing
from the governing spec in `05d68ed95bc5b154df6be54f0279bc18856eab96`. Ten current compatibility/control cases identify actual
consumers, existing regression protection and retirement conditions. Remaining
legacy/schema/control references are explicitly unreconciled; T15.5 stays active.

## Contracts Added or Changed

The [assigned signature ledger](../../../plan/code-quality-remediation-slice-15-signatures.csv)
contains 40 original rows: 26 implemented and 14 retained as cohesive five-input
contracts, each with its operation-specific reason. Production signatures now
have at most five parameters and zero exact debts. The configured intermediate
ceiling remains six; final enforcement and removal of the empty debt manifest
belong to Slice 16. Closeout imports shared state and contenthash within the module.

The host still registers process identity before permitting execution, joins
wait/close/finalization errors, uses the original five-second detached process
cleanup and rechecks authority and frozen inputs after final observation.
Already-closed attachment, failed receipt, interrupted recovery and confirmed
completion remain distinct. No retained field, schema, tag or state spelling
changes.

## Files of Interest

- `internal/core/workflow/closeout/`: policy, execution and narrow host ports.
- Closing admission, command, proof, projection and persistence owners.
- Existing Store method, declaration, authority/dependency and ownership ledgers.

## Validation

Task checks passed; final slice validation remains pending. Closeout extraction
passed 18 real tests and four focused races; closing exchanges passed 12 and two.
Capacity/continuation passed 27 real tests and five focused races, control/receipt
passed 24 and five, and amendment/frontier/upgrade passed 31 and five. These are
separate task selections, not a unique-test total or the full gate. The later
closing-history extraction passed 30 real tests and five focused races. All passed
without skips; owned-fixture cleanup reported no leaked descendants.

- `go test ./internal/core/workflow -run '^Test(WorkflowClosing(Retry|Empty|ActualMembers|Changes|UseCase|Reopen|SetAdvance|Budget|Recovery|Journal|PostObservation|Stop)|NativeClosing(Gate|Divergence|Evidence|PublicationRace)|WorkflowParallelClosingRecovery|QualityRecordWire)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow ./internal/core/workflow/closeout -run '^Test(WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery|Execute)$' -count=1 -timeout=10m -json`
- `go test ./internal/core/workflow -run '^Test(WorkflowClosing(Retry|Empty|ActualMembers|Changes|Budget|Recovery|Journal|Stop)|NativeClosing(Gate|Divergence|Evidence|Lost))$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowClosingStop|WorkflowParallelClosingRecovery)$' -count=1 -timeout=10m -json`
- `go test ./internal/core/workflow -run '^Test(WorkflowCapacity(Bounds|Queue|QueueSlices|Aliases|Reduction|Later|Overlap|Orphan|Continuation|StoppedWait|RetainedLock|RoleWait)|WorkflowContinuation(Tasks|Bounds|Delivery|Recovery|Corruption|Terminal)|WorkflowPreparationRetryAuthority|Native(ProfileUpgrade|LegacySnapshot|MergerDescendants)|WorkflowReviewModes|NativePreparationMethods|TelegramControlLifecycle(Journal|Fence)|QualityRecordWire)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowCapacity(Overlap|Orphan|StoppedWait)|WorkflowContinuationRecovery|TelegramControlLifecycleFence)$' -count=1 -timeout=10m -json`
- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(TelegramControl(ReceiptMissing|Journal|LifecycleJournal|LifecycleFence|ClaimScope|StopOwner|LifecycleAck|ReceiptRetry|ReceiptRedelivery|ReceiptAuthority|ReceiptProjection|StaleReceipt|DecisionGuards|Decisions|StaleReview)|WorkflowControlCAS|WorkflowTerminalControl|OperationReceiptGuards|GuidedUseCaseControl(Native|LostAck|Admission|AdmissionAuthority)|TUIWorkflowControls|WorkflowCLIControls)$' -count=1 -timeout=15m -json`
- `go test -race -p=1 ./internal/core/workflow ./internal/app -run '^Test(WorkflowControlCAS|TelegramControlLifecycleFence|GuidedUseCaseControlLostAck|TelegramControlReceiptProjection|TelegramControlReceiptAuthority)$' -count=1 -timeout=10m -json`
- `go test ./internal/core/workflow -run '^Test(WorkflowAmendment(Admission|Gate|Rejects|LiveOwner|Recovery|Evidence)|WorkflowPlanOrder(Amendment|ActiveAmendment)|WorkflowUpgrade(Policy|Preserves|Adoption|Rejects|Ambiguous|Recovery|AdoptionRecovery|Edited|Owned|Native|Path|Restart)|HarnessSelectionUpgradeFence|Native(ProfileUpgrade|LegacySnapshot|ReceiptFrontier|ReceiptFrontierBlock)|WorkflowFrontier(Order|Reject|Recovery|Race)|WorkflowMembershipFrontier|WorkflowClosingRetry)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowAmendmentLiveOwner|WorkflowFrontierRace|WorkflowUpgradeOwned|WorkflowUpgradeRestart|WorkflowPlanOrderActiveAmendment)$' -count=1 -timeout=10m -json`

- `go test ./internal/core/workflow -run '^Test(WorkflowClosing.*|NativeClosing.*|WorkflowProcess(Birth|Edit|Signal|Observation)|WorkflowCapacity(Process|Overlap|Orphan)|QualityRecordWire)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery|WorkflowCapacityOverlap|QualityRecordWire)$' -count=1 -timeout=10m -json`

- `go test ./internal/core/workflow/closeout -count=1`
- `go test -short ./internal/core/workflow/... ./internal/app ./scripts/goquality`
- `make check-go-style` — 9,366 production declarations, zero exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality` — zero issues.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./scripts/goquality` — zero issues after the final private-result conversions.
- `git diff --check`

The vocabulary-only update checked all relative links, six definitions, ten case
rows and every named regression declaration against the current source. T15
tracker columns and whitespace checks pass. Runtime checks were not repeated for
that documentation change.

## Risks and Follow-ups

Remaining closeout policy extraction, method-level Store policy ownership,
final bridge removal and naming/compatibility reconciliation remain active
Slice 15 work. Shared guided parity, final ordered exact-head controller
validation, PR publication and PR-reference follow-up are pending. No final
slice acceptance is claimed by these task checks.
