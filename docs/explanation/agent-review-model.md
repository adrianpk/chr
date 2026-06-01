# Agent Review Model

ChrysoCoord does not model agent review as a direct copy of human pull request
review.

Human review exists for many reasons that do not transfer cleanly to agentic
work: attention limits, ownership boundaries, team communication, fatigue, and
shared accountability. Agents fail in different ways. The important risks are
usually contextual rather than social: incorrect context, drift from the spec,
unexpected scope expansion, weak evidence, broken contracts, and incoherent
integration after several small slices.

For that reason, review is treated as an integration checkpoint.

The useful question is not whether an agent can approve another agent's work.
It is whether the accumulated changes still match the spec, stayed within
scope, preserved contracts, and left enough evidence for a human to decide what
gets promoted.

## Slice Flow

Small slices should stay fast.

Each slice can rely on local quality gates, execution evidence, and ticket
state. Tests, linters, hooks, and validation commands are quality controls, not
human-style review. They prevent obviously broken work from moving forward
without requiring a human to inspect every small branch.

This keeps the agent loop useful. A workflow that requires human review every
few minutes turns the user into a bottleneck and reduces the value of
orchestration.

## Integration Checkpoint

The review point belongs around accumulated integration work:

- before promoting `dev` to `main`;
- when a user asks for an explicit review pass;
- after a batch of slices completes;
- before tagging or publishing a meaningful state.

At that point, a review pass should look at the whole batch:

- included tickets and slices;
- changed areas;
- possible scope drift;
- contract or persistence changes;
- missing or weak evidence;
- validation results;
- recommended files or behaviors for human inspection.

The output should help the human review less and better. It should not pretend
to be a final authority.

## Role of Agent Review

An agent reviewer is useful as an auditor of coherence and drift.

It can compare the spec, plan, tickets, execution reports, and code changes. It
can point out suspicious changes, missing evidence, or mismatches between what
the ticket claimed and what the implementation changed.

That review should be evidence for promotion, not a human ritual recreated with
models.
