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

`daemon status` is read-only. It reports `stopped` when no daemon state exists,
or prints the stored daemon status, process ID, lifecycle timestamps, failure
reason, and user-level state paths.

`daemon start` and `daemon stop` require `--dry-run` until lifecycle state and
process ownership are delivered. Without `--dry-run`, they return an explicit
error instead of writing partial daemon state.

`daemon start --dry-run` evaluates the single-instance guard and reports whether
a future start would be allowed.

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

## State Persistence

Daemon lifecycle state is persisted in the user-level SQLite database.

The daemon state repository stores one singleton row for the local user
environment. It supports:

- read current daemon state;
- upsert daemon state;
- update heartbeat time;
- clear daemon state.

The repository does not start, stop, or signal processes. It only persists the
lifecycle state that later daemon slices use for process ownership and status
reporting.

## Single-Instance Guard

The daemon start guard prevents duplicate local daemon ownership.

The guard allows start when:

- daemon state is missing;
- daemon state is `stopped`;
- daemon state is `failed`;
- daemon state is already classified as `stale`;
- active daemon state has not heartbeated within the stale threshold.

The guard rejects start when daemon state is freshly `starting`, `running`, or
`stopping`.

The default stale threshold is five minutes. Stale classification in this slice
is a decision result only; persisted stale-state mutation is handled by a later
recovery slice.

## Project Loading

Daemon project loading reads registered projects from the user-level project
registry.

Only enabled projects are returned for daemon work.

Each enabled project is classified before work starts:

- `ready`: project path exists, is a directory, and contains `.chc/`;
- `missing`: project path does not exist;
- `invalid`: project path is not usable for daemon work.

Invalid classifications include non-directory project paths and missing
workspace directories. Later daemon run-loop slices can skip or report these
projects without failing the whole daemon process.

## Foreground Run Loop

`chc daemon start` runs daemon coordination in the foreground.

The foreground run:

- evaluates the single-instance guard;
- writes running daemon state with the current process ID;
- updates heartbeat state once per iteration;
- loads enabled registered projects;
- skips projects that are not ready;
- refreshes ticket indexes for ready projects;
- stops after `--max-iterations`;
- writes stopped daemon state on normal completion.

`--max-iterations` is required to be greater than zero and defaults to `1`.

The foreground loop does not run as an OS service and does not observe stop
requests yet. Graceful stop handling is delivered by the stop and stale recovery
slice.

## Status Reporting

`chc daemon status` reads the persisted daemon state.

When state is missing, the output includes:

- `daemon: stopped`;
- `state: missing`.

When state exists, the output includes:

- daemon status;
- `state: stored`;
- PID;
- start time;
- heartbeat time;
- stop request time;
- failure reason;
- update time;
- data and state directories.
