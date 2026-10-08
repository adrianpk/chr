# Slice 4: Owned Isolated Task Execution

Status: delivered
Delivery set: ticket-native-workflow
Plan: [Ticket-Native Work Delivery](../../../plan/ticket-native-workflow.md)
Tracker: [Ticket-Native Work Tracker](../../../tracker/ticket-native-workflow.md)
Branch: `feat/slice-4-workflow-execution`
PR: [#367](https://forge.adrianpk.com/adrianpk/chr/pulls/367)
Implementation candidate: `df517f43250d1f7e52e7599984c9e33d0b920850`

## Purpose

Execute one owned task in a retained isolated lane, verify its candidate on the
host and preserve safe manual handoff, stop and recovery boundaries.

## Delivered Behavior

- Equivalent starts attach to a durable invocation. Frozen selection, policy,
  target and task/set images guard admission and attempts. Overlapping task and
  hierarchy claims reject across registered linked checkouts and namespaces.
- Native and legacy starts retain shared repository serial capacity and existing
  operation locks. Private credentials and increasing generations fence result,
  checkpoint and release mutations. Status excludes owner tokens.
- Each task receives immutable input and a retained cumulative Git worktree.
  Host validation binds ordered gates or an explicit waiver to a clean branch,
  full commit/tree, execution contract and unchanged prerequisite evidence.
  Workers cannot certify status through their output.
- Exact grouped-ticket work validates only that task, reports integration
  pending and creates no partial PR. Broader slice requests reuse unchanged
  task evidence and lanes and stop at the integration barrier. Validated tasks
  remain `in_progress`; integration owns completion.
- Pause and cancellation persist intent before ownership release. Quiescent
  checkpoints retain input, lane, output, attempts and budgets. Resume restores
  the retained frontier or integration barrier; cancellation is terminal.
- Linux process birth is journaled before launch permission. A surviving Bash
  supervisor retains capacity and enforces lifetime after host loss. A shared
  durable process-group ledger keeps children counted after supervisor loss,
  including workers that close inherited descriptors. Signals bind process
  birth rather than an unverified reused PID.
- Explicit dead-owner reconciliation checks process-group quiescence, held
  capacity and every owned lock before journaling a recovery checkpoint.
  Replacement retains generations, attempt budgets and prior ownership history.
- Manual preparation, supervised harness execution and explicit Git-candidate
  reconciliation use the same claims, revisions, prerequisites, host gate and
  result writer. Harness success alone and direct status edits do not advance
  work. Accepted manual replay checks retained inputs, evidence and clean Git
  state without another attempt or gate.

## Implementation Notes

Invocation files live in the existing run extension. Per-attempt input, command
output, result and checkpoint files carry immutable references and hashes.
Task/result/cursor transitions share the guarded file journal; partial matching
writes replay, while incompatible edits remain intact and reject reconciliation.
The central application service delegates to the core owner, including when
called from a registered alias. SQLite is not execution authority.

Finite policy defaults are 3600 seconds and three attempts per task. Explicit
limits cap lifetime at 86400 seconds, attempts per task at 128, retained invocation
attempts at 4096 and commands per attempt at 17. Command argv is bounded at 64
elements and 4096 bytes; each output stream retains a 16 KiB prefix. Validation
also respects the private claim lease. Frozen files retain the 1 MiB bound.

## Contracts Added or Changed

- `WorkflowStartRequest`, `WorkflowInvocation`, `WorkflowClaim`, `WorkflowAttempt`
  and shared start/read/claim/check/checkpoint operations retain owned progress.
- `WorkflowLane`, `WorkflowTaskInput`, `WorkflowTaskResult` and
  `ExecuteWorkflowTask` bind isolated execution to host evidence and integration
  pending status. Input/result `execution_mode` distinguishes automatic/manual.
- `WorkflowProcessIdentity`, `WorkflowProcess`, `StopWorkflow`, `ResumeWorkflow`
  and `ReconcileWorkflow` retain process, stop, replacement and recovery facts.
- `WorkflowManualTask`, `WorkflowManualCandidate`, `PrepareWorkflowManualTask`,
  `ExecuteWorkflowManualTask` and `ReconcileWorkflowTask` expose the shared
  manual protocol with full Git identities and a retained private claim.

## Files of Interest

- `internal/core/workflow_invocation*.go`, `workflow_claim*.go`: admission,
  guarded task ownership, budgets and shared serial exclusion.
- `internal/core/workflow_lane.go`, `workflow_execution*.go`: retained lanes,
  frozen inputs, subprocesses, host gates, result journals and evidence reuse.
- `internal/core/workflow_stop*.go`, `workflow_recovery*.go`,
  `workflow_process*.go`: durable intent, process identity/capacity and recovery.
- `internal/core/workflow_manual*.go`: manual handoff, candidate verification
  and immutable accepted-result replay.
- `internal/app/service_workflow_execution.go`: adapter-neutral operations.
- [Workflow execution reference](../../../../../docs/reference/workflow-execution.md).

## Validation

Successful final task implementation checks:

- `make verify` — strict lint and full repository tests passed.
- `go test ./internal/core ./internal/app -run '^(TestWorkflowManual|TestWorkflowProcessObservation|TestWorkflowStopBarrier|TestServiceWorkflowManual)' -count=1` — focused manual/process regression tests passed.
- `go test -race ./internal/core ./internal/app -run '^(TestWorkflowManual|TestWorkflowTask|TestWorkflowStop|TestWorkflowRecovery|TestWorkflowProcess|TestServiceWorkflowManual|TestServiceWorkflowTask|TestServiceWorkflowClaim)' -count=1` — focused concurrency and real-process tests passed.
- `git diff --check` — passed.

Real fixtures cover competing starts, overlapping scopes, linked worktrees,
exact grouped requests, cumulative reuse, independent integration barriers,
changed task/set/input/prerequisite/evidence images, dirty or moved candidates,
partial claim/result/stop journals, terminal retry, lease expiry, stale releases,
host and supervisor loss, descriptor-closing children and retained output.
Manual fixtures additionally cover a human-created commit, harness-only progress,
candidate assertions, failing gates, live validation stop and accepted replay.

The first T4.4 focused race run failed in `TestWorkflowStopBarrier` because a
process vanished during `/proc` observation and Linux returned `ESRCH`. The
observation path now treats that as process disappearance while retaining
fail-closed handling for other errors. Updated focused tests, repository gate
and focused race command passed. Go toolchain: `go1.27.1 linux/amd64`.

Repository gates use the retained owned CLI dump-curation executable documented
in the [Slice 2 report](slice-2-workflow-materialization.md). It intercepts that
fixed fixture input, delegates all other calls and has no service to stop.

Canonical PR #367 merged into `dev` at `517fdbb3322b4e02b951e5301b8f6c8e04740f57` on 2026-10-08
at 19:45:19 UTC. All four task tickets are solved. Controller evidence passed
`make verify`, `go test ./internal/core ./internal/app -count=1`,
`go test -race ./internal/core ./internal/app -count=1` and `git diff --check`
in that order on the clean integrated head. Its immutable-head JSON/log retain
those results. Acta merged event: `01M4EGRYV77Q7CJ4ZJ00TE0JS0`.

Acceptance: A06 exact grouped task; A09 retained serial ownership; A12 lease,
process capacity and fencing; A17 shared manual validation; A21 execution-stop
retention; A24 finite execution and retained attempt budgets. Later publication
and integration stages compose the remaining parts of those criteria.

## Risks and Follow-ups

Process supervision and birth observation require Linux procfs and Bash.
Unregistered aliases, live reservations, incompatible record/lock images or
changed retained candidates require explicit reconciliation. Serial exclusion
remains effective. Slice 5 owns aggregate candidate and delivery receipts;
publication, continuation, parallel capacity and transport controls retain
their recorded later slices. The CLI curation fixture is a validation dependency.
