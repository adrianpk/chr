# Slice 9: Independent parallel runners

Status: reviewing

Delivery set: ticket-native-workflow

Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)

Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)

Branch: `feat/slice-9-workflow-concurrency`

PR: [#372](https://forge.adrianpk.com/adrianpk/chr/pulls/372).

## Purpose

Allow configured independent native lanes to overlap while retaining ordered
frontiers, fenced task results, bounded admission and canonical publication
exclusion across one registered repository coordinator.

## Delivered Behavior

Worker capacity defaults to one and accepts opt-in limits of 1–8. Namespaces and
linked host checkouts share process-held slots in the Git common directory and
must resolve the same policy. The ready queue defaults to 64 unfinished
integration lanes, bounded to 1–1024. Overflow preserves accepted work;
equivalent requests attach, and cancellation frees admission without rewriting
history. Reduced capacity still counts reservations above the new limit.

A live worker or orphaned process group retains its assigned slot after lease,
host or supervisor loss. Native claims no longer create coarse operation files.
Legacy runners and adoption remain exclusive across the complete pool; retained
legacy and historical coarse files block native starts until their own release
or matching fenced recovery. Unknown or replaced files are never removed by
admission. Quiescent counted-owner recovery reacquires its assigned slot and a
shared legacy guard, permitting unrelated slots to continue. An open or reused
assigned slot blocks that recovery. Historical unbound attempts retain conservative exclusive recovery.

Public coordinator registration is journaled with explicit native authoring.
Private host and common-directory identity rejects copied authority and
unsupported separate clones or hosts. Linked roots and namespaces share one
owner; moving the original directory preserves its identity. Execution,
closure and candidate boundaries verify committed canonical registration.
Worker-owned lane copies cannot redefine the host coordinator. Task, slice,
ordered-set and use-case frontier checks enforce exclusive ownership on shared
start paths, including manual handoffs and legacy operations. Replaced or
expired credentials cannot apply results or release a newer reservation.

Capacity waits retain their prior task, aggregate, review, preparation, repair
or closing phase without consuming attempts or changing accepted evidence.
Explicit retry and continuation recheck capacity. Repeated unchanged waits do
not grow checkpoints; pause/resume retains the prior phase and cancellation
stays terminal. Human review and publication waits hold no worker slot.

Publication retains exclusion for the canonical remote target independently of
worker capacity. Actual publisher processes across namespaces and linked
aliases never submit together. Unknown outcomes retain exclusion through
cancellation; trusted unpublished evidence releases it. Target movement rejects
stale combined validation and review. A fresh combined tree, gate and review
must precede publication, and a conflicting lane preserves another lane's
confirmed receipt.

## Implementation Notes

`WorkflowStore` owns admission, reservations, coordinator/frontier validation,
wait checkpoints and recovery. The central application service retains the
shared execution boundary. Kernel-held slot descriptors and a durable actual
process-group ledger count live work; cleanup closes only owned descriptors.
The existing durable submission journal and atomic Git guards provide target
publication exclusion. Real source and publisher subprocess barriers establish
overlap and exclusion rather than inferring either from race-detector success.

## Contracts Added or Changed

- `work.max_active_workers` and `work.ready_queue_capacity` resolve from the
  host checkout's `.chr/default/config.yml`; namespace and retained lane copies
  cannot create extra pools.
- `chr-coordinator/v1` supplies public opaque registration with a private local
  owner binding. Unregistered or copied authority blocks work.
- Invocation attempts, process observations and closing attempts retain bounded
  `capacity_slot` bindings. Zero preserves conservative historical behavior.
- Invocation `capacity_resume_status` retains a waiting role's prior phase.
  Wait, pause, cancellation and continuation reuse frozen evidence and budgets.

## Files of Interest

- `internal/core/workflow_capacity_slot.go`, `workflow_capacity_process.go` and
  `workflow_capacity_config.go`: shared slot allocation, process counting and
  queue/configuration limits.
- `internal/core/workflow_coordinator.go` and
  `workflow_coordinator_admission.go`: local authority and committed target
  registration.
- `internal/core/workflow_claim_lock.go`, `workflow_capacity_wait.go` and
  `workflow_capacity_recovery.go`: conservative legacy exclusion, resumable
  waits and independent metadata recovery.
- `internal/core/workflow_parallel_runner_test.go`,
  `workflow_parallel_recovery_test.go` and
  `workflow_parallel_publication_test.go`: actual overlap, expiry, replacement,
  unrelated blockers and serialized target publication.
- `docs/reference/workflow-execution.md`, `docs/how-to/user-guide/index.md` and
  `CHANGELOG.md`: ownership boundary, capacity configuration and product notes.

## Validation

T9.1–T9.4 passed focused process/race checks and repository task gates. Exact
prior task commands and corrections are recorded in their reviewing tickets.
T9.4 passed:

- `make verify`
- `go test -race ./internal/core -run '^(TestWorkflowParallel|TestWorkflowOwnerExpiry|TestWorkflowRecovery|TestWorkflowCapacityRoleWait|TestWorkflowCapacityContinuation|TestWorkflowCapacityStoppedWait|TestWorkflowCapacityRetainedLock|TestWorkflowClosingRecovery)' -count=1`
- `git diff --check`

Go 1.27.1, the retained CLI curation fixture and a scoped uncached 30-minute test
timeout were used. Strict lint reported zero issues. The first complete gate
exposed recovery of a still-held expired reservation. Recovery was corrected to acquire the exact
assigned slot before fencing; the focused race checks and complete retry then
passed. Source-process barriers proved capacity-one serialization, capacity-two overlap, third-lane waits,
legacy exclusion and independent failure boundaries. Expired live groups
retained capacity and rejected results. Recovery preserved generations and
budgets beside unrelated native work. Capacity retries passed for all worker
roles and repeated continuation retained its existing wait checkpoint.

Required final clean PR-head checks, recorded through the controller, use this
order:

- `make verify`
- `go test ./internal/core ./internal/app -count=1`
- `go test -race ./internal/core ./internal/app -count=1`
- `git diff --check`

These candidate checks are distinct from the delivery set's later full gate.

## Risks and Follow-ups

T9.3 had one canonical fetch failure during its repository gate. Ten targeted
traced repetitions and the complete traced retry passed without a source
change; its cause remains unresolved. Provider publication acceptance uses
controlled Git remotes and provider fixtures. Live-provider proof remains in
the approved assurance slice.
