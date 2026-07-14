---
title: "Merger Agent Workflow"
project_id: "chryso"
case_slug: "ticket-execution-loop"
status: "accepted"
source_request_id: "chat-2026-06-02-merger-agent-role"
source_response_id: "codex-spec"
created_at: "2026-06-02T00:00:00Z"
mirror: ".chr/default/specs/merger-agent-workflow.md"
---

# Merger Agent Workflow

Chryso should formalize an `@merger` role after ticket review.

This role is an operational integration role. It is not a second review phase
and must not reinterpret code quality after `@reviewer` has passed a ticket.

This spec is mirrored in `.chr/default/specs/merger-agent-workflow.md` while the
project evaluates self-hosted Chryso workflow adoption.

## Problem

Ticket review can move work from `ready_for_review` to `ready_for_merge`, but
the merge step still needs a formal handoff and objective preflight.

Without a merger role, merge authority can remain implicit in the development
agent, the reviewer, or repository automerge settings. That weakens audit
trail clarity and leaves no stable extension point for later merge policy,
provider checks, or integration auditing.

## Role Boundary

The agent roles are:

- `@dev` produces branch and pull request work.
- `@reviewer` inspects the ticket pull request and records pass or findings.
- `@merger` verifies objective merge readiness and executes integration.

`@merger` may be operated by the same underlying agent model as the other
roles. The value is the protocol boundary: the role constrains what the agent
is allowed to decide.

## Core Rule

A passed review is an instruction to merge, subject only to objective merge
preflight failures.

`@merger` must not send a passed ticket back to review or development because
of style preference, subjective judgment, or a second reading of the diff.
It may block only when preflight exposes an objective integration problem.

## Handoff

The `@reviewer -> @merger` handoff is deterministic:

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

Unlike the `@reviewer -> @dev` repair handoff, the normal passed-review path has
no round trip. A passed review becomes merge authorization.

## Expected Preflight

The merger preflight should check:

- ticket status is `ready_for_merge`;
- ticket review outcome is passed;
- ticket has branch and pull request metadata;
- provider can resolve the pull request;
- pull request is open;
- pull request targets the expected base branch;
- pull request head matches the ticket branch;
- pull request is mergeable;
- configured review policy is satisfied;
- no earlier repair blocks serial workflow;
- configured merge strategy is allowed.

The preflight is simpler than review preflight. It checks integrability, not
quality.

## Outcomes

When preflight passes:

- merge the pull request using the configured merge strategy;
- update ticket merge metadata;
- move the ticket to `done`;
- keep local comments quiet unless the workflow explicitly requires an audit
  note.

When preflight fails:

- do not merge;
- record the objective blocker;
- keep the ticket in `ready_for_merge` when the blocker is operational;
- return the ticket to `ready_for_dev` only when the blocker requires code or
  branch repair;
- do not create a new Board column for merger failure.

Examples of operational blockers:

- missing provider credentials;
- provider unavailable;
- pull request closed or missing;
- missing ticket PR metadata;
- wrong PR base branch;
- configured merge strategy unavailable.

Examples of repair blockers:

- merge conflict;
- branch stale and requiring code rebase or regeneration;
- PR head no longer matches the ticket branch.

## Provider Integration

The merger role should reuse the provider model introduced for ticket review.

Forgejo and Codeberg remain the priority providers. GitHub remains useful for
ubiquity. Provider logic should normalize only the fields needed for merge
readiness and execution before adding broader abstractions.

Provider operations include:

- resolve pull request by ticket metadata;
- read open/closed state;
- verify base and head branches;
- inspect mergeability when the provider exposes it;
- execute configured merge strategy;
- report provider-specific failures as machine-readable reasons.

## Noise Policy

Successful merges should not create ceremonial comments.

The ticket status and merge metadata are enough for normal audit. Local comments
are reserved for blocked merge attempts, unexpected provider failures, or
explicitly configured audit mode.

## Non-Goals

This workflow does not include:

- a new code review pass after `@reviewer` passes;
- parallel agent execution in one repository workspace;
- replacing the separate `dev -> main` release alignment gate;
- a new Board column for merger failures;
- automatic merge of unrevised work unless the configured workflow mode
  explicitly allows trusted auto-merge.

## Documentation Impact

Durable documentation should explain:

- the role boundary between `@dev`, `@reviewer`, and `@merger`;
- the `@reviewer -> @merger` handoff;
- merger preflight reasons;
- how blocked merge attempts are represented in ticket metadata;
- how this differs from the `dev -> main` release gate.
