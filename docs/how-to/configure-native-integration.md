# Configure Native Integration

## Inspect and select the publication profile

From the intended project, inspect the effective policy before admitting work:

```sh
chr work config --json
```

Native publication defaults to `disabled`. Initializing or upgrading a project,
selecting version-one defaults, or selecting trusted automatic review does not
authorize rewriting a source branch. For an existing project, first complete the
explicit upgrade in [Workflow configuration](../reference/workflow-configuration.md);
it preserves the prior review choice and leaves an absent publication profile
disabled.

Edit the existing `work` mapping in `.chr/default/config.yml` when the project
authorizes the native profile. Keep other explicit settings in place:

```yaml
work:
  state_authority: tickets
  workspace_mode: worktree
  merge_strategy: rebase
  max_active_workers: 1
  ready_queue_capacity: 64
  integration:
    publication_profile: git_atomic_dual_ref
    max_prepared_per_target: 2
```

This profile permits rewriting the exact workflow-owned source branch to the
prepared result and advancing the canonical target to that same result in one
atomic Git transaction. It preserves the configured merge, rebase or squash
method; it does not grant ownership of another branch.

Run `chr work config --json` again. Confirm that
`work.integration.publication_profile` is `git_atomic_dual_ref` with source
`project`, and that the authority, workspace, method and review mode match the
project's approved policy. A malformed, duplicate or unknown setting rejects the
configuration. Inspection reads policy; it does not certify the remote server.

## Verify support before admitting work

The configured trusted publisher must verify the actual canonical repository and
both full refs before native work is admitted: the source is in the same
repository and rewrite-authorized, the target permits the authorized
fast-forward, the configured method is allowed, and the server supports the
atomic dual-ref transaction. Its current forge observation contract must certify
the original PR as merged when that transaction is observed.

The baseline uses host review of the exact prepared candidate, a writable
repository, and unprotected source and target branches without incompatible
enforced rules. The current executable rejects protected source or target branches and
nonempty enforced-rule responses. Supporting those policies requires equivalent
certified enforcement; retain the existing rules while publication is blocked. See
[Native integration](../reference/native-integration.md) for the supported policy
and explicit blockers. Selecting the profile alone supplies neither publisher
credentials nor certification. A source-only merge API is insufficient.

After those checks, materialize or adopt the approved native plan and preview
its exact extent:

```sh
chr work slice-set --preview --json <slice-set-id>
chr work slice-set --human-review --max-steps 1 <slice-set-id>
chr work status --json <execution-id>
```

The preview does not execute work. Use `--human-review` when that invocation
requires explicit approval. Review policy and native publication are independent
choices. An unsupported profile, changed canonical binding or uncertain
submission remains a blocker; the operator resolves the displayed required
action before continuing.

## Choose bounded parallel or token operation

Keep `work.max_active_workers: 1` for serial operation. To opt in to independent
parallel lanes, set it to a value from 2 through 8 and inspect the effective
configuration again. All agent roles share those slots. Tasks in one slice and
ordered frontiers stay serial; publication remains serialized per canonical
target. `max_prepared_per_target` bounds prepared candidates for that target
without adding workers.

An optional positive invocation token ceiling is currently a shared service
admission input, `WorkflowStartRequest.Policy.AgentTokenBudget`. The public
`work` admission commands and project YAML expose no token-ceiling setting.
A host caller selecting finite-token mode must provide a certified harness that
enforces and reports whole-call input plus output usage for executor, merger,
reviewer and closing roles. Unsupported metering blocks an agent launch.
Omitting a ceiling retains bounded attempts and time and displays token usage
as unmetered.

Inspect retained role, queue, attempt and token facts with `work status`.
A positive [operator grant](../reference/workflow-controls.md#native-budget-grants-and-decisions)
can extend an existing allowance without erasing consumed actions. Unknown
usage retains its reservation. Editing configuration, restarting or approving a
candidate does not refill the invocation.

Continue with [Run ticket-native workflows](run-ticket-workflows.md) and
[Workflow controls](../reference/workflow-controls.md) for retained execution,
displayed approval guards, decisions and explicit reconciliation.
