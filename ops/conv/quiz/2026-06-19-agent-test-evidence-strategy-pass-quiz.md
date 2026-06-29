# Agent Test/Evidence Strategy Pass Quiz

Status: complete

Date: 2026-06-19

Topic: Agent test and evidence strategy pass.

Raw capture:
`ops/conv/quiz/2026-06-19-agent-test-evidence-strategy-pass-quiz-raw.md`

Spec: `ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md`

## Summary

Chryso should add a mandatory Test/Evidence Strategy stage to the
development worker lifecycle. The stage should happen before code changes, be
persisted in the ticket tasklist/checklist, and be enforced by the worker prompt
so the agent treats it as an operational contract.

The goal is evidence-first development, not universal TDD. The agent must decide
what evidence is appropriate for the ticket, including whether new tests,
Gherkin, validation, or runtime proof are required. The user may edit or override
the checklist, but the normal workflow should not require the user to author the
strategy manually.

## Decisions

- Every development worker run must begin with a Test/Evidence Strategy stage
  before modifying project files.
- The primary persisted strategy belongs in the ticket tasklist/checklist.
- Execution reports may duplicate details as trace evidence, but they are not
  the primary operational source.
- Free-form comments are acceptable only when Chryso can recognize their
  semantics later.
- The worker prompt must explicitly instruct the agent to read, obey, update,
  and reconcile the strategy.
- The checklist should be highly parseable while remaining readable for humans.
- The user may edit the strategy, but the agent owns strategy creation by
  default.
- If Gherkin is active and applies to observable behavior, the agent must create
  or update the `.feature` artifact.
- Missing Docker or Codex runtime is a severe operational blocker.
- A ticket must not move to `ready_for_review` with unresolved linting,
  compilation, runtime, or required validation failures.
- Configuration should exist from the start with preferred sane defaults.
- Documentation must explain how a project changes the default behavior.

## Preferred Checklist Shape

The ticket checklist should have stable labels and compact natural-language
values. A first draft shape:

```md
- [ ] Evidence strategy: <short decision>
- [ ] Gherkin: <required|not-applicable|skipped> - <reason>
- [ ] Tests: <none|unit|table|integration|fuzz|property|mutation|runtime> - <reason>
- [ ] Validation: <commands/evidence>
- [ ] Evidence summary: pending
```

The exact syntax may change during implementation, but the parseable labels and
status semantics should remain stable.

## Operational Contract

The development worker should use this lifecycle:

```text
ticket picked
-> read ticket and existing checklist
-> write or update Test/Evidence Strategy checklist items
-> implement
-> validate
-> update Evidence Summary
-> move ticket only if closure conditions pass
```

The strategy is not advisory prose. It is part of the worker contract.

## Open Questions

- Exact Markdown syntax for the persisted checklist.
- Exact parser behavior for existing user-edited checklist items.
- Exact config keys for strategy mode, Gherkin policy, and validation strictness.
- Exact execution report location and whether the report should link back to the
  checklist task IDs.
- How Chryso detects that Gherkin is active for a project.
- Whether the first implementation should only support inline strategy or also
  include a disabled separate-pass mode for future use.
