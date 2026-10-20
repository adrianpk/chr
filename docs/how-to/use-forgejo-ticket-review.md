# Use Forgejo Ticket Review And Merge

## Goal

Configure and run the Forgejo-first ticket review workflow for a ticket branch
and pull request, then merge the reviewed ticket.

## Prerequisites

- `tea` is installed.
- A Forgejo API token is available in `FORGEJO_TOKEN`.
- The ticket implementation branch has a pull request targeting the configured
  integration branch.
- The workspace satisfies the configured work branch policy before starting
  development or repair work.

## Configure Review

Set review configuration in `.chr/default/config.yml`:

```yml
review:
  mode: agentic_review
  forgejo_enabled: true
  forgejo_cli: tea
  forgejo_token_env: FORGEJO_TOKEN
  github_enabled: true
  github_cli: gh
  github_token_env: GITHUB_TOKEN
  non_blocking_comment_limit: 3
```

Check the active mode:

```sh
chr ticket review-mode status
```

## Attach the Pull Request

Attach branch and pull request metadata to the ticket:

```sh
chr ticket review-ref CHR-1 \
  --branch feat/slice-1-example \
  --provider forgejo \
  --url https://forge.example.com/owner/repo/pulls/123 \
  --id 123
```

Move the ticket to `ready_for_review` after implementation is complete:

```sh
chr ticket status CHR-1 ready_for_review
```

## Check Provider Access

Run diagnostics against the pull request URL:

```sh
chr ticket review-mode diagnostics --url https://forge.example.com/owner/repo/pulls/123
```

The diagnostic output should report provider mapping, CLI availability, token
availability, and pull request readability.

## Apply Review Result

For a clean pass:

```sh
chr ticket review CHR-1 --result pass
```

For blocking findings:

```sh
chr ticket review CHR-1 \
  --result needs_changes \
  --finding "Handle the nil input path before calling the parser."
```

When changes are requested, chr writes finding comments to the pull
request, stores `review_comments_url`, and moves the ticket back to
`ready_for_dev`.

When review passes, chr stores passed review metadata and moves the
ticket to `ready_for_merge`.

## Merge Reviewed Work

Preview the merge preflight without changing the pull request or ticket:

```sh
chr ticket merge CHR-1 --dry-run
```

Merge the reviewed pull request:

```sh
chr ticket merge CHR-1 --strategy rebase
```

The merge command verifies the ticket handoff, pull request readability, target
base branch, head branch, mergeability, serial repair blockers, and merge
strategy before calling the provider merge operation.

On success, chr moves the ticket to `done` and records merge metadata.
Clean merges do not write ceremonial pull request comments.

When the merge is blocked, chr keeps the ticket in `ready_for_merge`
with compact `merge_state`, `merge_result`, and `merge_blocker` metadata. The
Board does not need a separate merge-failure column.

## Repair

Run agent work normally:

```sh
chr work ticket CHR-1
```

The worker context includes the existing branch, pull request URL, and review
comments URL. The worker uses the remote pull request comments as the source of
detailed findings.

If another `ready_for_dev` ticket is selected while a review repair is pending,
work selection blocks and points at the pending repair ticket.
