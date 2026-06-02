# Review Workflow Configuration

## Purpose

Review workflow configuration controls how ticket or slice work is reviewed
before integration into `dev`.

The setting is independent from the separate `dev` to `main` promotion gate.

## Location

Review workflow settings live in `.chc/user/config.yml` under `review:`.
The same provider settings are used by ticket review diagnostics, remote review
comments, and reviewed ticket merge execution.

```yml
review:
  mode: human_review
  non_blocking_comment_limit: 3
  forgejo_enabled: true
  forgejo_cli: tea
  forgejo_token_env: FORGEJO_TOKEN
  github_enabled: true
  github_cli: gh
  github_token_env: GITHUB_TOKEN
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
- `forgejo_enabled`: whether Forgejo-compatible review providers are allowed.
- `forgejo_cli`: CLI executable used for Forgejo-compatible provider access.
- `forgejo_token_env`: environment variable that must contain the Forgejo or
  Codeberg API token for agentic review diagnostics.
- `github_enabled`: whether GitHub review provider access is allowed.
- `github_cli`: CLI executable used for GitHub provider access.
- `github_token_env`: environment variable that must contain the GitHub API
  token for agentic review diagnostics.

Defaults:

- `mode`: `human_review`
- `non_blocking_comment_limit`: `3`
- `forgejo_enabled`: `true`
- `forgejo_cli`: `tea`
- `forgejo_token_env`: `FORGEJO_TOKEN`
- `github_enabled`: `true`
- `github_cli`: `gh`
- `github_token_env`: `GITHUB_TOKEN`

Clean agentic review passes do not write ceremonial remote comments.

Review workflow behavior is described in
`docs/reference/ticket-review-workflow.md`.

Reviewed ticket merge behavior is described in
`docs/reference/ticket-merge-workflow.md`.

## Provider Diagnostics

Provider diagnostics verify that a pull request URL maps to a supported review
provider and that the configured CLI, token environment variable, and remote
pull request read path are available.

Forgejo-compatible providers use URLs shaped like:

```text
https://forge.example.test/owner/repo/pulls/123
```

GitHub uses URLs shaped like:

```text
https://github.com/owner/repo/pull/123
```

SSH remotes prove Git transport access only. They do not prove API access for
pull request review comments or pull request metadata.

Run diagnostics for one pull request:

```sh
chc ticket review-mode diagnostics --url https://forge.example.test/owner/repo/pulls/123
```

Preview provider mapping and local CLI/token checks without reading the remote
pull request:

```sh
chc ticket review-mode diagnostics --url https://forge.example.test/owner/repo/pulls/123 --dry-run
```

The command prints machine-readable check lines such as `provider_url`,
`provider_enabled`, `cli_available`, `token_available`, and `pr_readable`.

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
