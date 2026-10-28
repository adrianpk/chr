# Slice 14: Forge publication and review stages

Status: reviewing
Delivery set: code-quality-remediation
Plan: [Approved plan](../../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../../tracker/code-quality-remediation.md)
Branch: `refactor/slice-14-quality-publication`
PR: #451

## Purpose

Give forge publication and review policy independent owners while retaining exact
candidate authority, explicit approval, host publication, journals and receipt-backed recovery.

## Delivered Behavior

The forge stage owns provider reference parsing, bounded PR create/read/branch
protocol, review snapshot/diff/comment/check and merge protocol, native protection
and identity policy, capability revisions and canonical observation/source attribution.
Configured host adapters retain enabled endpoint resolution, prerequisite diagnostics,
process supervision, current trusted grants and the live publication callback.
Optional Forgejo checks retain absence; inline comments retain their existing general
fallback. Command and response bounds, partial results and error order are preserved.

The review stage owns immutable input encoding/hashing, strict decision parsing,
repair/policy checks, native candidate/owner/Git/capability binding and bounded
completion. Read-only proof composition uses the existing host validators without
sibling-stage imports. Current entry/source/generation fences, claims, explicit
human approval and atomic decision acceptance remain host operations.

Named operation/result values distinguish dispatch permission, uncertain transport
response, canonical observation, revision fences and receipt evidence. Original
five-second detached completion/retention bounds and three-minute native capability/
observation bounds remain. A late reply cannot replace definitive observation;
uncertainty cannot permit another dispatch. `Handled` and `Observed` retain their
partial/error barriers and do not claim finalization.

Seven distinct named string state types cover submission, response, review,
dispatch, Git application, transport and canonical observation. Original stored
strings, zero values, unknown-value decoding, JSON order/tags and hash inputs remain.
Application display and IPC projections convert at their existing string boundary.
One-shot versus polling observation is an explicit host mode; both keep the same
guards, caller context and receipt checks.

All 61 assigned signature rows are reconciled: 54 implemented and seven specific
cohesive five-parameter retentions. All satisfy the final parameter/result policy;
no Slice 14 declaration remains in the 24-entry temporary debt manifest.
Ticket-native execution and forge-native publication have distinct
[code-backed terms](../../../../../docs/explanation/architecture.md#publication-and-review-terms).

## Implementation Notes

Predecessor PR #450 is verified merged into dev at
`53b2adb23b9b1d20529d00bbe3b32b3be5990b4e`. The primary checkout stays on dev;
Slice 14 uses its dedicated worktree and recorded branch.

- `42253fe479a1a47d9a465c7359bba1a0df3f3586` closes Slice 13 and introduces this report and the 61-row signature ledger.
- `ee47cb41ee02ee2df60d475a29128783b0e9ef37` and `3e053ee5515e9943ab6ddac1a8024f5fb3f73d13` mechanically move eight review and six forge declarations before separate API naming.
- `09431ee3293fdd5a279e5d3b7170dbf7037070a6`, `09b9e4c262dee11a83b7a34d5c886c2ceadb5907`, `932090eb2531dba22e53b2ec57264c38c49d9429` and `187a9ec4c15fef69e137c158fe574163aaae5497` introduce bounded review/submission flows and precise operation/partial-result values.
- `f22302613d1462d2bee9d20f522bc1cb15127040`, `f026941182c708d20a5a36979dd8e71e3cd54ba7` and `41595fd23b14704199d3b7b94ae0022bd8b10cff` extract PR protocol, retained native values/reference parsing, native policy and canonical observation. Eleven PR, sixteen retained/reference and nineteen policy/revision normalized declaration streams match their preimages.
- `1685cf0959165446dbc9e51d681197bd270c21fc` owns legacy review/merge protocol in forge. Thirty-three complete normalized streams match; actual host configuration/process responsibilities remain local.
- `70ec732745fa1706ef9ae6e196a1818b4e83fa4d` owns native capability assembly through read-only provider, current-policy and privately bound transport ports.
- `281b0f9fae1b00eaa39c60178ec598109022ee23` extracts native review-input policy, removes the unused invocation operand, pairs independent observer inputs and documents the glossary. The remaining seven five-input contracts keep their distinct receipt, current-entry, approval, control and transaction operands explicit.
- `2adac31efaf111d6a15115afd3b76ad62e3d8bbf` names retained publication/review states and observation modes without changing their encoding or transitions.

The [143-entry extraction ledger](../../../plan/code-quality-remediation-slice-14-moves.csv)
distinguishes pure relocation, split constant members, input transformations and
host/leaf flow extraction; flow entries do not claim whole-method token equivalence.
Current images reconcile all 908 Store methods, 2,918 original relocations, 146
Slice 13 extractions, 100 authority boundaries and 79 dependency identities.
Original source IDs/hashes and earlier implementation commits remain frozen.
The external `cmp.Compare[int]` identity remains outside repository source hashing.

## Contracts Added or Changed

`review.Execution`, `Completion`, `Result` and `Decisions` express one role's
input, observed output/failure and guarded acceptance. `NativeProofs` exposes
immutable record/preparation validation only. `forge.Submitter` dispatches an already
journaled intent; `Responses` retains its reply against exact invocation/submission
identity. `Commands` is the single direct provider/review command IO port.

Native capability/observation ports expose canonical PR/branch, effective provider
policy, identity, current trusted grants, transport advertisement and expected-query/
observed-proof checking. Callback-bearing `GitPublication`, Store, claims, config,
locks, journals, process handles and receipt mutations stay in the host. Forge and
review each import only shared state and contenthash within the module; dependency
guards continue to reject coordinator, configuration and mutation imports.

`SubmissionIntent.Dispatch`, `NativePublicationResult.Transport` and the private
barrier/retention/owner/observer values preserve their distinct partial outcomes.
The duplicate host command-runner interface is removed in favor of `forge.Commands`;
all direct service, ticket, control, continuation and adapter consumers use actual
owners without compatibility aliases.

The remaining boolean operands are factual projection inputs: whether a current
decision authorizes domain images or a verified delivery completes the selected
extent. They confer no authority. Native permission booleans remain independent
explicit grants. Local error identities and messages remain stable; the native
history-divergence sentinel has one state owner. The read-only profile close
suppression retains its existing explanation; no bare suppression or swallowed
publication/durability/process-cleanup error is introduced.

## Files of Interest

- `internal/core/workflow/forge/{client,pull_request,review_client,native_capability_observe,native_observation}.go`.
- `internal/core/workflow/review/{input,parse,policy,execution,native}.go`.
- `internal/core/workflow/state/{publication_state,native_publication,native_records}.go`.
- Host review/native capability/proof assembly, publication operation values and guarded journal/receipt owners.
- Stage READMEs, the architecture glossary, signature/extraction/ownership/method/authority/dependency/port ledgers.

## Validation

Task-level successful checks follow. Real Git/provider/process selections use the
existing fixture-owning runner; it cleans only owned descendants after completion.
Live-forge certification is an explicit opt-in fixture and is not established by
these local selections. Earlier compiling-move and operation checks remain in
the task commits and Acta records.

- `go test ./internal/core ./internal/core/workflow/... ./internal/app ./scripts/goquality -run '^Test(RemoteReviewAdapter.*|RunReviewProviderDiagnostics.*|RunTicketMerge.*|RunTicketReview.*|QualityRecordWire|LeafDependencies|WorkflowPRAdapter|NativeMergePolicy)$' -count=1 -timeout=5m -v` — PASS: 28 provider, diagnostic, ticket review/merge, wire and dependency checks.
- `go test ./internal/core/workflow ./internal/core/workflow/forge ./internal/app ./scripts/goquality -run '^Test(Native(Capability|CapabilityDenials|SourceCapability|Host(Publication|Denials)|Review(Modes|Inputs|Policy|Observer|PolicyIO|Fence|Capabilities)|PreparationFence|DispatchGuard)|Service(NativeProvider|NativeReview)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS: 19 real capability, review, dispatch and application checks.
- `go test -race ./internal/core/workflow ./internal/core/workflow/forge ./scripts/goquality -run '^Test(NativeReviewFence|NativeDispatchConcurrent|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS: four focused review-fence/dispatch and wire/dependency race checks.
- `go test ./internal/core/workflow ./internal/core/workflow/review ./internal/app ./scripts/goquality -run '^Test(Native(Review(Modes|Policy|Inputs|Observer|PolicyIO|Capabilities|Fence)|Submission(Recovery|Responses|Fences|Applied|NonApplication)|Stop(Pending|BeforeSend|Rejected|Receipt)|Fanout(Evidence|Barrier|Crash)|Receipt(Recovery|Restart)|DispatchGuard)|WorkflowReview(Approval|Modes|ChangedInputs|Policy|Restriction|Repairs|Journal)|WorkflowSubmission(Methods|LostResponse|Journal)|Service(NativeReview|WorkflowReview)|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS: 37 real review, human approval, submission/lost-response/journal, stop/fanout/receipt, application and wire/dependency checks.
- `go test -race ./internal/core/workflow ./internal/core/workflow/review ./scripts/goquality -run '^Test(NativeReviewFence|WorkflowReviewConcurrent|NativeDispatchConcurrent|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=15m -v` — PASS: five focused review/dispatch and wire/dependency race checks.
- `go test ./internal/core/workflow ./internal/core/workflow/forge ./scripts/goquality -run '^Test(WorkflowGit(Packets|MalformedAdvertisement|Reports)|NativeDispatchEvidence|NativeMergePolicy|QualityRecordWire|LeafDependencies)$' -count=1 -timeout=3m -v` — PASS: seven protocol, dispatch evidence, policy, frozen-wire and dependency checks on typed states.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core ./internal/core/workflow/... ./internal/app ./scripts/goquality` — PASS: zero lint issues.
- `make check-go-style` — PASS: 9,359 production signatures, ceiling six, 24 exact debt entries and one external exception.

The controller records `make verify`, the focused real-boundary selection, the
focused race selection and `git diff --check` in that order against the final
clean pushed PR head. Its exact command list, head and retained log are the final
readiness evidence; the canonical PR carries the same ordered commands.

## Risks and Follow-ups

Canonical dev integration remains controller-owned. This report and the ticket
stay reviewing until that integration is verified. The retained native profile
requires its existing independently certified provider/method configuration.
