# Agentic Loop Review and Merge Workflow

Status: draft

Date: 2026-06-18

Quiz: `ops/conv/quiz/2026-06-18-agentic-loop-review-merge-quiz.md`

Raw quiz capture: `ops/conv/quiz/2026-06-18-agentic-loop-review-merge-quiz-raw.md`

Related:

- `docs/reference/ticket-loop.md`
- `docs/explanation/agent-review-model.md`
- `docs/reference/ticket-review-workflow.md`
- `docs/reference/ticket-merge-workflow.md`
- `docs/reference/work-readiness.md`

## Summary

ChrysoCoord should support a foreground loop that continues from ticket
development into pull request evidence and merge, while preserving predictable
ordering and explicit project policy.

The default path for new projects is automatic merge after strict development
validation. In this path, ChrysoCoord does not launch separate reviewer and
merger agent roles. The development agent implements the ticket, creates or
updates the branch and pull request, produces validation evidence, and merges the
pull request when repository policy allows it.

The formal reviewer/merger role cycle remains a supported configurable
alternative. It is useful for projects that want provider-backed review comments,
separate review artifacts, or a more traditional pull request process.

The default workflow is:

```text
select next ticket
develop
create or update branch and pull request evidence
validate through the project validation command
merge automatically when policy allows it
continue
```

The formal alternative is:

```text
select next ticket
develop
mark ready for review
review
request changes or approve
merge approved work
continue
```

## Goals

- Make automatic loop execution preserve canonical order unless the user
  explicitly selects a different ordering policy.
- Use ticket as the first implementation unit: one ticket maps to one branch,
  and one branch maps to one pull request.
- Make automatic merge the default loop path for new projects.
- Keep pull requests as inspectable evidence even when the loop merges
  automatically.
- Rely on the project `validation_command` as the merge gate, while encouraging
  initial project setup to establish a strong validation baseline.
- Keep the formal reviewer/merger path configurable, documented, and heavily
  tested.
- Keep review evidence local-first, with forge provider comments and remote diffs
  as optional integrations.
- Prevent infinite development-review repair cycles with explicit round limits
  and blocked outcomes.
- Preserve the ability to run one explicit ticket outside the automatic loop.

## Non-Goals

- Do not implement slice-level or batch-level integration units in the first
  workflow.
- Do not require the formal reviewer/merger roles in the default automatic merge
  path.
- Do not require GitHub-specific or Forgejo-specific provider APIs for the
  default path.
- Do not replace human review for `dev` to `main` promotion.
- Do not let a reviewer role mutate code produced by the development role.
- Do not make the merger role re-run validation only for ceremony.
- Do not allow automatic merge when branch policy, pull request state, or
  validation evidence is unreadable.
- Do not rely on agent memory to know where the loop is. The loop must rehydrate
  its state from durable artifacts.

## Desired Behavior

### Canonical Loop Order

Automatic loop mode should process work in deterministic order.

Default ordering:

1. plan order;
2. slice order inside the plan;
3. ticket creation/order inside the slice;
4. ticket key as final tie-breaker.

Priority, severity, type, area, tag, and use-case filters may narrow the
eligible set. Filtering must not silently replace canonical order. For example,
`--priority high` selects only high-priority work and then preserves canonical
order among that subset.

Explicit alternative ordering is allowed only when requested through a visible
policy or command flag, such as `--order severity`.

Running one explicit ticket remains separate from the loop:

```text
chc work ticket CHC-12
```

That command executes the requested ticket subject to normal readiness and
repair-precedence checks. It does not redefine the automatic loop order.

### Ticket Integration Unit

The first implementation unit is the ticket.

The expected mapping is:

```text
ticket -> branch -> pull request -> merge
```

Slice names may appear naturally in ticket keys, branch names, or pull request
titles, but slice does not become a separate integration unit in the first
workflow.

Future policy may add slice-level or batch-level integration units. That future
policy must be explicit and visible in preview and run summaries.

### Default Automatic Merge Path

Automatic merge is the default path for new projects.

In this path:

- the development agent implements the ticket;
- the development agent creates or updates the ticket branch and pull request;
- the pull request remains available as evidence for human inspection;
- ChrysoCoord records branch, pull request, validation, and merge evidence
  locally;
- ChrysoCoord merges the pull request automatically when repository policy,
  validation evidence, and provider preflight allow it;
- no separate reviewer or merger agent role is launched.

The development agent must be strict about implementation quality. A mergeable
implementation includes, as applicable to the project:

- unit tests;
- table-driven tests;
- fuzz or property-style tests when useful;
- stronger quality checks such as CRAP or mutation-oriented checks when feasible;
- linting;
- validation through the project `validation_command`;
- evidence that the application can run when runability is relevant.

The `validation_command` remains the immediate enforcement mechanism. During
project initialization or baseline setup, ChrysoCoord should help projects define
the strongest practical validation command the stack can support.

### Development Completion

When development creates or updates a pull request for the selected ticket,
ChrysoCoord must attach the branch and review reference to the ticket metadata.

In the default automatic merge path, the ticket may proceed directly to merge
preflight after validation evidence is available.

When automatic merge is disabled, unavailable, or blocked by policy, the ticket
must move to `ready_for_review` if a pull request is waiting for review.

The move to `ready_for_review` must fail if the ticket lacks:

- implementation branch;
- review reference;
- execution evidence;
- validation result, unless policy explicitly allows review with missing
  validation.

### Formal Reviewer/Merger Mode

The formal path is configurable and supported, but it is not the default path.

The formal path uses operational roles:

- `dev`: implement the selected ticket and create or update the configured review
  surface.
- `reviewer`: inspect the result and decide whether it passes, needs changes,
  lacks evidence, or needs a human decision.
- `merger`: merge only already-approved work after objective preflight.

The same model provider may power all phases, but ChrysoCoord must treat them as
separate operations. Each phase gets a bounded prompt and a narrow decision.

The reviewer may run read-only validation after verifying it is on the expected
branch. The reviewer must not mutate code. A future git worktree model should be
considered so users can keep working in the original repository while agent roles
operate in isolated workspaces.

### Review Criteria

The formal reviewer must evaluate the ticket against structured criteria.

Baseline criteria:

- the implementation satisfies the ticket objective;
- the diff stays within the intended scope;
- required tests or validation ran and passed;
- generated artifacts are not used as unsupported evidence;
- branch and pull request metadata are readable;
- the implementation does not contradict the active spec, plan, or known
  constraints;
- documentation or changelog changes are present when the ticket requires them.

Project-specific criteria may be configured later. They should be ChrysoCoord
native and may be produced from a future quiz/checklist workflow. The reviewer
must include the exact criteria version or source in its evidence.

### Review Evidence

Review output should be persisted as a local ChrysoCoord review artifact.

The artifact should record:

- reviewed ticket;
- reviewer role and run id;
- source branch;
- pull request reference;
- diff or commit range inspected;
- validation evidence inspected;
- criteria applied;
- findings;
- outcome;
- next handoff.

Review artifacts are operational ChrysoCoord artifacts and should live under
`./.chc`. Versioning should be configurable, with versioned review artifacts as
the default.

Local ticket comments or metadata should stay compact and point to the review
artifact instead of embedding long review text directly in the ticket.

Provider comments are optional. If configured, ChrysoCoord may publish blocking
findings to a pull request and may use provider-native diff or review surfaces.
Formal provider integrations should aim to support Forgejo through `tea` and
GitHub through `gh`, while preserving local artifacts as the handoff record.

### Review Outcomes

Supported reviewer outcomes:

```text
pass
needs_changes
insufficient_evidence
needs_human_decision
```

`pass` moves the ticket to `ready_for_merge`.

`needs_changes` returns the ticket to `ready_for_dev` with a repair handoff. The
same ticket, branch, and pull request are reused. The handoff must point to the
local review artifact and may also point to provider comments when available.

`insufficient_evidence` blocks automatic merge and records what evidence is
missing.

`needs_human_decision` stops the loop and records why the agent should not keep
deciding.

### Repair Rounds

The workflow must prevent endless loops between `dev` and `reviewer`.

Each ticket tracks review rounds. The default maximum is one repair cycle.

When the maximum is reached, ChrysoCoord stops automatic processing and marks the
ticket blocked with:

- latest review artifact;
- latest findings;
- attempted round count;
- recommended human action.

The loop stops when a ticket blocks. ChrysoCoord should not skip the blocked
ticket and continue advancing later slicing work in the same automatic run.

### Merge Phase

The merge phase handles objective integration.

For the default automatic merge path, merge preflight runs after development
validation and pull request evidence are available.

For formal reviewer/merger mode, merge preflight runs only after a review pass.

The merge phase checks:

- ticket is eligible for merge under the selected mode;
- configured branch policy is satisfied;
- pull request reference is readable;
- pull request is open;
- base branch matches configured expectation;
- head branch matches ticket metadata;
- provider reports mergeability when available;
- no earlier repair blocker prevents serial work;
- configured merge strategy is valid.

The merge phase does not re-run validation only to add ceremony. It relies on
valid prior evidence and performs objective merge/provider integration checks.

On success, the ticket moves to `done` with merge metadata.

On failure, the ticket remains in the appropriate pre-merge or review state with
compact blocker metadata. The merge phase does not create another subjective
review phase.

### Loop Modes

The loop should expose explicit phase coverage.

Suggested modes:

```text
auto_merge
formal_review_merge
dev_only
```

Mode behavior:

- `auto_merge`: default; implement, validate, create pull request evidence, merge
  when policy allows it, and continue.
- `formal_review_merge`: implement, review, merge approved work, and continue.
- `dev_only`: implement eligible tickets and stop after review-ready handoff.

The selected mode must appear in dry-run output, event streams, and run reports.

### Accumulated Plans and Tickets

Accumulating multiple plans, slices, or tickets before execution is valid when
the loop can determine canonical order and eligibility.

The loop should not assume every pending artifact is executable. It should select
only eligible tickets, preserve canonical order, and report skipped or blocked
tickets with reasons.

## Acceptance Criteria

- Loop preview shows selected mode, integration unit, ordering policy, and
  eligible tickets in deterministic order.
- Default automatic loop ordering preserves plan, slice, and ticket order before
  applying ticket-key tie-breaks.
- Metadata filters narrow the eligible set without silently changing order.
- The first implementation uses ticket as the integration unit.
- Development completion attaches branch and pull request reference metadata to
  the ticket.
- Default auto-merge mode does not launch separate reviewer or merger agent roles.
- Default auto-merge mode requires valid development evidence before merge.
- Tickets with pull requests waiting for review move to `ready_for_review` when
  auto-merge is disabled, unavailable, or blocked.
- Formal reviewer execution produces a local ChrysoCoord-native review artifact.
- Formal review pass moves the ticket to `ready_for_merge`.
- Formal review findings move the ticket back to `ready_for_dev` with repair
  handoff metadata pointing to the local review artifact.
- The same ticket, branch, and pull request are reused for repair.
- Review round limits stop repeated repair loops and produce a blocked outcome.
- A blocked ticket stops the automatic loop.
- Merge execution accepts only tickets eligible under the selected mode.
- Successful merge moves the ticket to `done` and records merge metadata.
- Blocked merge preserves the appropriate handoff state and records an objective
  blocker.
- The default workflow works without requiring GitHub-specific or Forgejo-specific
  APIs as the canonical review data source.
- Formal provider-backed review can integrate with Forgejo through `tea` and
  GitHub through `gh` when configured.
- Existing explicit single-ticket execution remains available.
- Tests cover loop ordering, filter-preserved ordering, default auto-merge,
  validation evidence gating, review-ready transition, formal review pass, formal
  review changes requested, max review rounds, loop stop on blocked ticket, merge
  success, and merge blocked behavior.

## Open Questions

- Whether the configuration should call the default mode `auto_merge`,
  `dev_review_merge`, or another name.
- Whether loop mode configuration belongs under `work:`, `review:`, or a new
  `loop:` configuration section.
- Which exact `.chc` path should store review artifacts.
- Whether versioned review artifacts should be committed automatically or only
  made visible as versionable files.
- How much validation baseline generation belongs in `chc init` versus a later
  project-hardening command.
- Whether provider comments should be disabled by default until local review
  artifacts are mature.
- How future slice-level and batch-level integration units should map back to
  individual ticket Board statuses.
