# Agent Sandbox and Runtime

ChrysoCoord uses Docker as the sandboxed runtime boundary for ticket execution.

The host-side `chc` process owns coordination. It resolves the project, selects
the ticket, builds the execution context, computes runtime paths, writes context
artifacts, and invokes the configured harness. The container receives explicit
inputs through the mounted project workspace and environment variables.

The container does not decide what project context matters. That decision is
made before the harness starts, and the selected context is materialized as
Markdown and supporting artifacts.

## Runtime Shape

A managed project contains runtime configuration and a project-owned runtime
Dockerfile. The default runtime is optimized for Go projects and can be edited by
the project when additional tools or dependencies are needed.

During ticket execution, the runtime receives:

- the selected ticket identity;
- the execution ID;
- the materialized worker context;
- paths for result, runtime log, and validation artifacts;
- the configured validation command;
- inherited credentials required by the harness, such as `OPENAI_API_KEY`.

The harness can then run Codex inside the container, modify the mounted
workspace, run validation, and write the result envelope expected by
ChrysoCoord.

## Harness

The harness is the runtime mechanism that performs ticket work. The current
Codex worker harness reads the worker context, executes the agent workflow, and
writes a result artifact with status, summary, and changed-file information.

Ralph loops are used where repeated validation attempts are needed. In plain
terms, a Ralph loop feeds the failure evidence from one attempt back into the
next attempt until the ticket succeeds or a configured limit is reached.

## Evidence

Execution leaves durable evidence rather than relying on transient terminal
output. Evidence can include:

- execution records;
- runtime logs;
- validation logs;
- result envelopes;
- changed-file summaries;
- ticket comments or blocker notes;
- previous-attempt context for later retries.

This evidence lets the user inspect what happened, where the agent got stuck,
and what should happen next.

## Current Boundary

The current execution model mounts the existing project workspace into the
container. The host project and the runtime view are therefore the same working
tree through the Docker mount.

This keeps the model simple and inspectable. It also means the user should avoid
editing the project while an agent run is active. Clone-based execution and
PAT-backed push from inside the container are later runtime-workspace concerns.
