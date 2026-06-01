# Review Workflow Configuration

## Purpose

Review workflow configuration controls how ticket or slice work is reviewed
before integration into `dev`.

The setting is independent from the separate `dev` to `main` promotion gate.

## Location

Review workflow settings live in `.chc/user/config.yml` under `review:`.

```yml
review:
  mode: human_review
  non_blocking_comment_limit: 3
```

## Modes

`human_review` stops for a human to inspect and approve the diff.

`agentic_review` requires an agentic review over ticket and pull request
evidence before merge. When remote pull request access is unavailable, later
review execution treats that as insufficient evidence.

`trusted_auto_merge` trusts the configured process and does not add a review
stop before integration.

## Fields

- `mode`: one of `human_review`, `agentic_review`, or `trusted_auto_merge`.
- `non_blocking_comment_limit`: maximum non-blocking remote comments a review
  operation may emit. Non-blocking comments never block merge.

Defaults:

- `mode`: `human_review`
- `non_blocking_comment_limit`: `3`

Clean agentic review passes do not write ceremonial remote comments.

## Status Command

Inspect the active review workflow mode:

```sh
chc ticket review-mode status
```

Preview the same information without performing review work:

```sh
chc ticket review-mode status --dry-run
```

The command prints the active mode, validation status, mode implications, clean
pass remote comment policy, and non-blocking comment limit.
