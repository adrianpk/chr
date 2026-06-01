# Use Forgejo Ticket Review

## Goal

Configure and run the Forgejo-first ticket review workflow for a ticket branch
and pull request.

## Prerequisites

- `tea` is installed.
- A Forgejo API token is available in `FORGEJO_TOKEN`.
- The ticket implementation branch has a pull request targeting `dev`.
- The workspace is on local `dev`, clean, and up to date with `origin/dev`
  before starting development or repair work.

## Configure Review

Set review configuration in `.chc/user/config.yml`:

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
chc ticket review-mode status
```

## Attach the Pull Request

Attach branch and pull request metadata to the ticket:

```sh
chc ticket review-ref CHC-1 \
  --branch feat/slice-1-example \
  --provider forgejo \
  --url https://forge.example.com/owner/repo/pulls/123 \
  --id 123
```

Move the ticket to `ready_for_review` after implementation is complete:

```sh
chc ticket status CHC-1 ready_for_review
```

## Check Provider Access

Run diagnostics against the pull request URL:

```sh
chc ticket review-mode diagnostics --url https://forge.example.com/owner/repo/pulls/123
```

The diagnostic output should report provider mapping, CLI availability, token
availability, and pull request readability.

## Apply Review Result

For a clean pass:

```sh
chc ticket review CHC-1 --result pass
```

For blocking findings:

```sh
chc ticket review CHC-1 \
  --result needs_changes \
  --finding "Handle the nil input path before calling the parser."
```

When changes are requested, ChrysoCoord writes finding comments to the pull
request, stores `review_comments_url`, and moves the ticket back to
`ready_for_dev`.

## Repair

Run agent work normally:

```sh
chc work ticket CHC-1
```

The worker context includes the existing branch, pull request URL, and review
comments URL. The worker uses the remote pull request comments as the source of
detailed findings.

If another `ready_for_dev` ticket is selected while a review repair is pending,
work selection blocks and points at the pending repair ticket.

