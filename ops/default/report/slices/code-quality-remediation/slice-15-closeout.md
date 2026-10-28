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

Closing admission pairs an optional reserved claim with the existing retained
closure result. Delivery proof and its verification hash use one named result
for ticket-native and legacy evidence, including partial decoded proof on error.
A supervised command value binds its exact argv, process, permit, output pair and
output base. Unused context forwarding through closing projection and commit
helpers is removed without changing lock or journal ownership.

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
retains five source/extraction images. Current identities and hashes are updated
in the original declaration, method, dependency and ownership ledgers while their
frozen source fields remain unchanged. There are still 908 Store methods; that
count and file movement do not establish final policy/persistence separation.

## Contracts Added or Changed

The [assigned signature ledger](../../../plan/code-quality-remediation-slice-15-signatures.csv)
contains 40 original rows: three implemented and 37 pending. Exact temporary
signature debt decreases from 24 to 21 with the six-parameter intermediate
ceiling unchanged. Closeout imports only shared state within the module.

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

Task checks passed; final slice validation remains pending. At the stage
extraction, 18 selected real boundary tests and four top-level focused race tests
passed without skips. After closing value conversion, 12 selected real tests and
two focused stop/recovery race tests passed without skips. The owned-fixture
runner completed cleanup with no leaked descendants.

- `go test ./internal/core/workflow -run '^Test(WorkflowClosing(Retry|Empty|ActualMembers|Changes|UseCase|Reopen|SetAdvance|Budget|Recovery|Journal|PostObservation|Stop)|NativeClosing(Gate|Divergence|Evidence|PublicationRace)|WorkflowParallelClosingRecovery|QualityRecordWire)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow ./internal/core/workflow/closeout -run '^Test(WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery|Execute)$' -count=1 -timeout=10m -json`
- `go test ./internal/core/workflow -run '^Test(WorkflowClosing(Retry|Empty|ActualMembers|Changes|Budget|Recovery|Journal|Stop)|NativeClosing(Gate|Divergence|Evidence|Lost))$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowClosingStop|WorkflowParallelClosingRecovery)$' -count=1 -timeout=10m -json`

- `go test ./internal/core/workflow/closeout -count=1`
- `go test -short ./internal/core/workflow/... ./scripts/goquality`
- `make check-go-style` — 9,364 production declarations, 21 exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./scripts/goquality` — zero issues.
- `git diff --check`

## Risks and Follow-ups

Remaining closeout policy extraction, 37 signature dispositions, method-level
Store policy ownership, final bridge removal and naming/compatibility
reconciliation remain active Slice 15 work. Shared guided parity, final ordered
exact-head controller validation, PR publication and PR-reference follow-up are
pending. No final slice acceptance is claimed by these task checks.
