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
| `ClaimWorkflow` | Claim the next eligible task under repository serial capacity and existing operation locks; journal a new fenced attempt. |
| `CheckWorkflowClaim` | Check the current private credential, generation, lease and frozen record images. |
| `CheckpointWorkflow` | Yield an unexpired, quiescent claim without validating the task; retain its input, lane, attempt and checkpoint. |
| `ExecuteWorkflowTask` | Freeze worker inputs, supervise its argv in the isolated lane, run the frozen host gate and journal the guarded task result. |

The caller retains the returned `*WorkflowClaim`; status cannot reconstruct it.
The token is excluded from JSON and only its hash is persisted. One claim covers
the task and its owning hierarchy. Native and legacy starts share the Git common
directory's process-held serial reservation across registered linked checkouts.
Successful task validation leaves task status `in_progress` and phase
`validated`, with integration pending. An exact grouped-ticket invocation
completes only that task. A whole slice or independent ticket stops at
`waiting_integration`; local validation does not mark work `done`.

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

The invocation completes when its selected members end at this slice. A broader
set/use-case invocation retains its current slice receipt barrier and leaves
later members untouched. Slice finalization does not close the owning set.
The complete multi-image journal remains subject to the 1 MiB transition bound;
an oversized fan-out fails before any task completion write.

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
or expired owner only after process-group quiescence, capacity acquisition and
verification of every owned operation lock. It preserves lane, input, logs,
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
