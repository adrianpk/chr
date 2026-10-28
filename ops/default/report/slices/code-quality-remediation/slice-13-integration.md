# Slice 13: Candidate validation and merge stages

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-13-quality-integration`
PR: pending

## Purpose

Give candidate validation and merge/repair policy independent owners while
preserving guarded workflow persistence, evidence and recovery.

## Delivered Behavior

`workflow/integrate` owns candidate validation execution, exact gate checks,
source projection, Git attribution/replay/signing metadata, bounded preparation
identity retries, complete retained proof and observed capability checks.
`workflow/merge` owns the bounded broker, contained model tools, immutable
input/result/scope/decision validation, semantic repair selection and retained
tree/regeneration/position/replay-history checks. Actual native and application
callers use the package-qualified APIs directly; no forwarding aliases remain.

The validation port still binds the authenticated claim, prepared input and
provider to Store-owned validation and atomic acceptance. A separate immutable
repair-check port composes merge policy in its original order without sibling
stage imports. The host retains Git/signing IO, certified process supervision,
checkout lifetime, current source/target authority and every claim/reservation,
lock, journal, atomic multi-record write and recovery operation. Contained model
file operations keep the same opened checkout and claim guard.

All 87 assigned signature rows are reconciled: 64 implemented replacements and
23 precise cohesive five-parameter contracts. Named input/result values preserve
partial values, absence, errors, ordering and cancellation. Typed committer,
merge-tree and path modes replace internal control choices. Signature-required
and membership booleans remain observed data; eight owned close suppressions
have specific cleanup/error explanations.

Integration entry/transition/blocker phases, merger proposal outcomes, repair
plan outcomes and host acceptance outcomes now use distinct string-backed types.
Their exact values, field order/tags, hashes, nil/empty representation, journal
checks and rejection behavior remain intact. Twelve additional zero/empty/
populated snapshots freeze the committed pre-change record shapes. Accepted
repairs still require the complete exact candidate gate and confer no receipt
or publication authority.

## Implementation Notes

Slice 12 PR #449 is verified integrated into dev at
`a49c32c9f5e81e1bde9f585a25c6d5cc669d7a9e` with its exact-head controller evidence
`EVFKMQZNCCYQDPLZLRXOYO2S7M`.

The [move ledger](../../../plan/code-quality-remediation-slice-13-moves.csv)
resolves 146 current extraction images and preserves their source identities,
raw hashes and introduction commits. Compiling mechanical moves precede API
naming and explicit contract changes. Token comparisons match all 55 immutable
protocol declarations, all 30 attribution/signing declarations, all 17 repair
policy/record declarations and all 12 complete-proof/capability declarations.
The initial broker move matches 27 of 28 declarations; its response formatter
inlines the same original 4096-byte failure bound.

The 908 Store methods retain one local receiver owner. Current hashes and
signatures resolve alongside all 100 original authority boundaries and 79 frozen
dependency references, including the intentional standard-library comparator.
Each stage imports exactly two module owners: shared state and contenthash.
Dependency checks reject coordinator, configuration and mutation-authority
imports. The exact repository debt manifest remains at ceiling six with 62
entries, one external exception and no assigned Slice 13 signature allowance.

## Contracts Added or Changed

`integrate.ValidationInput`, `Completion` and `Validation` bind exact candidate
execution to guarded Store acceptance without exporting claim resources.
`integrate.RepairProof` receives only immutable proof data; its stateless native
adapter checks merge repairs then replay history. Execution and stop contexts
remain explicit parameters.

`merge.Call`, `BrokerResult`, `RepairPosition` and `RepairHistory` name the actual
bounded execution, proposal/usage and observed proof values. Private admission,
claim, workspace, repair, budget, scheduler and delivery values preserve the
original partial/error results. Shared `PreparationQuery`, `MergerRepairPlan`
and `ErrAtomicGuard` have direct native/stage consumers and preserve their
original encoding or sentinel identity. The
[boundary ledger](../../../plan/code-quality-remediation-slice-13-boundaries.csv)
records every typed-state, mode, retained-data and suppression disposition.

## Files of Interest

- `internal/core/workflow/integrate/` and `internal/core/workflow/merge/`.
- `internal/core/workflow/workflow_stage_ports.go` and the existing guarded Store operations.
- `internal/core/workflow/state/` and `testdata/quality-records/`.
- The signature, extraction, authority, Store method and stage-port ledgers linked by the tracker.

## Validation

The controller records the final ordered checks against the exact clean pushed
PR head after the PR-reference commit: `make verify`, focused real workflow
checks, focused claim/process races and `git diff --check`. Their current outcome
is retained with that head by the controller and the canonical PR. Task-level
successful commands and evidence follow; they do not substitute for that final
head-bound validation.

- `go test ./internal/core/workflow ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|GateFailure|Stop|Guard|Journal)|NativePreparation(Methods|Failure|Stop)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; ten top-level checks, including real Git, stop, failure, journal/restart and frozen wire evidence. The existing runner owns and cleans its fixtures.

- `go test ./internal/core/workflow/integrate -count=1`: pass; deadline versus stop, detached acceptance, error preservation and accepted candidate retention.
- `make check-go-style`: pass after the named preparation contracts; ceiling six, 101 exact debts, one external exception.
- `scripts/check-package-readmes.sh`: pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./scripts/goquality`: zero issues.

- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/app ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|GateFailure|Stop|Guard|Journal|Unsupported|Conflict)|NativePreparation(Methods|Failure|Stop|Admission|Fence|Generation)|Native(DriftDuringGate|SourceDuringGate|CandidateRepair|CandidateClean|TargetDrift)|QualityRecordWire|LeafDependencies|ValidationCompletion)$' -count=1 -timeout=15m -v`: pass; 22 top-level checks through the existing owned-fixture runner. Real target/source drift, claim generations/fences, repair, normal/failed/cancelled preparation, journal recovery, app wire and stage/import contracts pass.

- `go test ./internal/core/workflow ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(NativeMerger(Execution|Input|Results|Malformed|Tools|Descendants|Stop)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; ten top-level checks through the existing owned-fixture runner. Actual child execution, protocol/input/result rejection, stop/descendant cleanup, filesystem aliases and frozen wire/import bounds pass.

- `go test -race ./internal/core/workflow ./internal/core/workflow/merge ./internal/core/workflow/integrate ./scripts/goquality -run '^Test(WorkflowPreparationStop|NativePreparationFence|NativeMerger(Execution|Tools|Malformed|Results|Descendants|Stop)|ValidationCompletion|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; 11 top-level checks, including real preparation stop/fence, child execution/cleanup, broker rejection, tools and stage/dependency contracts. No owned descendants remain.

- `go test ./internal/core/workflow ./internal/core/workflow/merge ./internal/core/workflow/state ./internal/app ./scripts/goquality -run '^(Test(NativeMerger(Protocol|Malformed|Fence|Usage|Input|Scope|Results|Decisions|IntentDecision|Regeneration|Unmetered|Settlement|MeterGuards|Execution|Admission|Acceptance|AcceptanceReject|AcceptanceStale|HostRegeneration|HostRegenerationFailure)|NativeRepositoryAliases|NativeBudget(Bounds|Grants|Recovery)|QualityRecordWire|LeafDependencies)|FuzzNativeMergerJSON)$' -count=1 -timeout=15m -v`: pass; 28 top-level checks after the mechanical protocol extraction. Existing protocol/fuzz seeds, repository aliases, real acceptance/regeneration, budget and wire/import contracts pass.
- `go test ./internal/core/workflow ./internal/core/workflow/merge -run '^(TestNativeMerger(Protocol|Malformed|Fence|Usage|Input|Scope|Results)|FuzzNativeMergerJSON)$' -count=1 -timeout=3m`: pass after package-qualified stage naming.
- `go test ./internal/core/workflow ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(NativeMerger(Execution|Admission|Acceptance|AcceptanceReject|AcceptanceStale|HostRegeneration|HostRegenerationFailure|Workspace|Repair|RepairGuards|RepairProof|ReplaySteps|Binary|Semantic|Fix|Descendants|Timeout|Stop|CheckpointGuards|CheckpointBinding)|NativeBudget(Recovery|Grants|Bounds)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v`: pass; 26 top-level checks after named merger admission/claim/repair contracts. Real acceptance/rejection/staleness, regeneration/config protection, both Git object formats, repair proof/replay, budget and supervised cleanup pass. The existing fixture-owning runner completes cleanup.
- `go test -race ./internal/core/workflow ./internal/core/workflow/merge ./scripts/goquality -run '^Test(NativeMerger(Execution|Admission|AcceptanceStale|HostRegenerationFailure|Descendants|Stop|Workspace)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; nine top-level checks after the same contract changes. Real child admission, stale acceptance, failed regeneration, descendants/stop cleanup and workspace isolation pass with the race detector; no owned fixtures remain.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality`: zero issues on the named merger contracts.
- `make check-go-style`: pass on the current contracts; ceiling six, 87 exact debt entries, one external exception.
- `scripts/check-package-readmes.sh`: pass.

- `go test ./internal/core/workflow ./internal/app ./scripts/goquality -run '^Test(NativeBudget(Bounds|Grants|Recovery|Legacy|Unknown)|NativeSchedule(Order|Waits|Tasks|Prerequisite|NoReplay|Inputs)|NativeDecision(Answer|Guards|Retry|Defer|Replay|IntegrationFix|Batch)|NativeReplay(Proof|Guard|Fix|Blocked|Observation)|NativePreparation(Signing|Methods)|NativeSigningBlocked|NativeCandidate(Repair|Clean)|WorkflowPreparation(Methods|Unsupported|Conflict|Journal)|WorkflowIntegration(Methods|Reject|WrongMethod|ChangedInput|History|HostConfig|Cancelled)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v`: pass; 41 top-level checks after named evidence/budget/scheduler and remaining high-arity contracts. Real budget/grant/recovery, selection exclusions, decisions/retry, signing/replay, candidate repair, preparation and delivery pass.
- `go test -race ./internal/core/workflow ./scripts/goquality -run '^Test(NativeBudgetRace|NativeScheduleWaits|NativeDecisionRetry|WorkflowPreparationRetryAuthority|NativeHostPreparationRetry|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=10m -v`: pass; seven top-level checks for budget concurrency, scheduler waits, decision retry and authorized interrupted-preparation retry. The owned-fixture runner completes cleanup.
- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/app ./scripts/goquality -run '^Test(WorkflowPreparationMethods|WorkflowIntegrationMethods|NativePreparationSigning|NativeSigningBlocked|NativeReplay(Proof|Guard|Fix)|NativePreparationCollision|NativeCollisionBudget|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; 12 top-level checks after the 30-declaration mechanical move, including actual Git signing, replay, collision bounds, method proof and frozen wire/dependency checks.
- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(NativePreparationSigning|NativeSigningBlocked|NativeReplay(Proof|Fix)|NativeMerger(ReplaySteps|RepairProof|Semantic|Fix|HostRegeneration|HostRegenerationFailure|Execution|Descendants|Stop)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; 16 top-level checks after candidate naming, explicit modes and narrow predecessor observation. Real signing/replay, semantic repair, host regeneration failures and supervised process/descendant cleanup pass. No owned fixtures remain.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality`: zero issues on the current candidate naming/mode contracts.
- `make check-go-style`: pass; ceiling six, 62 exact debt entries, one external exception. All 87 assigned Slice 13 rows satisfy the final five-parameter/result policy.
- `scripts/check-package-readmes.sh`: pass.

- `go test ./internal/core/workflow ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(NativeMerger(Repair|RepairGuards|RepairProof|ReplaySteps|Semantic|Fix|HostRegeneration|HostRegenerationFailure|Acceptance|AcceptanceReject|AcceptanceStale|CheckpointGuards|CheckpointBinding|Descendants|Stop)|NativeDecision(Replay|IntegrationFix|Retry)|NativeReplay(Proof|Guard|Fix)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v`: pass; 23 top-level checks for semantic repair, regeneration, replay/decision, stale acceptance and supervised cleanup. The existing runner owns and cleans its fixtures.
- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(NativeEntry(Transitions|Admission|Fences|Sequence|Blocker|Pause)|NativeSchedule(Waits|Order|Tasks|Prerequisite|NoReplay|Inputs)|NativeMerger(Protocol|Malformed|Results|Semantic|Fix|Acceptance|AcceptanceReject|AcceptanceStale|RepairProof|HostRegenerationFailure|Stop)|NativeDecision(Replay|IntegrationFix)|NativeReplay(Proof|Guard)|NativeSubmission(Recovery|LostAck)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v`: pass; 29 top-level checks for typed entry phases and merger outcomes through real admission, scheduling and repair. The existing runner owns and cleans its fixtures.
- `go test ./internal/core/workflow ./internal/app ./scripts/goquality -run '^Test(NativeEntry(Candidate|Malformed|Unknown|Alias|Recovery|Retries)|NativeCanonicalLost|NativeReceipt(Lost|Recovery|Uncertainty)|NativeSubmissionRecovery|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v`: pass; 14 top-level checks for malformed records, unknown fields, recovery/lost acknowledgements and all 12 direct record snapshots. The existing runner owns and cleans its fixtures.
- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|ProviderPaths|Guard|Unsupported)|NativeCombinedProof|NativeReplay(Proof|Guard|Fix)|NativeMerger(RepairProof|ReplaySteps|Semantic|Fix)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=12m -v`: pass; 15 top-level checks for consumer-owned repair proof composition and provider guards. The existing runner owns and cleans its fixtures.
- `go test ./internal/core/workflow ./internal/core/workflow/integrate ./internal/core/workflow/merge ./internal/app ./scripts/goquality -run '^Test(WorkflowPreparation(Methods|ProviderPaths|Guard|Unsupported|GateFailure|Stop|Journal)|WorkflowIntegration(Methods|Reject|WrongMethod|ChangedInput|History|HostConfig|Cancelled)|NativeCombinedProof|NativePreparation(Signing|Methods)|NativeSigningBlocked|NativeReplay(Proof|Guard|Fix|Blocked)|NativeMerger(RepairProof|ReplaySteps|Semantic|Fix|AcceptanceStale|Stop)|NativeCanonicalLost|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v`: pass; 32 top-level checks for complete candidate/provenance/signing/capability policy, delivery rejection/cancellation and lost acknowledgement. The existing runner owns and cleans its fixtures.

## Risks and Follow-ups

None recorded. Integration into dev remains a controller-owned transition;
the report and tracker close after the canonical merge is verified.
