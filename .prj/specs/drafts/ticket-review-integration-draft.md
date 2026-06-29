---
title: "Ticket Review Integration"
project_id: "chryso"
case_slug: "ticket-execution-loop"
status: "draft"
source_request_id: "chat-2026-06-01-ticket-review-integration"
source_response_id: "codex-capture"
created_at: "2026-06-01T00:00:00Z"
mirror: ".chr/user/specs/ticket-review-integration-draft.md"
---

# Ticket Review Integration

Chryso should support review of tickets that reached `ready_for_review`
after agent execution. This flow focuses on the review surface for a single
ticket or pull request.

This draft is mirrored in `.chr/user/specs/ticket-review-integration-draft.md`
while the project evaluates self-hosted Chryso workflow adoption.

## Problem

Agent execution can move quickly, but human review still needs a concrete
surface: the ticket, the implementation diff, validation evidence, and the pull
request where review comments live.

Remote review comments are forge-specific. Forgejo, Codeberg, and GitHub expose
pull requests and comments through different APIs and tooling. Chryso
should not hard-code the ticket model to one forge, but the first useful target
is Forgejo/Codeberg.

## Expected Behavior

- Every executable ticket should be worked in a dedicated branch.
- Every ticket branch should have a pull request as its review surface.
- A ticket moved to `ready_for_review` should have a stable pull request
  reference when a pull request exists.
- The agent or command that creates the pull request must write that reference
  back to the ticket as part of the workflow.
- Ticket review should inspect local evidence first: ticket content, execution
  result, changed files, validation logs, and current diff or PR branch range.
- When a remote review integration is enabled, ticket review should also inspect
  pull request metadata and comments.
- In `agentic_review` mode, the reviewer must be able to read the remote pull
  request. Missing remote access is an evidence failure, not a reason to perform
  a local-only review.
- Chryso must not invent remote review state when credentials, tooling, or
  provider support are unavailable.

## Review Reference

The ticket needs formal, indexable workflow metadata. The branch is lifecycle
metadata. The pull request is the review surface.

Recommended shape:

```yaml
branch: "feat/ticket-CHR-123-review-integrations"
review_refs:
  - type: "pull_request"
    provider: "forgejo"
    url: "https://forge.example.com/owner/repo/pulls/123"
    id: "123"
    status: "open"
```

If the current ticket metadata layer cannot support structured lists cleanly in
the first slice, a single `pr_url` field is acceptable as a transitional form.
The workflow rule matters more than the exact first representation: reviewable
pull request work must leave a durable URL on the ticket.

## Workflow Consistency

Chryso should treat branch and pull request metadata as part of the ticket
execution contract, not as optional notes.

Expected state rules:

- `ready_for_dev` may exist without branch or pull request metadata.
- `in_progress` should have `branch`.
- `ready_for_review` should have `branch` and a pull request review reference.
- `ready_for_merge` should have a pull request reference whose remote state is
  merge-ready when the provider integration can verify it.
- `done` should have a merged pull request reference or an explicit non-PR
  exception.

The agent that creates the branch or pull request must update the ticket before
claiming the workflow step is complete. A reviewer should be able to depend on
these fields being in stable locations. If a reviewer cannot find the expected
branch or pull request reference, the correct result is an evidence failure such
as `missing_branch`, `missing_review_ref`, or `state_mismatch`.

This is especially important when one agent implements and a separate reviewer
agent evaluates the result. The reviewer should not need to infer the review
surface from chat history, branch naming guesses, or remote repository scans.

## Providers

Forgejo and Codeberg are the highest-priority provider targets.

GitHub should also be supported because it is ubiquitous and useful for projects
that mirror or host work there. GitHub support should not shape the Forgejo
design as the default assumption.

The provider model should leave room for later providers without creating broad
empty abstractions before Forgejo and GitHub behavior is real.

## Authentication And Tooling

Review integrations should be explicitly enabled by the user.

PAT-based API access is the default credential model for review data. SSH can
prove Git transport access, but it is not enough for pull request metadata,
review comments, or checks.

When a provider integration is active, Chryso should verify the required
local tooling and authentication state:

- Forgejo/Codeberg:
  - `tea` is installed when the configured integration path depends on it.
  - A Forgejo-compatible PAT is available or `tea` is authenticated.
  - The ticket pull request URL can be resolved and read.
- GitHub:
  - `gh` is installed when the configured integration path depends on it.
  - A GitHub token is available or `gh auth status` passes.
  - The ticket pull request URL can be resolved and read.

Direct HTTP API access may be preferable for structured behavior. `tea` and
`gh` remain valuable as diagnostics, setup checks, and fallbacks.

## Review Result

Ticket review should produce a compact result:

```text
review: pass | needs_changes | insufficient_evidence | needs_human_decision
remote_review: available | unavailable | disabled | unsupported
recommended_status: ready_for_review | in_progress | blocked | ready_for_merge
```

Findings should be concrete and evidence-based. The reviewer should avoid
generic style commentary unless a repository rule, ticket requirement, or
explicit review mode asks for that.

## Reviewer Loop Behavior

A reviewer loop should select tickets in `ready_for_review` and process each
ticket through a bounded evidence pipeline:

1. Load the ticket and structured workflow metadata.
2. Verify that `branch` and a pull request review reference exist.
3. Resolve the provider from the review reference.
4. Load local evidence: execution result, changed files, validation logs, and
   relevant ticket context.
5. Load remote evidence when integration is enabled: pull request metadata,
   diff, comments, review state, and checks when the provider exposes them.
6. Compare the claimed ticket outcome with the actual changed surface.
7. Produce a compact review result and optionally write provider comments.

The reviewer may use the remote API to read pull request metadata, diffs,
comments, review state, and checks. It may write comments only when configured
to do so. The first implementation should support a non-writing dry-run mode.

Ticket review is an inspection workflow, not an execution workflow. The normal
review path should not start a new Docker runtime, launch another Codex worker,
re-implement the ticket, or rerun expensive validation. It should inspect the
ticket, pull request diff, execution result, validation logs, checks, and remote
comments that already exist. A deeper verification mode may exist later, but it
must be explicit.

Provider comments should be high-signal only:

- correctness bugs;
- broken or missing validation;
- mismatch with the ticket;
- missing branch or pull request metadata;
- accidental scope expansion;
- contract, persistence, CLI, or user-facing behavior regressions;
- merge-blocking evidence gaps.

The reviewer should not post style nits, naming preferences, speculative
refactors, or broad architectural opinions unless a strict review mode is
explicitly requested. Low-confidence observations should remain in the local
review artifact instead of becoming remote PR comments.

To preserve agentic development speed, the reviewer should have a finding
budget and severity threshold. For example, remote comments may be limited to
blocking or high-confidence findings, while lower-severity notes stay in the
ticket review report.

The reviewer should fail closed on missing evidence but not become noisy. A
ticket with insufficient metadata should produce `insufficient_evidence` or a
specific machine-readable reason, not a long list of guesses.

## Remote Comment Policy

In `agentic_review` mode, the pull request is the primary review discussion
surface. The ticket should not duplicate detailed findings.

The reviewer should write remote comments only when it finds something worth
surfacing. A clean pass should update local ticket metadata but should not leave
ceremonial "looks good" comments in the pull request.

When comments are needed, inline comments are preferred when the provider can
anchor them reliably to the diff. Inline comments make human supervision easier
because the supervising user can inspect the review in the same place as a
human pull request review. If inline placement is not available or reliable, the
reviewer may leave a general pull request comment with numbered findings.

Blocking findings should use the provider's blocking review mechanism when the
provider supports it. Non-blocking findings may be posted only when they are
clearly useful for human supervision, and they must not block merge or mark the
review as failed.

Suggested remote comment limits:

```text
comment_on_clean_pass: false
max_blocking_comments: 5
max_non_blocking_comments: 2
non_blocking_comments_block_merge: false
prefer_inline_comments: true
```

Local ticket comments should stay minimal. A rejected review may add a short
local note that points to the pull request comments, but detailed findings
belong in the remote review surface.

## Anti-Ping-Pong Policy

The review workflow should avoid endless back-and-forth between implementer and
reviewer agents.

Reviewer output should be finite, prioritized, and actionable. Each finding
should include the evidence, the expected correction, and whether it blocks the
ticket from advancing. The reviewer should not debate implementation taste or
defend a position through repeated comments.

An implementer response cycle should either:

- address the finding;
- mark it as not applicable with concrete evidence;
- escalate to `needs_human_decision`.

After one implementer response, the reviewer should re-check only the specific
findings and newly changed diff. If the same disagreement remains unresolved,
the ticket should move to `needs_human_decision` instead of continuing an agent
argument loop.

Suggested loop limits:

```text
max_review_rounds: 2
max_remote_comments_per_round: 5
unresolved_repeat_result: needs_human_decision
```

## Review Rejection And Work Scheduling

When the reviewer leaves blocking findings, Chryso must signal that the
ticket needs more implementation work. Remote comments alone are not enough.

A blocking review result should update local ticket state and review metadata
so future agents can pick the work back up:

```yaml
status: "in_progress"
review_state: "changes_requested"
review_round: 1
review_result: "needs_changes"
```

The exact state transition can be policy-driven, but the scheduler must not
treat a rejected ticket as complete or advance to dependent follow-up work as if
the slice had landed successfully.

For planned slice or ticket series execution:

- the next slice or dependent ticket should not start while the current slice
  has an unresolved blocking review;
- the work loop should prioritize resolving `changes_requested` tickets before
  selecting new ready work from the same series;
- if the rejection requires human judgment, the ticket should move to
  `needs_human_decision` or an equivalent blocked state and the series should
  pause.

The reviewer should record enough local metadata for the scheduler and next
agent to know that PR comments need attention. Detailed findings should remain
in the remote pull request unless later workflow experience shows that a richer
local artifact is necessary.

Suggested structured fields:

```yaml
review_state: "pass | changes_requested | insufficient_evidence | needs_human_decision"
review_round: 1
review_comments_url: "https://forge.example.com/owner/repo/pulls/123"
blocking_findings:
  - "RF-1"
```

## Serialized Slice Execution

Chryso should not assume a swarm model where many agents work on dependent
branches at the same time. The default rule for planned slice execution is
serialized delivery.

Parallel agents may look appealing in theory, but dependent slice work creates
coordination friction around stale branches, unresolved review comments,
unmerged pull requests, and unclear ownership of fixes. For now, the safer
operating model is one active slice or dependent ticket at a time per series.

Serialized work also means one active agent operation per repository workspace
by default. A reviewer agent and a development agent should not operate on the
same repository at the same time. Review, repair, implementation, merge, and
next-work selection should acquire the same workspace-level execution lock.

The scheduler must not start the next slice in a series until the previous slice
is effectively integrated into `dev`.

Required evidence:

- the previous slice or ticket has a registered pull request;
- the pull request is merged;
- the merged commit or PR head is contained in `dev`;
- local `dev` is updated from the canonical remote;
- the previous ticket/slice is closed as integrated in local metadata;
- the working tree is clean before creating the next branch.

If any condition fails, the scheduler should stop with a precise reason such as
`previous_slice_not_merged_to_dev`, `local_dev_not_updated`, or
`previous_review_unresolved`.

Review passing is not enough to start the next slice. The prerequisite work must
be merged into `dev`, and the next branch must be created from that updated
`dev`.

## Work Selection Preflight

Every agent work selection should pass through a scheduler preflight before
context is handed to the development agent. The agent should not infer whether
to continue, repair, or start new work from chat history or branch names.

The preflight should:

- detect unresolved `changes_requested` tickets;
- detect active series membership and sequence order;
- verify prerequisite ticket and pull request integration state;
- verify that local `dev` contains the prerequisite work before creating a new
  branch;
- select repair work before later fresh work from the same series;
- return a machine-readable block reason when the next requested item is not
  eligible.

Example scheduler result:

```text
next_work: CHR-102
reason: unresolved_review_repair
series: onboarding-flow
blocks: CHR-103
branch_policy: use_existing_branch
```

Only after this preflight succeeds should Chryso materialize the worker
context. The prompt can explain the decision, but it must not be the primary
enforcement mechanism.

When a slice ticket fails review, the development agent should treat that ticket
as the active work item again. It should not merely block the next slice and
wait. The scheduler should requeue the same ticket for repair before selecting
any later slice in the same series.

Example:

1. Slice 1 passes review and is merged to `dev`.
2. Slice 2 is implemented and reaches `ready_for_review`.
3. The reviewer returns `changes_requested`.
4. Slice 2 moves back to the development queue with review metadata attached.
5. The development agent resumes Slice 2 repair on the existing branch and pull
   request.
6. Slice 3 remains ineligible until Slice 2 is reviewed, merged, and present in
   updated `dev`.

The work loop should therefore prefer unresolved review repairs from the active
series over fresh ready work from later slices.

## State Transition Policy

The initial implementation should not automatically move a ticket from
`ready_for_review` to `ready_for_merge`.

It should attach or record review evidence first. Later slices can decide
whether a passing review may suggest or apply `ready_for_merge` under an
explicit policy.

## Review Mode Configuration

The review phase should be configurable. Different users and projects may want
different tradeoffs between speed, risk, and human involvement.

Supported policy modes should include:

- `agentic_review`: Chryso runs the ticket reviewer on `ready_for_review`
  tickets before the work can advance.
- `human_review`: Chryso stops after creating the pull request and waits
  for a human to inspect the diff and approve or merge.
- `trusted_auto_merge`: Chryso trusts the execution evidence and configured
  checks enough to merge ticket or slice pull requests into `dev` automatically.

These modes should be project configuration, eventually editable through CLI,
TUI, and YAML config. The configuration should make the tradeoff explicit:

- `agentic_review` adds token and integration cost but can reduce human review
  load.
- `human_review` preserves human control but makes the user the bottleneck.
- `trusted_auto_merge` maximizes agent throughput but requires strong confidence
  in the execution, validation, and rollback strategy.

The mode should apply to ticket or slice integration into `dev`. It should not
change the separate dev-to-main promotion gate.

## Degraded Modes

If a ticket has no review reference, remote review should return
`insufficient_evidence` or `missing_review_ref` for PR-backed work.

If the provider is unsupported or credentials are missing in `agentic_review`
mode, Chryso should not perform a local-only substitute review. It should
mark the review as `insufficient_evidence` or `remote_review_unavailable` so
the integration problem is visible.

If remote comments are available but cannot be correlated with the local diff,
the review should surface that as a limitation rather than treating the comments
as complete evidence.

## Open Questions

- Should the first representation be structured `review_refs` or transitional
  `pr_url`?
- Should pull request creation be owned by ticket execution, a separate publish
  command, or a forge integration command?
- Should remote review comments be copied into the ticket, referenced by URL, or
  stored in a separate review artifact?
- What policy allows automatic transition from `ready_for_review` to
  `ready_for_merge`, if any?
- Which Forgejo API path should be primary when both direct HTTP and `tea` are
  available?
