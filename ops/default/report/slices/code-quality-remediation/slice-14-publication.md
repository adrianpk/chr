# Slice 14: Forge publication and review stages

Status: drafting
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-14-quality-publication`
PR: pending

## Purpose

Give forge publication and review policy independent stage owners while retaining
native execution authority, guarded host publication, journals and receipt-backed
recovery.

## Delivered Behavior

The review stage owns bounded input encoding and hashing, strict decision parsing,
repair-decision checks, review-policy validation and decision-state selection.
`review.Complete` joins role IO and parsing failures, observes interruption and
requests guarded acceptance through the existing `Decisions` port. It retains the
partial binding and exact role-selected repair task key.

The Store retains command supervision, claims, current source/provider observations,
explicit approval, repository locks, journals and atomic acceptance. Completion
acceptance keeps its five-second detached cleanup bound; host-directed repair keeps
the original operation context. Both seven-parameter acceptance methods now use
bound operation/completion inputs, taking three and four parameters respectively.

The [assigned signature ledger](../../../plan/code-quality-remediation-slice-14-signatures.csv)
records 16 implemented rows and 45 pending rows. The exact debt manifest
has 52 entries, with its six-parameter ceiling and external exception unchanged.

Named host values now bind review observations to their inputs, retained inputs to
decisions, and claims to partial prepared results. The selected command/harness and
hashed evidence references remain cohesive. Application dependency assembly and
all direct control/continuation callers consume one Store/provider result. Two
six-parameter review claim/policy methods now take four inputs and return one
claimed-review value with its error.

The forge stage owns submission response/state checks, native versus legacy PR
identity matching and sealed-unpublished validation. `forge.Submit` performs one
already journaled provider dispatch under the watched context and retains its
reply within the original five-second detached cleanup bound. Provider acceptance
remains an uncertain response until canonical observation confirms delivery. The
host's response adapter retains exact invocation/submission identity and the rule
that a definitive concurrent observation wins over a late reply.

## Implementation Notes

Predecessor PR #450 is verified merged into dev at
`53b2adb23b9b1d20529d00bbe3b32b3be5990b4e`. Controller evidence
`5CL5UBXU2HGBPKW5NNUL6O3PMP` passed its four exact-head commands. The primary
checkout remains on dev; implementation uses the recorded slice branch.

- `42253fe479a1a47d9a465c7359bba1a0df3f3586` closes Slice 13 and freezes all 61 assigned signature rows.
- `ee47cb41ee02ee2df60d475a29128783b0e9ef37` moves eight pure review declarations. All eight normalized token
  streams match their frozen preimages; compilation, wire and dependency checks pass.
- `09431ee3293fdd5a279e5d3b7170dbf7037070a6` applies concise review API names and introduces bounded completion
  through the existing private host adapter. Fresh claim/execution/input checks and
  the failed-review checkpoint retain their order under the Store lock.

- `09b9e4c262dee11a83b7a34d5c886c2ceadb5907` converts 13 more assigned review rows; 42 real review, preparation,
  candidate-repair and application-control checks plus six focused races pass.
- `3e053ee5515e9943ab6ddac1a8024f5fb3f73d13` moves six pure forge declarations with all six normalized token
  streams preserved; compilation, frozen wire and dependency checks pass.
- `932090eb2531dba22e53b2ec57264c38c49d9429` applies concise forge API names and extracts dispatch/response flow.
  The private response operation retains its guarded journal transaction.

The [extraction ledger](../../../plan/code-quality-remediation-slice-14-moves.csv)
keeps 14 mechanical preimages plus two explicit flow extractions.
The flow entries record their original host methods without claiming whole-method
moves or token equivalence. Current images reconcile all 908 Store methods, 2,918 prior
relocations, 146 Slice 13 extractions, 100 authority boundaries and 79 dependencies;
original source identities and hashes remain frozen. The standard-library
`cmp.Compare[int]` reference remains an external dependency identity.

## Contracts Added or Changed

`review.Execution` carries one immutable input and observed role output/failure.
`review.Result` carries acceptance's partial binding and the exact repair key.
`Completion` and `Decisions` retain their existing acceptance shape. The private
host adapter binds the claim, prepared input and provider without exposing Store.

The review package imports only shared state and contenthash within the module.
Forge imports only shared state. Its `Submitter` is a direct consumer-owned host IO
port; `PublicationProvider` includes that unchanged guarded submission method.
`Responses` stays bound to the private host operation, which now receives one
reply/failure value instead of parallel response parameters.
The dependency checker permits that actual hash dependency and continues to reject
coordinator, configuration and mutation imports. Wire/hash encodings, persisted
state spellings and public service signatures remain unchanged in this unit.
Ticket-native execution and forge-native publication remain distinct terms.

## Files of Interest

- `internal/core/workflow/review/{input,parse,policy,execution,decision}.go`.
- `workflow_review.go`, `workflow_review_result.go` and `workflow_stage_ports.go`.
- `forge/{identity,policy,dispatch,response}.go` and `workflow_review_operation.go`.
- `internal/app/service_workflow_review.go` and its direct control/continuation callers.
- The assigned-signature, extraction, Store-method, owner and stage-port ledgers.

## Validation

These are successful task checks for the recorded review and forge units.
The real fixture selections ran through the existing owned-process QA runner,
which completed cleanup after each command.

- `go test ./internal/core/workflow/... ./internal/app ./scripts/goquality -run '^Test(QualityRecordWire|LeafDependencies)$' -count=1 -timeout=3m` — PASS for the mechanical relocation.
- `go test ./internal/core/workflow ./internal/core/workflow/review ./internal/app ./scripts/goquality -run '^Test(WorkflowReview(Modes|Approval|Policy|Restriction|ChangedInputs|Revocation|DecisionChecks|Repairs|DuringIO|Stop|Journal|RoleJournal|RepairJournal)|NativeReview(Modes|Inputs|Policy|Observer|PolicyIO|Fence|Capabilities)|NativeAgentReview|NativeCandidateReviewRepair|ServiceWorkflowReview|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, 26 top-level checks covering review modes, approval, policy,
  source/provider changes, revocation, decision checks, repairs, IO, stop and journals.
- `go test -race ./internal/core/workflow ./internal/core/workflow/review ./scripts/goquality -run '^Test(WorkflowReview(Concurrent|DuringIO|Stop)|NativeReviewFence|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, six top-level checks including concurrent review, IO, stop,
  native provider fences, frozen wire values and dependency boundaries.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality` — PASS with zero issues after correcting two required blank lines.
- `make check-go-style` — PASS: 9,324 production declarations, ceiling six,
  60 exact debt entries and one external exception.
- Eight current pure-policy declaration token streams match their frozen sources
  after reversing only package API names. Provenance counts and frozen source
  fields were checked directly against the versioned ledgers and current AST.
- `git diff --check` — PASS.

- `go test ./internal/core/workflow ./internal/app ./scripts/goquality -run '^Test(WorkflowReview(Modes|Approval|Policy|Restriction|ChangedInputs|Revocation|DecisionChecks|Repairs|DuringIO|Stop|Journal|RoleJournal|RepairJournal)|NativeReview(Modes|Inputs|Policy|Observer|PolicyIO|Fence|Capabilities)|NativeAgentReview|NativeCandidate(ReviewRepair|Substitution|RepairGuard)|WorkflowPreparation(Methods|Guard|IO|Approval|Journal)|Native(PreparationFence|SourceDrift|DispatchGuard)|Service(WorkflowReview|NativeReview|NativeProvider|IntegrationGrant|IntegrationApproval|WorkflowSubmission|WorkflowManual)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, 42 top-level checks for the review value conversion.
- `go test -race ./internal/core/workflow ./scripts/goquality -run '^Test(WorkflowReview(Concurrent|DuringIO|Stop)|NativeReviewFence|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, six top-level checks for the converted claim/input paths.
- `go test ./internal/core/workflow ./internal/core/workflow/forge ./internal/app ./scripts/goquality -run '^Test(WorkflowSubmission(Methods|LostResponse|Budget|Policy|AtomicGuard|Exclusion|ProjectionLoss|Stop|Journal|Corruption)|NativeSubmission(Unknown|Recovery)|NativeObservationProof|ServiceWorkflowSubmission|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, 17 top-level checks for submission methods,
  lost response, budgets/policy, atomic guards, exclusion, projection loss, stop,
  journals/corruption, native recovery/observation proof and the application path.
- `go test -race ./internal/core/workflow ./scripts/goquality -run '^Test(WorkflowSubmission(Stop|Exclusion|LostResponse)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, five top-level checks including submission
  stop, exclusion, lost response, frozen wire and dependency boundaries.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality` — PASS with zero issues for both completed units.
- `make check-go-style` — PASS on the current source: 9,325 production declarations,
  ceiling six, 52 exact debt entries and one external exception.
- Six current forge policy token streams match their frozen preimages after
  reversing only package API names. All current images and frozen source fields
  in the extraction, signature, method, authority and dependency ledgers resolve.
- `git diff --check` — PASS after report, tracker and ledger updates.

## Risks and Follow-ups

PR creation/observation and native publication/review binding policy extraction,
the remaining 45 assigned signature rows and local stored-state/flag/error/
suppression/glossary reconciliation are active Slice 14 work. Final `make verify`,
remaining focused publication/receipt checks and exact-head controller validation
must finish before the canonical PR is ready; no Slice 14 PR is open yet.
