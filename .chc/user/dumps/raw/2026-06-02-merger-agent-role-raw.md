# Merger Agent Role Raw Dump

Date: 2026-06-02
Topic: Merger agent role

## Raw Capture

The ticket review instance has been resolved. A possible next role is an
`@merger` instance, analogous to the existing `@reviewer` role.

The role is not meant to add another model, another quality review, or a large
amount of extra ceremony. In practice, ChrysoCoord works serially and the same
agentic system may operate under different responsibilities. The repository
could also be configured for automerge, so the immediate technical need is not
strong.

The value of formalizing `@merger` is the responsibility boundary. It puts a
separate operational hat on the integration step:

- `@dev` produces the change.
- `@reviewer` checks whether there are findings.
- `@merger` checks whether integration preconditions hold and executes the
  merge.

This role can stay small today while becoming an extension point later. In
larger or more mature contexts, `@merger` could become a single agent with
additional responsibility, including audit-oriented behavior.

The important distinction is that `@merger` should not re-review code. It
should perform a basic checkup and integration action:

- confirm the ticket is ready for merge;
- confirm the ticket has branch and pull request metadata;
- confirm the pull request targets the expected base;
- confirm the pull request is open and mergeable;
- confirm the configured review policy is satisfied;
- confirm no earlier repair blocks the serial workflow;
- execute the configured merge strategy;
- update the ticket after merge or leave a minimal blocking note if merge is
  not possible.

The role should avoid inventing a new Board column or a complex state machine.
The existing `ready_for_merge` status is enough as the pickup signal. A failed
merge preflight can leave a minimal local comment such as:

```text
@merger: Merge blocked. PR is not mergeable: base branch is stale.
```

The central preference is to keep the role mechanical and auditable. It should
not become a third review phase, and it should not add noise to tickets when
everything is healthy.

## Follow-Up Capture

A separate quiz is not needed for this topic. The requirements are already
clear enough to move into a formal spec, plan, and tracker.

The important missing piece is the handoff between `@reviewer` and `@merger`,
similar to the handoff already implemented between `@dev` and `@reviewer`.
This handoff is simpler because a passed review is merge authorization. There
should be no subjective second review and no normal round trip back to review.

The rule can be stated as:

```text
A passed review is an instruction to merge, subject only to objective merge
preflight failures.
```

The merger preflight is still needed, but it should be simpler than reviewer
preflight. It checks objective integration readiness:

- ticket status is `ready_for_merge`;
- review outcome is passed;
- ticket has pull request, provider, and branch metadata;
- pull request is open;
- pull request targets the expected base;
- pull request head matches the ticket branch;
- pull request is mergeable;
- no earlier repair blocks serial flow;
- configured merge strategy is allowed.

If preflight passes, `@merger` merges the pull request and marks the ticket
done. If preflight fails, `@merger` does not re-review the code. It records the
objective blocker and leaves the ticket in the appropriate attention or repair
path.
