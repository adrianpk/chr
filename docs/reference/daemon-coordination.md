# Daemon Coordination

The local foreground daemon coordinates enabled registered projects. Its
user-level SQLite state records coordinator status and heartbeats. Ticket-native
invocations remain authoritative in their shared project-local workflow records.

## Commands

```sh
chr daemon status
chr daemon start --dry-run
chr daemon start --max-iterations 1
chr daemon stop --dry-run
chr daemon stop
chr daemon workflow status --dir /path/to/project --json execution-id
chr daemon workflow pause --dir /path/to/project execution-id
chr daemon workflow cancel --dir /path/to/project execution-id
chr daemon workflow resume --dir /path/to/project execution-id
```

`--max-iterations` must be positive and defaults to one. Start runs in the
foreground and finishes after that many coordination iterations. Dry-run start
reports the single-instance guard without writing daemon or invocation state.
Dry-run stop prints the stop plan without applying it.

`daemon workflow` also supports `continue`, `approve` and `reconcile`. These
commands use the same arguments, read-only previews, revision checks, frozen
limits and exact approval bindings as [Workflow controls](workflow-controls.md).
They address one retained invocation; they do not start a new extent.

## Retained ticket-native work

For registered projects with `work.state_authority: tickets`, each iteration
observes retained invocations in stable ID order through shared read-only pages.
Startup never creates an invocation or selects a new backlog extent. A user must
first authorize an extent through a work command or a confirmed TUI/Web action. Configuration changes to tracker compatibility disable
automatic native continuation; explicit retained status remains observable.

An iteration attempts at most 16 runnable invocations per ready project, with at
most one shared checkpoint per invocation. It retains each invocation's selector,
membership policy, target, consumed attempts and limits. Explicit later-ticket
admission follows that already authorized policy within the same retained
extent; see [Workflow selection](workflow-selection.md). Later passes can
continue from those checkpoints. Independent invocations use the same repository
capacity and frontier claims as other adapters; the daemon has no execution queue.

A live owner remains exclusive. Recovery fences a dead owner through common
reconciliation and resumes only the resulting recovery pause with no operator
stop intent. An explicit operator pause, cancellation, blocker or human approval
wait remains in place. Capacity waits use their retained phase when capacity is
available. Publication and integration use the shared effective review policy;
startup never grants human approval.

The output reports observed invocations, waits and checkpoint steps. An attempted
activation prints the shared invocation identity, revision, extent, validation and
integration counts, members, candidate/PR bindings and evidence. Ownership,
capacity and stale-revision failures remain visible. A completed checkpoint or
successful manual harness does not certify integrated delivery.

Projects using tracker compatibility retain ticket-index refresh behavior.
Native-only projects do not require the legacy ticket directory or manufacture it.
Disabled, missing and invalid projects do not execute workflow work.

## Stop and recovery

`daemon stop` persists a coordinator stop request. The foreground loop checks it
between project/invocation checkpoints and exits after any current owned
checkpoint finishes. This request preserves invocation authorization and does not
cancel its work. Use `daemon workflow pause` or `cancel` to request the shared
invocation stop watcher to terminate an owned group and retain its evidence.
Ownership is released only after that group is quiescent. Resume must be explicit
for an operator pause; terminal cancellation cannot restart.

The daemon updates its heartbeat once per second during owned work and at each
iteration boundary. The heartbeat watcher is stopped before its state repository
closes. Normal completion writes stopped state; unexpected failures retain failed
coordinator state and the existing workflow checkpoint.

A fresh active coordinator state prevents another start. The default stale
threshold is five minutes. `daemon status` classifies stale active state and
persists that classification. This coordinator recovery does not recover or
replace a workflow owner; workflow reconciliation proves process exclusion
separately.

## Coordinator state

The singleton user-level state supports `stopped`, `starting`, `running`,
`stopping`, `stale` and `failed`. Active states require a positive PID and start
time. Heartbeats cannot precede startup. Stop request timestamps are valid for
stopping or stopped state.

`daemon status` reports missing state as stopped. Stored output includes status,
PID, start/heartbeat/stop/update times, failure reason and user-level data/state
paths. `daemon workflow status` instead reads the exact shared invocation,
including bounded member pagination, without modifying coordinator or workflow
state. Manual, CLI, TUI and Web status read that same evidence.
