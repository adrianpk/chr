---
title: "Merger Agent Workflow"
project_id: "chryso"
case_slug: "ticket-execution-loop"
status: "accepted"
source_request_id: "chat-2026-06-02-merger-agent-role"
source_response_id: "codex-spec"
created_at: "2026-06-02T00:00:00Z"
mirror: ".prj/specs/specs/merger-agent-workflow.md"
---

# Merger Agent Workflow

This spec mirrors `.prj/specs/specs/merger-agent-workflow.md` during
Chryso self-adoption. The `.prj` spec remains the operational source for
planned implementation.

## Summary

Chryso should formalize an `@merger` role after ticket review.

`@merger` verifies objective merge readiness and executes integration. It is not
a second review phase and must not reinterpret code quality after `@reviewer`
has passed a ticket.

## Core Rule

A passed review is an instruction to merge, subject only to objective merge
preflight failures.

## Handoff

```text
@reviewer pass
  -> ticket status: ready_for_merge
  -> review outcome metadata: passed
  -> pull request reference remains authoritative

@merger pickup
  -> selects ready_for_merge tickets
  -> reads ticket pull request metadata
  -> verifies objective merge preflight
  -> merges the pull request
  -> marks the ticket done
```

## Expected Preflight

- Ticket status is `ready_for_merge`.
- Ticket review outcome is passed.
- Ticket has branch and pull request metadata.
- Provider can resolve the pull request.
- Pull request is open.
- Pull request targets the expected base branch.
- Pull request head matches the ticket branch.
- Pull request is mergeable.
- Configured review policy is satisfied.
- No earlier repair blocks serial workflow.
- Configured merge strategy is allowed.

## Operational Source

- `.prj/specs/specs/merger-agent-workflow.md`
