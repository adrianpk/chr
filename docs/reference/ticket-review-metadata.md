# Ticket Review Metadata

## Purpose

Ticket review metadata records the branch, pull request review surface, and
agentic review state associated with a ticket.

The metadata lives in ticket Markdown frontmatter. SQLite stores indexed copies
for board, search, and workflow selection, but Markdown remains authoritative.

## Fields

`branch` stores the repository branch used for the ticket implementation.

`review_refs` stores external review surfaces. The first supported reference
type is a pull request:

```yaml
review_refs:
  - type: "pull_request"
    provider: "forgejo"
    url: "https://forge.example.com/owner/repo/pulls/123"
    id: "123"
    status: "open"
```

`review_state` stores the current review state:

```text
pass
changes_requested
insufficient_evidence
needs_human_decision
```

`review_round` stores the numeric review round. Empty or zero means no review
round has been recorded.

`review_result` stores the last review result:

```text
pass
needs_changes
insufficient_evidence
needs_human_decision
```

`review_comments_url` stores the remote pull request or review comments URL that
contains the detailed review discussion.

`blocking_findings` stores machine-readable identifiers for blocking findings,
such as `RF-1`.

## Board Behavior

Review metadata does not create additional board columns.

Tickets with `review_state: "changes_requested"` remain represented by their
lifecycle `status`, but they are surfaced as attention items with the reason
`review changes requested`.

Tickets with `review_state: "insufficient_evidence"` or
`review_state: "needs_human_decision"` are also surfaced as attention items.

## Validation

Review references require `type`, `provider`, and `url`.

`review_round` cannot be negative.

Empty `blocking_findings` values are rejected.
