# Observe Agent Work

Use this guide when an agent ticket run or ready-ticket queue is active and you
want to see what chr is doing.

## Open Dash

From the managed project:

```sh
chr
```

Open Dash with `F8`.

Dash shows:

- work readiness;
- ready-ticket queue;
- current execution or loop state;
- current or latest integration phase;
- latest event message or stop reason;
- recent human-readable agent events;
- recent execution and loop outcomes.

Dash refreshes from local runtime state. Changing views does not start work,
stop work, or mutate tickets by itself.

## Start Work From Dash

Use Board or Dash actions to start work:

- `enter`: start the selected ticket.
- `a` or `R`: run all visible `ready_for_dev` tickets.
- `p`: cycle priority filter for the ready queue.
- `o`: cycle ready-queue order.
- `s`: request agent stop.

Work actions ask for confirmation before starting or stopping work.

## Repair Readiness From Dash

If Dash reports blocked work readiness, it blocks ticket execution and run-all
before the runtime worker starts.

When Dash offers safe repair:

1. Press `x`.
2. Confirm the repair.
3. Wait for Dash to reload readiness.
4. Retry the same work action.

When Dash says manual action is required, inspect the same readiness state from
the CLI:

```sh
chr work doctor
```

Apply safe repair from the CLI when available:

```sh
chr work doctor --fix-safe
```

## Watch The Event Timeline

Dash renders recent agent events as readable progress messages. During normal
execution, expect events for ticket selection, execution creation, runtime
preparation, worker start and finish, validation, result application, and ticket
status updates.

In serialized integration modes, also expect events for:

- `cycle_rehydrated` after the ticket result is applied;
- `review_started` and `review_finished`;
- `repair_started` and `repair_finished` when review requests changes;
- `merge_started` and `merge_finished` when the ticket reaches merge.

If the screen appears quiet, check whether readiness is blocked, whether the
current ticket is waiting for the Docker/Codex worker, or whether the execution
has already moved into report/log evidence.

## Find Reports

List execution and loop reports:

```sh
chr report list --kind execution
chr report list --kind loop
```

Filter by ticket or status:

```sh
chr report list --ticket CHR-1
chr report list --kind execution --status succeeded
```

Show a report:

```sh
chr report show --execution <execution-id>
chr report show --loop <loop-id>
```

Reports are the best first stop after work finishes. They summarize status,
runtime artifacts, validation, changed files, and stop reasons.

## Inspect Expanded Runtime Evidence

Locate logs and event stream paths for one execution:

```sh
chr work logs <execution-id>
```

Then inspect the printed paths directly. The runtime log is the closest
expanded view of what happened inside the Docker/Codex worker. The validation
log records the configured validation command output.

Summarize an event stream:

```sh
chr work events ops/default/run/<artifact-id>/evidence/events.jsonl
```

Runtime artifacts live in the active shared namespace. Prefer the paths
printed by `chr work logs <execution-id>` when following an active run.

For run-all, summarize the loop stream when one is present:

```sh
chr work events ops/default/run/<loop-artifact-id>/evidence/events.jsonl
```

The event summary is useful when a stream is partially corrupted because it
counts malformed lines and still reports the latest readable state.

## CLI-Only Flow

Preview ready work:

```sh
chr work run-all --dry-run --order priority_severity
```

Run the visible ready queue:

```sh
chr work run-all --order priority_severity
```

Run one ticket:

```sh
chr work ticket CHR-1
```

Use Dash for live status while the command is active, and use reports, logs, and
event streams for post-run inspection.
