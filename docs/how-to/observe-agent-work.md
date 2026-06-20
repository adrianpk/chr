# Observe Agent Work

Use this guide when an agent ticket run or ready-ticket queue is active and you
want to see what ChrysoCoord is doing.

## Open Dash

From the managed project:

```sh
chc
```

Open Dash with `F8`.

Dash shows:

- work readiness;
- ready-ticket queue;
- current execution or loop state;
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
chc work doctor
```

Apply safe repair from the CLI when available:

```sh
chc work doctor --fix-safe
```

## Watch The Event Timeline

Dash renders recent agent events as readable progress messages. During normal
execution, expect events for ticket selection, execution creation, runtime
preparation, worker start and finish, validation, result application, and ticket
status updates.

If the screen appears quiet, check whether readiness is blocked, whether the
current ticket is waiting for the Docker/Codex worker, or whether the execution
has already moved into report/log evidence.

## Find Reports

List execution and loop reports:

```sh
chc report list --kind execution
chc report list --kind loop
```

Filter by ticket or status:

```sh
chc report list --ticket CHC-1
chc report list --kind execution --status succeeded
```

Show a report:

```sh
chc report show --execution <execution-id>
chc report show --loop <loop-id>
```

Reports are the best first stop after work finishes. They summarize status,
runtime artifacts, validation, changed files, and stop reasons.

## Inspect Expanded Runtime Evidence

Locate logs and event stream paths for one execution:

```sh
chc work logs <execution-id>
```

Then inspect the printed paths directly. The runtime log is the closest
expanded view of what happened inside the Docker/Codex worker. The validation
log records the configured validation command output.

Summarize an event stream:

```sh
chc work events .chc/user/runtime/runs/CHC-1/<execution-id>/events.jsonl
```

In `worktree` execution mode, runtime artifacts live under the ticket
worktree's `.chc/user/runtime/runs/...` directory. Prefer the paths printed by
`chc work logs <execution-id>` when following an active run from outside that
worktree.

For run-all, summarize the loop stream when one is present:

```sh
chc work events .chc/user/runtime/runs/<loop-id>/events.jsonl
```

The event summary is useful when a stream is partially corrupted because it
counts malformed lines and still reports the latest readable state.

## CLI-Only Flow

Preview ready work:

```sh
chc work run-all --dry-run --order priority_severity
```

Run the visible ready queue:

```sh
chc work run-all --order priority_severity
```

Run one ticket:

```sh
chc work ticket CHC-1
```

Use Dash for live status while the command is active, and use reports, logs, and
event streams for post-run inspection.
