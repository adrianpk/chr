# Ticket Execution Context Hardening

Status: initial

Sources:

- `/home/adrian/Downloads/context_composition_and_document_transitions.md`
- `/home/adrian/Downloads/chc_ticket_execution_context_and_ralph_loop.md`
- `.prj/drafts/chc-feature-slicing-set-map.md`

## Purpose

Define the post-cutoff hardening block that strengthens autonomous ticket
execution before notification and remote command surfaces.

The operational cutoff is already reached: `chc` can resolve projects, select
tickets, run Codex through the runtime container path, capture results, apply
ticket updates, execute foreground loops, and coordinate daemon state.

This hardening block addresses a different gap: the semantic quality and
auditability of the context given to the agent before it edits code.

## Problem

The current system has the execution mechanics, but the policy for composing
agent context is not yet fully formalized.

Without that policy, a ticket execution can still rely too much on the agent
rediscovering project structure, code relationships, validation expectations,
and workflow rules at execution time.

## Scope

This hardening block includes:

- layered ticket execution context;
- document-to-spec-to-ticket transitions;
- ticket resolution lifecycle;
- semantic code context extraction;
- Ralph Loop retry policy;
- sandbox and runtime control policy;
- ticket loop execution policy refinements;
- execution evidence and handoff contracts.

## Related Specs

- `.prj/specs/specs/ticket-execution-context.md`
- `.prj/specs/specs/context-composition-pipeline.md`
- `.prj/specs/specs/ticket-resolution-lifecycle.md`
- `.prj/specs/specs/semantic-code-context.md`
- `.prj/specs/specs/ralph-loop.md`
- `.prj/specs/specs/agent-sandbox-and-runtime.md`
- `.prj/specs/specs/ticket-loop-execution-policy.md`
- `.prj/specs/specs/execution-evidence-and-handoff.md`

## MVP Relationship

This is post-cutoff work, but it is still MVP-hardening work.

It does not invalidate the cutoff. It improves the reliability of the product
shape already reached by making ticket execution more reproducible, explainable,
and less dependent on ad hoc prompt construction.

## Implementation Direction

Likely implementation work should become a dedicated slicing before
`Notification and Remote Commands`.

Candidate implementation themes:

- richer worker context materialization;
- ticket readiness checks;
- document reference loading;
- code context selection and summarization;
- `gograph` or Go AST integration;
- Ralph Loop attempt records and stop rules;
- evidence aggregation for ticket and PR handoff.

## Out of Scope

This hardening block does not include:

- Telegram or email transport;
- remote command parsing;
- multi-user execution;
- hosted execution services;
- replacing the existing ticket, runtime, or daemon contracts.

## Open Questions

- Which specs in this block should be implemented first?
- Should code context extraction be introduced before Ralph Loop records?
- Should context hardening produce durable artifacts before every execution?
- How much of the materialized context should be stored in execution records
  versus runtime artifacts?
