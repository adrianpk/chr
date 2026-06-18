# Configurable Work Readiness Model

Status: draft

Date: 2026-06-18

Spec: `ops/specs/2026-06-18-configurable-work-readiness.md`

## Purpose

This model defines the project-local work readiness policy and diagnostic
contracts used before ChrysoCoord starts agent work.

The model is configuration-first. It should let projects keep ChrysoCoord's
default `dev` workflow or choose another Git posture without patching code.

## Workspace Config

`WorkspaceConfig` gains a `Work` section.

```go
type WorkspaceConfig struct {
    TicketKeyPrefix string
    Runtime         RuntimeConfig
    Notifications   NotificationConfig
    Review          ReviewWorkflowConfig
    Work            WorkConfig
}
```

## Work Config

```go
type WorkConfig struct {
    BranchPolicy WorkBranchPolicy
}
```

## Work Branch Policy

```go
type WorkBranchPolicy struct {
    Mode                  WorkBranchPolicyMode
    IntegrationBranch     string
    CanonicalRemote       string
    RequireCleanWorktree  bool
    RequireRemoteTracking bool
    AllowSafeRepair       bool
    AllowBranchCreation   bool
}
```

## Branch Policy Modes

Initial modes:

| Mode | Semantics |
| --- | --- |
| `integration_branch` | Work starts only from the configured integration branch. |
| `current_branch` | Work may start from the current branch, subject to cleanliness checks. |
| `no_git_enforcement` | Git branch and cleanliness checks are skipped for work selection. |

Unknown modes should be treated as configuration errors by readiness services
that enforce work policy.

## Defaults

Default work config:

| Field | Default |
| --- | --- |
| `mode` | `integration_branch` |
| `integration_branch` | `dev` |
| `canonical_remote` | `origin` |
| `require_clean_worktree` | `true` |
| `require_remote_tracking` | `true` |
| `allow_safe_repair` | `true` |
| `allow_branch_creation` | `false` |

These defaults preserve the current ChrysoCoord workflow while making it
editable.

## Config File Shape

```yaml
work:
  mode: integration_branch
  integration_branch: dev
  canonical_remote: origin
  require_clean_worktree: true
  require_remote_tracking: true
  allow_safe_repair: true
  allow_branch_creation: false
```

The parser may also accept nested `branch_policy` in a later iteration, but the
first implementation should prefer the flat `work:` section to match the
existing simple config parser.

## Work Readiness Diagnostic

```go
type WorkReadinessDiagnostic struct {
    Code           WorkReadinessDiagnosticCode
    Severity       WorkReadinessSeverity
    Detail         string
    CurrentBranch  string
    ExpectedBranch string
    Remote         string
    RemoteRef      string
    Repair         WorkReadinessRepair
}
```

## Diagnostic Codes

Initial codes:

- `wrong_branch`
- `missing_branch`
- `dirty_worktree`
- `runtime_artifacts_visible`
- `missing_remote_ref`
- `branch_behind_remote`
- `git_state_in_progress`
- `git_unavailable`
- `policy_invalid`

## Severity

Initial severities:

- `info`
- `warning`
- `blocking`

Work execution may proceed only when there are no blocking diagnostics.

## Repair Model

```go
type WorkReadinessRepair struct {
    Available bool
    Code      WorkReadinessRepairCode
    Safe      bool
    Detail    string
}
```

Initial repair codes:

- `ignore_runtime_artifacts`
- `fetch_remote`
- `fast_forward_branch`
- `switch_branch`
- `create_branch`
- `manual`

Safe repair execution must be idempotent when practical and must not discard
user changes.

## Invariants

- Defaults must preserve existing `dev` plus `origin` behavior.
- ChrysoCoord-owned runtime artifacts must not make work readiness appear dirty
  once safe repair or the ignore contract has been applied.
- A missing remote tracking ref is a diagnostic, not an unclassified Git error.
- `no_git_enforcement` must skip branch, remote, and worktree checks but must
  not skip ticket review-repair ordering.
- `current_branch` must not require a specific branch name.
- No repair may run a destructive Git command.
