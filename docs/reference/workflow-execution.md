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
`automatic` or `manual`. It binds invocation/attempt IDs, generation, task
revision, source and contract hashes, body, execution contract, target, scope,
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
