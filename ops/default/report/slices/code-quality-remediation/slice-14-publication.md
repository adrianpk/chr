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
records 50 implemented rows and 11 pending category B rows. All assigned
category A/C rows are implemented. The exact debt manifest has 24 entries, with its six-parameter ceiling and external exception unchanged.

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

Submission and publication values now distinguish journaled intent and dispatch
permission, sanitized transport output, canonical observation, next revision fence
and receipt evidence. `SubmissionIntent` preserves the original retry decision;
`NativePublicationResult` keeps the transport result separate from delivery proof.
Private callback snapshots and owner/service bundles replace parallel parameters
while contexts and guarded Store transitions remain explicit.

`Handled` and `Observed` preserve the existing native-path and publication barriers,
including partial/error paths. They do not claim that finalization succeeded.
Partial journal, owner, PR, transport, observation, fence, proof and receipt values
survive the same errors and deferred cleanup as before. The existing closeout
reader adapts the native proof/hash value while retaining its own current API.


The forge client now owns provider-specific request arguments, all-state PR
pagination, structured PR creation/read/branch IO and bounded JSON decoding.
Native policy/identity reads consume this same request protocol. The host adapter
retains enabled configuration, canonical URL/repository resolution, prerequisite
diagnostics and the existing bounded process runner. Shared PR query, request and
content values live in state; direct application and host callers use that owner.
The separate transport input boolean is replaced by the request body's nil state.

Native retained policy, transport replies/advertisements, observation bindings,
receipts and the history-divergence sentinel now have a shared state owner.
Provider reference parsing and enum values have a forge owner. Direct consumers,
including ticket merge diagnostics, use those owners without compatibility aliases.

Forge owns effective repository/branch policy reads and decoding, publisher API
identity hashing, capability sealing, immutable publication revision encoding and
canonical native observation/classification. The existing native policy test moved
beside its private implementation without removing assertions. The host retains
configuration/grant composition, bounded processes and the callback-bearing Git
request; only immutable identity crosses the stage boundary.

Canonical PR/branch and Git observers are read-only ports. The native proof port
passes both expected query and observed proof to the unchanged integration/history
validator. Exact source/result attribution, uncertain outcomes, receipt descendants,
history divergence, partial values and the three-minute observation bound remain.

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

- `187a9ec4c15fef69e137c158fe574163aaae5497` converts 34 additional assigned rows, including all remaining
  submission/publication result tuples and the three high-arity native callbacks.
  All direct consumers retain their assertion and authority behavior. Original
  named/unnamed error scope and conditional-initializer variable scope are preserved.

- `f22302613d1462d2bee9d20f522bc1cb15127040` extracts provider PR protocol and relocates shared inputs.
  Eleven pure declaration token streams match after naming changes. Four PR flow
  entries retain their original host configuration gates; provider IO does not gain
  store, claim, approval or receipt authority.

- `f026941182c708d20a5a36979dd8e71e3cd54ba7` moves eight neutral retained values/errors and eleven provider reference
  declarations. Sixteen complete declaration streams match after owner naming;
  the split provider enum group retains exact values and host diagnostic constants.
- `41595fd23b14704199d3b7b94ae0022bd8b10cff` moves native provider policy/identity and canonical observation into
  forge. Nineteen additional declaration streams match after recorded API/field/
  type naming. Exact query/proof composition, callback ownership and immutable
  revision projection remain explicit; 47 real local checks and five focused races pass.

The [extraction ledger](../../../plan/code-quality-remediation-slice-14-moves.csv)
keeps 91 entries. Pure stream checks, explicit input/port transformations,
split constant members, inlined paths and host/leaf flow extractions are identified
individually. Flow entries do not claim whole-method moves or token equivalence. Current images reconcile all 908 Store methods, 2,918 prior
relocations, 146 Slice 13 extractions, 100 authority boundaries and 79 dependencies;
original source identities and hashes remain frozen. The standard-library
`cmp.Compare[int]` reference remains an external dependency identity.

## Contracts Added or Changed

`review.Execution` carries one immutable input and observed role output/failure.
`review.Result` carries acceptance's partial binding and the exact repair key.
`Completion` and `Decisions` retain their existing acceptance shape. The private
host adapter binds the claim, prepared input and provider without exposing Store.

The review package imports only shared state and contenthash within the module.
Forge imports shared state and contenthash. Its `Submitter` is a direct consumer-owned host IO
port; `PublicationProvider` includes that unchanged guarded submission method.
`Responses` stays bound to the private host operation, which now receives one
reply/failure value instead of parallel response parameters.
The dependency checker permits those actual hash dependencies and continues to reject
coordinator, configuration and mutation imports. Wire/hash encodings and persisted
state spellings remain unchanged. Service/adapter callers now consume the shared
state PR input types directly; no compatibility aliases are introduced.
`forge.Commands` exposes only bounded command Output/Input IO to `Client.Request`.
The host resolves configuration before constructing an endpoint; leaf modules
import only shared state and contenthash within the module. PR creation is metadata
IO, not permission to dispatch integration or proof of delivered work.
Native observation consumes `NativePullRequests`, `NativeGit`, optional
`PreparedNativeGit` and `NativeProofs`; the host supplies configured read-only IO
and stateless exact method/history validation. The leaf receives no Store, claim,
lock, process or dispatch callback. The capability observation still uses current
host grants, and canonical observation issues no receipt.
Ticket-native execution and forge-native publication remain distinct terms.

## Files of Interest

- `internal/core/workflow/review/{input,parse,policy,execution,decision}.go`.
- `workflow_review.go`, `workflow_review_result.go` and `workflow_stage_ports.go`.
- `forge/{identity,policy,dispatch,response,client,pull_request}.go` and `workflow_review_operation.go`.
- `state/{pull_request,native_publication}.go` and `forge/{reference,native_*}.go`.
- `workflow_publication_identity.go` and native observation/config/proof assembly.
- `workflow_publication_operation.go`, native dispatch/observation/receipt owners
  and the existing closeout proof reader.
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

- `go test ./internal/core/workflow ./internal/app ./scripts/goquality -run '^Test(WorkflowSubmission(Methods|LostResponse|AtomicGuard|Journal)|WorkflowPublication(Modes|Unknown|LateResponse|Crash)|NativeSubmission(Unknown|Recovery|Responses|Fences|Applied)|NativeDispatch(|Guard|Policy|Evidence)|NativeGitOutcomes|NativeCanonical(Outcomes|Poll|Fence|Lost)|NativeReceipt(Proof|Delivery|Recovery)|NativeFinalization(|Ticket)|NativeClosing(Publication|Evidence)|NativeHost(Publication|Denials)|ServiceWorkflowSubmission|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, 35 top-level checks for converted native dispatch,
  callbacks/observation, receipts/finalization, legacy publication/recovery and
  application dependency paths. The existing owned-process runner completed cleanup.
- `go test -race ./internal/core/workflow ./scripts/goquality -run '^Test(NativeDispatchConcurrent|NativeCanonicalFence|NativeFinalizationConcurrent|WorkflowPublicationLateResponse|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, 6 top-level checks for changed concurrent dispatch,
  observation/finalization and late-response paths, plus frozen wire/dependency guards.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality` — PASS with zero issues for this conversion.
- `make check-go-style` — PASS: 9,325 production declarations, ceiling six,
  24 exact debt entries and one external exception. No assigned Slice 14 declaration
  remains in temporary exact debt; the 11 pending rows still require their five-input
  cohesion disposition.
- Current AST images, immutable source columns, report/ticket/tracker links and
  package READMEs resolve; all 908 Store methods retain their guarded owner.
- `git diff --check` — PASS after documentation/provenance updates.


- `go test ./internal/core/workflow ./internal/app ./scripts/goquality -run '^Test(WorkflowPR(Adapter|AdapterBounds|Branch|Command|CommandDeadline)|WorkflowSlicePR(Concurrent|Reuse|Observation|Block|Existing|Repair|Source)|Native(Capability|CapabilityDenials|PolicyDenials|Observation|ObservationProof|HostPublication|HostDenials)|ServiceWorkflowSlicePR|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, 22 top-level real local checks for exact adapter protocol, response
  bounds, command deadlines, PR reuse/uncertainty/concurrency/source binding,
  native capability/observation and the application PR path. The additional
  `TestNativePolicyDenials` requires explicit live fixture opt-in and skipped;
  this command does not establish live-forge certification.
- `go test -race ./internal/core/workflow ./scripts/goquality -run '^Test(WorkflowSlicePRConcurrent|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, three top-level checks for concurrent PR intent, frozen wire values
  and leaf dependencies.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality` — PASS with zero issues for the PR transport extraction.
- `make check-go-style` — PASS: 9,333 production declarations, ceiling six,
  24 exact debt entries and one external exception.
- Eleven pure PR wire decoder/content/input declarations match their frozen
  token preimages after naming; the existing `PullRequestSnapshot` declaration
  remains byte-identical. Frozen wire fixtures retain nested PR binding bytes.
- Current hashes, frozen source columns and report/ticket/tracker links resolve.
- `git diff --check` — PASS after report and provenance reconciliation.


- `go test ./internal/core ./internal/core/workflow ./internal/app ./scripts/goquality -run '^Test(DetectReviewProviderReference|RunReviewProviderDiagnostics.*|RemoteReviewAdapter.*|RunTicketMerge(Preflight|Command).*|QualityRecordWire|LeafDependencies|WorkflowPRAdapter|NativeMergePolicy)$' -count=1 -timeout=5m -v` — PASS, 29 top-level checks for provider URL/diagnostic and remote/ticket-merge
  consumers, immutable nested wire values, PR adapter/policy and leaf dependencies.
- `go test ./internal/core/workflow ./internal/core/workflow/forge ./internal/app ./scripts/goquality -run '^Test(Native(Capability|CapabilityDenials|MergePolicy|Observation|LaterSource|ObservationProof|Source(Result|ResultGuards|Repair|Later|Contract|Capability|Receipt)|Host(Publication|Denials)|Canonical(Outcomes|Poll|Fence|Deadline|Timeout|Lost)|Receipt.*|Closing(Publication|Evidence))|Service(NativeReview|NativeProvider|WorkflowSubmission)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, 47 top-level real local checks for native capability/denials, exact
  observation, source attribution, canonical deadlines/fences/lost responses,
  receipt proof/recovery/descendants/history divergence and application callers.
- `go test -race ./internal/core/workflow ./internal/core/workflow/forge ./scripts/goquality -run '^Test(NativeCanonicalFence|NativeFinalizationConcurrent|NativeReceiptTransition|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS, five top-level checks for canonical observation fences, concurrent
  finalization, receipt transitions, frozen wire values and dependency guards.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/core/workflow/... ./internal/app ./scripts/goquality` — PASS with zero issues.
- `make check-go-style` — PASS: 9,344 production declarations, ceiling six,
  24 exact debt entries and one external exception.
- Sixteen complete retained/reference and nineteen additional policy/revision
  declaration streams match recorded naming/type/field normalizations. The original
  publication revision payload is unchanged; the host projection excludes live callbacks.
- Ninety-one extraction images, source hashes, split provider/diagnostic constant
  lineage, current method/declaration images, links and package READMEs resolve.
- `git diff --check` — PASS after documentation/provenance reconciliation.

## Risks and Follow-ups

Remaining review transport, host native capability composition and review-input
binding policy extraction,
the remaining 11 category B cohesion dispositions and local stored-state/flag/error/
suppression/glossary reconciliation are active Slice 14 work. Final `make verify`,
remaining focused publication/receipt checks and exact-head controller validation
must finish before the canonical PR is ready; no Slice 14 PR is open yet.
