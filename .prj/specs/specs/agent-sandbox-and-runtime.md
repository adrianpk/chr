# Agent Sandbox and Runtime

Status: initial

Sources:

- `/home/adrian/Downloads/chr_ticket_execution_context_and_ralph_loop.md`
- `docs/reference/ticket-execution.md`

## Purpose

Define the boundary between `chr`, the runtime container, and Codex during
ticket execution.

## Execution Planes

### External Plane

The external `chr` process owns:

- project resolution;
- ticket selection;
- execution record creation;
- runtime plan computation;
- context materialization;
- context source collection, pruning, and summarization;
- semantic code context query execution;
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

The internal plane should not rediscover the full project from scratch when
`chr` can provide durable context artifacts. It may inspect source files during
execution, but project protocol, ticket intent, related documents, semantic code
context, validation expectations, and Ralph Loop history should be prepared by
the external plane and mounted into the container as artifacts.

## Context Composer Compatibility

The context composer is an external-plane responsibility.

It must produce Docker-compatible artifacts before the runtime container starts:

- `context.md`, the main prompt consumed by `codex exec`;
- optional source inventory artifacts used to justify selection;
- selected project protocol excerpts, including applicable `AGENTS.md` files;
- selected `.chr/` workflow and runtime context;
- selected ticket/use-case document context;
- selected semantic code context, initially from `gograph` CLI output;
- previous Ralph Loop attempt summaries and validation failures when retrying;
- output contract and artifact paths.

The Docker runtime should receive these artifacts through the existing project
workspace mount and path environment variables. The composer should not require
`chr` to run inside the container for the initial implementation.

The runtime remains responsible for executing Codex against the mounted
workspace, while `chr` remains responsible for deciding what context is
materialized and how much of it is exposed to the worker.

## Runtime Injection Options

Possible ways to make `chr` behavior available in the container include:

- mount the `chr` binary;
- install `chr` inside the image;
- mount `chr` source code;
- provide a minimal bridge command;
- pass all required context and artifact paths at container startup.

The initial preference is to keep orchestration outside the container and pass a
materialized context plus artifact paths into the container.

## Secrets

Secrets and credentials must not be written into durable context artifacts.

The runtime may need access to:

- Codex credentials;
- `OPENAI_API_KEY`;
- Git credentials;
- forge credentials;
- package registry credentials.

Any secret injection must be explicit, scoped, and excluded from logs and
result artifacts.

For the local Docker runtime, Codex credentials may be provided by read-only
mounts such as `~/.codex/auth.json` and `~/.codex/config.toml`. Environment
variables that carry secrets should be inherited by name, not rendered with
their values into durable command artifacts.

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

Because Docker is the sandbox boundary, Codex may run inside the container with
its own filesystem sandbox disabled when nested namespace creation is not
available. This avoids bubblewrap failures while preserving the outer Docker
boundary.

## Mounted Workspace Ownership

In mounted workspace mode, Docker does not clone the repository. It mounts an
existing host workspace into the container and Codex edits that mounted
filesystem directly.

Operationally, the mounted workspace belongs to the agent while execution is in
progress. The user should not edit the same workspace until the execution
finishes.

Concurrent user edits are a worktree concurrency problem, not a Docker or Codex
problem. The initial MVP may rely on this operational rule. Later hardening can
add guardrails:

- detect dirty worktrees before execution starts;
- create an execution lock while the agent is running;
- record pre-run git status and diff;
- compare post-run status against the pre-run snapshot;
- detect external drift during execution;
- abort or mark execution blocked when unexpected concurrent edits appear;
- use an ephemeral git worktree for stronger isolation when needed.

## Git and PR Safety

The runtime may modify files, but branch, commit, push, and PR behavior should
remain controlled by `chr` policy.

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

- Should `chr` ever run inside the container, or should it always remain the
  external orchestrator?
- What secret injection mechanism is acceptable for local use?
- Should PR creation be allowed from inside the runtime?
- Should runtime validation run inside the same image as Codex execution?
