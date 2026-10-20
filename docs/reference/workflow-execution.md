# Workflow Execution

`WorkflowStore` and the central application service share the ticket-native
invocation, claim, execution and recovery operations below. Selection follows
[Workflow selection](workflow-selection.md). Files own workflow state; tracker
edits and SQLite projections do not authorize execution or certify success.

## Invocation and ownership

| Operation | Contract |
| --- | --- |
| `StartWorkflow` | Freeze an explicit selection, target and execution policy. Equivalent requests attach to the retained invocation. Overlapping retained task/hierarchy ownership rejects. |
| `ReadWorkflow` | Return invocation status, cursor, attempts, lane, process identities and retained evidence without private credentials. |
| `InspectWorkflow` | Read current invocation, publication policy and task records without replaying pending transitions or reconciling ownership. |
| `ClaimWorkflow` | Claim the next eligible task under repository worker capacity and exclusive task/hierarchy ownership; journal a new fenced attempt. |
| `CheckWorkflowClaim` | Check the current private credential, generation, lease and frozen record images. |
| `CheckpointWorkflow` | Yield an unexpired, quiescent claim without validating the task; retain its input, lane, attempt and checkpoint. |
| `ExecuteWorkflowTask` | Freeze worker inputs, supervise its argv in the isolated lane, run the frozen host gate and journal the guarded task result. |
| `ContinueWorkflow` | Run bounded transitions on an existing invocation; preserve its frozen selection, policy, task proof, PR and delivery checkpoints. |

The caller retains the returned `*WorkflowClaim`; status cannot reconstruct it.
`WorkflowClaim.Close` releases only that host's capacity descriptors after its
worker group becomes quiescent. It does not record a result or checkpoint and
cannot release an inherited descriptor held by a live descendant.
The token is excluded from JSON and only its hash is persisted. One claim covers
the task and its owning hierarchy. Native and legacy starts share the Git common
directory's process-held reservations across registered linked checkouts.
Successful task validation leaves task status `in_progress` and phase
`validated`, with integration pending. An exact grouped-ticket invocation
completes only that task. A whole slice or independent ticket stops at
`waiting_integration`; local validation does not mark work `done`.

## Worker capacity

The default is one active worker. Opt in to independent parallel lanes in
`.chr/default/config.yml`:

```yaml
work:
  max_active_workers: 2
  ready_queue_capacity: 64
```

`max_active_workers` accepts 1–8; `ready_queue_capacity` accepts 1–1024 and
defaults to 64 unfinished integration lanes. Invalid values reject admission.
Namespaces and registered linked host checkouts share one Git-common-directory
pool and must resolve the same configuration. Retained worker-lane snapshots
cannot change its limits. Reducing capacity counts all existing reservations;
it does not stop live work or omit slots above the new limit.

Independent native workers use distinct process-held slots and isolated lanes.
Tasks within a slice, and the frontier of the same ordered set or use case,
remain serial. Publication has separate exclusion for each canonical remote
target, including uncertain submissions. A review or publication wait holds no
worker slot. `waiting_capacity` retains `capacity_resume_status`, task evidence
and consumed budgets. An explicit claim, role retry or continuation rechecks
capacity; repeating an unchanged wait does not create an attempt or checkpoint.
Pause/resume retains the prior phase, and cancellation remains terminal.

Native claims do not create `agent-operation.lock`. Legacy runners and adoption
retain exclusive access across the whole pool. A retained legacy or historical
coarse file also blocks new native claims until its own release or matching
fenced recovery. Native admission never removes an unknown or replaced file.
This coordination requires the registered local host and Git common directory;
separate clones or hosts cannot form another pool from copied records.

## Coordinator registration

Explicit native creation, materialization or adoption journals
`ops/<namespace>/coordinator.json` with the first authority records. The marker
contains a `chr-coordinator/v1` schema and an opaque ID. The matching registration
in the Git common directory is host-owned. It binds that ID to the machine and
the actual common directory, without publishing private paths or host identity.
Linked worktrees and namespaces share this registration. Moving the original
directory preserves its identity; cloning or copying it does not grant ownership.

Before initial execution, commit the coordinator marker to the configured
canonical target. Native execution and candidate acceptance require that target
to retain the registered ID. Markers for other namespaces must agree. A work lane
cannot bootstrap a missing target registration by executing or publishing itself.
Unsupported separate clones or hosts return `ErrWorkflowCoordinator`. Existing
unbound authority, a missing host registration or conflicting markers require
explicit reconciliation; work starts cannot silently create another owner.

Task and hierarchy ownership are rechecked before each reservation and fenced
result application. Worker-owned lane copies do not redefine the host
coordinator; their committed marker is checked when accepting their own result.
Legacy Git runners validate the same repository capacity
policy and reject copied registered authority. Their conservative exclusion
continues to apply across all native slots. Web and daemon code without native
work-start controls cannot create execution ownership; shared controls use the
application operation service.

Invocation admission can retain an immutable target before its objects are
hydrated. It verifies the local registered owner and grants no worker authority.
The execution and closure boundaries verify the available canonical target.

## Frozen inputs and evidence

The retained lane uses a deterministic `refs/heads/chr/workflow/<identity>`
branch and isolated worktree. Each next task descends from the accepted lane
head. Dirty, moved or unregistered lanes require reconciliation.

`WorkflowTaskInput` uses schema `chr-task-input/v1`. Its `execution_mode` is
`automatic`, `manual` or host-only `aggregate`. It binds invocation/attempt IDs,
generation, task revision, source and contract hashes, body, execution contract, target, scope,
lane, worker argv and prerequisite record/evidence images. Workers receive this
JSON on stdin, their input-file location through `CHR_WORKFLOW_INPUT`, and the
lane as cwd. They do not receive the owner token.

New executor inputs include host-owned `instructions` to implement only the
bound task, run its required validation, commit the task changes on the assigned
lane branch and leave a clean worktree before returning success. Pushing,
merging and changing workflow state remain host operations. These instructions
are retained with the frozen input and its hash; aggregate validation and
review inputs do not receive the executor's commit instruction.

Host validation runs the contract's ordered commands or records its explicit
waiver. Success binds the clean branch, full commit and tree to
`chr-task-validation/v1` evidence. The result also retains execution mode, input
reference/hash, attempt/generation, bounded worker/gate output and finish time.
Task evidence, expected images and cursor advance in one replayable journal.
Unchanged validated task prefixes reuse their lane and proof rather than rerun
workers. Changed inputs or incompatible journal images reject.

## Aggregate slice candidates

| Operation | Contract |
| --- | --- |
| `ExecuteWorkflowSlice` | Run remaining tasks in order on their retained cumulative lane, then host-validate one complete slice and stop at its integration barrier. |
| `ValidateWorkflowSlice` | Run only the aggregate acceptance commands or approved waiver. Reuse an unchanged accepted candidate without another attempt or gate. |
| `ReadWorkflowSliceCandidate` | Recheck retained candidate, input, contract, ordered task proof and actual clean Git image. |
| `RepairWorkflowSlice` | Explicitly invalidate the affected task and its ordered successors, clear aggregate acceptance and resume from that task on the same branch. |

Aggregate validation requires every actual task in the slice to be present in
the frozen extent and host-validated. Exact grouped-ticket and loose-ticket
requests cannot produce partial-slice candidates. Ordered slice execution stops
before starting another slice, pending integration of the current slice.

An aggregate input retains the set and slice identity, immutable contract
reference/hash, ordered task evidence references/hashes, current task images,
prerequisite proofs and accepted target/lane. The host freezes the full clean
commit and tree before the first command and verifies them before and after
each command and before journaling success. A gate that commits or dirties the
lane cannot establish acceptance.

`chr-slice-candidate/v1` retains these bindings and host observations in the
aggregate attempt's immutable `result.json`. Its result uses
`chr-slice-validation/v1`, which cannot serve as task-validation proof.
Invocation `candidate_ref` and `candidate_hash` point only to an accepted host
candidate; explicit empty values clear the binding during repair. Aggregate
success leaves task phase `validated`, status `in_progress` and integration
pending. It does not create delivery proof or mark tasks `done`.

Aggregate failure preserves validated task results and the cumulative lane.
An unchanged retry runs only the aggregate gate. Repair clears affected evidence
references without deleting historical result files. Ordered successor tasks
are affected because their inputs depend on earlier task results; unchanged
prefix evidence remains reusable. Changed input, proof, membership or branch
images require reconciliation rather than silent reuse.

Attempts with purpose `slice_validation` share task leases, fenced ownership,
supervised processes, stop intent and recovery. Pausing or recovering an
aggregate attempt preserves validated tasks and resumes at
`waiting_integration`. Journaled repairs retain a separate `slice_repair`
attempt with outcome `repaired`, allowing lock recovery after host loss. Task,
aggregate and repair attempts consume separate per-anchor budgets under the
same configured attempt limit and total invocation bound. None resets earlier
worker attempts.

## Independent ticket candidates

`PrepareWorkflowTicketCandidate` binds one independent ticket's accepted host
task result to its own integration candidate. The central application service
exposes the same operation. It requires a validated clean lane and unchanged
task/input/result images. It executes no additional source gate or worker.

`chr-ticket-candidate/v1` stores a `ticket` input with `unit.kind: ticket`, the
exact key, execution-contract hash and one task proof. The candidate has no
`slice` input, set ID, aggregate contract reference or fabricated hierarchy.
Its immutable path is
`run/<execution-id>/extension/chr/ticket-candidate/<candidate-id>.json`; the ID
binds the invocation and accepted evidence hashes. Candidate and invocation
bindings share one journal. An unchanged retry reuses them without another
attempt, including after interruption of either replacement.

The existing PR, preparation, review, repair, publication and finalization
operations accept this actual ticket unit. Their established evidence envelopes
retain the actual `ticket` unit reference. Combined-result validation runs the
ticket's frozen validation contract against the method-specific prepared target.
Human, agentic and trusted policies apply with the same exact candidate/base
guards, bounded repair budgets, uncertainty handling and canonical observation.
Repair affects only this ticket and retains its branch and canonical PR.

Source validation and PR creation leave integration pending. Only verified
canonical method/history proof and journaled receipt finalization move the
independent task to `done` with phase `integrated`. Finalization writes its task
and invocation images without a set descriptor. An exact ticket invocation then
completes with outcome `ticket_integrated`; loose work retains its frontier for
subsequent queue advancement. Closed/unmerged PRs require explicit reconciliation.
Grouped exact-ticket requests still finish at task validation and cannot publish
a partial slice.

`ContinueWorkflow` can execute an exact ticket or prepare an independent ticket
candidate in separate bounded steps. Its selector, filters, policy and original
request identity remain frozen. Accepted task results are reused on restart.

## Loose queue reselection

Independent loose tickets share the candidate and canonical delivery lifecycle.
`AdvanceWorkflowFrontier` retains each actual ticket unit in an immutable
`chr-unit-frontier/v1` image, clears its active bindings and hydrates the next
execution target from canonical Git. The original selection target and request
hash remain unchanged. Loose work creates no set or set-closing gate.

`ReconcileWorkflowMembership` selects again when the current lane has cleared.
The receipt-confirmed prefix stays in place. Pending admissions are filtered and
ordered using their current native classification, the frozen AND filters and
rank thresholds, and the requested rank/time/key order. A live ticket cannot be
preempted. Prerequisite readiness is observed against the hydrated target before
claiming its successor; dependencies do not expand the selection.

The invocation's `loose_queue` retains admission-order keys separately from the
current pending order. Its `cursor`, `eligible` count and per-key `excluded`
reasons describe the last quiescent selection. Excluded admissions remain in
`members`; they do not advance the validated or receipt-confirmed cursors. When
no eligible pending members remain, the invocation completes with
`loose_exhausted`, preserving its actual deliveries and excluded admissions.

A revision/source-guarded `WorkflowStore.Write` can update an inactive loose
member's classification or pending status and its expected invocation image in
one journal. It proves that the execution fields, contract and description are
unchanged, including native Markdown fields and bodies. Changes to a live ticket,
execution-owned fields, descriptions or raw unguarded files still conflict with
the frozen image. Admission revisions and source hashes remain retained.

Later matching ready independent tickets are counted under `excluded.later` by
default. The explicitly frozen `IncludeLaterTickets` policy can instead append
their admissions and order them after the confirmed prefix. Slice-owned work,
nonmatching or unready arrivals and terminal invocations cannot extend that
queue. Record/member bounds remain 4096; journal and evidence images retain the
1 MiB bound.

## Loose invocation stops and resume

A failed worker or task gate leaves its invocation paused, with the same lane,
cursor and retained attempt evidence. Direct claims and bounded continuation
cannot execute another task. `ResumeWorkflow` permits another attempt on that
ticket without resetting its finite worker budget. Cooperative checkpoints
retain their established fenced replacement behavior.
Existing exact-ticket manual claims can retry their own failed task with its
retained lane and consumed attempt budget.

A prerequisite refusal persists a paused invocation with outcome
`prerequisite_blocked` and the typed dependency, reason and next-action owner.
Queue reselection stays dormant while paused, so a pending rank change cannot
bypass the blocked cursor. Neither a retry nor resume adds the prerequisite to
the invocation. Resume observes readiness and integration proof again; status
text alone cannot clear the stop.

When an outside prerequisite integrated after the retained target, an inactive
independent cursor can resume with trusted `WorkflowResumePorts`. The central
service supplies its canonical provider, Git observer and configured integration
branch. Hydration must agree with the provider's current branch and prove that
the previous execution target is an ancestor. The exact ticket's prerequisites
must be proven on that target before the invocation changes. A retained lane or
a changed target branch cannot use this refresh.

`dependency_targets` retains at most 128 `chr-dependency-target/v1` checkpoints.
Each binds the invocation, ticket, cursor, previous and observed target, branch,
original blocker and immutable receipt hashes. Its image at
`run/<execution-id>/extension/chr/dependency-target/<sequence>.json` and the
resumed invocation share one journal. Later claims verify those images and
history. Resume preserves the original selection target, filters, admitted
members, confirmed cursor and policy; subsequent delivery advances its usual
frontier from the refreshed execution base.

Initial admission reads all bounded pages rather than truncating at 128 query
members. Each continuation call still obeys its frozen step limit. Equivalent
unfinished requests attach; a reused key with different policy conflicts.
Completed and cancelled invocations remain terminal on retry, including with
accepted later-ticket policy. Exhaustion never subscribes to later arrivals.

## Compatibility entrypoint authority

`work.state_authority` accepts `tickets` and `tracker_compat`. Absence retains
the existing transitional behavior. Versioned default activation is separate
from selecting explicit authority for these entrypoints.

| Entrypoint | `tickets` | `tracker_compat` |
| --- | --- | --- |
| `work run-all`, ready-work service operations | Start or attach to a finite loose invocation with the requested status, AND filters and order. Report excluded slice-owned work. | Select the existing unadopted legacy queue and report its authority and scope. |
| `work ticket`, `ticket start`, selected-ticket service operations | Select one exact adopted task. A grouped task validates without starting its parent extent; an independent task follows the common candidate/PR lifecycle. | Retain legacy single-ticket execution; refuse adopted keys and projections before a worker starts. |
| `work run-all --dry-run`, `ticket start --dry-run` | Read the common selection and display authority, extent and exclusions without allocating a run, worker or PR. | Retain the existing legacy preview and display transitional scope. |

Ticket-mode `--max-iterations` supplies a per-call checkpoint-step limit of
`1..128`; zero uses the retained invocation's limit. It counts accepted task,
candidate and delivery transitions rather than promising completed tickets in
one call. The ready-work service uses one step. Waiting review, failed gates and
stop intent retain the same checkpoint and prevent successor execution.

The entrypoint reads `work.canonical_remote` as a credential-free public HTTPS
repository identity and observes `work.integration_branch` through the trusted
target port. SSH aliases and local transports must use a configured public HTTPS canonical
remote for this entrypoint; no persistent repository URL is guessed. Object hydration leaves checkout
branches and files unchanged. An equivalent unfinished entrypoint reuses its
original target and request; a newly observed target does not replace frozen
inputs or create a second worker. Policy or selector changes do not attach as
equivalent work. The explicit invocation API retains its exact retry-key rules.

The native path uses the host-supervised JSON worker in the owned lane and the
frozen task gate. It does not initialize the legacy runtime. Both modes retain
the shared repository reservations. Legacy selection partitions actual source
images before preflight; a second guard under the operation lock prevents an
adoption between preview and worker start from granting two owners.

`PublishWorkflowSource` is the common continuation boundary before PR binding.
It checks the immutable candidate before and after trusted host IO and publishes
only its owned `refs/heads/chr/workflow/...` source. A new source uses an empty
lease; a repaired bound candidate uses its retained previous source head. A
different remote head requires reconciliation. Retrying an already published
exact source performs observation only, including recovery after a lost push
response. Uncertain PR creation continues by observation without a source
rewrite. Source publication neither creates a delivery receipt nor moves the
integration target. The central service supplies this source port to the shared
continuation operation; adapters do not schedule a separate PR lifecycle.

## Canonical slice pull requests

`EnsureWorkflowSlicePullRequest` accepts a retained invocation, target branch
and bounded initial title/body. The shared service uses its configured Forgejo
or GitHub adapter. The provider must observe the validated source commit on the
same canonical repository branch before creation; an unpublished or different
source blocks creation.

The host records `chr-slice-pr/v1` intent before provider IO and records
`creating` before the create call. It searches all PR states for the source
branch and adopts only one matching repository, source branch/commit and target
branch. Once bound, retries read that same PR URL. A closed unmerged PR,
ambiguous identity, fork head or changed candidate blocks the invocation and
retains its immutable provider observation for explicit reconciliation.

A timeout, malformed create response or host crash retains uncertain creation.
Retries observe the original source branch without issuing another create. If
no PR can yet be observed, `ErrWorkflowAwaitingObservation` preserves the
uncertain intent. Worker expiry, restart and ordinary retries cannot clear it.

The set-owned slice stores one branch, candidate reference and PR URL. All its
tasks enter `waiting_slice` / `ready_for_review` in the same replayable journal;
their task evidence remains valid and they share the set-owned PR. A provider
merge observation enters `waiting_integration` for host verification and never
marks a task `done`.

Explicit repair retains the branch and PR identity, clears aggregate review
readiness, preserves unchanged prefix evidence and invalidates the affected
suffix. A newly validated candidate rebinds only that same open PR after its
source commit is observed. Uncertain or blocked PR creation requires
reconciliation before repair.

Provider IO has a one-minute deadline and holds repository mutation exclusion
without a worker capacity reservation. Discovery stops only after an empty
page, or after proving ambiguity; it rejects incomplete pagination, more than
4096 PR records or 128 pages. Immutable observations remain within the 1 MiB
file bound; at most 128 distinct observations are retained per binding, with
identical repeats reusing their observation. Titles allow at most 256 bytes,
bodies 16 KiB and branch names 256 bytes.

## Combined-result preparation

Preparation imports result objects without rewriting the original worker's
`FETCH_HEAD`. Its owned validation lane leaves worker and user checkout heads,
indexes, staged/unstaged content and original evidence unchanged.

New merge and squash preparation retains `WorkflowIntegrationProof.Provenance` (`provenance` in JSON)
with schema `chr-integration-provenance/v1`. It records the object format, every
original commit after the frozen origin in topological order, original authors and Git dates, exact
message hashes, parent and tree identities, and the prepared commit's author
and committer metadata. Squash uses the host as author and preserves source
attribution through the complete original manifest. Its attribution policy is
`host_author_with_original_source_manifest`.

`combined_diff_hash` covers the deterministic binary-capable diff from B to P.
The provenance joins the existing history hash and candidate binding. Host
preparation and publication guards re-read actual Git objects and reject changed
attribution or diff evidence even if its enclosing hash is recomputed. The
original source manifest must contain the same ordered original IDs. Limits
remain 4096 commits and 1 MiB per evidence file and Git output. Historical proofs
without provenance retain their original hash representation and read rules.

Rebase uses the attribution policy `preserve_source_authorship` and retains
one prepared commit per original linear source commit. Authors, author dates,
message bytes and encoding remain unchanged; rewritten commits record the host
as committer and exclude original signature headers. The `replays` list binds
each original/prepared pair to `nonempty`, `source_empty` or `already_applied`.
Empty replays remain in order with their original messages. Merge-containing
sources and wholly empty integration units require explicit operator resolution.

Declared `integration_fixes` append separate attributed commits after the
original replay sequence. Each retains an actionable finding, exact parent and
tree, author and committer metadata. The complete sequence and finding enter the
candidate/history hash and must pass combined validation and review. Missing
repair declarations, cosmetic findings, altered replay metadata and false empty
decisions fail host verification before publication.

New host preparation also records `preparation` metadata: initial and effective
committer dates, identity, no-op collision count, and required signature policy.
A result equal to the original source is rebuilt with the recorded committer
timestamp advanced by one second, at most three times. Authors, messages, trees
and replay count remain unchanged; no unrelated commit is inserted.

`WorkflowPublicationCapability.required_signatures` and `signing_profile` bind
the exact trusted signing policy. Required signatures use an explicitly supplied
`WorkflowSSHCommitSigning` host key through `WithSSHCommitSigning`; its public
Ed25519 key and principal identify the certified profile. Missing or mismatched
host capability blocks preparation before validation. Every rewritten object,
including declared integration fixes, is signed in the Git namespace. Host
verification checks the actual signature against the exact public policy and
records the signature hash per prepared commit. Private key paths and secrets
are excluded from candidates and worker inputs. Unsigned provider profiles do
not acquire signing authorization from a local key alone.

Native publication observation re-proves the retained history version through
`WorkflowPreparedIntegrationGit.VerifyPreparedWorkflowIntegration`. The trusted
host fetches the original source and canonical target, verifies every mapped
commit and declared fix, and compares the retained history hash. Rich proofs
require this port; legacy hashes remain unchanged rather than acquiring new
metadata during observation.

| Operation | Contract |
| --- | --- |
| `PrepareWorkflowSliceIntegration` | Resolve the current method and atomic provider capability, prepare an isolated combined result, and run the frozen slice gate against that exact clean lane. |
| `CheckWorkflowPublicationGuard` | Recheck combined validation, exact review, policy, canonical PR/source/base and current protection; return the immutable guard required by a publisher. |

`WorkflowPublicationGuardProvider` extends canonical PR observation with a trusted
capability for the exact repository, PR and configured method. The capability
must guarantee both `atomic_source_base` and `exact_result` in the publication
transaction. A source-only merge API, an earlier target observation or a local
operation lock cannot establish that guarantee. The existing Forgejo and GitHub
merge adapter paths report their source-only capability and return
`ErrWorkflowAtomicGuard` through native preparation. They produce a durable
`publication_guard_unsupported` blocker without invoking the legacy merge API or
changing the method. Provider contract fixtures are local evidence; they do not
establish live publication support.

The trusted `WorkflowPreparationGit` port prepares one configured `merge`,
`rebase` or `squash` result. `OSWorkflowPreparationGit` fetches the retained source
and actual canonical target into an owned temporary object store. A moved base,
conflicting tree or non-linear rebase source blocks preparation. Merge retains
source and base parents; squash has one combined commit on the base; rebase
retains every ordered source change as a linear commit on the base. Each result
is checked against the same method-specific tree and history verifier used for
delivery proof.

The prepared commits enter a separate registered validation worktree. Source
lane, canonical PR and remote target remain unchanged. Before Git IO, a fenced
`integration_validation` attempt journals ownership and the preparation round.
Host commands then use the shared supervisor, lease and repository capacity.
The frozen slice gate runs against the prepared lane, with its exact commands
or explicit waiver. Failed, mutating or truncated gates retain a stale
preparation and resumable checkpoint without rewinding task validation.

An accepted `chr-integration-candidate/v1` binds source/task/contract proof,
observed base, policy, protection revision, result tree/history, validation input
and host command evidence. Its immutable bytes live at
`run/<id>/extension/chr/preparation/<round>.json`, with an identical retained
attempt result. The invocation's `preparation_binding` guards these bytes and
`preparation_count` limits retained rounds to 128. Candidate replacement, host
death and resume do not reset that count. Git IO has a three-minute deadline;
history and workflow files retain the existing 4096-record and 1 MiB limits.

Preparation invalidates an earlier source-only review grant. The next review
input additionally binds `preparation_ref`, `preparation_hash` and `prepared_git`;
manual approval must name that new exact input. A changed base, method, policy,
provider protection, prepared lane or evidence requires fresh validation and
review. Unchanged retries recheck actual Git and retained evidence and reuse the
accepted preparation without another gate. Neither operation submits a merge
or creates a delivery receipt.

## Merger proposal protocol

`WorkflowMergerInput` freezes protocol version 1, entry/unit identity, candidate
revision/hash, claim generation, canonical source/target refs, exact S/B, selected
method, scope hash and an opaque host-owned workspace identity. Evidence stays
under the owning invocation with immutable SHA-256 references. Source intent,
intervening target changes, original source manifest and the validation contract
are required. Each declared conflict names its path, kind and preserved source
and target requirements. A rebase checkpoint retains the exact original replay
step, prepared parent and ordered predecessor mappings.

`SealWorkflowMergerInput` validates the host-assembled contract and hashes all
fields without changing the caller's input. It admits only preparation conflicts,
combined-gate regressions or actionable review findings. Attempts must be positive
and bounded by 10; the deadline must remain within the admission-time worker
runtime, which is bounded by the existing 86400-second policy ceiling. An optional
token allowance requires whole-call metering and the same deadline. The absence
of an allowance does not claim token enforcement. Capacity and budget reservation
identities are included in the immutable input; sealing never acquires a slot or
charges a call.

`CheckWorkflowMergerEvidence` verifies actual retained file hashes, source
manifest ownership and replay position through the contained workflow store.
References to another invocation, noncanonical relative paths, changed images
and symlink traversal are rejected. `ParseWorkflowMergerResult` requires the
current host fence and exact input hash. Input/output images are bounded by
1 MiB, reference/path lists by 128 entries and JSON nesting by 32 levels. Unknown
fields, duplicate keys, case aliases, invalid UTF-8, trailing JSON and incompatible
versions are rejected.

Results are untrusted `resolved`, `needs_decision` or `failed` proposals. They
record changed paths, per-conflict rationale, preserved requirements, supplementary
local-check references and optional whole-call usage. A resolved proposal must
retain every required intent; generated, lockfile, binary, submodule and LFS
resolutions require declared deterministic regeneration and frozen validation
commands. Decision
requests contain a precise issue, affected requirements, two to five options and
an authority owner. Expired inputs never answer a decision. Nil usage remains
unknown and cannot release a finite reservation.

Allowed paths use exact file or directory boundaries. Protected validation paths,
workflow records, agent contracts and ticket/specification scope cannot be edited
by the role. The result has no approval, receipt, method override or publication
grant. Protocol validation does not run a harness, accept actual changes or
replace host validation and candidate-bound review.

## Owned merger execution

`ExecuteWorkflowMerger` admits a host-assembled conflict, gate-regression or
actionable-review input only for the current merging queue entry. It rechecks
source/target observations, generation, frozen scope, configured method, exact
unit validation commands and retained evidence before reserving the shared agent
slot. The existing integration ledger charges a finite merger attempt before
launch and retains its reservation with the immutable role input.

`WorkflowMergerHarness` selects an absolute driver executable and exact arguments.
Its retained `WorkflowMergerHarnessCertificate` binds the executable checksum,
broker protocol and capability revision. The trusted driver must disable built-in
filesystem, shell, network and MCP tools and expose model tool requests solely
through `chr-merger-broker/v1`. An unrestricted worker CLI is not compatible.
The driver is a trusted host component; the broker is a model tool boundary,
not an operating-system sandbox for that component.

The host creates a private Git repository from the current source or verified
prepared candidate. It has no remotes, credential helpers, hooks or shared writable
object store. Driver environment and working directory are private and omit
inherited provider credentials. The model receives hashed intent/conflict images
and can only read, write or delete permitted regular files through the broker.
Protected paths, symlink/hardlink aliases, traversal, shell commands, publication,
approval, scope changes and receipt creation are denied. Local validation processes
are also denied because repaired code could bypass the broker; the exact host gate
remains required after independent diff acceptance.

Every file request rechecks ownership and live source/target fences. JSON-line
messages retain the protocol's strict parsing bounds, with at most 128 messages
and 8 MiB per exchange. The existing process supervisor records birth identity
before granting execution, owns the process group and keeps the shared slot until
all descendants stop. Timeout, pause and cancellation retain failure and consumed
attempt evidence; a paused invocation resumes in integration.

`WorkflowMergerAttempt` retains the proposal and host-observed process evidence
without committing or accepting edits. Only separately certified driver accounting
can settle a finite token reservation, with exact whole-call capability and
input/output ceilings. Unknown or partial usage stays charged. The model's usage
fields cannot settle the ledger. Git method repair, host diff acceptance, combined
validation and candidate-bound review remain separate operations.

## Merger history repair

`BeginWorkflowMergerRepair` creates a private Git checkpoint from exact S/B.
Actual conflicts retain their combined tree or original replay step, prepared
parent and verified predecessor sequence. `ContinueWorkflowMergerRepair` consumes
a host-accepted tree repair, preserves the original author, author date, encoding
and message, and resumes subsequent source steps. A later conflict retains its own
checkpoint. Merge and squash preserve their configured result parents. A clean
preparation returns method proof without requiring a merger call.

`PlanWorkflowMergerRepair` resolves semantic findings against actual original Git
diffs. A unique source owner selects that replay step;
`BeginWorkflowMergerSemanticRepair` preserves its exact prefix and rebuilds the
suffix. Ambiguous ownership requires a concrete decision or an explicitly declared
actionable integration fix. `PrepareWorkflowMergerIntegrationFix` retains that
finding and host attribution in a separate commit while preserving source replay
mappings. Cosmetic findings cannot authorize this fix path.

`WorkflowIntegrationTreeRepair` binds frozen input and per-conflict rationale to
the original source/parent, prepared parent, baseline and repaired trees. Method
verification recomputes actual Git trees and diffs; undeclared paths, missing
conflict changes, incorrect replay positions and disguised binary/submodule/LFS
types fail. Generated and other procedure-bound resolutions require host
observations of the exact regeneration and frozen validation commands and their
input/output trees. Declared incompatible requirements or missing procedures
produce bounded maintainer decisions tied to the input. These trusted Git ports
consume accepted repairs; candidate validation and review remain required before
publication.

## Merger host acceptance

`RetainWorkflowMergerCheckpoint` verifies a host-prepared private Git checkpoint
and seals it to the current entry, unit, scope, candidate revision and claim
generation. Its immutable reference is part of `WorkflowMergerInput.InputHash`.
`ExecuteWorkflowMergerRepair` checks that image and its actual Git prefix before
claiming capacity, seeds a separate broker checkout at the exact baseline tree,
and retains `WorkflowMergerAcceptance` with the owned attempt.

After the certified driver stops, the host reconstructs the tree independently of
its index. It compares every actual changed path with the declared paths and
frozen scope, rejects aliases, protected/control edits, missing conflict changes,
unresolved markers, undeclared file types and whole-side copies (including copies
that only change surrounding whitespace). Per-conflict rationale and preserved
requirement IDs remain evidence, rather than a semantic proof.

Procedure-bound paths must remain unchanged by the model. The host runs the
exact declared regeneration and frozen validation argv under the existing
birth-identity supervisor, capacity claim and finite attempt deadline. At most
eight procedures fit the existing 17-process bound including the driver. Command
failure, truncated output, changed local Git configuration, validation mutations
or escaped output diffs reject the result. Merger completion independently waits
for journaled group quiescence within the original deadline; supervisor exit
alone cannot finish an attempt while an orphan descendant remains alive. Evidence records the actual tree after
manual edits and every subsequent input/output tree; method verification checks
that chain against actual objects. Procedures do not dispatch another model call
or refill its consumed budget.

The host imports verified trees into the separate checkpoint object store and
commits through `WorkflowMergerRepairGit`. Replay continues at the owning original
step, preserving its prefix and attribution. Current S/B, generation, scope,
method, input image and gate are checked before procedure execution, before host
commit and afterward. Rejection retains command and failure evidence; a private
commit observed before a stale callback is recorded as rejected.

`accepted_pending_validation` and `merger_repair_pending_validation` retain a
prepared checkpoint without granting candidate validation, review, publication or
receipt authority. A later conflict retains another replay checkpoint. Complete
results still require the full host combined gate and candidate-bound review;
model local checks and declared procedure validation cannot replace them. These
checks run as trusted host commands with a private environment. They do not
establish an operating-system sandbox for repository code.

## Candidate review and approval

| Operation | Contract |
| --- | --- |
| `ReviewWorkflowSlice` | Re-resolve current review policy and bind it to the validated candidate, canonical PR and observed target. Human mode waits; trusted mode accepts host evidence; agentic mode supervises a separate review argv. |
| `ApproveWorkflowSlice` | Accept the displayed review input hash and recheck every bound input before recording explicit human approval. |
| `CheckWorkflowReviewGrant` | Recheck current policy, immutable evidence, actual clean lane, canonical PR and target against the authorized input hash. |

The owning namespace's `config.yml` supplies `review.mode` and
`work.merge_strategy`. Absence retains `human_review` and `rebase`. Invalid
values reject review. An invocation's optional `review_restriction` can require
`human_review`; it cannot relax project policy. The restriction is retained in
request identity, so an equivalent start or resume preserves it. Policy evidence
records the effective mode, method, source and configuration image hash.

`chr-workflow-review-input/v1` retains invocation/slice identity, aggregate
candidate reference/hash, ordered task proof, canonical PR repository/URL,
head/base branches and commits, observed target and current policy. Review inputs
and decisions are immutable files beneath the run's `extension/chr/review/`.
`review_binding` retains their paths and byte hashes. Historical rounds remain
available when changed inputs replace a current grant.

Human mode returns `waiting_approval`. It holds no worker reservation. Explicit
approval names the exact input hash and re-observes policy, provider facts and
host proof. A changed candidate, task/set image, policy, PR or target rejects
the old approval. Trusted mode records acceptance of required host validation
and enters `waiting_integration` with the review binding `authorized`.

Agentic mode requires a distinct bounded argv. It runs as purpose `slice_review`
through the shared fenced claim, repository capacity and subprocess supervisor.
Its stdin is `WorkflowTaskInput` with `review` and `review_hash` fields. The role
returns exactly one JSON `WorkflowReviewDecision`: `binding_hash`, structured
`result` using the existing agentic review contract, and an optional
`repair_task_key`. A pass with blocking findings, another input hash, a cosmetic
blocking finding or a repair key outside the slice rejects the decision.
The host rechecks all bound inputs after the role exits.

An agentic pass enters `authorized`. `insufficient_evidence` retains a review
stop and operator blocker. `human_required` retains the role decision and
permits exact-bound explicit approval; that approval extends the immutable
decision chain. `needs_changes` enters repair for its named task and ordered
successors, preserving the validated prefix and canonical PR. Three automatic
review repairs are permitted per slice within an invocation. Consumption is
journaled with the repair and survives candidate changes, restart and resume.
Exhaustion retains `repair_limit_exceeded` and returns
`ErrWorkflowReviewRepairBudget`.

Decision acceptance journals task `waiting_slice` / `ready_for_merge` and
set-owned slice `ready_for_merge` together with evidence and invocation state.
Review alone leaves integration pending. Review grant checks establish the
review boundary; guarded publication also requires prepared combined-result
validation and source/base protection.

At most 128 review rounds are retained per invocation, with at most two
decisions per round for an agentic human-required stop followed by approval.
The review role shares worker lease, argv, output and file bounds. Truncated or
malformed output fails; a failed or interrupted review retains a paused
`waiting_review` checkpoint and the original task/candidate evidence. Durable
cancel is terminal. Host death during input, claim, decision or approval writes
replays the guarded journal; accepted decisions are reused without another role.

## Durable submission and target exclusion

| Operation | Contract |
| --- | --- |
| `SubmitWorkflowSliceIntegration` | Recheck the exact publication guard, persist intent and exclusion, then call the guarded provider once. Existing uncertain intent enters observation. |
| `ReconcileWorkflowSubmission` | Verify actual canonical PR/Git delivery or a definitively unpublished operation without consuming a submission attempt. |
| `ReadWorkflowSubmission` | Read the authoritative repository journal, including intent retained before an interrupted invocation projection. |

`WorkflowPublicationProvider` extends the guarded capability with submission and
stable-ID observation. Submission must protect source, validated base and exact
prepared result in the same server transaction. The current source-only native
Forgejo/GitHub adapter paths reject submission without calling legacy merge.

One `chr-integration-journal/v1` file per integration unit/invocation lives in
`chr-workflow-publications/<key>.json` under the canonical Git common directory.
Its atomic image retains the stable submission ID, candidate identity/hash,
canonical PR, source/base/method/result proof, prepared lane, exact validation and
review bindings, current policy, attempt number, timestamps and target exclusion.
The invocation's `submission_binding` is a projection written after that image.
Host death between the two writes leaves authoritative intent and exclusion;
recovery never infers permission to submit from a missing projection.

Publication exclusion uses normalized canonical repository and target branch,
independent of namespace or project aliases. All linked worktrees share the same
coordinator. The durable common-directory journal continues to exclude the target
after cancellation, process death, worker lease expiry or loss of an invocation
projection. Preparation and review do not reserve target publication.

A provider response retains `awaiting_observation`, even when accepted with a
result commit. An open PR or missing result cannot prove failure. The provider's
unpublished observation must name this exact submission and guarantee it cannot
publish later; canonical PR and target observations are rechecked before sealing
it. Confirmed delivery requires the matching merged canonical PR and fetched,
reachable, method-specific Git proof. The exact prepared result confirms native
publication intent. A different result or base can confirm separately verified
external delivery; `external_delivery: true` retains that distinction in the
journal and its invocation projection. That observation proves the configured
method's actual combined tree and complete ordered history, without attributing
external execution to the guarded command. Submission observation does not clear
stop intent, create a receipt or mark any task done.

Submission IO and reconciliation have three-minute deadlines. Durable pause or
cancel cancels a live provider call while preserving uncertain exclusion. Each
integration unit/invocation permits three submissions by default; an explicit
execution-policy limit permits 1 through 10. An omitted policy field retains its
legacy request identity and uses the same effective limit. Sealed unpublished
attempts require fresh canonical observation before a guarded submission within
the remaining budget. Changed
candidates and repeated starts do not reset consumption. Exhaustion retains
`integration_submission_budget_exhausted` with an operator action owner.

Each journal remains within 1 MiB, holds at most ten attempts, and the repository
coordinator bounds retained journals to 4096. Invalid images, unknown schemas,
symlink traversal or exceeded bounds fail closed. Accepted responses and later
observation update the guarded journal; earlier sealed attempts remain retained.

## Shared publication operation

`AdvanceWorkflowSliceIntegration` is the central application operation over the
retained slice PR. It returns the current invocation and, after verified shared
finalization, its receipt. Each call observes canonical PR/Git facts before
running another phase. It performs at most one combined gate, one separate
review role and one guarded submission; it never starts task work or a later
slice. Human approval uses the displayed `review_binding.input_hash` through
`ApproveWorkflowSlice`.

- An actual merged PR enters method-specific verification and shared receipt
  finalization, including when automatic provider guards are unsupported or the
  invocation is paused or cancelled. Stop intent remains intact. A paused run's
  resume checkpoint remains at the integration barrier.
- An uncertain submission requires canonical confirmation or the provider's
  stable-ID unpublished seal before any new preparation, review, repair or
  publication. Cancellation and lease expiry cannot release it.
- A moved base or changed policy/protection requires a fresh isolated prepared
  result, host combined gate and exact review binding. Old approval cannot
  authorize that new result. Unchanged original task and aggregate proofs remain
  reusable; corrupt evidence fails closed.
- Current native provider APIs retain the atomic-guard blocker. The operation
  never substitutes a method or invokes their legacy merge endpoints.

The host rereads canonical PR identity, merged result and current target after
Git verification. A native receipt names the confirmed stable submission ID;
a verified external result uses the retained canonical PR binding identity.
Both use the same journaled task/set/receipt finalization. Completed retries
validate the retained receipt and publication projection without allocating new
work. Sealed attempts remain consumed history after a repaired candidate; the
new candidate requires fresh preparation/review and a new guarded attempt, or
separately verified external delivery. Prepared lanes retained in prior attempt
records share repository process
exclusion without copying operational locks into their validated trees.

## External integration proof

`VerifyWorkflowSliceIntegration` reads the bound canonical PR and fetches its
exact target branch through a trusted host Git port. It accepts only a merged
PR with matching repository, head repository, branches and validated source
commit. Worker success, manually written task status and a structurally valid
receipt cannot establish delivery. The application service uses the canonical
HTTPS repository as its Git transport; explicit repository/transport mappings
belong to trusted host wiring, outside task inputs.

The observer imports the retained source into an owned temporary bare repository
and fetches the canonical target there. It requires the provider result to be
reachable from that actual target and its integration base to contain the
frozen target. The provider's current PR base SHA is not an immutable base:
verification derives the base from the actual method-specific result history.

| Method | Required Git proof |
| --- | --- |
| `merge` | Exactly two result parents: target base, then the validated source. Its tree equals the clean combined source/base tree. |
| `rebase` | Linear source history and an equally sized ordered linear result on the target base. Each result tree equals that source commit's change replayed onto the preceding result. Original source ancestry alone is insufficient. |
| `squash` | Exactly one result commit on its target base. Its tree equals the combined source change applied to that base. |

The host rechecks candidate, membership, contracts and retained input/evidence
images after provider and Git IO. Successful verification journals immutable
`chr-slice-delivery-proof/v1` evidence and invocation `delivery_proof_ref` /
`delivery_proof_hash`. It retains the canonical PR observation, original
candidate, source/base/result trees, ordered history fingerprint and observed
target. Local temporary paths and transport details are excluded from evidence.
Repeated verification rechecks actual Git and reuses matching immutable proof;
later target commits preserve delivery when the verified result remains
reachable. Mismatched or edited proof fails closed. Tasks remain at the shared
finalization barrier until receipt finalization completes.

Verification holds repository mutation exclusion, has a three-minute shared
deadline, permits at most 4096 source/history commits and bounds Git stdout and
proof files to 1 MiB. Git replacement objects are disabled. Cancellation stops
the owned Git process group; the temporary repository is removed only after
the observer finishes. Conflicts, missing target history and unsupported method
proof do not trigger a method fallback or a target write.

## Shared receipt finalization

`FinalizeWorkflowSlice` requires the invocation's retained host delivery proof.
It accepts no caller-supplied receipt. Before the first finalization, it rechecks
the canonical merged PR, actual target reachability and method-specific proof,
then rechecks frozen member, set and candidate images under repository mutation
exclusion.

One `chr-workflow/v1` receipt beneath the owning run's `evidence/` binds the
integration unit, candidate/submission identity, canonical provider/repository/PR,
source/base/result, prepared and verified tree, history fingerprint, host proof
reference, aggregate/task validation references and confirmation time. Its
stable identity is derived from the run, candidate and retained PR binding.

Receipt creation and all task/set/invocation images form one guarded transition
journal. Every task enters `integrated` / `done` with the same
`slice_delivery_ref`; its `slice_delivery` projects the common PR, branch,
method, result commit and confirmation time while preserving task evidence.
The set-owned slice enters `integrated` with that same receipt. Invocation
`delivery_ref` / `delivery_hash` bind its exact immutable receipt bytes.

Host death at any replacement leaves the journal for replay before another
operation can read or admit successors. Retries after completed replay check
the receipt hash, host proof, frozen images and every finalization projection;
they return the same receipt without another write, worker or provider call.
Concurrent callers serialize and share that same result. An edited proof or
receipt, changed target proof or member image blocks finalization.

An exact slice invocation completes at receipt finalization. A broader
set/use-case invocation retains its current slice receipt barrier and leaves
later members untouched, including when its last slice is integrated.
Slice finalization does not close the owning set.
The complete multi-image journal remains subject to the 1 MiB transition bound;
an oversized fan-out fails before any task completion write.

## Ordered integration frontier

`AdvanceWorkflowFrontier` releases one finalized slice of a retained set or
use-case invocation. A task-validation result, merged PR or host delivery proof
without completed receipt finalization cannot release this barrier. Unsliced
use-case tickets remain excluded from the original selection.

The host observes the exact canonical repository and target branch, imports
its commit objects into the host repository without moving the checkout or
branch, and verifies the retained PR result against that current target. It
rechecks canonical branch, method-specific history, receipt, candidate and
frozen task/set images before journaling the release. Changed target history,
missing proof or stopped invocation blocks continuation. The operation has a
three-minute shared deadline and retains repository mutation exclusion.

An immutable `chr-unit-frontier/v1` image retains the unit's lane, candidate,
PR, delivery proof, receipt and fresh canonical target proof. The invocation's
`integrated_units` binds each image by reference and hash; `confirmed_cursor`
records receipt-finalized work. The original selection, request identity and
policy stay frozen. `execution_target` supplies the fresh base for later lanes.
Active slice bindings clear atomically with the retained image and cursor.
Restart reuses that checkpoint without restoring the previous active PR or
repeating workers. Task, aggregate and provider observation evidence remains
immutable and is checked before later work.

Each later slice owns a distinct cumulative lane and canonical PR binding.
PR observation files are scoped by binding, with the existing 128-observation
bound per binding; prior observation paths remain readable. The next slice in
the same set becomes `pending` after release. A set boundary, including the
last selected set, enters `waiting_integration` with outcome
`set_closure_required`. Claiming another task remains blocked until closing
validation releases that separate barrier.

## Bounded continuation checkpoints

`WorkflowContinuationRequest` supplies worker argv, initial PR content and an
optional step limit. It accepts no selector, filters, target or membership
override. Equivalent unfinished starts retain one invocation ID; continuation
uses that ID and its original request hash. Explicit scope or policy changes
through an existing request key reject.

The frozen execution policy permits a checkpoint step limit from 1 through 128.
An absent value uses 128, including earlier invocation files. A call may use a
smaller limit; it cannot increase the frozen bound. Each accepted task,
aggregate, PR, integration, frontier or set-closing operation consumes one step. Exhaustion
returns reason `step_limit` with resumable state. Awaiting authority, paused or
blocked work returns without executing successors. A set boundary returns a
closure checkpoint; the next continuation executes only that set's gate.
Cancelled and completed invocations stay terminal.

`chr-continuation-checkpoint/v1` records request identity, source revision,
validated and receipt-confirmed cursors, member/attempt counts, accepted steps,
reason and parent reference/hash. Immutable checkpoint creation and the
invocation's `continuation_checkpoint` binding share one replayable journal.
Restart retains consumed attempts, accepted task results, the canonical PR and
finalized receipts. Identical waits reuse their checkpoint without another
event. Checkpoint corruption blocks further claims. Checkpoint history is
bounded to 4096 images per invocation and rejects exhaustion before another
transition; every image retains the 1 MiB file bound.

## Membership and approved amendments

`WorkflowExecutionPolicy.IncludeLaterTickets` is false unless explicitly accepted
at admission. `ReconcileWorkflowMembership` reports matching later ready tasks in
`Excluded["later"]` under the default policy. With the accepted policy enabled,
it admits matching ready tasks within the original selector, retaining every
initial member image, request hash, budget and ordering rule. New work cannot
precede validated or receipt-confirmed progress. Loose filters belong to the
loose extent; ordered selectors retain their exact ticket, slice, set or use-case
identity. Completed sets and terminal invocations never reopen through admission.

Membership changes wait for the current unit's lane to clear. Continuation checks
membership at these quiescent boundaries, and each accepted membership change
consumes one bounded step. Direct task claims use the same reconciliation. A
changed ordering rule, external task reassociation or incompatible source image
blocks reconciliation. An unchanged discovery result does not write another
checkpoint. Appending a slice to the current set releases its closure barrier;
adding another use-case set keeps the current set's closure barrier intact.

`AmendWorkflowSet` requires the observed set revision and source hash, a changed
approved plan, and the native ticket-key allocator. It preserves existing task
identities, order and execution contracts, existing slice order and contracts,
and all delivered projections. It can append tasks to an unstarted slice or add
future slices. Live reservations, unresolved claims, retained active lanes and
publication bindings block amendment. An unfinished owner must use its owning
checkout. Set association and use-case order cannot change.

The changed plan and `chr-set-amendment/v1` receipt are immutable files. The
current set's `plan_ref` points to the new approved plan; the original plan
artifact keeps its bytes for prior task and closure references. New
tickets, descriptors, set revision and unfinished owner images share one
replayable journal. The set retains revision/reference/hash amendment bindings;
approval or copied-plan corruption blocks further work. Explicit `Reopen` is
required for a completed set and requires a changed plan. Reopening moves the
current closure reference to `closure_history`, preserving its hashed bytes,
prior task/slice receipts and terminal invocation files. Ordinary record writes
cannot change a completed set. Amendment and closure histories retain the
4096-entry and 1 MiB record bounds.

## Set closing validation

| Operation | Contract |
| --- | --- |
| `CloseWorkflowSet` | Require the observed set revision/hash and canonical repository/branch. Validate all actual members and the exact target, then retain the set-owned result. An optional invocation ID identifies its unfinished owner. |
| `CloseWorkflowInvocationSet` | Close only the next frozen set after its selected work finalizes, without changing admission or earlier delivery images. |
| `ReconcileWorkflowClosing` | Recover an expired or dead set-closing owner after group quiescence and acquisition of the appropriate legacy guard. Retain the interrupted attempt without confirming its gate. |

Closure checks every actual task and slice in the set, including later tickets
excluded from frozen invocation membership. Every task must be receipt-finalized
and every slice must retain its integration receipt and method-specific proof.
Fresh canonical PR and Git observations must agree with those receipts and the
current target. Pending actual members block closing before another attempt.
Empty sets still require an explicit approved gate or waiver.

`chr-set-closing-input/v1` freezes actual ticket IDs, revisions and source hashes,
ordered slice receipts and fresh proofs, the approved plan reference/hash,
closing contract/hash, canonical target commit/tree and prerequisite evidence
hashes. Commands run on a clean isolated checkout of that exact target, with the
input JSON on stdin and its reference in `CHR_WORKFLOW_INPUT`. The checkout stays
outside the operational namespace. The host checks member and evidence images,
stop intent, lease and freshly observed target again before confirmation.
Target movement, changed members or a gate that dirties its checkout cannot
establish completion.

Each set invocation retains a budget of 128 closing attempts; retries and restart
consume the same counter. Set history is bounded to 4096 attempts across explicitly
reopened cycles. Each attempt owns its fence, PID birth, lease, immutable input, bounded
logs and result under `plan/<set-id>/extension/chr/closing/<generation>/`.
Required gates use the contract's finite timeout; waivers retain their reason
and a one-minute observation lifetime. Closing uses the same repository worker
slots, descendant supervision and durable pause/cancel controls.
A live owner or descendant blocks recovery. `ReconcileWorkflow` also dispatches
to an invocation's active set-closing role.

Failure retains `closing_validation_failed`, leaves the set blocked and releases
its closing role after process quiescence. Retry executes closure alone; validated
and integrated task evidence, worker budgets and prior deliveries stay intact.
Interrupted owner recovery retains `interrupted` without a success receipt.
Changed frozen inputs require reconciliation before retry.

A successful `chr-set-closure/v1` receipt retains the input reference/hash, exact
snapshot, host command results or explicit waiver, and confirmation time. Receipt,
set `completed` projection and owning invocation's `set_closures` frontier share
one replayable transition. The next associated set starts from the confirmed
canonical target. The final selected set makes its invocation terminal. Historical
closure receipts and earlier terminal invocations remain immutable; explicit
approved reopening preserves the receipt in `closure_history`.

## Manual protocol

| Operation | Contract |
| --- | --- |
| `PrepareWorkflowManualTask` | Freeze the owned handoff once. Return `WorkflowManualTask` with input, reference, hash and lease deadline. |
| `ExecuteWorkflowManualTask` | Supervise the frozen harness argv once. Harness success retains the claim and does not advance the cursor. Failure or interruption uses the shared result/checkpoint writer. |
| `ReconcileWorkflowTask` | Accept `WorkflowManualCandidate` with input hash, full branch ref, full commit and tree; verify the actual clean Git image and run the frozen host gate before applying results. |

A harness is supplied as argv during preparation and launched through
`ExecuteWorkflowManualTask`, using the same process ownership and capacity as
automatic workers. Empty argv permits a human to edit and commit a quiescent
lane candidate without launching a worker. The caller then reports the exact
candidate through `ReconcileWorkflowTask` while retaining the original claim.

Reconciliation checks the lease, fence, frozen task/set revisions, prerequisites,
evidence bytes and candidate before and after validation. It does not accept a
caller-provided success flag. Direct ticket/status edits cannot advance work.
Retrying an accepted result with its original claim reuses proof only while its
inputs, prerequisites, evidence and checkout remain unchanged; it creates no
new attempt or validation process.

## Stop and recovery

`StopWorkflow(id, "pause" | "cancel")` records intent before releasing ownership.
A running command stops at a fenced checkpoint with retained bounded output and
evidence. `ResumeWorkflow` clears a completed pause and restores its retained
frontier or integration barrier without resetting attempt budgets. Cancellation
is terminal; an equivalent start attaches without restarting it.

Each command journals its Linux boot ID, PID and start ticks before launch
permission. A Bash supervisor retains the capacity descriptor, waits for its
process group and enforces the finite lifetime after host loss. A durable shared
process-group ledger keeps surviving children counted after supervisor loss.
Signals require the registered process birth; PID reuse does not authorize
killing another process. Linux `ENOENT` and `ESRCH` observations mean a process
vanished; other read failures block ownership recovery.

`ReconcileWorkflow(id)` explicitly replays matching journals and recovers a dead
or expired owner only after process-group quiescence and verification of every
retained owned operation lock. A counted native attempt reacquires its exact slot
and a shared legacy guard for metadata recovery. A held or reused assigned slot
blocks that recovery while unrelated slots can continue. Historical
attempts without a slot binding retain exclusive recovery. It preserves lane, input, logs,
checkpoint, consumed budget and replacement history. A live child or held
reservation blocks replacement; expired/replaced credentials cannot finalize or
release a newer claim.

## Bounds and errors

Worker policy defaults to 3600 seconds and three attempts per task. Explicit
bounds are 1–86400 seconds and 1–128 attempts, with at most 4096 retained invocation
attempts. Each attempt permits one worker and at most 16 host commands. Command
argv has at most 64 elements and 4096 bytes. Each output stream retains at most
16 KiB; failure text retains at most 4096 bytes. Input/result files remain within
the 1 MiB workflow-file bound. Validation deadlines also respect the claim lease.

`ErrWorkflowOwned` rejects overlap, invalid/expired credentials or nonquiescent
claim operations. `ErrWorkflowCapacity` reports counted live repository capacity.
`ErrWorkflowConflict` rejects changed frozen or candidate images.
`ErrWorkflowStop` identifies durable stop/interruption checkpoints. Invalid
arguments and host-command failures return their validation errors; gate failure
retains unsuccessful evidence and pauses the invocation without advancing work.
