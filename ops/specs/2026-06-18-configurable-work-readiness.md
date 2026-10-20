# Configurable Work Readiness

Status: draft

Date: 2026-06-18

Model: `ops/specs/2026-06-18-configurable-work-readiness-model.md`

## Summary

chr should prepare and validate agent work from a configurable project
policy instead of assuming one hardcoded Git shape.

The current work-selection preflight protects the user's repository before
running agent work, but the protection is too rigid and too opaque. It assumes
that work must happen on local `dev`, that `origin/dev` must exist, and that the
worktree must be clean. Those defaults are useful for chr's own delivery
workflow, but they should be project policy rather than product law.

The desired behavior is a work readiness layer that:

- keeps sane defaults for `dev` plus `origin`;
- lets project owners configure branch and remote expectations;
- diagnoses blocking repository state in user language;
- safely repairs non-destructive issues when policy allows it;
- never hides real user changes or performs destructive Git operations.

## Goals

- Move branch and remote assumptions into `.chr/default/config.yml`.
- Preserve the existing default policy: integration branch `dev`, canonical
  remote `origin`, clean worktree, and remote tracking.
- Support projects that prefer a different integration branch, current-branch
  work, or no Git enforcement for local experiments.
- Replace raw Git failures such as `exit status 128` with classified work
  readiness diagnostics.
- Add a `chr work doctor` surface for checking and explaining readiness.
- Add safe repair behavior for issues chr can resolve without losing
  user work.
- Make Dash show the same readiness information before agent work starts.

## Non-Goals

- Do not remove the default `dev` workflow from chr's own repository.
- Do not force every project to use `dev`.
- Do not automatically stash, reset, rebase, force-push, or resolve divergent
  histories.
- Do not invent a remote for a real user repository.
- Do not implement full branch-per-ticket orchestration in this first pass.
- Do not turn runtime log inspection into a full embedded terminal viewer.

## Desired Behavior

### Default Policy

New and existing projects that do not configure work policy use:

```yaml
work:
  branch_policy:
    mode: integration_branch
    integration_branch: dev
    canonical_remote: origin
    require_clean_worktree: true
    require_remote_tracking: true
    allow_safe_repair: true
```

This preserves the current safety posture while making the assumptions visible
and configurable.

### Configurable Policy

Projects can choose a different integration branch:

```yaml
work:
  branch_policy:
    mode: integration_branch
    integration_branch: develop
    canonical_remote: upstream
```

Projects that intentionally run work on whatever branch is active can choose:

```yaml
work:
  branch_policy:
    mode: current_branch
    require_clean_worktree: true
```

Local experiments can disable Git enforcement:

```yaml
work:
  branch_policy:
    mode: no_git_enforcement
```

### Work Readiness Diagnostics

Readiness checks should report concrete issues:

- current branch does not match configured policy;
- configured integration branch is missing;
- worktree has user changes;
- chr runtime artifacts are visible to Git;
- configured remote tracking ref is missing;
- integration branch is behind its configured remote;
- Git repository is absent or unreadable;
- merge, rebase, or cherry-pick state is in progress.

Diagnostics should include the configured policy values that caused the check.
For example:

```text
work readiness: blocked
issue: wrong_branch
current_branch: main
expected_branch: dev
configured_by: .chr/default/config.yml
repair: switch_to_branch
```

Raw Git errors remain available as details, but they are not the primary user
message.

### Safe Repair

Safe repair is allowed only when the configured policy permits it and the repair
does not discard user work.

Allowed safe repairs:

- add chr-owned runtime paths to a local ignore mechanism;
- remove stale chr operation locks when they are proven dead;
- fetch the configured remote;
- fast-forward the configured integration branch when it is cleanly behind;
- switch to an existing configured integration branch when the worktree is
  clean;
- create a missing integration branch from the current branch only when the
  project policy explicitly allows branch creation.

Disallowed automatic repairs:

- `git reset`;
- forced checkout that overwrites files;
- stash or stash pop of user changes;
- force push;
- creating or replacing a canonical remote for a normal user repository;
- resolving merge, rebase, or cherry-pick conflicts.

### CLI Surface

The first operator-facing surface should be:

```text
chr work doctor
chr work doctor --fix-safe
```

Later work actions may use the same service directly:

```text
chr work run-all --prepare
```

The doctor should be useful outside the TUI and should be safe to run before an
expensive agent loop.

### Dash Surface

When Dash cannot start work, it should show readiness diagnostics and available
repairs instead of only a one-line failure.

If safe repair is available, Dash should expose a clear repair action and retry
path. The user should not need to copy shell commands for normal recoverable
states.

## Acceptance Criteria

- Work branch policy is loaded from `.chr/default/config.yml` with default values
  matching the current `dev`/`origin` behavior.
- Work-selection preflight uses the configured policy instead of hardcoded
  `dev` and `origin/dev` names.
- Projects can configure at least `integration_branch`, `current_branch`, and
  `no_git_enforcement` modes.
- Preflight diagnostics classify wrong branch, dirty worktree, missing remote
  tracking ref, and branch-behind-remote states without surfacing raw
  `exit status 128` as the primary error.
- `chr work doctor` prints readiness status and diagnostics for the resolved
  project.
- Safe repair can resolve chr-owned runtime artifact dirtiness and at
  least one branch/remote readiness issue without destructive Git operations.
- Dash displays work readiness diagnostics when work cannot start.
- Existing default behavior remains protective for chr's own workflow.
- Tests cover default policy, custom branch policy, current-branch mode,
  no-Git-enforcement mode, diagnostics, and safe repair behavior.

## Open Questions

- Whether branch creation should default to `ask`, `never`, or `safe` for new
  projects.
- Whether `feature_branch_per_ticket` should become a first-class branch policy
  mode in a later iteration.
- Whether project initialization should write the default work policy explicitly
  or rely on documented implicit defaults.
- Whether safe repair should write runtime ignore rules to `.git/info/exclude`
  or project `.gitignore`.
