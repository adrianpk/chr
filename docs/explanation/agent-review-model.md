# Agent Review Model

chr does not model agent review as a direct copy of human pull request
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

## Ticket-Level Review

Ticket-level review is the lightweight checkpoint for one branch and pull
request before the ticket proceeds toward integration into the configured
integration branch.

This checkpoint is operational rather than ceremonial. The ticket stores the
branch, pull request reference, review mode, review outcome, and repair handoff
metadata. The remote pull request remains the place for provider-native review
comments when the configured provider supports them.

A clean review pass should avoid extra local comment noise. A review with
findings should mark the ticket as needing repair by returning it to
`ready_for_dev`, so the next agent work pass fixes that ticket before continuing
serial slice work that assumes the development branch has advanced.

The exact ticket metadata and state transitions are reference contracts:

- [Ticket review workflow](../reference/ticket-review-workflow.md)
- [Ticket review metadata](../reference/ticket-review-metadata.md)

## Merger Role

The merger role is a separate operational role after a clean ticket review
pass.

`@dev` creates the branch and pull request. `@reviewer` inspects that pull
request and either requests repair or records a pass. `@merger` receives only
passed tickets in `ready_for_merge`, checks objective integration conditions,
and merges the pull request.

This role boundary is useful even when the same underlying model operates each
role. It narrows the question the agent is allowed to answer:

- `@dev`: what change satisfies the ticket?
- `@reviewer`: did this branch satisfy the ticket with enough evidence?
- `@merger`: is the already-reviewed pull request objectively mergeable?

`@merger` should not re-run subjective code review, reopen style disagreements,
or create a second approval ritual. It should block only on concrete integration
problems such as unreadable provider state, wrong base branch, head mismatch,
non-mergeability, pending serial repair work, or invalid merge strategy.

The merge role gives the workflow a stable audit point without turning every
successful review into extra local comments. Passed reviews normally merge
quietly; blocked merges leave compact ticket metadata.

The merge contract is described in
[Ticket merge workflow](../reference/ticket-merge-workflow.md).

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

The future enforcement path for that evidence is described in
[Dev to main gate direction](dev-main-gate-direction.md).
