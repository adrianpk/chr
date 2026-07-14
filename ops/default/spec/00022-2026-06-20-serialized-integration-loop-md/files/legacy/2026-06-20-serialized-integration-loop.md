# Serialized Integration Loop

Status: planned

Date: 2026-06-20

Model: `ops/specs/2026-06-20-serialized-integration-loop-model.md`

Related:

- `ops/specs/2026-06-18-agentic-loop-review-merge.md`
- `ops/specs/2026-06-19-isolated-ticket-worktrees.md`
- `docs/reference/ticket-loop.md`
- `docs/reference/ticket-review-workflow.md`
- `docs/reference/ticket-merge-workflow.md`
- `docs/reference/review-workflow-config.md`
- `docs/explanation/agent-review-model.md`

## Summary

Chryso's ready-ticket loop must become a serialized integration loop.

The loop must not treat development as the whole unit of progress. It must treat
one ticket integration cycle as the unit:

```text
select next eligible ticket
develop
review or apply trusted-review policy
merge or block
rehydrate durable state
select the next eligible ticket
```

The outer loop may process many tickets, but it must process only one active
ticket integration cycle at a time. The next normal ticket may start only after
the current ticket is merged into the configured integration branch and recorded
as done, or after the current ticket reaches a blocked state.

This is a prerequisite for future task parallelization. Parallel workers may
exist later, but integration must remain serialized unless an explicit future
policy introduces a different integration unit.

## Problem

Current `run-all` behavior selects a ready queue and processes multiple
`ready_for_dev` tickets sequentially. That is serial development execution, but
it is not serialized integration.

The current loop can produce:

```text
ticket 1 developed
ticket 2 developed
ticket 3 developed
```

before ticket 1 has completed review and merge. That allows later tickets to be
implemented against an integration branch that has not yet absorbed the earlier
work. It also creates conflict risk, unclear ordering, and weak foundations for
future parallel execution.

The target behavior is:

```text
ticket 1 developed
ticket 1 reviewed
ticket 1 merged
ticket 2 developed
ticket 2 reviewed
ticket 2 merged
```

or:

```text
ticket 1 developed
ticket 1 reviewed
ticket 1 needs repair
ticket 1 repaired
ticket 1 reviewed again
ticket 1 merged
ticket 2 developed
```

If ticket 1 cannot pass review or merge, the loop stops before ticket 2.

## Goals

- Make the automatic loop a loop of ticket integration cycles.
- Preserve one active agent task at a time.
- Preserve one active merge into the integration branch at a time.
- Rehydrate ticket, branch, review, merge, and worktree state between cycles.
- Require each normal ticket to reach `done` before the next normal ticket
  starts.
- Stop the loop when a ticket requires human review, has insufficient evidence,
  exceeds repair limits, or cannot merge.
- Keep review high-level and integration-focused.
- Prevent review from blocking progress on cosmetic or preference-only issues.
- Preserve explicit single-ticket execution.
- Preserve configurable review modes.
- Make the selected loop mode visible in dry-run output, event streams, reports,
  and Dash.
- Leave human `dev -> main` promotion separate from ticket integration.

## Non-Goals

- Do not introduce parallel execution in this delivery.
- Do not introduce slice-level or batch-level integration units.
- Do not replace the `dev -> main` promotion gate.
- Do not require human review for every ticket.
- Do not let the review role mutate code.
- Do not let review become a style or nitpicking gate.
- Do not continue to later tickets after a failed review or blocked merge.
- Do not hide skipped or blocked tickets.
- Do not depend on agent memory to know which ticket is next.

## Loop Modes

The serialized integration loop uses existing review workflow configuration as
the policy source:

```yml
review:
  mode: human_review
```

Supported modes keep their existing names but gain explicit loop behavior.

### `human_review`

The loop develops one ticket, attaches branch and pull request evidence, moves
the ticket to `ready_for_review`, records the handoff, and stops.

The loop must not continue to the next normal ticket while human review is
pending.

### `agentic_review`

The loop develops one ticket, runs a high-level agentic review, applies the
review result, merges approved work, records merge metadata, rehydrates state,
and then selects the next ticket.

If review requests repair, the same ticket, branch, pull request, and worktree
are reused. The repair cycle is finite.

### `trusted_auto_merge`

The loop develops one ticket and skips the separate reviewer role, but it still
must merge the ticket before continuing to the next normal ticket.

This mode trusts the development evidence and validation command. It does not
allow a queue of unmerged developed tickets.

## Ticket Integration Cycle

One ticket integration cycle has these phases:

```text
selection
development
development result application
review handoff or review execution
repair when needed
merge preflight
merge execution
cycle finalization
state rehydration
```

Only one phase may mutate ticket state at a time.

The operation lock remains the workspace-level guard against concurrent
development, review, and merge mutation.

## Selection

The loop selects exactly one ticket at the beginning of each cycle.

Selection must:

- load tickets from durable ticket state;
- apply readiness policy;
- prioritize pending review repair for the same ticket before later work;
- preserve configured queue filters;
- preserve deterministic ordering inside the eligible set;
- report selected and skipped tickets.

The selected ticket list for one cycle has length one.

The loop must not keep using a stale multi-ticket queue after a ticket finishes.
It must rehydrate state and re-run selection after each completed or blocked
cycle.

## Ordering

Default ordering should preserve the existing behavior until plan-order metadata
is available:

1. plan order when recorded;
2. slice order when recorded;
3. ticket creation order;
4. ticket key as a final tie-breaker.

Priority, severity, type, area, tag, and use-case filters may narrow the
eligible set. They must not silently replace the canonical order unless the user
explicitly selects an alternate order flag.

## Development Phase

Development executes the selected ticket through the existing ticket execution
path.

Development must produce or preserve:

- execution record;
- runtime logs;
- validation output;
- result envelope;
- changed-file summary;
- ticket branch metadata;
- pull request review reference when configured for provider-backed review or
  merge.

The development phase must not be considered complete for loop continuation
unless the ticket has enough evidence for the selected review mode.

## Review Phase

Review is an integration-level gate. It is not a human-style nitpicking pass.

The reviewer asks:

- Did the implementation satisfy the ticket objective?
- Is the diff plausibly within scope?
- Is there credible validation evidence?
- Did the change avoid obvious contract, persistence, runtime, or data loss
  risks?
- Did the work contradict the active spec, plan, or repository invariants?
- Is the pull request and branch evidence readable enough for merge?

The reviewer must not block on:

- minor formatting that passed configured linting;
- naming preferences that do not violate an existing convention;
- wording preferences outside the ticket scope;
- optional refactors;
- speculative future improvements;
- stylistic comments that can be recorded as non-blocking notes.

Blocking findings must be objective and tied to correctness, scope, evidence,
contracts, safety, or mergeability.

Non-blocking comments may be recorded only when useful. They must not prevent
merge.

## Review Outcomes

Supported review outcomes:

- `pass`: ticket moves to `ready_for_merge`.
- `needs_changes`: ticket returns to `ready_for_dev` with repair handoff
  metadata.
- `insufficient_evidence`: loop stops with a blocked cycle and records missing
  evidence.
- `human_required`: loop stops with a blocked cycle when the reviewer cannot
  decide safely.

`needs_changes` must point to a review artifact or provider comments. The
repair worker must receive that handoff as context.

## Repair Limit

The default maximum repair cycle count is one.

The first repair cycle may return the same ticket to development. If the review
still fails after the allowed repair cycle, the ticket must become blocked and
the loop must stop.

The repair count must be durable. It cannot depend on process memory.

## Merge Phase

Merge phase runs only after the ticket is eligible under the selected mode.

For `agentic_review`, merge requires:

- `status: "ready_for_merge"`;
- review result `pass`;
- branch metadata;
- pull request reference;
- merge preflight pass.

For `trusted_auto_merge`, merge requires:

- development evidence;
- validation success;
- branch metadata;
- pull request reference when provider-backed merge is configured;
- merge preflight pass.

Merge success moves the ticket to `done` and records merge metadata.

Merge failure keeps the ticket in the appropriate pre-merge state, records an
objective blocker, and stops the loop before later work begins.

## State Rehydration

After every ticket cycle, the loop must reload durable state before selecting
the next ticket.

State rehydration must include:

- ticket index refresh;
- ticket file state;
- execution records;
- review metadata;
- merge metadata;
- worktree metadata;
- configured work and review policy.

This prevents stale in-memory queues from driving later ticket execution.

## Event Stream

The event stream must show both loop levels:

- outer loop events for queue selection and next-ticket decisions;
- inner cycle events for development, review, repair, merge, and finalization.

Human-readable Dash output should make clear whether the loop is:

- developing a ticket;
- reviewing a ticket;
- repairing a ticket;
- merging a ticket;
- waiting for human review;
- blocked;
- selecting the next ticket.

## Reports

Loop reports must include:

- selected mode;
- ordering policy;
- selected ticket sequence;
- per-ticket cycle status;
- review outcome;
- repair count;
- merge outcome;
- blocked reason when present;
- skipped ticket count and reason summary.

A successful loop means every processed ticket reached `done`.

A blocked loop means the active ticket could not safely finish its integration
cycle. Later tickets were intentionally not started.

## CLI Behavior

`chr work run-all` should use serialized integration cycles.

`chr work run-all --dry-run` should show:

- review mode;
- integration unit: `ticket`;
- cycle policy;
- first selected ticket;
- remaining eligible ticket count;
- whether the loop will stop for human review.

`chr work ticket <ticket-key>` remains explicit single-ticket development. It
does not silently redefine loop ordering. It may still be subject to repair
precedence and branch readiness preflight.

## TUI Behavior

Dash should show the active cycle phase.

Board should continue using existing ticket columns. The integration cycle does
not require new Board columns.

When the loop stops for human review, Dash should explain that this is expected
under `human_review` mode and should not present it as an execution failure.

## Acceptance Criteria

- `run-all` selects one ticket for one integration cycle at a time.
- The loop rehydrates durable state after each ticket cycle.
- The next normal ticket starts only after the previous ticket is `done`.
- `human_review` mode stops after creating the review handoff.
- `agentic_review` mode runs high-level review before merge.
- `trusted_auto_merge` mode still merges before continuing.
- Review blocks only on objective correctness, scope, evidence, contract,
  safety, or mergeability issues.
- Review does not block on minor style issues that pass configured validation.
- Review repair reuses the same ticket, branch, pull request, and worktree.
- Repair is finite and defaults to one repair cycle.
- A blocked review or merge stops the loop before later work begins.
- Event streams and reports expose outer loop and inner cycle phases.
- Tests prove the previous multi-ticket dev-only behavior no longer occurs in
  integrated loop mode.
- Tests prove a two-ticket queue processes ticket 2 only after ticket 1 is
  merged.
- Tests cover human-review stop, agentic review pass, agentic review repair,
  repair limit, trusted auto-merge, merge blocked, and state rehydration.

## Open Questions

- Should `run-all` always use serialized integration cycles immediately, or
  should a temporary compatibility flag preserve dev-only behavior?
- Should `chr work ticket <ticket-key>` remain development-only, or should a new
  command run exactly one full integration cycle?
- Should review artifacts live under `.chr/user/reviews/`, under the execution
  run directory, or both?
- Should project templates switch default review mode from `human_review` to
  `trusted_auto_merge` or `agentic_review` after this workflow is proven?
