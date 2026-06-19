# Ticket Review Workflow

## Purpose

Ticket review workflow connects ticket implementation, pull request review,
repair, and merge readiness without adding Board columns beyond the existing
ticket lifecycle statuses.

This workflow is separate from the `dev` to `main` promotion gate.

## Lifecycle

Ticket implementation uses one branch and one pull request per ticket or slice.
The ticket stores the branch and pull request reference in frontmatter.

Typical lifecycle:

```text
ready_for_dev
in_progress
ready_for_review
ready_for_merge
done
```

When a review requests changes, the ticket returns to `ready_for_dev` with
review metadata that marks it as repair work:

```yaml
status: "ready_for_dev"
review_state: "changes_requested"
review_result: "needs_changes"
review_comments_url: "https://forge.example.com/owner/repo/pulls/123"
```

The Board remains stable. Repair priority is represented by metadata and work
selection behavior, not by an additional Board column.

## Required Pull Request Reference

Before a ticket can move to `ready_for_review`, it must have:

- `branch`
- at least one `review_refs` entry with `type: "pull_request"`
- required evidence strategy checklist items completed when strict evidence
  gating is enabled.

Attach them with:

```sh
chc ticket review-ref CHC-1 \
  --branch feat/slice-1-example \
  --provider forgejo \
  --url https://forge.example.com/owner/repo/pulls/123 \
  --id 123
```

The evidence gate is configured under `.chc/user/config.yml` at
`evidence.strategy`. See
[Evidence strategy configuration](evidence-strategy-config.md).

Inspect evidence before review with:

```sh
chc ticket show CHC-1
chc ticket board
```

The Board selected-ticket detail shows the same derived checklist summary in a
human-readable form. CLI list and board output keep the status compact so the
normal Board columns remain stable.

## Review Command

`chc ticket review` applies an explicit review result to a
`ready_for_review` ticket.

Clean pass:

```sh
chc ticket review CHC-1 --result pass
```

Blocking findings:

```sh
chc ticket review CHC-1 \
  --result needs_changes \
  --finding "Handle the nil input path before calling the parser."
```

Clean passes update local review metadata and move the ticket to
`ready_for_merge`. They do not write remote ceremonial comments.

`needs_changes` writes finding comments to the pull request, updates local
review metadata, and moves the ticket back to `ready_for_dev`.

## Merge Handoff

A clean review pass moves the ticket to `ready_for_merge`. That status is the
handoff from `@reviewer` to `@merger`.

The merger role runs objective merge preflight and then merges the pull request
through the configured provider:

```sh
chc ticket merge CHC-1
```

Merge execution and blocked merge metadata are described in
[Ticket merge workflow](ticket-merge-workflow.md).

## Repair Handoff

Repair work reuses the existing branch and pull request.

When a development worker starts a `ready_for_dev` ticket with
`review_state: "changes_requested"` or `review_result: "needs_changes"`, the
worker context includes a Review Repair Handoff section with:

- branch
- pull request URL
- review comments URL
- review state and result
- review round

The handoff points the worker at remote review comments as the source of
detailed findings. ChrysoCoord does not copy remote comments into ticket
comments.

If repair metadata lacks `review_comments_url`, branch metadata, or the pull
request reference, worker context materialization stops before launching the
worker.

## Serialized Work Selection

Development, review, and merge operations are serialized per core.

Execution paths acquire `.chc/user/runtime/agent-operation.lock` before
mutating ticket, review, or merge state. A second development, review, or merge
operation in the same workspace is blocked while the lock is held.

Work selection prioritizes unresolved review repairs before later
`ready_for_dev` work. If a non-repair ticket is selected while a repair is
pending, selection fails with a machine-readable preflight blocker.

Git workspaces must satisfy the configured work branch policy before work
selection proceeds. By default that means the local integration branch is clean
and contains the configured canonical remote tracking branch.

## Provider Access

Agentic review relies on provider API access for pull request evidence and
remote comments. Forgejo-compatible providers use `tea`; GitHub uses `gh`.

Provider diagnostics are available through:

```sh
chc ticket review-mode diagnostics --url https://forge.example.com/owner/repo/pulls/123
```
