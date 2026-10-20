# Agent Sandbox and Runtime

chr uses Docker as the sandboxed runtime boundary for ticket execution.

The host-side `chr` process owns coordination. It resolves the project, selects
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
chr.

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

## Workspace Boundary

chr separates the user-visible project checkout from the execution
workspace used by the runtime plan.

In `project` mode, both paths are the same directory. The host project and the
runtime view are the same working tree through the Docker mount.

In `worktree` mode, chr prepares a ticket-specific Git worktree before
the worker starts. Docker mounts that worktree into the container, so agent
edits happen outside the user's active checkout while still using normal Git
branches and local artifacts.

The host-side `chr` process still owns coordination, readiness checks, event
streams, result capture, validation, and ticket updates. The container receives
the prepared workspace and explicit context; it does not choose a different
repository boundary by itself.
