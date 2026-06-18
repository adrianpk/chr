# Work Readiness

## Purpose

Work readiness is the preflight contract ChrysoCoord checks before starting
agent work.

It protects the managed project from starting a ticket execution or ready queue
from an unexpected repository state. The policy is project-local and
configurable, so projects do not need to follow ChrysoCoord's own branch names.

## Configuration

Work readiness configuration lives in `.chc/user/config.yml` under `work:`.

```yml
work:
  mode: integration_branch
  integration_branch: dev
  canonical_remote: origin
  require_clean_worktree: true
  require_remote_tracking: true
  allow_safe_repair: true
  allow_branch_creation: false
```

Generated config may omit these fields. Omitted fields use the default work
policy.

## Defaults

Default work policy:

| Field | Default |
| --- | --- |
| `mode` | `integration_branch` |
| `integration_branch` | `dev` |
| `canonical_remote` | `origin` |
| `require_clean_worktree` | `true` |
| `require_remote_tracking` | `true` |
| `allow_safe_repair` | `true` |
| `allow_branch_creation` | `false` |

The defaults preserve ChrysoCoord's own `dev` plus `origin` workflow while
making the policy explicit and replaceable per project.

## Modes

| Mode | Behavior |
| --- | --- |
| `integration_branch` | Agent work must start from the configured integration branch. |
| `current_branch` | Agent work may start from the current branch, subject to configured cleanliness checks. |
| `no_git_enforcement` | Branch, remote, and worktree checks are skipped. Review-repair ordering still applies. |

Unknown modes are configuration errors.

## Diagnostics

`chc work doctor` prints the effective policy, readiness status, diagnostics,
and repair guidance without starting agent work.

```sh
chc work doctor
```

Typical blocking diagnostics include:

- `wrong_branch`: current branch does not match the configured integration
  branch.
- `missing_branch`: the configured branch is missing locally.
- `dirty_worktree`: user-owned changes are visible.
- `runtime_artifacts_visible`: ChrysoCoord runtime artifacts are visible to
  Git and can be safely ignored.
- `missing_remote_ref`: the configured remote tracking ref is missing.
- `branch_behind_remote`: the local branch is behind the configured remote ref.
- `git_state_in_progress`: Git has an in-progress operation such as merge or
  rebase.
- `git_unavailable`: Git state could not be inspected.
- `policy_invalid`: work configuration is invalid.

Work may start only when there are no blocking diagnostics.

## Safe Repair

When a diagnostic has a non-destructive repair, `chc work doctor --fix-safe`
applies it and prints the resulting readiness report.

```sh
chc work doctor --fix-safe
```

Safe repair may:

- add ChrysoCoord runtime artifact paths to local ignore state;
- fetch the configured canonical remote;
- fast-forward the configured integration branch;
- switch to the configured branch when no user work would be lost;
- create the configured branch when policy allows branch creation.

Safe repair must not discard user changes or run destructive Git operations.
When repair is not safe, the diagnostic remains manual.

## Dash Behavior

Dash displays work readiness in the current work panel. When readiness is
blocked, Dash prevents ticket execution and run-all actions before starting the
runtime worker.

If safe repair is available, Dash exposes `x` to apply it after confirmation.
After repair, Dash reloads readiness so the user can retry the same work action.

## Work Selection

The same readiness policy is used by:

- `chc work doctor`;
- `chc work ticket <ticket-key>`;
- `chc work run-all`;
- TUI Dash work actions;
- ticket review-repair prioritization before later serial work.

Ticket review-repair ordering is not disabled by `no_git_enforcement`.

## Merge Base Defaults

`chc ticket merge` uses the configured work integration branch as its default
expected pull request base. Passing `--base` overrides that expectation for one
command invocation.
