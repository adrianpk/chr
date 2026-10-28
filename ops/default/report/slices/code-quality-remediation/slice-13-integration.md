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
All 87 assigned signature rows are reconciled: 64 have implemented replacements
and 23 retain specific cohesive five-parameter contracts. The exact repository
debt manifest now has 62 entries, with no Slice 13 signature entry, at ceiling
six and one external exception.

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

A subsequent mechanical batch moves 41 immutable input/result/intent/decision
policy declarations into merge and 14 shared lexical validators into state. All
55 normalized declaration token streams match the frozen source before and after
stage naming. Production callers use merge.SealInput, merge.CheckInput,
merge.ParseResult, merge.InputFence and merge.RequiredDecisions directly. There
is no forwarding layer. The merge dependency bound now permits only shared
state and contenthash for existing protocol checksums.

Named admission, claimed preparation and workspace results preserve original
partial values and errors. Private repair/regeneration inputs bind the same
claim, prepared task and observations; contexts remain explicit. Typed
exact/declared path checks replace internal boolean choices. All 20 newly
implemented signature rows have at most five parameters and bounded results.
Role reservation, lease/generation checks, immutable evidence reads, host Git
continuation, journal writes and resource cleanup retain their existing owner
and order.

Named evidence, budget/grant, scheduler, delivery, signing and retry results
preserve absent, present and failed outcomes, including partial values. The last
six high-arity inputs now bind their actual snapshot, replay or commit data.

The next mechanical extraction moves 30 attribution, provenance/replay,
preparation identity/collision and signing declarations into integrate. All 30
declaration token streams match after owner naming in the mechanical commit.
Production callers now use integrate.ParseCommit, integrate.SignaturePayload,
integrate.DistinctResult and the named provenance/signing checks directly. Git
reads/writes, private signer access, source/target observation and guarded
acceptance remain with their native host/Store owners. Typed committer and
merge-tree modes preserve their original rules. Eight owned production close
suppressions have precise explanations. The
[boundary ledger](../../../plan/code-quality-remediation-slice-13-boundaries.csv)
records these decisions and the stored-state work still pending.

## Implementation Notes

The predecessor PR #449 is verified merged into dev at
`a49c32c9f5e81e1bde9f585a25c6d5cc669d7a9e`. Its exact-head controller evidence
is `EVFKMQZNCCYQDPLZLRXOYO2S7M`; all four recorded commands passed.

Completed task commits:

- Candidate stage: `05a7a106029d64ee5aa1ba8b62703afc40437d72`.
- Preparation contracts: `c7a15133e0a07f8eb3fec735e36c967730aff038`.
- Mechanical merger broker: `962e03aef96bf7aefd2b075fe218041b9cd00dc6`.
- Stage naming and broker/driver contracts: `8e3630b9b074ec7b15d7b0de252429c3844dd283`.
- Mechanical immutable merger policy: `af02c72f04ec8fb1c8554f7fa777a2e18958dacc`.
- Merger stage naming: `a4fa20fe7344c49dbc164e726dcac9883b574c32`.
- Named merger claims and repair observations: `7f2d18b1a04821084af93f0c825d3f3376c791cf`.

- Named evidence/budget/scheduling contracts: `d9bfaa918823edcdeb698e7fff13139663ce9d63`.
- Mechanical integrate attribution/signing policy: `77b24972b7361457978889a8f7f69476fae8d860`.
- Candidate API names, explicit modes and close explanations: `40ab73aaaed653ee7606801c09f9ad43b2e933c1`.

All 908 Store methods retain one local receiver owner. Current method and move
hashes follow these changes while their original source identities/hashes remain
frozen. The current-file owner map records the actual stage files and private
preparation resource bindings.

## Contracts Added or Changed

`ValidationInput` carries immutable candidate identity, preparation, lane and
initial task result. The stage consumes `Validation`; it receives neither a
mutable store nor claim/process resources. Execution and stop contexts remain
explicit parameters. The integrate module dependency bound is shared state plus contenthash for
existing retained checksums; coordinator, configuration and mutation authorities
remain blocked. No alias
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

## Risks and Follow-ups

Remaining candidate proof/capability assembly and semantic/tree-repair policy
still need their final stage boundaries. The boundary ledger retains three
pending stored-state entries. The slice is active; final controller validation
and the canonical PR are pending.
