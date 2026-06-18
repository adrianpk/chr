# Ticket Merge Workflow

## Purpose

Ticket merge workflow integrates a reviewed ticket pull request into the
configured work integration branch after a passed ticket review.

This workflow formalizes the `@reviewer -> @merger` handoff. It is not a
second code review phase and it is separate from the `dev` to `main` promotion
gate.

## Lifecycle

The normal reviewed ticket lifecycle is:

```text
ready_for_dev
in_progress
ready_for_review
ready_for_merge
done
```

`ready_for_merge` means the ticket has passed review and is authorized for
objective merge preflight.

## Merge Command

Merge one reviewed ticket:

```sh
chc ticket merge CHC-1
```

Preview without provider or ticket mutation:

```sh
chc ticket merge CHC-1 --dry-run
```

Select the expected base branch and merge strategy:

```sh
chc ticket merge CHC-1 --base develop --strategy rebase
```

Supported strategies are:

```text
merge
rebase
squash
```

The default expected base is the configured work integration branch. The
default strategy is `rebase`.

## Required Handoff

The ticket must have:

- `status: "ready_for_merge"`;
- `review_state: "pass"`;
- `review_result: "pass"`;
- `branch`;
- a `review_refs` entry with `type: "pull_request"`.

The pull request reference is the provider-native merge surface.

## Preflight

Before calling the provider merge operation, ChrysoCoord checks:

- ticket merge handoff metadata;
- provider pull request readability;
- pull request open state;
- expected base branch;
- pull request head branch against ticket `branch`;
- explicit provider mergeability when available;
- pending serial review repair blockers;
- configured merge strategy.

Preflight checks integrability only. It does not reinterpret code quality after
review has passed.

## Provider Execution

Forgejo-compatible providers use `tea api`:

```text
tea api -X PUT /repos/<owner>/<repo>/pulls/<number>/merge -d @-
```

GitHub uses `gh pr merge` with the matching strategy flag:

```text
gh pr merge <pull-request-url> --rebase
```

Provider CLI and token configuration use the review provider settings described
in [Review workflow configuration](review-workflow-config.md).

## Successful Merge

On success, ChrysoCoord updates ticket metadata:

```yaml
status: "done"
merge_state: "merged"
merge_result: "merged"
merge_actor: "merger"
merge_commit: "abc123"
merge_blocker: ""
merged_at: "2026-06-02T12:00:00Z"
```

`merge_commit` is the provider merge SHA when the provider response includes
one. It may be empty when the provider does not return a merge SHA.

Successful merges do not write ceremonial pull request or ticket comments.

## Blocked Merge

When preflight or provider execution blocks the merge, ChrysoCoord keeps the
ticket in `ready_for_merge` and records compact blocker metadata:

```yaml
status: "ready_for_merge"
merge_state: "blocked"
merge_result: "blocked"
merge_actor: "merger"
merge_blocker: "not_mergeable"
merged_at: ""
```

Examples of objective blockers:

- missing pull request metadata;
- provider unreadable;
- closed pull request;
- wrong base branch;
- head branch mismatch;
- not mergeable;
- pending review repair;
- invalid merge strategy.

Blocked merge metadata does not create another Board column.

## Serialization

Real merge execution acquires the workspace operation lock with the `merge`
operation. A workspace should not run development, review, or merge operations
concurrently.

Dry-run merge checks do not acquire the merge lock and do not write state.
