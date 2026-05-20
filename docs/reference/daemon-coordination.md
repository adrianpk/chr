# Daemon Coordination

## Purpose

Daemon coordination is the user-level process boundary for long-running local
work across registered projects.

The daemon lives in the `chc` binary. It is a local coordinator, not a hosted
service and not a notification transport.

## Command Boundary

The daemon command group is:

```sh
chc daemon
```

Supported skeleton commands:

- `chc daemon status`;
- `chc daemon start --dry-run`;
- `chc daemon stop --dry-run`.

The skeleton resolves the user environment and prints the user-level data and
state directories that later lifecycle slices will use.

`daemon status` is read-only and reports `stopped` until lifecycle state is
delivered.

`daemon start` and `daemon stop` require `--dry-run` until lifecycle state and
process ownership are delivered. Without `--dry-run`, they return an explicit
error instead of writing partial daemon state.

## Out of Scope

The command skeleton does not:

- write daemon state;
- acquire process ownership;
- start a background process;
- stop a running process;
- load registered projects;
- execute ticket loops;
- send notifications.

Those behaviors are delivered by later daemon coordination slices.

## Lifecycle State

Daemon lifecycle state records the local coordinator status and process
identity.

Supported statuses:

- `stopped`;
- `starting`;
- `running`;
- `stopping`;
- `stale`;
- `failed`.

Lifecycle fields:

- `status`;
- `pid`;
- `started_at`;
- `heartbeat_at`;
- `stop_requested_at`;
- `failure_reason`;
- `updated_at`.

Running process states require a positive PID and a start time. Heartbeats
cannot precede the start time. Stop requests are valid only for stopping or
stopped daemon state.
