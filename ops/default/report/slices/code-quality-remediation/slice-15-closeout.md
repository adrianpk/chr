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

Closing input construction and actual-member selection now have closeout owners.
The InputSource port supplies finite host observations while BuildInput retains
contract hashing, partial returns, descriptor order and the original shared
evidence map. Native record, task, receipt, snapshot and delivery-proof values
have one state owner. Record/snapshot validation remains explicit in the host,
with original value copies and guarded writes preserved.

Invocation closing eligibility, ordered per-attempt process checks, frontier
selection and accepted outcome projection now execute in closeout. The host
supplies RunGuard for process quiescence and final input fences. Projection keeps
the live execution-target alias while retained closure history copies its value.
Retained closed-set attachment, invocation receipt/evidence inspection and
closing budget enumeration use Evidence with their distinct original error and
validation contracts. Twenty-seven public checkpoint records have exact state
owners; private token hashes, unknown JSON fields, source hashes and atomic
journal operations remain with the host.

Closing status aggregation now belongs to closeout. It verifies receipt fanout
before reading ordered publication heads, observing live capacity and checking
the retained final receipt. Duplicate blockers, partial status on error and gate
outcomes retain their order. The host supplies only those observations through
StatusSource; locked snapshots and private journals stay with the host. Capacity
and credential-free role records have direct shared-state consumers.

Adoption binding, amendment slice construction, materialization contract images
and projection registration are private coordinator functions with no Store
receiver. The Markdown reader is a private persistence function; readRecord
still resolves its contained path first. Six actual callers use these functions
directly, preserving the original bodies, partial outcomes and atomic guards.

`CloseSet` and `CloseInvocationSet` now sequence closing as actual coordinator
operations. A private closing host owns watched deadlines, isolated checkout IO
and process/permit/output lifetime. Its finite Workspace/Commands binding keeps
the original private claim. Store still owns guarded admission, process records,
credential/input/target fences, receipt confirmation and interrupted recovery.
Application and continuation callers use the new operations directly.

Task, slice and manual execution now use actual coordinator operations and a
private task host for lease deadlines, gate commands and partial results. The
stop watcher owns its ticker, goroutine and cancellation outside Store. Guarded
claim creation, input freezing, current-candidate checks, process journals,
verified replay and complete result/claim transactions retain their Store owner.
Manual handoff and successful worker exit still require explicit host-verified
candidate reconciliation before task acceptance.

Preparation, review, merger and submission execution now use actual coordinator
operations and private role hosts. Merger workspace handles, driver/procedure
lifetimes, trusted Git repair observation and partial acceptance evidence retain
their existing bounds and order. The exact provider and trusted repair Git ports
share one operation value; proposal-only execution cannot accept a repair.
Guarded claims, approval, budgets, process birth, candidate and complete result
or receipt transactions remain with Store. Accepted repair still requires the
original candidate validation and review before publication.

Review mode admission and human-only invocation restriction now have actual
`workflow/review` policy owners. Store reads current configuration and applies
waiting, trusted decision or agentic claim under the original review lock.
Ordinary and candidate-repair preparation Git effects use a private host under
the existing admission guard and lease. Accepted repair/evidence reads, current
source fences and complete claim/input/approval/result transactions retain Store.

Publication advancement, explicit native reconciliation and continuation now use
actual coordinator operations. The private continuation host owns only the
accepted bounded step loop. Each native observation still reconciles Git,
re-reads the current guard, selects one observation or bounded polling, re-reads
again and confirms only the retained receipt. The sole poll-only forwarding
method is removed; its caller supplies the same explicit poll mode. Acceptance
locks and complete journal, budget, decision, projection, receipt and checkpoint
transactions remain with Store.

Lifecycle control now has an actual coordinator for freezing and post-acceptance
continuation. Store owns the original minute-bounded acceptance lock and complete
action/receipt transaction. The existing frozen request/digest and partial
acceptance values cross that boundary. Acceptance cancellation still precedes
caller-context continuation; verified receipt replay returns without steps.

Canonical PR provider read/find/create sequencing now has a private host under
the unchanged minute-bounded repository lock. Store retains current-candidate
revalidation, creating intent before send and complete observation/binding/task/
set transactions. Exact PR identity and URL policy have actual forge owners and
direct consumers. Uncertain creation remains observation-only; neither absence
nor merged metadata can authorize a replacement PR or confirmed delivery.

Native single observation, bounded polling and Git reconciliation now use actual
coordinator functions. A private observation host sequences canonical provider
IO and detached retention; Store retains reconciliation snapshots, evidence
binding, callback fences and complete journal/projection transactions. Actual
continuation, closing and frontier consumers call these owners directly. The
persisted deadline, bounded backoff, cancellation and partial outcomes are exact.

Frontier target/receipt/branch IO and live capacity process/slot/closing-owner
aggregation now have private observers. A private integration role selector owns
ordered finishable validation, preparation, merger and agentic-review eligibility.
Store retains original coherent query, admission and frontier locks, current
record/candidate/target fences and complete receipt/entry/invocation transactions.
Unknown process/slot exclusion, partial observations, role priority and human or
uncertain-publication barriers preserve their original behavior.

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
retains 268 source/extraction/consolidation images. Current identities and hashes are updated
in the original declaration, method, dependency and ownership ledgers while their
frozen source fields remain unchanged. 101 former Store method rows now have actual
closeout policy, host evidence, coordinator, lifecycle or private-function owners. The original 908 method rows
retain frozen source evidence and current identities; 807 methods remain on Store. This count does not establish
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
from the governing spec in `05d68ed95bc5b154df6be54f0279bc18856eab96`. Twenty current compatibility/control cases identify actual
consumers, existing regression protection and retirement conditions. Remaining
legacy/schema/control references are explicitly unreconciled; T15.5 stays active.

Frozen input/shared-record extraction is in `c737a56d85e5fb0532d20b18a3c56c9d86a21fb1`. Nine owner/name moves
and five value-receiver validations match after explicit call/receiver
normalization. The input/member policy split uses real InputSource operations
and has semantic integration evidence. The internal attempt validator is private.

Closing slice/proof policy is in `93708ffd89d4eee64d111ad71424a5fc3fcfed19`. ProofSource supplies scoped
retained reads, native guards and canonical observations. Closeout now owns
ordered slice assembly, receipt/schema/member/canonical checks and bounded
hash collection. Shared immutable delivery predicates, task digest and the
observation-required error have direct state consumers without aliases. Twelve
shared/pure declarations match after owner/name qualification, with the verified
completed-ticket literal normalized for the finalized-member predicate. The
nine Store policy extractions use real finite port operations; they are not
unchanged whole-method token claims. Host native journal guards, integration
history policy and final post-observation authority/input checks retain their
original owners. The port ledger now records 20 live contracts.

Invocation closing extraction is in `e7255f915be24ad6e3fecda8d9ed5a76ecc35f8c`. Its dependency graph moves
27 public records, execution selection, the terminal predicate and the two
owned/stop error identities to their actual state owner. Thirty-six owner/name
comparisons and four explicit public-checkpoint/owner-root comparisons match.
The remaining six policy extractions have finite guard/evidence/projection
contracts, not unchanged whole-method token claims. Actual runtime, application,
TUI and web consumers use the new definitions directly. No migration aliases
replace them, and the empty closing-frontier shell is removed. Closing status aggregation also has an actual stage owner, as recorded below;
final Store API assessment remains active.

Closing status extraction is in `5f506a41dee1e172f8915f5712bbc9d6292a7bea`, with three exact declaration
comparisons and four frozen source images. The port map now records 22 actual
contracts. Final method inspection now records 788 body-specific decisions in
the existing 908-row method map: 687 guarded-persistence methods, 23 closing
policies, one host evidence adapter, five coordinator functions, one persistence
reader, 29 lifecycle owners and 42 coordinator operation mappings. The remaining
120 source rows stay explicitly pending final assessment.

Unused-receiver cleanup is in `3ed65239b85570183865b74e169bcbf5a62c1819`. Five complete declarations
match after receiver removal and the Markdown reader name; all 879 remaining
Store receivers had actual uses at that checkpoint. Source inspection adds explicit decisions for
contract/allocation reads, invocation/owner lookup, control receipts, process
records, publication journals and immutable upgrade recovery, plus every already
extracted closing method. None is accepted from its filename alone. The retained
upgrade snapshot case permits omitted publication/prepared-limit keys only at
their established defaults, with changed policy still rejected.

Closing lifecycle separation is in `b5b75ade5cbf21eb68ab3521b26545a20c5ee970`. Eleven full declarations
match after explicit coordinator, receiver, private host, port and caller
substitutions. Original admission and detached confirmation critical sections,
command ordering, partial results and cleanup/error joins are preserved. All
closing methods now have specific ownership decisions; tracker replacement,
registration and versioned-projection methods also retain individual source
bases. The v1/v2 tracker compatibility case preserves default schema, original
bytes, bound order-map requirements and hash-guarded replacement without granting
execution authority. Final repository-wide API acceptance remains active.

Task and stop lifecycle separation is in `0ce09547cb796604fb01ee41851a1e9a709cf02e`. Twenty-seven complete
production declarations and 85 test declarations match after explicit receiver,
coordinator, host and call substitutions. The original execution mutex, frozen
mode/grant, watched deadline, ordered gate, partial output and detached settlement
ordering remain. Eighty-six additional method decisions retain individual source
bases: actual task/manual/slice/stop owners are separate from guarded claim,
input, replay, current-image and complete result transactions. C15.14 records the
supported manual empty-worker and execute/reconcile distinction. Final API and
compatibility acceptance remain active.

Role/publication lifecycle separation is in `72253efe04a5f0ca41a2393b442f89c98aac8637`. Twenty-five production
and 135 test declaration streams match after explicit owner/call substitutions
and the exact MergerRepairPorts field pairing. Seventeen operations leave Store;
72 further transaction/authority/read methods retain individual source bases.
At that revision, review admission/restriction and preparation Git ownership
remained pending; the separation below resolves those specific responsibilities. C15.15 records
proposal-only execution versus trusted host repair acceptance, including declared
regeneration, pending validation and the absence of a publication grant.

Review-admission and preparation-Git separation is in `12eb09162d1a2dd2256bd5197b28e9be70c66312`.
Fifteen comparable production declarations and three test declarations match
after explicit owner/call normalization. Admission is a semantic extraction:
the trusted decision literal, waiting outcome and command error remain exact,
and only trusted admission reads the retained binding. Restriction keeps the
original human-only by-value override. Two Git methods leave Store; its remaining
review methods now name actual configuration reads and guarded admission writes.
Forty-seven new body-specific decisions also cover mutation receipt/allocation,
native record/submission evidence and integration-entry transactions. C15.8 binds
the actual stage policy to existing public restriction/upgrade contracts.

Publication/continuation separation is in `6e281666ac251b7f6cbdf997122dee845fff87bf`. Nineteen Store
methods now map to seventeen coordinator functions and one private loop; the
nineteenth source body was an exact poll-only forwarding method consolidated
into its sole caller. Twenty-two complete production declarations and 38 test
declarations match after explicit owner/call normalization, including that exact
call consolidation. Frozen bounds, membership refresh, operator preparation
retry, barriers, partial step/progress values and five-second detached settlement
retain original behavior. Seventy-one new source decisions include 52 concrete
persistence bases for publication checkpoints, submission intents, budget/agent
reservations and operator decisions. At that revision mixed lifecycle-control
ownership remained pending; the separation below resolves that responsibility. C15.16 preserves polling versus explicit reconciliation and
its retained observation contract/deadline, barrier and consumed budget.

Lifecycle-control separation is in `4d4f74804e857046b099ea3db91e87ec5d1258f6`. The original acceptance
timeout/withLock/cancel block is exact. Three production declaration streams
match after explicit acceptance-block recomposition and call normalization;
five changed test declarations match. This is a semantic responsibility split,
not a whole-method move: the source-method map follows the remaining guarded
Store acceptance, while the new source image follows the actual coordinator.
No frozen JSON, digest, partial acceptance, replay/zero-limit barrier or detached
settlement changes. Sixty-five new body-specific persistence decisions also
cover recovery, invocation admission, frontier, native dispatch/delivery evidence
and tracker order-source reads. Store still has 822 methods; count reduction is
not an ownership result. C15.17 records exact lifecycle request receipt/replay
and the independent post-acceptance continuation boundary.

Canonical PR ownership separation is in `85256e808df3824ddaf4b1eb230ff4a5ad8ae757`. All 16 affected
production declaration streams match after explicit owner/call normalization and
creation-intent helper recomposition. The remaining Store helper is individually
verified, and the original public timeout/lock ordering is exact. This is a
semantic responsibility split, not an unchanged whole-method move; the original
creation method row follows the guarded intent facet while its new source image
follows the actual provider host. No tests are edited. Seventy new source-specific
decisions cover these PR paths and retained plan-order, adoption, materialization
and receipt guards. Store has 821 methods; 324 source rows remain pending. C15.18
records retained uncertain creation and canonical identity without replacement.

Native observation separation is in `6afffc97a34bc31483d9f7f8acd4ff045e827bc3`. Eight production
declarations and 23 changed test declarations match after explicit receiver,
coordinator, host and call normalization. The original once/poll choice, 250ms
initial backoff capped at eight seconds, persisted deadline, three-minute IO
bound and five-second detached retention/error joins remain. Store retains all
original reconciliation and callback locks. One hundred five additional source
bases cover these actual owners plus native evidence/receipt/finalization,
transition/budget, coherent inspection/selection and coordinator registration.
Store has 816 methods; 219 source rows remain pending. C15.16 records the actual
observation owners and C15.19 resolves the explicit registration/read control.

Frontier/capacity observation and role selection are separated in
`e5fbc60981a76a7b25cf76e69b63ce6260140252`. All 16 affected production declaration streams match after
explicit receiver/private-owner/call normalization and two verified sole-Store-field
conversions; no tests are edited. Original
canonical proof order, current-record fences, partial target/proof/capacity
results, unknown slot/process exclusion, lease semantics, stable role priority,
human waits and target barriers remain exact. Ninety-nine new source-specific
decisions cover these owners and guarded repair/preparation, decisions, stop,
mutation, result, merger/guided admission and coordinator reads. Store has 807
methods; 120 source rows remain pending. Merger checkpoint Git seeding remains
pending separately and is not classified as persistence from its filename.
C15.20 resolves shared capacity/role priority and its observational boundary.

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

Frontier/capacity/role task checks passed all 34 selected real fixtures and six
focused races without skips or owned fixture leaks. They cover canonical and
receipt-backed frontier, role priority/prepared bounds/human waits, shared
process capacity and orphan/stop handling, closing publication barriers and
actual application operations. All 16 affected production streams match after
explicit owner/call normalization; no tests changed. Affected short tests,
signature/style checks pass. Strict lint first required two explicit private-owner
conversions in place of identical single-field literals. Both types retain exactly
the same sole Store pointer; normalized streams still match. Short compilation
and the same pinned strict lint command then passed. Real fixtures/races were not
repeated for those proven field-preserving conversions. Prior task failures remain
recorded below. Guided parity and final exact-head controller acceptance remain pending.

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(NativeSchedule.*|NativeRole(Priority|Pool|HumanWait|Blocked)|NativePrepared(Bound|Wait)|NativeProcessCapacity|NativeQueue(Capacity|DeadProcess|Closure)|NativeInspectionRole|WorkflowFrontier(Order|Reject|Recovery|Race)|NativeReceipt(Frontier|FrontierBlock)|WorkflowCapacity(RoleWait|Overlap|Orphan|StoppedWait)|NativeClosing(Gate|Evidence|Divergence|PublicationRace)|ServiceWorkflow(Frontier|Closing)|WorkflowMembershipFrontier)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(NativeRolePool|NativeProcessCapacity|NativeClosingPublicationRace|WorkflowCapacity(Overlap|RoleWait)|WorkflowFrontierRace)$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `go test -short ./internal/core/workflow` — after the private-owner conversions.
- `make check-go-style` — 9,410 declarations; zero exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/... ./scripts/goquality` — zero issues.
- `git diff --check`


Native observation task checks passed all 33 selected real fixtures. The initial
race command passed five cases; its closing-publication case failed the existing
three-second gate-start wait, with no data-race warning. That single case passed
an isolated uncached retry under unchanged fixture and timeout bounds. The first
failure is retained; its cause was not reproduced. No test or production change
was made between runs, and no passed race was repeated. Both commands cleaned
owned fixtures and reported no skips. They cover actual Git
application, canonical outcomes, polling/deadline/timeout, lost responses,
callback fences, dispatch crash recovery, receipt proof/descendants/divergence,
stop, fanout/restart and closing publication races. Affected short tests,
signature/style and pinned strict lint pass. Eight production and 23 test
streams retain their original tokens after explicit owner/call normalization.
These are task checks; guided parity and final exact-head controller acceptance
remain pending.

- `go test ./internal/core/workflow -run '^Test(NativeCanonical.*|NativeGit(Outcomes|Fence)|NativeReceipt.*|NativeStop(Rejected|Receipt)|NativeHostPublication|NativeClosing(Evidence|Divergence|PublicationRace)|NativeFaultCrashes)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(NativeCanonical(Poll|Fence|Lost)|NativeGitFence|NativeClosingPublicationRace|NativeReceiptFences)$' -count=1 -timeout=10m -json` — failed only the closing gate-start wait; five other races passed.
- `go test -race ./internal/core/workflow -run '^TestNativeClosingPublicationRace$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `make check-go-style` — 9,410 declarations; zero exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/... ./scripts/goquality` — zero issues.
- `git diff --check`


Canonical PR task checks passed all 19 selected real fixtures and four focused
races without skips or owned fixture leaks. They cover durable intent before
send, restart/reuse, uncertain creation, closed/ambiguous/fork/changed-head
observations, candidate/repair fences, crash recovery and actual application
consumers. Affected short tests and signature/style checks pass. Strict lint first
reported six import-group formatting errors; only imports were regrouped and the
same lint command then passed with zero issues. Executable declaration comparisons
still match; runtime fixtures were not repeated for import-only formatting.
Guided parity and final exact-head controller validation remain pending.

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(WorkflowSlicePR(Concurrent|Reuse|Observation|Block|Existing|Repair|Source|Crash)|WorkflowPublication(Revision|LateResponse|DuringIO)|WorkflowTicketClosedPR|WorkflowReview(Modes|Approval)|WorkflowFrontier(Reject|Recovery)|ServiceWorkflow(SlicePR|Review|Frontier))$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowSlicePR(Concurrent|Observation|Crash)|WorkflowPublicationDuringIO)$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `make check-go-style` — 9,410 declarations; zero exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/... ./scripts/goquality` — zero issues after import grouping.
- `git diff --check`


Lifecycle-control task checks passed all 20 selected real fixtures and four
focused races, with no skips or owned fixture leaks. They cover crash recovery,
receipt replay after later authority changes, lost acknowledgements, stale
revision/authorization fences, guided control and preparation retry. Affected
short tests, signature/style checks and pinned strict lint pass. The original
lock block and recomposed pipeline checks are distinct evidence. Guided parity
and final exact-head controller validation remain pending.

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(TelegramControl(Lifecycle(Journal|Fence|Ack)?|Journal|StopOwner|ClaimScope|Receipt(Missing|Retry|Redelivery|Authority|Projection))|NativeHostPreparationRetry|WorkflowContinuation(Bounds|Recovery)|WorkflowPreparationRetryAuthority|TelegramWorkContinuationGuard|GuidedUseCaseControl(Native|LostAck|Admission))$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(TelegramControlLifecycleJournal|TelegramControlLifecycleFence|TelegramControlStopOwner|NativeHostPreparationRetry)$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `make check-go-style`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `git diff --check`

Publication/continuation task checks passed all 27 selected real fixtures and six
focused races, with no test skips or owned fixture leaks. The process-only crash
helper was excluded explicitly; its owning crash fixture ran. Affected short
tests, style/signature checks and pinned strict lint pass. Production signatures
remain within five inputs. The 22 production/38 test declaration comparisons and
exact forwarding consolidation are distinct evidence. Guided parity and final
exact-head controller validation remain pending.

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(WorkflowPublication.*|WorkflowContinuation.*|NativeHost(Publication|PreparationRetry|Denials)|ServiceWorkflow(Publication.*|Continuation)|TelegramWorkContinuationGuard|TelegramControlLifecycle(Journal|Fence)|WorkflowPreparationRetryAuthority)$' -skip 'CrashProcess$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowPublicationDuringIO|WorkflowPublicationLateResponse|WorkflowPublicationCrash|WorkflowContinuationRecovery|TelegramControlLifecycleJournal|NativeHostPreparationRetry)$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `make check-go-style`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `git diff --check`

Review-admission/preparation task checks passed 16 selected real fixtures and four
focused races, with no test skips or owned fixture leaks. Affected short tests,
style/signature checks and pinned strict lint pass. The first lint attempt found
one missing blank line above the admission switch; after that formatting fix,
the unchanged lint command passed with zero issues. The semantic extraction and
15 production/three test comparisons are distinct evidence. These task checks do
not establish guided parity or final exact-head controller acceptance.

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(WorkflowReview(Modes|Approval|Policy|Restriction|ChangedInputs|DuringIO|Journal|Stop)|WorkflowPreparation(Methods|Guard|Journal)|NativeCandidate(Repair|Substitution|ReviewRepair)|ServiceWorkflow(Preparation|Review))$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowReviewDuringIO|WorkflowReviewConcurrent|WorkflowReviewStop|WorkflowPreparationJournal)$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `make check-go-style`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `git diff --check`

Task checks passed; final slice validation remains pending. Closeout extraction
passed 18 real tests and four focused races; closing exchanges passed 12 and two.
Capacity/continuation passed 27 real tests and five focused races, control/receipt
passed 24 and five, and amendment/frontier/upgrade passed 31 and five. These are
separate task selections, not a unique-test total or the full gate. The later
closing-history extraction passed 30 real tests and five focused races. All passed
without skips; owned-fixture cleanup reported no leaked descendants. The later
input/shared-record extraction passed 31 real tests and five focused races.
Affected core, authoring, app, TUI and web short tests and strict lint also pass.
The closing-proof extraction passed 42 real closing/finalization/frontier,
materialization and retained-receipt regressions plus six focused races, without
skips or leaked owned descendants. Workflow/app/checker short tests and strict
lint pass. An initial lint failure for exported comments/import grouping was
corrected before the successful strict check.

The invocation extraction passed 34 selected workflow regressions plus the
projection target-alias/history regression and seven focused races, without
skips or leaked owned descendants. These include real closing/continuation
fixtures and wire/retained-result checks. Affected broad core/app/TUI/web short
tests and strict lint pass.

The status extraction passed 36 selected closing, queue/query and inspection
regressions and six focused races, without skips or leaked owned descendants.
Affected workflow/app short tests, the production signature check and strict
lint pass. The final slice gate and shared guided parity remain pending.

The receiver cleanup passed 33 selected real adoption, amendment,
materialization, plan-order and record regressions and four focused recovery /
concurrency races, without skips or leaked owned descendants. Affected
workflow/app/checker short tests, unchanged production signature bounds and
strict lint pass.

The closing host change passed 31 selected workflow regressions plus one real
application closing fixture and six focused lifecycle races, without skips or
leaked owned descendants. Affected workflow/app/checker short tests, unchanged
signature bounds and strict lint pass. These are task checks; final slice
validation and guided parity remain pending.

Task/stop lifecycle changes passed 46 real workflow fixtures and 12 application
fixtures, plus 11 focused races, without skips or leaked owned descendants.
Affected short tests, unchanged signature bounds and strict lint pass. These are
task checks; guided parity and exact-head slice/controller acceptance remain open.

The original combined role/publication selection hit its 15-minute package cap
after 97 fixtures passed; it recorded no assertion failure and cleaned its owned
descendants. That package failure is retained. Only the ten unfinished submission
fixtures were selected again under the unchanged 15-minute limit, and all passed.
Across those commands, all 107 selected fixtures (104 workflow and 3
application) passed without test skips. The 11 focused races, affected short
tests, unchanged signature bounds and strict lint pass; owned fixtures are clean.
These are task checks; guided parity and exact-head controller validation remain
pending.

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

- `go test ./internal/core/workflow -run '^Test(WorkflowClosing.*|NativeClosing.*|WorkflowPlanAdoption(Rejects|Restart)|WorkflowMaterialization(Rejects|Recovery)|WorkflowRoundTrip|WorkflowRecord|WorkflowHierarchy|WorkflowIndex|QualityRecordWire)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery|WorkflowMaterializationConcurrent|QualityRecordWire)$' -count=1 -timeout=10m -json`

- `go test ./internal/core/workflow -run '^Test(WorkflowClosing.*|NativeClosing.*|WorkflowParallelClosingRecovery|WorkflowFinalization.*|WorkflowFrontier.*|NativeReceipt(Frontier|FrontierBlock|Proof|Legacy)|WorkflowMaterialization(Recovery|Rejects)|QualityRecordWire)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery|WorkflowFinalizationConcurrent|WorkflowFrontierRace|QualityRecordWire)$' -count=1 -timeout=10m -json`

- `go test -p=1 ./internal/core/workflow ./internal/core/workflow/closeout -run '^Test(WorkflowClosing.*|NativeClosing.*|WorkflowParallelClosingRecovery|WorkflowContinuation.*|WorkflowResult.*|NativeResult.*|WorkflowFrontier(Order|Recovery)|WorkflowRecord|WorkflowRoundTrip|QualityRecordWire|ProjectionTarget)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery|WorkflowContinuationRecovery|WorkflowFrontierRace|WorkflowControlCAS|QualityRecordWire)$' -count=1 -timeout=10m -json`

- `go test ./internal/core/workflow -run '^Test(NativeQueue.*|NativeInspection.*|NativeDecisionInspection|WorkflowClosing.*|NativeClosing.*|WorkflowParallelClosingRecovery)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(NativeQueue(Capacity|Closure|NoReplay)|WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery)$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/...`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/...`

- `go test ./internal/core/workflow -run '^Test(WorkflowPlanAdoption.*|WorkflowAmendment.*|WorkflowMaterialization.*|WorkflowPlanOrder.*|WorkflowLooseMarkdown|WorkflowRoundTrip|WorkflowAdoption|NativeRecord.*|QualityRecordWire)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowMaterializationConcurrent|WorkflowPlanAdoptionRestart|WorkflowAmendmentRecovery|WorkflowPlanOrderRecovery)$' -count=1 -timeout=10m -json`
- `go test -short ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app/... ./scripts/goquality`

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(WorkflowClosing.*|NativeClosing.*|WorkflowParallelClosingRecovery|WorkflowContinuation.*|WorkflowCapacityRoleWait|NativeQueueClosure|ServiceWorkflowClosing)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowClosingStop|NativeClosingPublicationRace|WorkflowParallelClosingRecovery|WorkflowContinuationRecovery|WorkflowCapacityRoleWait|NativeQueueClosure)$' -count=1 -timeout=10m -json`

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(WorkflowTask.*|WorkflowManual.*|WorkflowSlice(Aggregate|Prefix|Retry|Repair|Changed|GateMutation|Stop|Waiver|Partial|Budget|Fence|Journal|RepairJournal)|WorkflowStop.*|TelegramControlStopOwner|WorkflowClosingStop|WorkflowContinuationRecovery|WorkflowCapacityRoleWait|NativeQueueClosure|ServiceWorkflow.*Task|DaemonWorkflow.*)$' -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowTaskJournal|WorkflowTaskProcess|WorkflowManualStop|WorkflowManualJournal|WorkflowManualFencing|WorkflowSliceStop|WorkflowSliceJournal|WorkflowStop|TelegramControlStopOwner|WorkflowClosingStop|WorkflowCapacityRoleWait)$' -count=1 -timeout=10m -json`

- `go test -p=1 ./internal/core/workflow ./internal/app -run '^Test(Workflow(Preparation|Review|Submission).*|Native(Merger|Preparation|Review|Dispatch|Candidate).*|NativeHost(Publication|PreparationRetry)|NativeAgentReview|WorkflowContinuation(Recovery|PreparationRetryAuthority)|ServiceWorkflow(Preparation|Review|Submission).*)$' -skip 'CrashProcess$' -count=1 -timeout=15m -json` — package timeout after 97 passed fixtures; unfinished cases passed in the residual selection below.
- `go test ./internal/core/workflow -run ^(TestWorkflowSubmissionAtomicGuard|TestWorkflowSubmissionExclusion|TestWorkflowSubmissionProjectionLoss|TestWorkflowSubmissionStop|TestWorkflowSubmissionBudget|TestWorkflowSubmissionPolicy|TestWorkflowSubmissionCorruption|TestWorkflowSubmissionJournal|TestWorkflowSubmissionMethods|TestWorkflowSubmissionLostResponse)$ -count=1 -timeout=15m -json`
- `go test -race ./internal/core/workflow -run '^Test(WorkflowPreparationJournal|WorkflowReviewJournal|WorkflowReviewConcurrent|WorkflowReviewStop|WorkflowSubmissionLostResponse|WorkflowSubmissionStop|NativeDispatchConcurrent|NativePreparationStop|NativeMergerStop|NativeMergerHostRegeneration|NativeCandidateReviewRepair)$' -count=1 -timeout=10m -json`

- `go test -short ./internal/core/... ./internal/app/... ./internal/tui/... ./internal/webui/... ./scripts/goquality`
- `go test ./internal/core/workflow/closeout -count=1`
- `go test -short ./internal/core/workflow/... ./internal/app ./scripts/goquality`
- `make check-go-style` — 9,407 production declarations, zero exact debts and one external exception.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./internal/app ./scripts/goquality` — zero issues.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/workflow/... ./scripts/goquality` — zero issues after the final private-result conversions.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run --timeout=5m ./internal/core/... ./internal/app/... ./internal/tui/... ./internal/webui/... ./scripts/goquality` — zero issues after native input/record extraction.
- `git diff --check`

The vocabulary-only update checked all relative links, six definitions, ten case
rows and every named regression declaration against the current source. T15
tracker columns and whitespace checks pass. Runtime checks were not repeated for
that documentation change.

Closing-proof ledger reconciliation checks 4,182 current declaration hashes,
all 908 original method mappings, four signature ledgers, 21 new frozen source
preimages, 55 commit objects and 20 actual ports. Relative links, named regression
declarations, eleven compatibility rows, T15 table columns and whitespace pass.
Runtime checks are not repeated for this documentation-only update.

Invocation ledger reconciliation checks 4,228 current declaration hashes, all
908 original method mappings, four signature ledgers, 46 new frozen source
preimages, 57 commit objects and 21 actual ports. Relative links, named regression
declarations, T15 columns and whitespace pass. Source fields and original
introduction commits remain frozen. Runtime checks are not repeated for the
metadata-only update.

Receiver and ownership ledger reconciliation checks 4,237 current declaration
hashes, all 908 original method mappings, four signature ledgers, five new frozen
source preimages, 61 commit objects, 22 unchanged actual ports and 104 explicit
ownership decisions. Relative links, named regression declarations, twelve
compatibility rows, T15 columns and whitespace pass. Original source fields and
introduction commits stay frozen; runtime checks are not repeated for this
metadata-only update.

Closing host ledger reconciliation checks 4,248 current declaration hashes,
all 908 original method mappings, four signature ledgers, eleven new frozen
source preimages, 63 commit objects, 22 actual ports and 156 explicit ownership
decisions. Source images retain both their full revision and declaration ID,
keeping the earlier gate extraction distinct from this host extraction. Relative
links, named regression declarations, thirteen compatibility rows, T15 columns
and whitespace pass. Runtime checks are not repeated for this metadata update.

Task/stop ledger reconciliation checks 4,275 current declaration hashes, all
908 original method mappings, four signature ledgers, 27 new frozen source
preimages, 65 commit objects and 22 actual ports. It records 242 source-specific
ownership decisions with 666 pending rows. Frozen source/baseline/introduction
fields remain unchanged across 33 ledgers. Relative links, named regressions,
14 compatibility cases, T15 columns and whitespace pass. Runtime checks are not
repeated for this metadata-only update.

Role/publication ledger reconciliation verifies 4,300 current declaration
hashes, all 908 original method mappings, four signature ledgers, 25 new frozen
source preimages, 67 commit objects and 22 actual ports. It records 331 specific
ownership decisions with 577 pending rows. Frozen source/baseline/introduction
fields remain unchanged across 33 ledgers. Relative links, named regressions,
15 compatibility cases, T15 columns and whitespace pass. Metadata-only changes
do not repeat runtime checks. The combined package timeout and successful
unfinished-case recovery remain explicit in the validation evidence above.

Review-admission ledger reconciliation verifies 4,316 current declaration hashes,
all 908 original method mappings, four signature ledgers, 16 new frozen source
preimages, 69 commit objects and 22 actual ports. It records 378 body-specific
ownership decisions with 530 pending rows. Frozen source/baseline/introduction
fields remain unchanged across 33 ledgers. Current owner declarations, relative
links, named regressions, 15 compatibility cases, T15 columns and whitespace
pass. Runtime checks are not repeated for this metadata-only update.

Publication/continuation ledger reconciliation verifies 4,339 current declaration
hashes, all 908 original method mappings, four signature ledgers, 23 new frozen
source preimages, 71 commit objects and 22 actual ports. It records 449 specific
ownership decisions with 459 pending rows. Frozen source/baseline/introduction
fields remain unchanged across 33 ledgers. Current owner declarations, relative
links, named regressions, 16 compatibility cases, T15 columns and whitespace pass.
Runtime checks are not repeated for this metadata-only update.

Lifecycle-control ledger reconciliation verifies 4,342 current declaration
hashes, all 908 original method mappings, four signature ledgers, three new frozen
source preimages, 73 commit objects and 22 actual ports. It records 514 specific
ownership decisions with 394 pending rows. Frozen source/baseline/introduction
fields remain unchanged across 33 ledgers. Current owner declarations, relative
links, named regressions, 17 compatibility cases, T15 columns and whitespace pass.
Runtime checks are not repeated for this metadata-only update.

Canonical PR ledger reconciliation verifies 4,358 current declaration hashes,
all 908 original method mappings, four signature ledgers, 16 new frozen source
preimages, 75 commit objects and 22 actual ports. It records 584 specific ownership
decisions with 324 pending rows. Frozen source/baseline/introduction fields remain
unchanged across 33 ledgers. Current owner declarations, relative links, named
regressions, 18 compatibility cases, T15 columns and whitespace pass. Runtime
checks are not repeated for this metadata-only update.

Native observation ledger reconciliation verifies 4,366 current declaration
hashes, all 908 original method mappings, four signature ledgers, eight new frozen
source preimages, 77 commit objects and 22 actual ports. It records 689 specific
ownership decisions with 219 pending rows. Frozen source/baseline/introduction
fields remain unchanged across 33 ledgers. Current owner declarations, relative
links, named regressions, 19 compatibility cases, T15 columns and whitespace pass.
Runtime checks are not repeated for this metadata-only update.

Frontier/capacity/role ledger reconciliation verifies 4,382 current declaration
hashes, all 908 original method mappings, four signature ledgers, 16 new frozen
source preimages, 79 commit objects and 22 actual ports. It records 788 specific
ownership decisions with 120 pending rows. Frozen source/baseline/introduction
fields remain unchanged across 33 ledgers. Current owner declarations, relative
links, named regressions, 20 compatibility cases, T15 columns and whitespace pass.
Runtime checks are not repeated for this metadata-only update.

## Risks and Follow-ups

Remaining method-level Store policy ownership, final bridge removal and naming/compatibility reconciliation remain active
Slice 15 work. Shared guided parity, final ordered exact-head controller
validation, PR publication and PR-reference follow-up are pending. No final
slice acceptance is claimed by these task checks.
