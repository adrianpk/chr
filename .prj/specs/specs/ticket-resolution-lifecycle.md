# Ticket Resolution Lifecycle

Status: initial

Sources:

- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`
- `docs/reference/ticket-execution.md`
- `docs/reference/ticket-result-application.md`

## Purpose

Define what it means for `chc` to resolve a ticket end to end.

Resolving a ticket is not only code generation. It is the full operational
cycle of selecting work, building context, running an agent, validating changes,
recording evidence, and updating durable project state.

## Functional Unit

The ticket is the minimum functional work unit.

Use cases, dumps, quizzes, specs, plans, and trackers may inform work,
but implementation execution should eventually reduce to a ticket that the agent
can attempt.

## Lifecycle

An initial ticket resolution flow is:

1. Select or receive a ticket.
2. Resolve the project.
3. Load the ticket title, body, metadata, checklist, comments, and blockers.
4. Determine use-case and document references.
5. Load relevant project and protocol context.
6. Build code and abstraction context.
7. Build the materialized agent context.
8. Prepare the sandboxed runtime.
9. Create or select an execution branch.
10. Create an execution record.
11. Invoke Codex inside the runtime boundary.
12. Let Codex modify files inside the allowed workspace.
13. Run validation.
14. Iterate through the Ralph Loop when validation fails and retry is allowed.
15. Apply the execution result to ticket state.
16. Record comments, validation evidence, changed files, and artifacts.
17. Commit coherent changes.
18. Prepare or open a PR when the workflow requires it.
19. Update the ticket final state.

## Ticket Prompt Base

The ticket title and body are the core prompt.

They should describe:

- the problem;
- expected behavior;
- likely system area;
- constraints;
- checklist;
- acceptance criteria;
- evidence expected from the implementation.

The prompt base is not enough by itself. The execution context must add project,
runtime, protocol, validation, and code context.

## Branch and PR Boundary

Branch and PR creation are operational boundaries.

The agent may modify files during execution, but the system should decide:

- branch naming;
- PR target branch;
- commit granularity;
- PR body sections;
- whether PR creation happens inside or outside the runtime container.

For this repository, PRs target `dev`.

## Tool Readiness

Before invoking an agent, `chc` should know whether required tools are
available.

Tool readiness may include:

- Docker availability;
- runtime image build readiness;
- Codex availability inside the runtime;
- validation command availability;
- Git availability;
- forge or PR tooling when PR handoff is expected;
- semantic code context tooling such as `gograph`, `gopls`, or Go AST loaders.

Missing required tools should block execution with a clear reason. Missing
optional tools should degrade context quality explicitly rather than failing
silently.

## Completion

A ticket resolution is complete when:

- required changes are present;
- validation has been run or a clear blocker is recorded;
- result artifacts are written;
- ticket state is updated;
- evidence is recorded;
- commit and PR handoff are prepared when required.

## Failure and Blocking

A ticket can stop as blocked when:

- required context is missing;
- validation cannot run because tooling is unavailable;
- secrets or credentials are unavailable;
- the required change is outside allowed scope;
- repeated Ralph Loop attempts fail without new useful information.

Blocked state must include a human-readable reason and enough evidence for a
future retry.

## Open Questions

- Which parts of branch and PR management should happen inside the runtime
  container?
- Should every ticket resolution create exactly one execution record?
- Should checklist completion be automatic, agent-proposed, or human-reviewed?
- What is the minimum ticket body required for autonomous execution?
- Which tools are mandatory for execution versus optional context enrichment?
