# Agent Sandbox and Runtime

Status: initial

Sources:

- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`
- `docs/reference/ticket-execution.md`

## Purpose

Define the boundary between `chc`, the runtime container, and Codex during
ticket execution.

## Execution Planes

### External Plane

The external `chc` process owns:

- project resolution;
- ticket selection;
- execution record creation;
- runtime plan computation;
- context materialization;
- Docker command rendering and invocation;
- artifact collection;
- ticket result application;
- final validation and handoff.

### Internal Plane

The runtime container owns:

- running Codex;
- reading the materialized context;
- modifying the mounted workspace according to the ticket;
- writing result artifacts;
- writing logs.

The internal plane should not decide global workflow policy.

## Runtime Injection Options

Possible ways to make `chc` behavior available in the container include:

- mount the `chc` binary;
- install `chc` inside the image;
- mount `chc` source code;
- provide a minimal bridge command;
- pass all required context and artifact paths at container startup.

The initial preference is to keep orchestration outside the container and pass a
materialized context plus artifact paths into the container.

## Secrets

Secrets and credentials must not be written into durable context artifacts.

The runtime may need access to:

- Codex credentials;
- Git credentials;
- forge credentials;
- package registry credentials.

Any secret injection must be explicit, scoped, and excluded from logs and
result artifacts.

## Workspace Mounts

The mounted workspace should expose only what ticket execution needs.

The runtime must know:

- container workdir;
- host project path;
- writable paths;
- artifact paths;
- any read-only mounts.

Git worktrees can isolate repository changes but are not a security sandbox.
Docker remains the stronger initial process and filesystem boundary.

## Git and PR Safety

The runtime may modify files, but branch, commit, push, and PR behavior should
remain controlled by `chc` policy.

The system must define:

- whether the runtime can create commits;
- whether the runtime can push;
- whether PR creation happens outside the container;
- how target branches are chosen;
- how dirty worktrees are handled.

## Logs

Logs must support debugging without leaking secrets.

Runtime logs should be stored as execution artifacts and linked from execution
records. Sensitive environment values must be redacted.

## Open Questions

- Should `chc` ever run inside the container, or should it always remain the
  external orchestrator?
- What secret injection mechanism is acceptable for local use?
- Should PR creation be allowed from inside the runtime?
- Should runtime validation run inside the same image as Codex execution?
