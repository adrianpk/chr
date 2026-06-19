# Ticket Validation

## Purpose

Ticket validation checks Markdown ticket metadata before a ticket is accepted as
indexable state.

Markdown ticket files remain canonical. Validation protects the derived SQLite
index from malformed lifecycle metadata.

## Required Fields

The parser requires:

- `key`;
- `title`;
- `status`;
- `created_at`;
- `updated_at`.

Empty values for those fields are rejected.

## Status Values

Allowed status values are:

```text
backlog
ready_for_dev
in_progress
ready_for_review
ready_for_merge
done
open
ready
reviewing
blocked
solved
merged
closed
```

The first six values are normal Board statuses. The remaining values are legacy
statuses that still validate for existing tickets. `blocked` validates but is
rendered in the Board attention view rather than a normal column.

## Classification Values

`priority` and `severity` are optional. When present, each must be one of:

```text
low
normal
high
critical
```

`type` is optional. When present, it must be one of:

```text
task
bug
feature
cleanup
research
docs
```

## Tags

Tags are optional free-form grouping values.

Empty tag values are rejected.

## Review Metadata

Review metadata is optional.

Allowed `review_state` values are:

```text
pass
changes_requested
insufficient_evidence
needs_human_decision
```

Allowed `review_result` values are:

```text
pass
needs_changes
insufficient_evidence
needs_human_decision
```

`review_round` must be empty, zero, or a positive integer.

Each `review_refs` entry requires `type`, `provider`, and `url`.

Moving a ticket to `ready_for_review` requires both `branch` metadata and at
least one `review_refs` entry with `type: "pull_request"`.

When strict evidence strategy gating is enabled, `ready_for_review` also
requires a parseable completed evidence strategy checklist. The gate rejects
missing, malformed, blocked, or failing evidence before the status changes.

`blocking_findings` cannot contain empty values.

## Time Consistency

`created_at` and `updated_at` must parse as RFC3339 timestamps.

`updated_at` cannot be earlier than `created_at`.

## Scope

Validation runs when parsing ticket Markdown into the indexable ticket model.

Validation does not mutate ticket files.
