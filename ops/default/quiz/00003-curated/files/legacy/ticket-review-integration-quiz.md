# Ticket Review Integration Quiz

Status: active

Raw Capture: `.chr/default/quizzes/raw/ticket-review-integration-quiz-raw.md`
Spec Draft: `.chr/default/specs/ticket-review-integration-draft.md`
Mirror: `.prj/quizzes/curated/ticket-review-integration-quiz.md`

## Purpose

Clarify how chr should handle agentic review for tickets that reach
`ready_for_review`, especially the handoff between development agents, reviewer
agents, pull request comments, ticket metadata, and serialized slice execution.

## Decisions

- In `agentic_review` mode, a rejected ticket should return to the development
  loop without interrupting the human user.
- Missing remote pull request access is an evidence failure for agentic review.
  chr should not perform a local-only substitute review in the first
  design.
- The remote pull request is the primary surface for detailed review comments.
- The chr ticket should stay clean and record only state, metadata, and
  a minimal pointer to pull request comments.
- Inline pull request comments are preferred when the provider can anchor them
  reliably.
- General pull request comments are acceptable when inline comments are not
  available or reliable.
- The reviewer should comment remotely only when there are findings.
- Clean reviews should not leave ceremonial "looks good" comments.
- Non-blocking comments are allowed only when they are useful for human
  supervision and must never block merge or mark the review as failed.
- Non-blocking comments should have a low limit to prevent review noise.
- Do not introduce statuses such as `pass_with_notes` unless they help agent
  workflow. Use `pass` plus metadata or counters if needed.
- References such as `@dev`, `@reviewer`, and `@user` are conventional text for
  now, not formal user identities.
- chr may later search for actor references in Markdown or pull request
  comments without requiring a full user model first.
- Repair context should point to the remote pull request review surface rather
  than duplicating detailed findings into ticket comments.

## Constraints

- The ticket comment system should remain clean and low-noise.
- Human supervisors should be able to inspect agent behavior through Git, Forge,
  and the board without being interrupted by normal agent repair loops.
- Review comments should support human supervision without slowing normal
  agentic development.
- The remote pull request audit trail is important enough that local-only review
  is not acceptable for the first agentic review workflow.

## Open Questions

- If the development agent cannot read remote comments when resuming repair,
  should repair block just like review blocks without remote access?
- What exact metadata should record that remote review comments need attention
  without duplicating those comments locally?
- Which provider capabilities are required for inline comments in the first
  Forgejo implementation?
