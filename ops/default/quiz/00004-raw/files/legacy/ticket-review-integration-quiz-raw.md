# Ticket Review Integration Quiz Raw

Status: active

Spec Draft: `.chr/user/specs/ticket-review-integration-draft.md`
Mirror: `.prj/quizzes/raw/ticket-review-integration-quiz-raw.md`

## Q1: Repair After Changes Requested

Question: When a ticket receives `changes_requested`, should the same
development agent pick it up automatically on the next work selection, or
should the system wait for human confirmation before retrying?

Answer: In `agentic_review` mode, the rejected ticket should return to the
development loop without interrupting the human user. The human can observe the
agents through Git, Forge, and the board, but should not be notified as part of
the normal repair loop.

## Q2: Missing Remote Review Access

Question: If the reviewer cannot read the remote pull request because token,
`tea`, `gh`, or API access is missing, but local diff and execution evidence are
available, should it perform a degraded local review or block the ticket?

Answer: Avoid local-only review for now. It would add convoluted logic and lose
the audit trail that exists when the agent works through the remote pull
request. In `agentic_review` mode, missing remote access should be treated as
insufficient evidence or incomplete integration setup.

## Q3: Review Detail Location

Question: Should the agentic reviewer write comments directly in the remote
pull request, or should it initially write only a local artifact and change
ticket state?

Answer: The reviewer should leave the review in Forgejo or the active forge.
That avoids noise in the Chryso ticket comment system, which is limited
and should stay clean. The ticket should only record that comments were left in
the pull request.

## Q4: Inline Versus General Remote Comments

Question: In the remote pull request, should the reviewer prefer inline diff
comments when possible, or start with a general numbered findings comment?

Answer: Prefer inline comments when possible because that better emulates human
review workflow. Inline comments also help the supervising human lead inspect
whether Chryso is doing its review work well. If inline comments are not
available or reliable, a general comment is acceptable.

## Q5: Clean Review Signal

Question: When review passes, should the reviewer leave a visible remote signal
similar to "looks good to me"?

Answer: No. If the review passes, local ticket state and metadata are enough.
Avoid ceremonial comments, because repeated identical messages become tiring for
humans who occasionally supervise pull requests.

## Q6: Remote Comments On Clean Pass

Question: Should the reviewer only comment remotely when there are findings,
with "passed" comments disabled by default?

Answer: Yes. The reviewer should comment only when there are findings.

## Q7: Non-Blocking Findings

Question: If there are non-blocking findings, should the reviewer comment in
the pull request or keep them out of the remote surface?

Answer: Non-blocking comments are acceptable when they are genuinely useful to
the supervising human. They must not block operation. A pull request must not
fail or be prevented from advancing merely because a heuristic produced a
non-blocking comment.

## Q8: Non-Blocking Comment Limit

Question: Should non-blocking comments have a low maximum, such as one or two
per pull request, to prevent verbosity?

Answer: Yes. The reviewer should not overdo non-blocking comments.

## Q9: Pass With Notes Status

Question: Should a review with only non-blocking comments create a distinct
`pass_with_notes` status?

Answer: No. Do not invent statuses that do not help the agent workflow. Use
`pass` with supporting counters or metadata if needed.

## Q10: Actor References

Question: Should references such as `@dev`, `@reviewer`, and `@user` be treated
as configurable identities from the start, or as conventional text for now?

Answer: Treat them as conventional text for now. Chryso does not need to
institutionalize human or agentic users yet. Later it can gradually introduce
user concepts. Even without formal user entities, Chryso can search for
these references in Markdown and pull request comments.

## Q11: Repair Context Source

Question: When a development agent resumes a `changes_requested` ticket, should
Chryso inject the remote comments into the worker context, or should the
agent read them directly from the remote pull request?

Answer: Prefer the pull request as the detail source for now. Duplicating
review comments into ticket comments would reduce round trips, but it would add
noise to a text-based ticket system. For now, the ticket should point to the
remote review surface rather than duplicate findings.

## Pending Q12: Repair Without Remote Comment Access

Question: If the development agent cannot read remote comments when resuming
repair, should it block just like the reviewer blocks without remote access?

Answer: Pending.
